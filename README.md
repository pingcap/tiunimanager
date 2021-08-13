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
TOKEN=token-tidb-cloud-platform
CLUSTER_STATE=new
NAME_1=machine-1
HOST_1=127.0.0.1
CLUSTER=${NAME_1}=http://${HOST_1}:2380

THIS_NAME=${NAME_1}
THIS_IP=${HOST_1}
etcd --data-dir=data.etcd1 --name ${THIS_NAME} \
    --initial-advertise-peer-urls http://${THIS_IP}:2380 --listen-peer-urls http://${THIS_IP}:2380 \
    --advertise-client-urls http://${THIS_IP}:2379 --listen-client-urls http://${THIS_IP}:2379 \
    --initial-cluster ${CLUSTER} \
    --initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}
```

### Setup Jaeger (for opentracing)

Download binary from https://www.jaegertracing.io/download/.

```shell
$ jaeger-all-in-one --collector.zipkin.host-port=:9411
```

And visit the web interface from port 16686.

### Run Service

start micro-metadb
```
cd micro-metadb
$ go run main.go
```

start micro-cluster
```
cd micro-metadb
$ go run main.go 
```

start micro-api
```
cd micro-api
$ go run main.go 
```

### Try it out
via swagger : http://localhost:8080/swagger/index.html
or : http://localhost:8080/system/check

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
		log.Info("user:", u)
		if nil == bcrypt.CompareHashAndPassword([]byte(u.FinalHash), []byte(u.Salt+passwd)) {
			log.Debug("check success")
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
    log.Debug("init: sqlite.open")
    db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
    /* ... */
}
```

Basically, use `logger.WithContext(ctx)` to get the `log` if there is a valid ctx to inherit, or otherwise, use `logger.WithContext(nil)` to get the default log.

```go
func (d *Db) CheckUser(ctx context.Context, req *dbPb.CheckUserRequest, rsp *dbPb.CheckUserResponse) error {
	ctx = logger.NewContext(ctx, logger.Fields{"micro-service": "CheckUser"})
	log := logger.WithContext(ctx)
	e := models.CheckUser(ctx, req.Name, req.Passwd)
	if e == nil {
		log.Debug("CheckUser success")
	} else {
		log.Errorf("CheckUser failed: %s", e)
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}
```

Use something like `logger.NewContext(ctx, logger.Fields{"micro-service": "CheckUser"})` to create a new ctx from the old ctx with new customized log fields added.
