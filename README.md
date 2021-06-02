# TCP - TiDB Cloud Platform

## Contents

- addon - Addons.
- api - OpenAPI http handler.
- auth - Authentication.
- client - Client of go-micro service.
- config - Configurations.
- docs - Documentation.
- micro - Final main.go for different microservices.
- models - Models of database.
- proto - Protobuf definitions.
- router - Router of http handler.
- service - go-micro service implementation.

## Build and Run TCP

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
cd proto/common
protoc --proto_path=$GOPATH/src:. --micro_out=. --go_out=. *.proto
cd proto/db
protoc --proto_path=$GOPATH/src:. --micro_out=. --go_out=. *.proto
```

### Install And Setup local etcd service

Download the binary from https://github.com/etcd-io/etcd/releases/.

Setup:

```
TOKEN=token-tidb-cloud-platform
CLUSTER_STATE=new
NAME_1=machine-1
NAME_2=machine-2
NAME_3=machine-3
HOST_1=127.0.0.1
HOST_2=127.0.0.2
HOST_3=127.0.0.3
CLUSTER=${NAME_1}=http://${HOST_1}:2380,${NAME_2}=http://${HOST_2}:2380,${NAME_3}=http://${HOST_3}:2380

# For machine 1
THIS_NAME=${NAME_1}
THIS_IP=${HOST_1}
etcd --data-dir=data.etcd1 --name ${THIS_NAME} \
    --initial-advertise-peer-urls http://${THIS_IP}:2380 --listen-peer-urls http://${THIS_IP}:2380 \
    --advertise-client-urls http://${THIS_IP}:2379 --listen-client-urls http://${THIS_IP}:2379 \
    --initial-cluster ${CLUSTER} \
    --initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}

# For machine 2
THIS_NAME=${NAME_2}
THIS_IP=${HOST_2}
etcd --data-dir=data.etcd2 --name ${THIS_NAME} \
    --initial-advertise-peer-urls http://${THIS_IP}:2380 --listen-peer-urls http://${THIS_IP}:2380 \
    --advertise-client-urls http://${THIS_IP}:2379 --listen-client-urls http://${THIS_IP}:2379 \
    --initial-cluster ${CLUSTER} \
    --initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}

# For machine 3
THIS_NAME=${NAME_3}
THIS_IP=${HOST_3}
etcd --data-dir=data.etcd3 --name ${THIS_NAME} \
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

```shell
$ cd micro/common
$ cat cfg.toml
OpenApiPort = 4443
PrometheusPort = 8080

[Certificates]
  CrtFilePath = "../../config/example/server.crt"
  KeyFilePath = "../../config/example/server.key"

$ go run main.go --tidb-cloud-platform-conf-file cfg.toml --registry etcd --registry_address 127.0.0.1:2379,127.0.0.2:2379,127.0.0.3:2379
```

```bash
$ cd micro/db
$ cat cfg.toml
SqliteFilePath = "./tcp.sqlite.db"
OpenApiPort = 4444
PrometheusPort = 8081

[Certificates]
  CrtFilePath = "../../config/example/server.crt"
  KeyFilePath = "../../config/example/server.key"

$ go run main.go --tidb-cloud-platform-conf-file cfg.toml --registry etcd --registry_address 127.0.0.1:2379,127.0.0.2:2379,127.0.0.3:2379
```

### Run Client

```
$ curl --insecure -su "admin:admin" -vX GET \
    -H "Content-type: application/json"   \
    -H "Accept: application/json"   \
    -d '{"NamE":"bar"}'   \
    "https://127.0.0.1:4443/api/hello"

{"code":"200","data":{"greeting":"Hello bar"}}

$ curl --insecure -su "admin:dontknow" -vX GET \
    -H "Content-type: application/json"   \
    -H "Accept: application/json"   \
    -d '{"NamE":"bar"}'   \
    "https://127.0.0.1:4443/api/hello"

< HTTP/1.1 401 Unauthorized
< Content-Type: text/plain; charset=utf-8
< Www-Authenticate: Basic realm="Authorization Required"
< Date: Tue, 11 May 2021 10:47:49 GMT
< Content-Length: 14
<
not authorized
```

### Watch Traces

Visit the web interface of Jaeger from port 16686.

![opentrace1](docs/img/opentrace1.png)

![opentrace2](docs/img/opentrace2.png)

### Collect Promethus Merics

The Promethus metrics is exported at ":8080/metrics".

## Code Style Guide

See [Effective Go](https://golang.org/doc/effective_go) for the style guide.

## How to Log

For example:

```go
import(
	"github.com/pingcap/tcp/addon/logger"
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
    dbFile := "tcp.sqlite.db"
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
