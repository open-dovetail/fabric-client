/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package signcert

import (
	"errors"
	"strings"

	"github.com/project-flogo/core/data/coerce"
)

// Settings of the activity
type Settings struct {
}

// Input of the activity
type Input struct {
	OrgName  string `md:"orgName"`
	UserName string `md:"userName,required"`
}

// Output of the activity
type Output struct {
	Code    int         `md:"code"`
	Message string      `md:"message"`
	Result  interface{} `md:"result"`
}

// ToMap converts activity input to a map
func (i *Input) ToMap() map[string]interface{} {
	user := i.UserName
	if len(i.OrgName) > 0 {
		user += "@" + i.OrgName
	}

	return map[string]interface{}{
		"userName": user,
	}
}

// FromMap sets activity input values from a map
func (i *Input) FromMap(values map[string]interface{}) error {

	user, err := coerce.ToString(values["userName"])
	if err != nil {
		return err
	}
	tokens := strings.Split(strings.TrimSpace(user), "@")
	if len(tokens) == 0 {
		return errors.New("username is not specified")
	}
	i.UserName = strings.TrimSpace(tokens[0])
	if len(tokens) > 1 {
		i.OrgName = strings.TrimSpace(tokens[1])
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
