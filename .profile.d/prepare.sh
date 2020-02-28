#!/bin/bash

wget https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 -O jq
chmod +x jq
export MOTTAINAI_DB__DB_PATH="$(echo $VCAP_SERVICES | ./jq -r '."persi-nfs"[0]."volume_mounts"[0]."container_dir"')"
rm -rf jq

export MOTTAINAI_WEB__PORT=$PORT
export MOTTAINAI_WEB__LISTENADDRESS=$VCAP_APP_HOST
