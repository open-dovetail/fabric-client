/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package request

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This test requires to start the fabric test-network using
//    network.sh up createChannel
//    network.sh deployCC
// and set network config files as in setup of activity_test.go
const (
	connectorName = "test"
	channelID     = "mychannel"
	org           = "org1"
	user          = "User1"
	ccID          = "basic"
)

func TestClient(t *testing.T) {
	os.Setenv("CRYPTO_PATH", cryptoPath)
	networkConfig, err := ReadFile(testConfig)
	require.NoError(t, err, "failed to read config file %s", testConfig)

	entityMatcherOverride, err := ReadFile(testMatchers)
	require.NoError(t, err, "failed to read entity matcher file %s", testMatchers)

	fbClient, err := NewFabricClient(ConnectorSpec{
		Name:           connectorName,
		NetworkConfig:  networkConfig,
		EntityMatchers: entityMatcherOverride,
		OrgName:        org,
		UserName:       user,
		ChannelID:      channelID,
	})
	require.NoError(t, err, "failed to create fabric client %s", connectorName)
	logger.Infof("created fabric client %+v", fbClient)

	// initialize ledger
	result, _, err := fbClient.ExecuteChaincode(ccID, "InitLedger", [][]byte{}, nil)
	require.NoError(t, err, "failed to invoke %s", ccID)
	logger.Infof("InitLedger result: %s", string(result))

	// query original
	result, _, err = fbClient.QueryChaincode(ccID, "ReadAsset", [][]byte{[]byte("asset6")}, nil)
	require.NoError(t, err, "failed to query %s", ccID)
	logger.Infof("Query asset6 result: %s", string(result))
	origValue := result

	// update
	result, _, err = fbClient.ExecuteChaincode(ccID, "TransferAsset", [][]byte{[]byte("asset6"), []byte("Jose")}, nil)
	require.NoError(t, err, "failed to invoke %s", ccID)
	logger.Infof("Transfer asset6 result: %s", string(result))

	// query after update
	result, _, err = fbClient.QueryChaincode(ccID, "ReadAsset", [][]byte{[]byte("asset6")}, nil)
	require.NoError(t, err, "failed to query %s", ccID)
	logger.Infof("Query asset6 result: %s", string(result))
	assert.NotEqual(t, origValue, result, "original %s should different from %s", string(origValue), string(result))
}

func TestNetworkConfigYaml(t *testing.T) {
	os.Setenv("CRYPTO_PATH", cryptoPath)
	networkConfig, err := ReadFile(testConfig)
	require.NoError(t, err, "failed to read config file %s", testConfig)
	cs := ConnectorSpec{
		NetworkConfig: networkConfig,
	}
	f := orgFilter(cs).(*OrgFilter)
	assert.Equal(t, "Org1MSP", f.MSPID, "default MSPID should be 'Org1MSP'")

	cs = ConnectorSpec{
		NetworkConfig: networkConfig,
		OrgName:       "org2",
	}
	f = orgFilter(cs).(*OrgFilter)
	assert.Equal(t, "Org2MSP", f.MSPID, "org2's MSPID should be 'Org2MSP'")
}
