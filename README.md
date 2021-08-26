# TiEM - TiDB Enterprise Manager

## Contents

- library - Common components and resources.
- micro-api - OpenAPI http handler.
- micro-cluster - Core service.
- micro-metadb - Client of go-micro service.
- docs - Documentation.

## Build and Run TiEM

### Dependencies

Install the protobuf stuff

```
go get github.com/asim/go-micro/cmd/protoc-gen-micro/v3
go get google.golang.org/protobuf/cmd/protoc-gen-go

# Install protoc Refer to http://google.github.io/proto-lens/installing-protoc.html
PROTOC_ZIP=protoc-3.14.0-linux-x86_64.zip
curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/$PROTOC_ZIP
sudo unzip -o $PROTOC_ZIP -d /usr/local bin/protoc
sudo unzip -o $PROTOC_ZIP -d /usr/local 'include/*'
rm -f $PROTOC_ZIP
```

### Generate Protobuf files

```
cd micro-metadb/proto
protoc --proto_path=$GOPATH/src:. --micro_out=. --go_out=. *.proto
cd micro-cluster/proto
protoc --proto_path=$GOPATH/src:. --micro_out=. --go_out=. *.proto
```

### Install And Setup local etcd service

Download the binary from https://github.com/etcd-io/etcd/releases/.

Setup:

```
etcd --data-dir=data.etcd1 --name machine-1 \
    --initial-advertise-peer-urls http://127.0.0.1:4102 --listen-peer-urls http://127.0.0.1:4102 \
    --advertise-client-urls http://127.0.0.1:4101 --listen-client-urls http://127.0.0.1:4101 \
    --initial-cluster machine-1=http://127.0.0.1:4102 \
    --initial-cluster-state new --initial-cluster-token token-tiem
    &
```

### Setup Jaeger (for opentracing)

Download binary from https://www.jaegertracing.io/download/.

```shell
$ jaeger-all-in-one --collector.zipkin.host-port=:9411
```

And visit the web interface from port 16686.

### Run Service

start micro-metadb
```shell
$ cd micro-metadb
$ go run init.go main.go
```
or
```shell
$ cd micro-metadb
$ go run init.go main.go \
    --host=192.168.1.100 \
    --port=4100 \
    --registry-client-port=4101 \
    --registry-peer-port=4102 \
    --metrics-port=4121 \
    --registry-address=192.168.1.100:4101,192.168.1.101:4101,192.168.1.102:4101 \
    --tracer-address=192.168.1.100:4123 \
    --deploy-dir=/tiem-deploy/tiem-metadb-4100 \
    --data-dir=/tiem-data/tiem-metadb-4100 \
    --log-level=info
```

start micro-cluster
```shell
$ cd micro-cluster
$ go run main.go 
```
or
```shell
$ cd micro-cluster
$ go run init.go main.go \
    --host=192.168.1.100 \
    --port=4110 \
    --metrics-port=4121 \
    --registry-address=192.168.1.100:4101,192.168.1.101:4101,192.168.1.102:4101 \
    --tracer-address=192.168.1.100:4123 \
    --deploy-dir=/tiem-deploy/tiem-cluster-4110 \
    --data-dir=/tiem-data/tiem-cluster-4110 \
    --log-level=info
```

start micro-api
```shell
$ cd micro-api
$ go run main.go 
```
or
```shell
$ cd micro-api
$ go run init.go main.go \
    --host=192.168.1.100 \
    --port=4116 \
    --metrics-port=4121 \
    --registry-address=192.168.1.100:4101,192.168.1.101:4101,192.168.1.102:4101 \
    --tracer-address=192.168.1.100:4123 \
    --deploy-dir=/tiem-deploy/tiem-api-4115 \
    --data-dir=/tiem-data/tiem-api-4115 \
    --log-level=info
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
	"github.com/pingcap/tiem/addon/logger"
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
    dbFile := "tiem.sqlite.db"
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
