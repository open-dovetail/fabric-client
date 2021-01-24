/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package request

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
)

const (
	opInvoke = "invoke"
	opQuery  = "query"
)

// NetworkConfig is the content of fabric network config file
var NetworkConfig []byte

// EntityMatcher is the content of fabric local entity matcher file
var EntityMatcher []byte

// InitializeNetwork can be called to initialize Fabric network config
func InitializeNetwork(config, matcher []byte) {
	NetworkConfig = config
	if len(matcher) > 0 {
		EntityMatcher = matcher
	}
}

// Create a new logger
var logger = log.ChildLogger(log.RootLogger(), "activity-fabclient-request")

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func init() {
	_ = activity.Register(&Activity{}, New)
}

// Activity fabric request activity struct
type Activity struct {
	connectionName  string
	channelID       string
	chaincodeID     string
	transactionName string
	arguments       []*Attribute
	requestType     string
	userOrgOnly     bool
}

// New creates a new Activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	logger.Infof("Create request activity with InitContxt settings %v", ctx.Settings())
	if err := s.FromMap(ctx.Settings()); err != nil {
		logger.Errorf("failed to configure request activity %v", err)
		return nil, err
	}

	return &Activity{
		connectionName:  s.ConnectionName,
		channelID:       s.ChannelID,
		chaincodeID:     s.ChaincodeID,
		transactionName: s.TransactionName,
		arguments:       s.Arguments,
		requestType:     s.RequestType,
		userOrgOnly:     s.UserOrgOnly,
	}, nil
}

// Metadata implements activity.Activity.Metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	logger.Debugf("%v", a)

	// check input args
	input := &Input{}
	if err = ctx.GetInputObject(input); err != nil {
		return false, err
	}

	params := a.prepareParameters(input.Parameters)
	transientMap := prepareTransient(input.Transient)

	client, err := a.getFabricClient(input)
	if err != nil {
		output := &Output{Code: 500, Message: err.Error()}
		ctx.SetOutputObject(output)
		return false, err
	}

	// invoke fabric transaction
	var response []byte
	var status int
	if a.requestType == opInvoke {
		logger.Debugf("execute chaincode %s transaction %s timeout %d endpoints %v", a.chaincodeID, a.transactionName, input.TimeoutMillis, input.Endpoints)
		response, status, err = client.ExecuteChaincode(a.chaincodeID, a.transactionName, params, transientMap)
	} else {
		logger.Debugf("query chaincode %s transaction %s timeout %d endpoints %v", a.chaincodeID, a.transactionName, input.TimeoutMillis, input.Endpoints)
		response, status, err = client.QueryChaincode(a.chaincodeID, a.transactionName, params, transientMap)
	}

	if err != nil {
		msg := "Fabric request returned error"
		logger.Errorf("msg %+v", msg, err)
		output := &Output{Code: 500, Message: msg}
		ctx.SetOutputObject(output)
		return false, errors.Wrapf(err, msg)
	}

	logger.Debugf("Fabric response - status %d, response %s", status, string(response))

	var result interface{}
	if status < 300 && len(response) > 0 {
		if err := json.Unmarshal(response, &result); err != nil {
			logger.Warnf("failed to unmarshal fabric response %+v, error: %+v", response, err)
			result = response
		}
	}

	var msg string
	if len(response) > 0 {
		msg = string(response)
	} else {
		msg = "No data returned"
	}
	output := &Output{Code: status,
		Message: msg,
		Result:  result,
	}
	ctx.SetOutputObject(output)
	return true, nil
}

func (a *Activity) getFabricClient(input *Input) (*FabricClient, error) {
	if len(input.UserName) == 0 {
		logger.Error("user name is not specified")
		return nil, errors.New("user name is not specified")
	}

	return NewFabricClient(ConnectorSpec{
		Name:           a.connectionName,
		NetworkConfig:  NetworkConfig,
		EntityMatchers: EntityMatcher,
		OrgName:        input.OrgName,
		UserName:       input.UserName,
		ChannelID:      a.channelID,
		TimeoutMillis:  input.TimeoutMillis,
		Endpoints:      input.Endpoints,
		UserOrgOnly:    a.userOrgOnly,
	})
}

func prepareTransient(transData map[string]interface{}) map[string][]byte {
	if transData == nil {
		logger.Debug("no transient data is specified")
		return nil
	}
	transMap := make(map[string][]byte)
	for k, v := range transData {
		if jsonBytes, err := json.Marshal(v); err != nil {
			logger.Infof("failed to marshal transient data %+v", err)
		} else {
			transMap[k] = jsonBytes
		}
	}
	return transMap
}

func (a *Activity) prepareParameters(parameters map[string]interface{}) [][]byte {
	var result [][]byte
	for _, p := range a.arguments {
		// TODO: assuming string params here to be consistent with implementaton of trigger and chaincode-shim
		// should change all places to use []byte for best portability
		param := ""
		if v, ok := parameters[p.Name]; ok && v != nil {
			if param, ok = v.(string); !ok {
				pbytes, err := json.Marshal(v)
				if err != nil {
					logger.Errorf("failed to marshal input: %+v", err)
					param = fmt.Sprintf("%v", v)
				} else {
					param = string(pbytes)
				}
			}
			logger.Debugf("add chaincode parameter: %s=%s", p.Name, param)
		}
		result = append(result, []byte(param))
	}
	return result
}
