/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package signcert

import (
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
)

// Create a new logger
var logger = log.ChildLogger(log.RootLogger(), "activity-fabclient-signcert")

// NetworkConfig is the content of fabric network config file
var NetworkConfig []byte

// InitializeNetwork can be called to initialize Fabric network config
func InitializeNetwork(config []byte) {
	NetworkConfig = config
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func init() {
	_ = activity.Register(&Activity{}, New)
}

// Activity fabric signcert activity struct
type Activity struct {
}

// New creates a new Activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &Activity{}, nil
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

	user := input.UserName
	if len(input.OrgName) > 0 {
		user += "@" + input.OrgName
	}
	cert := UserCertificate(user)

	output := &Output{Code: 200,
		Message: "",
		Result:  cert,
	}
	ctx.SetOutputObject(output)
	return true, nil
}
