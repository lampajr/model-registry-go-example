#!/bin/bash

MLMD_PORT="${MLMD_PORT:-9090}"

docker run -p $MLMD_PORT:8080 --env METADATA_STORE_SERVER_CONFIG_FILE=/tmp/shared/conn_config.pb --volume ./run/:/tmp/shared --name mlmd-server gcr.io/tfx-oss-public/ml_metadata_store_server:1.14.0