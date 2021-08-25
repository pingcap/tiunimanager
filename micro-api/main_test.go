package main

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/firstparty/framework"
	"github.com/pingcap-inc/tiem/micro-api/route"
)

var g *gin.Engine

func TestMain(m *testing.M) {
	f := framework.NewUtFramework(framework.ApiService,
		InitGin)

	err := f.StartService()
	if err == nil {
		m.Run()
	} else {
		f.GetDefaultLogger().Error(err)
	}
}

func InitGin(d *framework.UtFramework) error {
	gin.SetMode(gin.TestMode)
	g = gin.New()

	route.Route(g)
	/*
		port := config.GetClientArgs().RestPort
		if port <= 0 {
			port = config.DefaultRestPort
		}
		addr := fmt.Sprintf(":%d", port)
		if err := g.Run(addr); err != nil {
			d.GetDefaultLogger().Fatal(err)
		}
	*/
	return nil
}
