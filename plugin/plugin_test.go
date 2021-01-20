/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package plugin

import (
	"os"
	"testing"

	"github.com/open-dovetail/fabric-client/activity/request"
	"github.com/stretchr/testify/assert"
)

func TestFabricGoFile(t *testing.T) {
	configFile = "../test-network/config.yaml"
	matcherFile = "../test-network/local_entity_matchers.yaml"

	networkConfig, err := request.ReadFile(configFile)
	assert.NoError(t, err, "read network config file should not thorw error")
	matchersConfig, err := request.ReadFile(matcherFile)
	assert.NoError(t, err, "read entity matchers file should not throw error")

	os.Mkdir("src", 0755)
	err = createFabricGoFile(networkConfig, matchersConfig)
	assert.NoError(t, err, "read sample contract should not throw error")
}
