# Flogo extension for Hyperledger Fabric client

This [Flogo](http://www.flogo.io/) extension is designed to allow developers to design and implement client apps to invoke Hyperledger Fabric chaincode in the Flogo visual programming environment. This extension supports the following release versions:

- [Flogo Web UI](http://www.flogo.io/)
- [Hyperledger Fabric 2.2](https://www.hyperledger.org/projects/fabric)

The Flogo extension supports the following activity to send request of `invoke` or `query` to a chaincode deployed on a Hyperledger Fabric network.

- [**Request**](activity/request): Configure request type in activity setting; Use Flogo CLI plugin `flogo configfabric` to specify Fabric network configuration.

With these Flogo extensions, Hyperledger Fabric client app can be designed and implemented by using the **Flogo Web UI** with zero code. The client app can use any other available Flogo triggers and activities implemented by the open-source community of Flogo.

## Getting Started

- Setup Fabric chaincode development environment as described in [README.md](https://github.com/open-dovetail/fabric-chaincode/blob/master/README.md).
- Build and run a sample Flogo app [demo](./samples/basic) as described in [README.md](./samples/basic/README.md). It is an HTTP service that invokes transactions of the Fabric sample [asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/master/asset-transfer-basic)

## Generate HTTP service to test smart contract

For smart contract developers who do not want to code, you can define chaincode transactions in a JSON file as illustrated [here](https://github.com/open-dovetail/fabric-chaincode/tree/master/contract).

You can generate an HTTP service for the contract JSON file, e.g., [sample-contract.json](./contract/sample-contract.json), and the test the chaincode for the contract by using HTTP requests as described in [README.md](./contract/README.md).

## View and edit Flogo model

You can view and edit the client app implementation in a web-browser. First, start the **Flogo Web UI**:

```bash
docker run -it -p 3303:3303 flogo/flogo-docker eula-accept
```

Open the **Flogo Web UI** in a web-browser by using the URL: `http://localhost:3303`.

Install the following Dovetail contributions, i.e., click the link `Install contribution` at the top-right corner of the UI, and then enter the following URL to install. If the installation fails, you can follow the `Troubleshoot` steps below to patch Flogo libs, and then retry the installation.

- github.com/open-dovetail/fabric-client/activity/request
- github.com/open-dovetail/dovetail-contrib/function/dovetail
- github.com/open-dovetail/dovetail-contrib/trigger/rest

You can then import a sample app by selecting the model file [samples/basic/demo_basic.json](./samples/basic/demo_basic.json).

## Troubleshoot

### Failed to import Flogo model

Make sure that you have installed required Flogo contributions listed above.

### Failed to install dovetail contributions in Web UI

Refer to the resolution [here](https://github.com/open-dovetail/fabric-chaincode#troubleshoot).
