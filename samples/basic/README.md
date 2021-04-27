# asset-transfer-basic

This example uses the [Open-source Flogo](http://www.flogo.io/) to implement an HTTP service app that invokes 2 of the transactions in the [Hyperledger Fabric](https://www.hyperledger.org/projects/fabric) sample chaincode [asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/master/asset-transfer-basic).

## Prerequisite

Set up development environment by following the **Getting Started** instructions in [README.md](../../README.md).

## Build and deploy chaincode to Hyperledger Fabric

The Flogo model [demo_basic.json](./demo_basic.json) is the HTTP service implementation. In a terminal console, type the command `make`, to build it into an executable `demo_basic_app`.

## Start Fabric test-network and test the HTTP service

Execute following steps to start the **Fabric test-network** and send HTTP request to test the HTTP service:

```bash
# start Fabric test-network and deploy the sample chaincode
make start

# start HTTP service
make run

# in another terminal, send HTTP request
make test
```

## Shutdown test-network

After successful test, you may shutdown the **Fabric test-network**:

```bash
make shutdown
```

## View and edit Flogo model

You can view and edit the chaincode implementation in a web-browser. First, start the **Flogo Web UI**:

```bash
docker run -it -p 3303:3303 yxuco/flogo-ui eula-accept
```

Open the **Flogo Web UI** in a web-browser by using the URL: `http://localhost:3303`. Then import the app by selecting the model file [demo_basic.json](./demo_basic.json).

For problems of importing the model, refer the troubleshoot instructions [here](../../README.md).
