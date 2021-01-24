# Fabric Request activity

This Flogo activity contribution can be configured to send `invoke` or `query` request from a client app to a specified chaincode deployed on a Fabric network. Most of the request operations are demonstrated in the [contract example](../../contract).

## Configuration and Inputs

This operation can be configured to support all types of client requests for Fabric chaincode, e.g.,

```json
    "activity": {
        "ref": "#request",
        "settings": {
            "connectionName": "=$property[\"NETWORK\"]",
            "channelID": "=$property[\"CHANNEL\"]",
            "chaincodeID": "=$property[\"CHAINCODE\"]",
            "transactionName": "queryMarblesByOwner",
            "parameters": "owner",
            "requestType": "query",
            "userOrgOnly": false
        },
        "input": {
            "parameters": "=$flow.parameters",
            "transient": {},
            "userName": "=$flow.user",
            "timeoutMillis": 0,
            "endpoints": []
        }
    }
```

Notes on the configuration and input parameters:

- **connectionName** identifies a Fabric network, e.g., `test-network`. The network configuration and local entity matchers patterns are not configured by the activity. Instead, they are provided when the application is built by using the command `flogo configfabric`. This late binding approach provides more flexibility for building an app model for multiple chaincode deployments.
- **parameters under settings** contain a comma-delimited names of parameters of the specified transaction. It defines the sequence of the parameters in the input.
- **requestType** is `invoke` or `query`. You may use `query` for read-only operations, and so it will not go through the endorsment process.
- **userOrgOnly** specifies an end-point filter. When it is turned on, the request will be sent to only the peers of the user's organization.
- **transient** specifies transient data that should not be sent to distributed ledger, nor orderer processes.
- **userName** specifies `user@org` that is used to invoke chaincode transactions. The `user` must be a valid blockchain user with CA crypto data accessible by the HTTP server. The `org` is optional, which specifies the user's organization as specified in the Fabric network config file. If `org` is not specified, the `user` is assumed to be part of the client organization specified by the Fabric network configuration.
- **timeoutMillis** specifies the wait time for responses from the Fabric network.
- **endpoints** is a list of peers to send the request to. It is typically left blank, and so the SDK will randomly choose an available peer to send the Fabric request. This list, if specified, overrides the settings for `userOrgOnly`.
