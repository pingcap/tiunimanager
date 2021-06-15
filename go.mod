module github.com/pingcap/ticp

go 1.16

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/asim/go-micro/plugins/logger/logrus/v3 v3.0.0-20210517071652-f48911d2c3ef
	github.com/asim/go-micro/plugins/registry/etcd/v3 v3.0.0-20210517071652-f48911d2c3ef
	github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3 v3.0.0-20210517071652-f48911d2c3ef
	github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3 v3.0.0-20210517071652-f48911d2c3ef
	github.com/asim/go-micro/v3 v3.5.1
	github.com/gin-gonic/gin v1.7.2
	github.com/golang/protobuf v1.5.2
	github.com/jackdoe/gin-basic-auth-dynamic v0.0.0-20201112112728-ede5321b610c
	github.com/mgechev/revive v1.0.2
	github.com/micro/cli/v2 v2.1.2
	github.com/opentracing/opentracing-go v1.2.0
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pingcap/errors v0.11.5-0.20201126102027-b0a155152ca3
	github.com/pingcap/failpoint v0.0.0-20200702092429-9f69995143ce
	github.com/prometheus/client_golang v1.10.0
	github.com/sasha-s/go-deadlock v0.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/swaggo/swag v1.6.6-0.20200529100950-7c765ddd0476
	github.com/uber/jaeger-client-go v2.29.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/tools v0.0.0-20200225230052-807dcd883420
	google.golang.org/protobuf v1.26.0
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.10
	gorm.io/plugin/opentracing v0.0.0-20210506132430-24a9caea7709
)
