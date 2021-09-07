#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.deploy_dir}}

cd "${DEPLOY_DIR}" || exit 1

LANG=en_US.UTF-8 \
exec env GODEBUG=madvdontneed=1 bin/metadb-server \
    --host="{{.tiem_metadb_servers.host}}" \
    --port="{{.tiem_metadb_servers.port}}" \
    --registry-client-port="{{.tiem_metadb_servers.registry_client_port}}" \
    --registry-peer-port="{{.tiem_metadb_servers.registry_peer_port}}" \
    --metrics-port="{{.global.metrics_port}}" \
    --registry-address="{{.registry_endpoints}}" \
    --tracer-address="{{.tracer_servers.host}}:{{.tracer_servers.port}}" \
    --deploy-dir="{{.tiem_metadb_servers.deploy_dir}}" \
    --data-dir="{{.tiem_metadb_servers.data_dir}}" \
    --log-level="{{.global.log_level}}"
