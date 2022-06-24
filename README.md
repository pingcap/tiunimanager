<p align="center">
  <br>
  <img src="./docs/img/tiunimanager.svg" alt="logo" height="80px">
  <br>
  <br>
</p>

# TiUniManager

[![License](https://img.shields.io/badge/license-Apache--2.0-green?style=flat-square)](./LICENSE)

TiUniManager is a database management platform built for operating and managing TiDB, a distributed database.

It allows users to manage TiDB clusters through OpenAPI or [TiUniManager UI](https://github.com/pingcap/tiunimanager), the web-based UI.

## Contents

- library - Common components and resources.
- micro-api - OpenAPI http handler.
- micro-cluster - Core service.
- docs - Documentation.

## Deploy TiUniManager

### Prerequisites

The followings are required for developing TiUniManager UI:

- [Git](https://git-scm.com/downloads)
- [Go 1.17+](https://go.dev/doc/install)

### Preparation

Before you can actually run the service, you need to prepare TiUP and certs.

#### Prepare TiUP

1. Install TiUP by following instructions in [Deploy TiUP on the control machine](https://docs.pingcap.com/tidb/stable/production-deployment-using-tiup#step-2-deploy-tiup-on-the-control-machine).

2. Prepare `TIUP_HOME` which will be used by TiUniManger.

   ```shell
   $ mkdir -p /home/tidb/.tiup/bin
   $ TIUP_HOME=/home/tidb/.tiup tiup mirror set https://tiup-mirrors.pingcap.com
   ```

#### Prepare certs

You need to prepare the following certs in the directory `./bin/cert`, where `aes.key` contains a 32 character string.

- aes.key
- etcd-ca.pem
- etcd-server-key.pem
- etcd-server.pem
- server.crt
- server.csr
- server.key

You can follow [GENERATE CERTS](./build_helper/GENERATE_CERTS.md) to generate these certs.

### Build and Run TiUniManager

TiUniManager can be compiled and used on Linux, OSX, CentOS, It is as simple as:
```
make
```

start cluster-server
```shell
./bin/cluster-server --host=127.0.0.1 --port=4101 --metrics-port=4104 --registry-address=127.0.0.1:4106 --elasticsearch-address=127.0.0.1:4108 --skip-host-init=true --em-version=InTesting --deploy-user=tidb --deploy-group=tidb
```

start openapi-server
```shell
./bin/openapi-server --host=127.0.0.1 --port=4100 --metrics-port=4103 --registry-address=127.0.0.1:4106
```

### Try it out

Now you can check API using Swagger: http://127.0.0.1:4100/swagger/index.html, and you can [use TiUniManager to deploy TiDB clusters](./build_helper/DEPLOY_TIDB.md).

## Interested in contributing?

Read through our [contributing guidelines](./CONTRIBUTING.md) to learn about our submission process and more.

## License

Copyright 2022 PingCAP. All rights reserved.

Licensed under the [Apache 2.0 License](./LICENSE).