# HTTP service for contract sample

This sample shows how to generate an HTTP service app for a [blockchain contract](./sample-contract.json), and use use the service to test the corresponding chaincode.

The [sample contract](./sample-contract.json) is defined based on the [contract schema](https://github.com/open-dovetail/fabric-chaincode/blob/master/contract/contract-schema.json), and it can be built, installed and run on a Fabric network as described in [README.md](https://github.com/open-dovetail/fabric-chaincode/blob/master/contract/README.md).

## Prerequisite

Set up development environment and install the smart contract on the Fabric test-network as described in [README.md](https://github.com/open-dovetail/fabric-chaincode/blob/master/contract/README.md).

## Build an HTTP service app for the contract

In a terminal console, change to this directory, and type the command `make`, which will perform the following steps:

- Use `flogo contract2rest` CLI extension to read the [sample-contract.json](./sample-contract.json), and geneate a Flogo HTTP service app `sample_rest.json`;
- Build the Flogo model, `sample_rest.json`, to an executable `sample_rest_app`.

## Start the HTTP service and test the smart contract

Execute following steps to start the HTTP service and invoke the **sample_cc** chaincode that is deployed on the Fabric test-network by the prerequisite steps:

```bash
# start HTTP service
make run

# test the HTTP service in another terminal
make test
```

## View and edit the Flogo model

You can view and edit the generated HTTP service app in a web-browser. First, start the **Flogo Web UI**:

```bash
docker run -it -p 3303:3303 yxuco/flogo-ui eula-accept
```

Open the **Flogo Web UI** in a web-browser by using the URL: `http://localhost:3303`. Then install the required Flogo extensions as listed in [README](https://github.com/open-dovetail/fabric-client#view-and-edit-flogo-model), and import the app by selecting the generated model file `sample_rest.json`.

If you have license to [Flogo Enterprise](https://edelivery.tibco.com/storefront/eval/tibco-flogo-enterprise/prod11810.html), you can uncomment the line for `FE` in the [Makefile](./Makefile), and then execute `make` to generate model file for Flogo Enterprise. You can then start the Flogo Enterprise with `/path/to/flogo/2.10/bin/start-webui.sh`. Import the generated model, `sample_rest.json`, you can then use the Web UI to edit the model, which is quite a bit more user-friendly than the open-source version of the Flogo Web UI.
