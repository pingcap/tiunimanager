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
    --registry-peer-port="{{.tiem_metadb_servers.registry_peer_port}}" \
    --metrics-port="{{.MetricsPort}}" \
    --registry-address="{{.registry_endpoints}}" \
    --tracer-address="{{.tracer_servers.host}}:{{.tracer_servers.port}}" \
    --deploy-dir="{{.DeployDir}}" \
    --data-dir="{{.DataDir}}" \
    --log-level="{{.global.log_level}}"
