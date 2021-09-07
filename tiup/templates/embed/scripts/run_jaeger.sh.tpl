# !/bin/bash
set -e

DEPLOY_DIR={{.DeployDir}}
cd "${DEPLOY_DIR}" || exit 1

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!

exec > >(tee -i -a "{{.LogDir}}/prometheus.log")
exec 2>&1

exec bin/jaeger-all-in-one \
    --collector.zipkin.host-port=":{{.tracer_servers.port}}"
