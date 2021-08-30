package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/route"
	"testing"
)

var g *gin.Engine

func TestMain(m *testing.M) {
	f := framework.InitBaseFrameworkForUt(framework.ApiService)

	gin.SetMode(gin.ReleaseMode)
	g = gin.New()

	route.Route(g)
	//
	//port := f.GetServiceMeta().ServicePort
	//
	//addr := fmt.Sprintf(":%d", port)
	//
	//go func() {
	//	if err := g.Run(addr); err != nil {
	//		f.GetLogger().Fatal(err)
	//	}
	//}()

	m.Run()
	f.Shutdown()

}
