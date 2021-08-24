package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/pingcap/tiem/docs"
	"github.com/pingcap/tiem/library/firstparty/client"
	"github.com/pingcap/tiem/library/firstparty/config"
	"github.com/pingcap/tiem/library/firstparty/framework"
	"github.com/pingcap/tiem/micro-api/route"
	cluster "github.com/pingcap/tiem/micro-cluster/proto"
)

// @title TiEM UI API
// @version 1.0
// @description TiEM UI API

// @contact.name zhangpeijin
// @contact.email zhangpeijin@pingcap.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1/
func main() {
	framework.NewDefaultFramework(framework.ClusterService,
		initClient,
		initGinEngine,
	)

	//f.StartService()
}

func initGinEngine(d *framework.DefaultServiceFramework) error {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()

	route.Route(g)

	port := config.GetClientArgs().RestPort
	if port <= 0 {
		port = config.DefaultRestPort
	}
	addr := fmt.Sprintf(":%d", port)
	if err := g.Run(addr); err != nil {
		d.GetDefaultLogger().Fatal(err)
	}

	return nil
}

func initClient(d *framework.DefaultServiceFramework) error {
	srv := framework.ClusterService.BuildMicroService(d.GetRegistryAddress()...)
	client.ClusterClient = cluster.NewClusterService(framework.ClusterService.ToString(), srv.Client())
	return nil
}