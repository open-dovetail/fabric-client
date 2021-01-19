/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package request

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/project-flogo/core/data/coerce"

	jschema "github.com/xeipuuv/gojsonschema"
)

// Attribute describes a name and data type
type Attribute struct {
	Name string `md:"name"`
	Type string `md:"type"`
}

// Settings of the activity
type Settings struct {
	ConnectionName  string       `md:"connectionName,required"`
	ChannelID       string       `md:"channelID,required"`
	ChaincodeID     string       `md:"chaincodeID,required"`
	TransactionName string       `md:"transactionName,required"`
	Arguments       []*Attribute `md:"arguments"`
	RequestType     string       `md:"requestType,required"`
}

// Input of the activity
type Input struct {
	OrgName       string                 `md:"orgName"`
	UserName      string                 `md:"userName,required"`
	Parameters    map[string]interface{} `md:"parameters"`
	Transient     map[string]interface{} `md:"transient"`
	TimeoutMillis int                    `md:"timeoutMillis"`
	Endpoints     []string               `md:"endpoints"`
}

// Output of the activity
type Output struct {
	Code    int         `md:"code"`
	Message string      `md:"message"`
	Result  interface{} `md:"result"`
}

// construct Attribute from map of name and type
func toAttribute(name, value string) *Attribute {
	jsonType := jschema.TYPE_STRING
	if strings.EqualFold(value, "true") || strings.EqualFold(value, "false") {
		jsonType = jschema.TYPE_BOOLEAN
	} else if matched, err := regexp.MatchString(`\d+\.\d*`, value); err == nil && matched {
		jsonType = jschema.TYPE_NUMBER
	} else if matched, err := regexp.MatchString(`\d+`, value); err == nil && matched {
		jsonType = jschema.TYPE_INTEGER
	}
	return &Attribute{
		Name: name,
		Type: jsonType,
	}
}

func (p *Attribute) String() string {
	return fmt.Sprintf("(%s:%s)", p.Name, p.Type)
}

// FromMap sets activity settings from a map
func (h *Settings) FromMap(values map[string]interface{}) error {
	var err error
	if h.ConnectionName, err = coerce.ToString(values["connectionName"]); err != nil {
		return err
	}
	if h.ChannelID, err = coerce.ToString(values["channelID"]); err != nil {
		return err
	}
	if h.ChaincodeID, err = coerce.ToString(values["chaincodeID"]); err != nil {
		return err
	}
	if h.TransactionName, err = coerce.ToString(values["transactionName"]); err != nil {
		return err
	}
	if h.RequestType, err = coerce.ToString(values["requestType"]); err != nil {
		return err
	}

	params, err := coerce.ToString(values["parameters"])
	if err != nil {
		return err
	}
	if len(params) == 0 {
		return nil
	}
	args := strings.Split(strings.TrimSpace(params), ",")
	for _, v := range args {
		pt := strings.Split(strings.TrimSpace(v), ":")
		if len(pt) == 0 || len(strings.TrimSpace(pt[0])) == 0 {
			continue
		}
		value := ""
		if len(pt) > 1 {
			value = strings.TrimSpace(pt[1])
		}
		if attr := toAttribute(strings.TrimSpace(pt[0]), value); attr != nil {
			h.Arguments = append(h.Arguments, attr)
		}
	}
	return nil
}

// ToMap converts activity input to a map
func (i *Input) ToMap() map[string]interface{} {
	var eps []interface{}
	for _, p := range i.Endpoints {
		eps = append(eps, p)
	}

	return map[string]interface{}{
		"orgName":       i.OrgName,
		"userName":      i.UserName,
		"timeoutMillis": i.TimeoutMillis,
		"endpoints":     eps,
		"parameters":    i.Parameters,
		"transient":     i.Transient,
	}
}

// FromMap sets activity input values from a map
func (i *Input) FromMap(values map[string]interface{}) error {

	var err error
	if i.OrgName, err = coerce.ToString(values["orgName"]); err != nil {
		return err
	}
	if i.UserName, err = coerce.ToString(values["userName"]); err != nil {
		return err
	}
	if i.TimeoutMillis, err = coerce.ToInt(values["timeoutMillis"]); err != nil {
		return err
	}
	if i.Parameters, err = coerce.ToObject(values["parameters"]); err != nil {
		return err
	}
	if i.Transient, err = coerce.ToObject(values["transient"]); err != nil {
		return err
	}

	var eps interface{}
	if eps, err = coerce.ToAny(values["endpoints"]); err != nil {
		return err
	}
	switch v := eps.(type) {
	case []interface{}:
		for _, d := range v {
			p := strings.TrimSpace(d.(string))
			if len(p) > 0 {
				i.Endpoints = append(i.Endpoints, p)
			}
		}
	case string:
		p := strings.TrimSpace(v)
		if len(p) > 0 {
			i.Endpoints = []string{p}
		}
	}
	return nil
}

// ToMap converts activity output to a map
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"code":    o.Code,
		"message": o.Message,
		"result":  o.Result,
	}
}

// FromMap sets activity output values from a map
func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	if o.Code, err = coerce.ToInt(values["code"]); err != nil {
		return err
	}
	if o.Message, err = coerce.ToString(values["message"]); err != nil {
		return err
	}
	if o.Result, err = coerce.ToAny(values["result"]); err != nil {
		return err
	}

	return nil
}
