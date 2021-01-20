#!/bin/bash
#
# Copyright (c) 2020, TIBCO Software Inc.
# All rights reserved.
#
# SPDX-License-Identifier: BSD-3-Clause-Open-MPI
#
# build fabric client app from model.json. executable will be in the same directory as the model.json
# usage:
#   ./build.sh model-file network-config-file [ entity-matchers-file ]
# e.g.,
#   ./build.sh ../samples/marble/marble_client.json config.yaml
#
# To build executable for different OS, set GO environment variables, e.g.,
#   export GOOS=darwin
#   export GOARCH=amd64

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"; echo "$(pwd)")"

if [ "$#" -lt 2 ]; then
  echo "Usage: ./build.sh model-file network-config-file [ entity-matchers-file ]"
  exit 1
fi

MODEL_DIR="$(cd "$(dirname "$1")"; echo "$(pwd)")"
MODEL=${1##*/}
NAME="${MODEL%.*}"
NETWORK="$(cd "$(dirname "$2")"; pwd -P)/$(basename "$2")"
MATCHER=""
if [ "$#" -gt 2 ]; then
  if [ -f "$3" ]; then
    MATCHER="$(cd "$(dirname "$3")"; pwd -P)/$(basename "$3")"
    MATCHER="-m $MATCHER"
  fi
fi

# create and build source code
cp ${SCRIPT_DIR}/template.mod ${MODEL_DIR}/go.mod
cd ${MODEL_DIR}
flogo create -f ${MODEL} -m go.mod ${NAME}

# create fabric network config Go file
cd ${NAME}
flogo configfabric -c ${NETWORK} ${MATCHER}
cd src
go mod tidy

cd ..
flogo build -e --verbose

# copy executable
if [ -f "bin/${NAME}" ]; then
  cp bin/${NAME} ${MODEL_DIR}/${NAME}_app
  echo "Created executable ${MODEL_DIR}/${NAME}_app"
else
  echo "failed to build app"
  exit 1
fi

# cleanup build files
cd ${MODEL_DIR}
if [ -f "${MODEL_DIR}/${NAME}_app" ]; then
  echo "cleanup build files"
  rm -R ${MODEL_DIR}/${NAME}
  rm ${MODEL_DIR}/go.mod
fi