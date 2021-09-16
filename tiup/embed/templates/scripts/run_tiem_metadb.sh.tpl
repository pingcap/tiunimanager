#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}

cd "${DEPLOY_DIR}" || exit 1

LANG=en_US.UTF-8 \
exec env GODEBUG=madvdontneed=1 bin/metadb-server \
    --host="{{.Host}}" \
    --port="{{.Port}}" \
    --registry-client-port="{{.ClientPort}}" \
    --registry-peer-port="{{.PeerPort}}" \
    --metrics-port="{{.MetricsPort}}" \
    --registry-address="{{.RegistryEndpoints}}" \
    --tracer-address="{{.TracerAddress}}" \
    --deploy-dir="{{.DeployDir}}" \
    --data-dir="{{.DataDir}}" \
    --log-level="{{.LogLevel}}"
