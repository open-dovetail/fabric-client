/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/open-dovetail/fabric-chaincode/plugin/contract"
	"github.com/pkg/errors"
	"github.com/project-flogo/cli/common" // Flogo CLI support code
	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/core/trigger"
	"github.com/project-flogo/flow/definition"
	"github.com/spf13/cobra"
	jschema "github.com/xeipuuv/gojsonschema"
)

var enterprise bool
var contractFile string
var restRoot string
var appFile string

func init() {
	contract2rest.Flags().StringVarP(&contractFile, "contract", "c", "contract.json", "specify a contract.json to create Flogo app from")
	contract2rest.Flags().StringVarP(&restRoot, "name", "n", "", "specify the root path of REST APIs")
	contract2rest.Flags().StringVarP(&appFile, "app", "o", "app.json", "specify the output file app.json")
	contract2rest.Flags().BoolVarP(&enterprise, "fe", "e", false, "user Flogo Enterprise")
	common.RegisterPlugin(contract2rest)
}

var contract2rest = &cobra.Command{
	Use:              "contract2rest",
	Short:            "generate REST app from contract specification",
	Long:             "This plugin reads a contract spec, and generate Flogo REST service to invoke the transactions in the spec",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Create Flogo REST service from ", contractFile)
		spec, err := contract.ReadContract(contractFile)
		if err != nil {
			fmt.Printf("Failed to read and parse contract file %s: %+v\n", contractFile, err)
			os.Exit(1)
		}
		app, err := createRESTApp(spec, enterprise)
		if err != nil {
			fmt.Printf("Failed to create REST service from contract file %s: %+v\n", contractFile, err)
			os.Exit(1)
		}
		if err = contract.WriteAppConfig(app, appFile); err != nil {
			fmt.Printf("Failed to write app config file %s: %+v\n", appFile, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully written service app %s\n", appFile)
	},
}

// generate Flogo REST app from the first contract in a contract spec
func createRESTApp(spec *contract.Spec, fe bool) (*app.Config, error) {
	if len(spec.Contracts) == 0 {
		return nil, errors.New("No contract is defined in the spec")
	}

	var name string
	var con *contract.Contract
	for k, v := range spec.Contracts {
		name = k
		con = v
		break
	}
	ac := &app.Config{
		Name:        name + "-service",
		Type:        "flogo:app",
		Version:     spec.Info.Version,
		Description: "REST service for " + con.Name,
		AppModel:    "1.1.1",
		Imports: []string{
			"github.com/open-dovetail/fabric-client/activity/request",
			"github.com/project-flogo/contrib/activity/actreturn",
			"github.com/open-dovetail/dovetail-contrib/trigger/rest",
			"github.com/open-dovetail/dovetail-contrib/function/dovetail",
			"github.com/project-flogo/flow",
		},
		Properties: fabricSampleProperties(),
	}
	if fe {
		// convert and cache app schemas for Flogo Enterprise
		if err := spec.ConvertAppSchemas(); err != nil {
			fmt.Printf("failed to convert app schema: %v\n", err)
		}
	}

	// create REST trigger with one handler per transaction
	trig := createRESTTrigger(con, fe)
	ac.Triggers = []*trigger.Config{trig}

	// create a flow resource per transaction
	resources := make(map[string]*definition.DefinitionRep)
	for _, tx := range con.Transactions {
		var schm *trigger.SchemaConfig
		if fe {
			schm = handlerSchema(trig, tx.Name)
		}
		id, res, err := createResource(tx, schm)
		if err != nil {
			return nil, err
		}
		resources[id] = res
	}

	if fe {
		// collect app schema for Flogo Enterprise
		if schm, err := contract.GetAppSchemas(); err == nil {
			ac.Schemas = schm
		} else {
			fmt.Printf("failed to collect app schemas: %v\n", err)
		}
	}

	// serializes resources
	contract.SetAppResources(ac, resources)

	return ac, nil
}

// search trigger config for a schema config of a specified transaction name
func handlerSchema(trigConfig *trigger.Config, txName string) *trigger.SchemaConfig {
	for _, h := range trigConfig.Handlers {
		if h.Name == txName {
			return h.Schemas
		}
	}
	return nil
}

func fabricSampleProperties() []*data.Attribute {
	var props []*data.Attribute
	props = append(props, data.NewAttribute("PORT", data.TypeInt64, 8989))
	props = append(props, data.NewAttribute("NETWORK", data.TypeString, "test-network"))
	props = append(props, data.NewAttribute("CHANNEL", data.TypeString, "mychannel"))
	props = append(props, data.NewAttribute("CHAINCODE", data.TypeString, "basic"))
	props = append(props, data.NewAttribute("APPUSER", data.TypeString, "Admin"))
	return props
}

// create REST trigger with handlers specified by transactions in a contract
func createRESTTrigger(c *contract.Contract, fe bool) *trigger.Config {
	trig := &trigger.Config{
		Id:  "receive_http_message",
		Ref: "#rest",
		Settings: map[string]interface{}{
			"port": `=$property["PORT"]`,
		},
	}

	path := rootRESTPath(c.Name)
	for _, tx := range c.Transactions {
		handler := createRESTHandler(tx, path, fe)
		trig.Handlers = append(trig.Handlers, handler)
	}
	return trig
}

func rootRESTPath(contractName string) string {
	if len(restRoot) > 0 {
		return "/" + restRoot
	}
	exp := regexp.MustCompile(`[\W\d]`)
	tokens := exp.Split(strings.TrimSpace(contractName), -1)
	for _, s := range tokens {
		if len(s) > 0 {
			return "/" + strings.ToLower(s)
		}
	}
	return ""
}

// create REST trigger handler for a contract transaction
func createRESTHandler(tx *contract.Transaction, path string, fe bool) *trigger.HandlerConfig {
	handler := &trigger.HandlerConfig{
		Name: tx.Name,
	}

	handler.Settings = map[string]interface{}{
		"method": "POST",
		"path":   path + "/" + strings.ToLower(tx.Name),
	}

	// generate flow action
	res := "res://flow:" + contract.ToSnakeCase(tx.Name)
	// map all parameters as a single object

	input := map[string]interface{}{
		"user": "=dovetail.httpUser($.headers)",
	}
	if len(tx.Parameters) > 0 {
		input["parameters"] = "=$.content"
	}
	if len(tx.Transient) > 0 {
		input["transient"] = "=$.content.transient"
	}
	output := map[string]interface{}{
		"code": "=$.code",
		"data": "=$.data",
	}
	action := &trigger.ActionConfig{
		Config: &action.Config{
			Ref:      "#flow",
			Settings: map[string]interface{}{"flowURI": res}},
		Input:  input,
		Output: output,
	}
	handler.Actions = []*trigger.ActionConfig{action}
	if fe {
		// set handler schema for Flogo enterprise
		handler.Schemas = createHandlerSchema(tx)
	}
	return handler
}

func parameterSchemaDef(tx *contract.Transaction) *schema.Def {
	if len(tx.Parameters) == 0 {
		return nil
	}
	props := make(map[string]interface{})

	for _, p := range tx.Parameters {
		props[p.Name] = p.Schema
	}
	ps := map[string]interface{}{
		"type":       jschema.TYPE_OBJECT,
		"properties": props,
	}
	if pbytes, err := json.Marshal(ps); err == nil {
		return &schema.Def{
			Type:  "json",
			Value: string(pbytes),
		}
	}
	return nil
}

func transientSchemaDef(tx *contract.Transaction) *schema.Def {
	if len(tx.Transient) == 0 {
		return nil
	}

	ts := map[string]interface{}{
		"type":       jschema.TYPE_OBJECT,
		"properties": tx.Transient,
	}
	if pbytes, err := json.Marshal(ts); err == nil {
		return &schema.Def{
			Type:  "json",
			Value: string(pbytes),
		}
	}
	return nil
}

func createHandlerSchema(tx *contract.Transaction) *trigger.SchemaConfig {
	result := &trigger.SchemaConfig{}

	if len(tx.Returns) > 0 {
		// convert returns schema
		rs := map[string]interface{}{
			"type": jschema.TYPE_OBJECT,
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type": jschema.TYPE_STRING,
				},
				"result": tx.Returns,
			},
		}
		if _, err := contract.ExpandRef(rs); err == nil {
			if rbytes, err := json.Marshal(rs); err == nil {
				result.Reply = map[string]interface{}{
					"data": &schema.Def{
						Type:  "json",
						Value: string(rbytes),
					},
				}
			}
		}
	}

	// convert parameters schema
	ps := make(map[string]interface{})
	if len(tx.Parameters) > 0 {
		for _, p := range tx.Parameters {
			ps[p.Name] = p.Schema
		}
	}

	// convert transient schema
	if len(tx.Transient) > 0 {
		if _, err := contract.ExpandRef(tx.Transient); err == nil {
			ts := map[string]interface{}{
				"type":       jschema.TYPE_OBJECT,
				"properties": tx.Transient,
			}
			ps["transient"] = ts
		}
	}

	content := map[string]interface{}{
		"type":       jschema.TYPE_OBJECT,
		"properties": ps,
	}

	if pbytes, err := json.Marshal(content); err == nil {
		result.Output = map[string]interface{}{
			"content": &schema.Def{
				Type:  "json",
				Value: string(pbytes),
			},
		}
	}
	return result
}

// create REST flow resource for a contract transaction
func createResource(tx *contract.Transaction, schm *trigger.SchemaConfig) (string, *definition.DefinitionRep, error) {
	id := "flow:" + contract.ToSnakeCase(tx.Name)

	input := map[string]data.TypedValue{
		"user": data.NewAttribute("user", data.TypeString, nil),
	}
	if len(tx.Parameters) > 0 {
		input["parameters"] = data.NewAttribute("parameters", data.TypeObject, nil)
	}
	if len(tx.Transient) > 0 {
		input["transient"] = data.NewAttribute("transient", data.TypeObject, nil)
	}
	rAttr := data.NewAttribute("data", data.TypeAny, nil)
	includeSchema := false
	if schm != nil {
		// add schema info for Flogo Enterprise
		includeSchema = true
		if len(tx.Parameters) > 0 {
			if sc := contract.ExtractFlowSchema(parameterSchemaDef(tx)); sc != nil {
				input["parameters"] = data.NewAttributeWithSchema("parameters", data.TypeObject, nil, sc)
			}
		}
		if len(tx.Transient) > 0 {
			if sc := contract.ExtractFlowSchema(transientSchemaDef(tx)); sc != nil {
				input["transient"] = data.NewAttributeWithSchema("transient", data.TypeObject, nil, sc)
			}
		}
		if sc := contract.ExtractFlowSchema(schm.Reply["data"]); sc != nil {
			rAttr = data.NewAttributeWithSchema("data", data.TypeAny, nil, sc)
		}
	}

	md := &metadata.IOMetadata{
		Input: input,
		Output: map[string]data.TypedValue{
			"code": data.NewAttribute("code", data.TypeInt64, 0),
			"data": rAttr,
		},
	}

	res := &definition.DefinitionRep{
		Name:     tx.Name,
		Metadata: md,
	}

	// add fabric request and return task resources
	res.Tasks = append(res.Tasks, fabricRequestTask(tx, includeSchema))
	res.Tasks = append(res.Tasks, returnTask())

	// add links
	link := &definition.LinkRep{
		FromID: "request_1",
		ToID:   "actreturn_1",
	}
	res.Links = append(res.Links, link)

	return id, res, nil
}

// create Fabric-request task resource from transaction spec
func fabricRequestTask(tx *contract.Transaction, includeSchema bool) *definition.TaskRep {
	actCfg := &activity.Config{
		Ref: "#request",
	}
	params, _ := tx.ParameterDef()
	reqType := "invoke"
	if isReadOnly(tx) {
		reqType = "query"
	}
	actCfg.Settings = map[string]interface{}{
		"chaincodeID":     `=$property["CHAINCODE"]`,
		"channelID":       `=$property["CHANNEL"]`,
		"connectionName":  `=$property["NETWORK"]`,
		"parameters":      params,
		"requestType":     reqType,
		"transactionName": tx.Name,
	}

	actCfg.Input = map[string]interface{}{
		"userName": "=$flow.user",
	}
	schemaInput := make(map[string]interface{})
	if len(tx.Parameters) > 0 {
		actCfg.Input["parameters"] = "=$flow.parameters"
		if includeSchema {
			pdef := parameterSchemaDef(tx)
			schemaInput["parameters"] = &contract.FlowSchema{
				SchemaType:  pdef.Type,
				SchemaValue: pdef.Value,
			}
		}
	}
	if len(tx.Transient) > 0 {
		actCfg.Input["transient"] = "=$flow.transient"
		if includeSchema {
			tdef := transientSchemaDef(tx)
			schemaInput["transient"] = &contract.FlowSchema{
				SchemaType:  tdef.Type,
				SchemaValue: tdef.Value,
			}
		}
	}

	if includeSchema {
		actCfg.Schemas = &activity.SchemaConfig{
			Input: schemaInput,
		}
		rs := map[string]interface{}{
			"result": tx.Returns,
		}
		if _, err := contract.ExpandRef(rs); err == nil {
			if rbytes, err := json.Marshal(rs["result"]); err == nil {
				actCfg.Schemas.Output = map[string]interface{}{
					"result": &contract.FlowSchema{
						SchemaType:  "json",
						SchemaValue: string(rbytes),
					},
				}
			}
		}
	}

	return &definition.TaskRep{
		ID:             "request_1",
		Name:           "Fabric Request",
		ActivityCfgRep: actCfg,
	}
}

// returns true if transaction does not call #put nor #delete
func isReadOnly(tx *contract.Transaction) bool {
	for _, r := range tx.Rules {
		for _, a := range r.Actions {
			if a.Activity == "#put" || a.Activity == "#delete" {
				return false
			}
		}
	}
	return true
}

// create return task resource
func returnTask() *definition.TaskRep {
	actCfg := &activity.Config{
		Ref: "#actreturn",
	}
	actCfg.Settings = map[string]interface{}{
		"mappings": map[string]interface{}{
			"code": "=$activity[request_1].code",
			"data": map[string]interface{}{
				"mapping": map[string]interface{}{
					"message": "=$activity[request_1].message",
					"result":  "=$activity[request_1].result",
				},
			},
		},
	}

	return &definition.TaskRep{
		ID:             "actreturn_1",
		Name:           "Return",
		ActivityCfgRep: actCfg,
	}
}
