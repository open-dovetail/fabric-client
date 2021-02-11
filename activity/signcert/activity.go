/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package signcert

import (
	"github.com/open-dovetail/fabric-client/activity/request"
	"github.com/pkg/errors"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
)

// Create a new logger
var logger = log.ChildLogger(log.RootLogger(), "activity-fabclient-signcert")

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func init() {
	_ = activity.Register(&Activity{}, New)
}

// Activity fabric signcert activity struct
type Activity struct {
	connectionName string
	userOrgOnly    bool
}

// New creates a new Activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	logger.Infof("Create signcert activity with InitContxt settings %v", ctx.Settings())
	if err := s.FromMap(ctx.Settings()); err != nil {
		logger.Errorf("failed to configure signcert activity %v", err)
		return nil, err
	}

	return &Activity{
		connectionName: s.ConnectionName,
		userOrgOnly:    s.UserOrgOnly,
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

	spec, err := a.getConnectorSpec(input)
	if err != nil {
		output := &Output{Code: 500, Message: err.Error()}
		ctx.SetOutputObject(output)
		return false, err
	}

	user := input.UserName
	if len(input.OrgName) > 0 {
		user += "@" + input.OrgName
	}
	cert := request.UserCertificate(spec, user)

	output := &Output{Code: 200,
		Message: "",
		Result:  cert,
	}
	ctx.SetOutputObject(output)
	return true, nil
}

func (a *Activity) getConnectorSpec(input *Input) (*request.ConnectorSpec, error) {
	if len(input.UserName) == 0 {
		logger.Error("user name is not specified")
		return nil, errors.New("user name is not specified")
	}

	return &request.ConnectorSpec{
		Name:           a.connectionName,
		NetworkConfig:  request.NetworkConfig,
		EntityMatchers: request.EntityMatcher,
		OrgName:        input.OrgName,
		UserName:       input.UserName,
		UserOrgOnly:    a.userOrgOnly,
	}, nil
}
