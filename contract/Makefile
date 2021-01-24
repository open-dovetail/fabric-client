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
CHAINCODE     := sample_cc
PORT          := 7979

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
	FLOGO_APP_PROP_RESOLVERS=env FLOGO_APP_PROPS_ENV=auto PORT=$(PORT) CHAINCODE=$(CHAINCODE) FLOGO_LOG_LEVEL=DEBUG FLOGO_SCHEMA_SUPPORT=true FLOGO_SCHEMA_VALIDATION=false CRYPTO_PATH=$(FAB_PATH)/test-network/organizations $(SRC_PATH)/$(APP_NAME)_app

.PHONY: test
test:
	curl -u broker: -X POST -H 'Content-Type: application/json' -d '{"name":"marble20","color":"blue","size":35,"owner":"tom"}' http://localhost:$(PORT)/marble/createmarble
	sleep 3
	curl -u tom: -X POST -H 'Content-Type: application/json' -d '{"name":"marble20","newOwner":"jerry"}' http://localhost:$(PORT)/marble/transfermarble
	sleep 3
	curl -u tom: -X POST -H "Content-Type: application/json" -d '{"name":"marble20"}' "http://localhost:$(PORT)/marble/gethistory"
	curl -u jerry@org2: -X POST -H 'Content-Type: application/json' -d '{"owner":"jerry"}' http://localhost:$(PORT)/marble/querymarblesbyowner
	curl -u jerry@org2: -X POST -H 'Content-Type: application/json' -d '{"color":"blue","newOwner":"tom"}' http://localhost:$(PORT)/marble/transfermarblesbasedoncolor
	sleep 3
	curl -u tom: -X POST -H "Content-Type: application/json" -d '{"name":"marble20"}' "http://localhost:$(PORT)/marble/gethistory"
