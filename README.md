<p align="center">
  <br>
  <img src="./docs/img/tiunimanager.svg" alt="logo" height="80px">
  <br>
  <br>
</p>

# TiUniManager

[![License](https://img.shields.io/badge/license-Apache--2.0-green?style=flat-square)](./LICENSE)

## Contents

- library - Common components and resources.
- micro-api - OpenAPI http handler.
- micro-cluster - Core service.
- docs - Documentation.

## Build and Run TiUniManager

TiUniManager can be compiled and used on Linux, OSX, CentOS, It is as simple as:
```
make
```

After building TiUniManager, it is good idea to test it using:
```
make test
```

## Preparation

Before you can actually run the service, you need to prepare TiUP and certs.

### Prepare TiUP

1. Install TiUP by following instructions in [Deploy TiUP on the control machine](https://docs.pingcap.com/tidb/stable/production-deployment-using-tiup#step-2-deploy-tiup-on-the-control-machine).

2. Prepare `TIUP_HOME` which will be used by TiUniManger.

   ```shell
   $ mkdir -p /home/tidb/.tiup/bin
   $ TIUP_HOME=/home/tidb/.tiup tiup mirror set https://tiup-mirrors.pingcap.com
   ```

### Prepare certs

You need to prepare the following certs in the directory `./bin/cert`, where `aes.key` contains a 32 character string.

- aes.key
- etcd-ca.pem
- etcd-server-key.pem
- etcd-server.pem
- server.crt
- server.csr
- server.key

You can follow [GENERATE CERTS](./build_helper/GENERATE_CERTS.md) to generate these certs.

### Run Service
If you want to run the service locally, try the following commands, enjoy it!

start cluster-server
```shell
$ ./bin/cluster-server --host=127.0.0.1 --port=4101 --metrics-port=4104 --registry-address=127.0.0.1:4106 --elasticsearch-address=127.0.0.1:4108 --skip-host-init=true --em-version=InTesting --deploy-user=tidb --deploy-group=tidb
```

start openapi-server
```shell
$ ./bin/openapi-server --host=127.0.0.1 --port=4100 --metrics-port=4103 --registry-address=127.0.0.1:4106
```

### Try it out
via swagger : http://localhost:4116/swagger/index.html
or : http://localhost:4116/system/check

### Watch Traces

Visit the web interface of Jaeger from port 16686.

![opentrace1](docs/img/opentrace1.png)

![opentrace2](docs/img/opentrace2.png)

### Collect Promethus Merics

The Promethus metrics is exported as http resource and is located at ":${PrometheusPort}/metrics" which the `PrometheusPort` is defined in the configure file.

## Code Style Guide

See [Effective Go](https://golang.org/doc/effective_go) for the style guide.

## How to Log

For example:

```go
import(
	"github.com/pingcap/tiunimanager/addon/logger"
)
```

```go
func CheckUser(ctx context.Context, name, passwd string) error {
	log := logger.WithContext(ctx).WithField("models", "CheckUser").WithField("name", name)
	u, err := FindUserByName(ctx, name)
	if err == nil {
		getLogger().Info("user:", u)
		if nil == bcrypt.CompareHashAndPassword([]byte(u.FinalHash), []byte(u.Salt+passwd)) {
			getLogger().Debug("check success")
			return nil
		} else {
			return fmt.Errorf("failed")
		}
	} else {
		return err
	}
}
```

```go
func init() {
    var err error
    dbFile := "tiunimanager.sqlite.db"
    log := logger.WithContext(nil).WithField("dbFile", dbFile)
    getLogger().Debug("init: sqlite.open")
    db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
    /* ... */
}
```

Basically, use `logger.WithContext(ctx)` to get the `log` if there is a valid ctx to inherit, or otherwise, use `logger.WithContext(nil)` to get the default getLogger().

```go
func (d *Db) CheckUser(ctx context.Context, req *dbPb.CheckUserRequest, rsp *dbPb.CheckUserResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "CheckUser"})
	log := logger.WithContext(ctx)
	e := models.CheckUser(ctx, req.Name, req.Passwd)
	if e == nil {
		getLogger().Debug("CheckUser success")
	} else {
		getLogger().Errorf("CheckUser failed: %s", e)
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}
```

Use something like `logger.NewContext(ctx, logger.Fields{"micro-service": "CheckUser"})` to create a new ctx from the old ctx with new customized log fields added.

## Interested in contributing?

Read through our [contributing guidelines](./CONTRIBUTING.md) to learn about our submission process and more.

## License

Copyright 2022 PingCAP. All rights reserved.

Licensed under the [Apache 2.0 License](./LICENSE).