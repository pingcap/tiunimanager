#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}

cd "${DEPLOY_DIR}" || exit 1

#if [ -f "~/.bash_profile" ]; then
#  source ~/.bash_profile
#
#fi

LANG=en_US.UTF-8 \
exec env GODEBUG=madvdontneed=1 bin/cluster-server \
    --host="{{.Host}}" \
    --port="{{.Port}}" \
    --metrics-port="{{.MetricsPort}}" \
    --registry-client-port="{{.ClientPort}}" \
    --registry-peer-port="{{.PeerPort}}" \
    --registry-address="{{.RegistryEndpoints}}" \
    --em-cluster-name="{{.ClusterName}}" \
    --em-version="{{.ClusterVersion}}" \
    --elasticsearch-address="{{.ElasticsearchAddress}}" \
    --tracer-address="{{.TracerAddress}}" \
    --deploy-user="{{.DeployUser}}" \
    --login-host-user="{{.LoginHostUser}}" \
    --login-private-key-path="{{.LoginPrivateKeyPath}}" \
    --login-public-key-path="{{.LoginPublicKeyPath}}" \
    --deploy-group="{{.DeployGroup}}" \
    --deploy-dir="{{.DeployDir}}/bin" \
    --data-dir="{{.DataDir}}" \
    --log-level="{{.LogLevel}}"
