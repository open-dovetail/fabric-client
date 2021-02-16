/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package request

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	yaml "gopkg.in/yaml.v2"
)

const (
	configType = "yaml"
)

// cached Fabric client connections
var clientMap = map[string]*FabricClient{}

// FabricClient holds fabric client pointers for chaincode invocations.
type FabricClient struct {
	name          string
	sdk           *fabsdk.FabricSDK
	client        *channel.Client
	timeoutMillis int
	endpoints     []string
	filter        fab.TargetFilter
}

// ConnectorSpec contains configuration parameters of a Fabric connector
type ConnectorSpec struct {
	Name           string
	NetworkConfig  []byte
	EntityMatchers []byte
	OrgName        string
	UserName       string
	ChannelID      string
	TimeoutMillis  int
	Endpoints      []string
	UserOrgOnly    bool
}

// OrgFilter implements TargetFilter interface for target peers
type OrgFilter struct {
	MSPID string
}

// Accept implements fab.TargetFilter interface
func (f *OrgFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.MSPID
}

func (c *FabricClient) setOrgFilter(config ConnectorSpec) {
	var data map[interface{}]interface{}
	yaml.Unmarshal(config.NetworkConfig, &data)
	msps := make(map[string]string)
	defaultOrg := execYamlPath(data, "client.organization")
	for k, v := range data["organizations"].(map[interface{}]interface{}) {
		mspid := v.(map[interface{}]interface{})["mspid"]
		msps[k.(string)] = mspid.(string)
	}

	orgName := defaultOrg.(string)
	if len(config.OrgName) > 0 {
		orgName = config.OrgName
	}

	if mspid, ok := msps[orgName]; ok && len(mspid) > 0 {
		c.filter = &OrgFilter{MSPID: mspid}
	}
}

// return value at path c1.c2.c3 from yaml file, does not handle arrays
func execYamlPath(node interface{}, path string) interface{} {
	tokens := strings.Split(path, ".")
	result := node
	var ok bool
	for _, name := range tokens {
		result, ok = yamlChildNode(result, name)
		if !ok {
			return nil
		}
	}
	return result
}

func yamlChildNode(parent interface{}, name string) (interface{}, bool) {
	data, ok := parent.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}
	c, ok := data[name]
	return c, ok
}

// NewFabricClient returns a new or cached fabric client
func NewFabricClient(config ConnectorSpec) (*FabricClient, error) {
	clientKey := fmt.Sprintf("%s.%s.%s.%t", config.Name, config.UserName, config.OrgName, config.UserOrgOnly)
	if fbClient, ok := clientMap[clientKey]; ok && fbClient != nil {
		fbClient.timeoutMillis = config.TimeoutMillis
		fbClient.endpoints = config.Endpoints
		return fbClient, nil
	}
	sdk, err := fabsdk.New(networkConfigProvider(config.NetworkConfig, config.EntityMatchers))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create new SDK")
	}

	opts := []fabsdk.ContextOption{fabsdk.WithUser(config.UserName)}
	if config.OrgName != "" {
		opts = append(opts, fabsdk.WithOrg(config.OrgName))
	}
	client, err := channel.New(sdk.ChannelContext(config.ChannelID, opts...))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create new client of channel %s", config.ChannelID)
	}
	fbClient := &FabricClient{
		name:          config.Name,
		sdk:           sdk,
		client:        client,
		timeoutMillis: config.TimeoutMillis,
		endpoints:     config.Endpoints,
	}
	if config.UserOrgOnly {
		fbClient.setOrgFilter(config)
	}
	clientMap[clientKey] = fbClient

	return fbClient, nil
}

func networkConfigProvider(networkConfig []byte, entityMatcherOverride []byte) core.ConfigProvider {
	configProvider := config.FromRaw(networkConfig, configType)

	if len(entityMatcherOverride) > 0 {
		return func() ([]core.ConfigBackend, error) {
			matcherProvider := config.FromRaw(entityMatcherOverride, configType)
			matcherBackends, err := matcherProvider()
			if err != nil {
				fmt.Printf("failed to parse entity matchers: %+v\n", err)
				// return the original config provider defined by configPath
				return configProvider()
			}

			currentBackends, err := configProvider()
			if err != nil {
				fmt.Printf("failed to parse network config: %+v\n", err)
				return nil, err
			}

			// return the combined config with matcher precedency
			return append(matcherBackends, currentBackends...), nil
		}
	}
	return configProvider
}

// Close closes Fabric client connection
func (c *FabricClient) Close() {
	c.sdk.Close()
}

// QueryChaincode sends query request to Fabric network
func (c *FabricClient) QueryChaincode(ccID, fcn string, args [][]byte, transient map[string][]byte) ([]byte, int, error) {
	opts := []channel.RequestOption{channel.WithRetry(retry.DefaultChannelOpts)}
	if c.timeoutMillis > 0 {
		//		fmt.Printf("set request timeout: %d ms\n", c.timeoutMillis)
		opts = append(opts, channel.WithTimeout(fab.Query, time.Duration(c.timeoutMillis)*time.Millisecond))
	}
	if c.endpoints != nil && len(c.endpoints) > 0 {
		//		fmt.Printf("set target endpoints: %s\n", strings.Join(c.endpoints, ", "))
		opts = append(opts, channel.WithTargetEndpoints(c.endpoints...))
	} else if c.filter != nil {
		opts = append(opts, channel.WithTargetFilter(c.filter))
	}
	response, err := c.client.Query(channel.Request{ChaincodeID: ccID, Fcn: fcn, Args: args, TransientMap: transient}, opts...)
	if err != nil {
		return nil, 500, err
	}
	return response.Payload, int(response.ChaincodeStatus), nil
}

// ExecuteChaincode sends invocation request to Fabric network
func (c *FabricClient) ExecuteChaincode(ccID, fcn string, args [][]byte, transient map[string][]byte) ([]byte, int, error) {
	opts := []channel.RequestOption{channel.WithRetry(retry.DefaultChannelOpts)}
	if c.timeoutMillis > 0 {
		//		fmt.Printf("set request timeout: %d ms\n", c.timeoutMillis)
		opts = append(opts, channel.WithTimeout(fab.Execute, time.Duration(c.timeoutMillis)*time.Millisecond))
	}
	if c.endpoints != nil && len(c.endpoints) > 0 {
		//		fmt.Printf("set target endpoints: %s\n", strings.Join(c.endpoints, ", "))
		opts = append(opts, channel.WithTargetEndpoints(c.endpoints...))
	} else if c.filter != nil {
		opts = append(opts, channel.WithTargetFilter(c.filter))
	}
	response, err := c.client.Execute(channel.Request{ChaincodeID: ccID, Fcn: fcn, Args: args, TransientMap: transient}, opts...)
	if err != nil {
		return nil, 500, err
	}
	return response.Payload, int(response.ChaincodeStatus), nil
}

// ReadFile returns content of a specified file
func ReadFile(filePath string) ([]byte, error) {
	f, err := os.Open(Subst(filePath))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open file: %s", filePath)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read file stat: %s", filePath)
	}
	s := fi.Size()
	cBytes := make([]byte, s)
	n, err := f.Read(cBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read file: %s", filePath)
	}
	if n == 0 {
		fmt.Printf("file %s is empty\n", filePath)
	}
	return cBytes, err
}

// Subst replaces instances of '${VARNAME}' (eg ${GOPATH}) with the variable.
// Variables names that are not set by the SDK are replaced with the environment variable.
func Subst(path string) string {
	const (
		sepPrefix = "${"
		sepSuffix = "}"
	)

	splits := strings.Split(path, sepPrefix)

	var buffer bytes.Buffer

	// first split precedes the first sepPrefix so should always be written
	buffer.WriteString(splits[0]) // nolint: gas

	for _, s := range splits[1:] {
		subst, rest := substVar(s, sepPrefix, sepSuffix)
		buffer.WriteString(subst) // nolint: gas
		buffer.WriteString(rest)  // nolint: gas
	}

	return buffer.String()
}

// substVar searches for an instance of a variables name and replaces them with their value.
// The first return value is substituted portion of the string or noMatch if no replacement occurred.
// The second return value is the unconsumed portion of s.
func substVar(s string, noMatch string, sep string) (string, string) {
	endPos := strings.Index(s, sep)
	if endPos == -1 {
		return noMatch, s
	}

	v, ok := os.LookupEnv(s[:endPos])
	if !ok {
		return noMatch, s
	}

	return v, s[endPos+1:]
}
