#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}

cd "${DEPLOY_DIR}" || exit 1

LANG=en_US.UTF-8 \
exec env GODEBUG=madvdontneed=1 bin/web-server \
    --host="{{.Host}}" \
    --port="{{.Port}}" \
    --registry-address="{{.registry_endpoints}}" \
    --deploy-dir="$DEPLOY_DIR"

