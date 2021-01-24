#
# Copyright (c) 2020, TIBCO Software Inc.
# All rights reserved.
#
# SPDX-License-Identifier: BSD-3-Clause-Open-MPI
#
MAKEFILE_THIS := $(lastword $(MAKEFILE_LIST))
SRC_PATH      := $(patsubst %/,%,$(dir $(abspath $(MAKEFILE_THIS))))
APP_FILE      := sample_rest.json
APP_NAME      := sample_rest
CONTRACT      := sample-contract.json

REPO_PATH     ?= $(SRC_PATH)/..
FAB_PATH      ?= $(REPO_PATH)/../hyperledger/fabric-samples

# Fabric network config files
NETWORK       := $(REPO_PATH)/test-network/config.yaml
MATCHER       := $(REPO_PATH)/test-network/local_entity_matchers.yaml

# uncomment this line to build for Flogo Enterprise
FE            := -e

.PHONY: all
all: build

.PHONY: clean
clean:
	-rm -Rf $(SRC_PATH)/$(APP_NAME)
	-rm $(SRC_PATH)/$(APP_NAME)_app

.PHONY: build
build: $(CONTRACT) clean
	flogo contract2rest $(FE) -c $(CONTRACT) -o $(APP_FILE)
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
