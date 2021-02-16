MAKEFILE_THIS := $(lastword $(MAKEFILE_LIST))
SRC_PATH      := $(patsubst %/,%,$(dir $(abspath $(MAKEFILE_THIS))))
APP_FILE      := demo_basic.json
APP_NAME      := demo_basic

REPO_PATH     ?= $(SRC_PATH)/../..
FAB_PATH      ?= $(REPO_PATH)/../hyperledger/fabric-samples

# Fabric network config files
NETWORK       := $(REPO_PATH)/test-network/config.yaml
MATCHER       := $(REPO_PATH)/test-network/local_entity_matchers.yaml

.PHONY: all
all: build

.PHONY: clean
clean:
	-rm -Rf $(SRC_PATH)/$(APP_NAME)
	-rm $(SRC_PATH)/$(APP_NAME)_app

.PHONY: build
build: $(APP_FILE) clean
	$(REPO_PATH)/scripts/build.sh $(APP_FILE) $(NETWORK) $(MATCHER)

.PHONY: run
run:
	FLOGO_APP_PROP_RESOLVERS=env FLOGO_APP_PROPS_ENV=auto PORT=7979 FLOGO_LOG_LEVEL=DEBUG FLOGO_SCHEMA_SUPPORT=true FLOGO_SCHEMA_VALIDATION=false CRYPTO_PATH=$(FAB_PATH)/test-network/organizations $(SRC_PATH)/$(APP_NAME)_app

.PHONY: test
test:
	curl -u Admin: -X PUT "http://localhost:7979/asset/initledger"
	sleep 5
	curl -u Admin: -X GET "http://localhost:7979/asset/readasset?id=asset1"
	curl -u Admin: -X GET "http://localhost:7979/asset/readasset?id=asset2"

.PHONY: start
start:
	cd $(FAB_PATH)/test-network && ./network.sh up createChannel
	cd $(FAB_PATH)/test-network && ./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go

.PHONY: shutdown
shutdown:
	cd $(FAB_PATH)/test-network && ./network.sh down
