#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}

cd "${DEPLOY_DIR}" || exit 1

LANG=en_US.UTF-8 \
exec env GODEBUG=madvdontneed=1 bin/openapi-server \
    --host="{{.Host}}" \
    --port="{{.Port}}" \
    --metrics-port="{{.MetricsPort}}" \
    --registry-address="{{.registry_endpoints}}" \
    --tracer-address="{{.tracer_servers.host}}:{{.tracer_servers.port}}" \
    --elasticsearch-address="{{.elasticsearch_servers.host}}:{{.elasticsearch_servers.port}}" \
    --deploy-dir="{{.DeployDir}}" \
    --data-dir="{{.DataDir}}" \
    --log-level="{{.global.log_level}}"
