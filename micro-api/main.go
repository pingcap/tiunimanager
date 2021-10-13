package main

import (
	"fmt"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"time"

	"github.com/pingcap-inc/tiem/library/thirdparty/etcd_clientv2"

	"github.com/pingcap-inc/tiem/library/knowledge"

	"github.com/pingcap-inc/tiem/micro-api/metrics"

	"github.com/asim/go-micro/v3"
	"github.com/gin-gonic/gin"
	_ "github.com/pingcap-inc/tiem/docs"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/route"
)

// @title TiEM UI API
// @version 1.0
// @description TiEM UI API

// @contact.name zhangpeijin
// @contact.email zhangpeijin@pingcap.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:4116
// @BasePath /api/v1/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	f := framework.InitBaseFrameworkFromArgs(framework.ApiService,
		loadKnowledge,
		defaultPortForLocal,
	)

	f.PrepareClientClient(map[framework.ServiceNameEnum]framework.ClientHandler{
		framework.ClusterService: func(service micro.Service) error {
			client.ClusterClient = clusterpb.NewClusterService(string(framework.ClusterService), service.Client())
			return nil
		},
	})

	f.PrepareService(func(service micro.Service) error {
		return initGinEngine(f)
	})

	//f.StartService()
}

func initGinEngine(d *framework.BaseFramework) error {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()

	monitor := metrics.NewPrometheusMonitor(common.TiEM, d.GetServiceMeta().ServiceName.ServerName())
	g.Use(monitor.PromMiddleware())

	route.Route(g)

	port := d.GetServiceMeta().ServicePort

	addr := fmt.Sprintf(":%d", port)
	// openapi-server service registry
	serviceRegistry(d)

	if err := g.Run(addr); err != nil {
		d.GetRootLogger().ForkFile(common.LogFileSystem).Fatal(err)
	}

	return nil
}

// serviceRegistry registry openapi-server service
func serviceRegistry(f *framework.BaseFramework) {
	etcdClient := etcd_clientv2.InitEtcdClient(f.GetServiceMeta().RegistryAddress)
	address := f.GetClientArgs().Host + f.GetServiceMeta().GetServiceAddress()
	key := common.RegistryMicroServicePrefix + f.GetServiceMeta().ServiceName.ServerName() + "/" + address
	// Register openapi-server every TTL-2 seconds, default TTL is 5s
	go func() {
		for {
			err := etcdClient.SetWithTtl(key, "{\"weight\":1, \"max_fails\":2, \"fail_timeout\":10}", 5)
			if err != nil {
				f.Log().Errorf("regitry openapi-server failed! error: %v", err)
			}
			time.Sleep(time.Second * 3)
		}
	}()
}
func loadKnowledge(f *framework.BaseFramework) error {
	knowledge.LoadKnowledge()
	return nil
}

func defaultPortForLocal(f *framework.BaseFramework) error {
	if f.GetServiceMeta().ServicePort <= 0 {
		f.GetServiceMeta().ServicePort = common.DefaultMicroApiPort
	}
	return nil
}
