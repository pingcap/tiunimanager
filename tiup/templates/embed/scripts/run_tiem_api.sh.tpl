#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.deploy_dir}}

cd "${DEPLOY_DIR}" || exit 1

LANG=en_US.UTF-8 \
exec env GODEBUG=madvdontneed=1 bin/openapi-server \
    --host="{{.tiem_api_servers.host}}" \
    --port="{{.tiem_api_servers.port}}" \
    --metrics-port="{{.global.metrics_port}}" \
    --registry-address="{{.registry_endpoints}}" \
    --tracer-address="{{.tracer_servers.host}}:{{.tracer_servers.port}}" \
    --elasticsearch-address="{{.elasticsearch_servers.host}}:{{.elasticsearch_servers.port}}" \
    --deploy-dir="{{.tiem_api_servers.deploy_dir}}" \
    --data-dir="{{.tiem_api_servers.data_dir}}" \
    --log-level="{{.global.log_level}}"
