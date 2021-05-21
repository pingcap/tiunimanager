module github.com/pingcap/tcp

go 1.16

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/asim/go-micro/plugins/logger/logrus/v3 v3.0.0-20210517071652-f48911d2c3ef
	github.com/asim/go-micro/plugins/registry/etcd/v3 v3.0.0-20210517071652-f48911d2c3ef
	github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3 v3.0.0-20210513120725-4c1f81dadb47
	github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3 v3.0.0-20210511075819-32cb1b435b9b
	github.com/asim/go-micro/v3 v3.5.1
	github.com/gin-gonic/gin v1.7.1
	github.com/golang/protobuf v1.5.2
	github.com/jackdoe/gin-basic-auth-dynamic v0.0.0-20201112112728-ede5321b610c
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/uber/jaeger-client-go v2.28.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	google.golang.org/protobuf v1.26.0
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.10
	gorm.io/plugin/opentracing v0.0.0-20210506132430-24a9caea7709
)
