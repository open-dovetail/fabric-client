/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package signcert

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/open-dovetail/fabric-client/activity/request"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

var cryptoPath = "../../../hyperledger/fabric-samples/test-network/organizations"
var testConfig = "../../test-network/config.yaml"
var testMatchers = "../../test-network/local_entity_matchers.yaml"

func setup() error {
	logger.Info("Setup network config")

	os.Setenv("CRYPTO_PATH", cryptoPath)
	netConfig, err := request.ReadFile(testConfig)
	if err != nil {
		return err
	}
	netMatchers, err := request.ReadFile(testMatchers)
	if err != nil {
		return err
	}
	request.InitializeNetwork(netConfig, netMatchers)
	return nil
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		logger.Errorf("FAILED %v", err)
		os.Exit(1)
	}
	logger.Info("Setup successful")
	status := m.Run()
	if status > 0 {
		logger.Info("You must start Fabric test-network and deploy chaincode:")
		logger.Info("   network.sh up")
	}
	os.Exit(status)
}

func TestSigncert(t *testing.T) {
	logger.Info("TestSigncert")

	// configure request activity
	settings := map[string]interface{}{
		"connectionName": "test-network",
	}
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	ctx := test.NewActivityInitContext(settings, mf)
	act, err := New(ctx)
	assert.NoError(t, err, "create activity instance should not throw error")

	tc := test.NewActivityContext(act.Metadata())

	// input data
	req := `{
		"userName": "Admin"
	}`
	var data map[string]interface{}
	err = json.Unmarshal([]byte(req), &data)
	assert.NoError(t, err, "input data should be valid JSON object")

	input := &Input{}
	err = input.FromMap(data)
	assert.NoError(t, err, "create input from map should not throw error")
	assert.Equal(t, "Admin", input.UserName, "username should be 'Admin'")

	err = tc.SetInputObject(input)
	assert.NoError(t, err, "setting action input should not throw error")

	// process request
	done, err := act.Eval(tc)
	assert.True(t, done, "action eval should be successful")
	assert.NoError(t, err, "action eval should not throw error")

	// verify activity output
	output := &Output{}
	err = tc.GetOutputObject(output)
	logger.Infof("output: %v", output)
	assert.NoError(t, err, "action output should not be error")
	assert.Equal(t, 200, output.Code, "output status code should be 200")
	assert.Contains(t, output.Result.(string), "CN=Admin@org1.example.com", "no data should be returned by this transaction")
}
