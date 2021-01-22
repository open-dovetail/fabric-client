/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package plugin

import (
	"fmt"
	"testing"

	"github.com/open-dovetail/fabric-chaincode/plugin/contract"
	"github.com/stretchr/testify/assert"
)

var testContract = "../contract/sample-contract.json"

func TestContractToREST(t *testing.T) {
	fmt.Println("TestContractToREST")
	spec, err := contract.ReadContract(testContract)
	assert.NoError(t, err, "read sample contract should not throw error")
	config, err := createRESTApp(spec, true)
	assert.NoError(t, err, "generate REST app should not throw error")

	err = contract.WriteAppConfig(config, "rest-app.json")
	assert.NoError(t, err, "write app config should not throw error")
	//assert.Fail(t, "test")
}
