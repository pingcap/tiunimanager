#!/bin/bash
set -e

DEPLOY_DIR={{.DeployDir}}
cd "${DEPLOY_DIR}" || exit 1

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!

exec > >(tee -i -a "{{.LogDir}}/jaeger.log")
exec 2>&1

exec bin/jaeger-all-in-one \
    --collector.zipkin.host-port=":{{.Port}}" \
    --query.http-server.host-port=":{{.WebPort}}" \
    --processor.zipkin-compact.server-host-port=":{{.ZipkinCompactPort}}" \
    --processor.jaeger-compact.server-host-port=":{{.JaegerCompactPort}}" \
    --processor.jaeger-binary.server-host-port=":{{.JaegerBinaryPort}}" \
    --http-server.host-port=":{{.JaegerHttpPort}}" \
    --collector.http-server.host-port=":{{.CollectorHttpPort}}" \
    --admin.http.host-port=":{{.AdminHttpPort}}" \
    --collector.grpc-server.host-port=":{{.CollectorGrpcPort}}" \
    --query.grpc-server.host-port=":{{.QueryGrpcPort}}"