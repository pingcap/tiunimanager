#!/usr/bin/env bash

# This script refers to the README.md in root folder of TiEM project
# etcd --data-dir=data.etcd1 --name machine-1 --initial-advertise-peer-urls http://127.0.0.1:2480 --listen-peer-urls http://127.0.0.1:2480  --advertise-client-urls http://127.0.0.1:2479 --listen-client-urls http://127.0.0.1:2479 --initial-cluster machine-1=http://127.0.0.1:2480  --initial-cluster-state new --initial-cluster-token token-tidb-cloud-platform &
cd /usr/local/tiem/bin
./metadb-server --registry etcd --registry_address 127.0.0.1:2479 &
./cluster-server --registry etcd --registry_address 127.0.0.1:2479 &
./openapi-server --registry etcd --registry_address 1127.0.0.1:2479 &