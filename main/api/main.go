package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/addon/tracer"
	"github.com/pingcap/ticp/api/controller"
	hostapi2 "github.com/pingcap/ticp/api/controller/hostapi"
	instanceapi2 "github.com/pingcap/ticp/api/controller/instanceapi"
	userapi2 "github.com/pingcap/ticp/api/controller/userapi"
	_ "github.com/pingcap/ticp/docs"

	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	"log"
)

// @title TiCP UI API
// @version 1.0
// @description TiCP UI API

// @contact.name zhangpeijin
// @contact.email zhangpeijin@pingcap.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/
func main()  {
	{
		gin.SetMode(gin.ReleaseMode)
		g := gin.New()
		g.Use(logger.GenGinLogger(), gin.Recovery())
		g.Use(tracer.GinOpenTracing())
		g.GET("/api/hello", controller.Hello)
		g.GET("/api/user/login", userapi2.Login)
		g.GET("/api/user/logout", userapi2.Logout)

		g.GET("/api/instance/create", instanceapi2.Create)
		g.GET("/api/instance/query", instanceapi2.Query)

		g.GET("/api/host/query", hostapi2.Query)

		g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		addr := fmt.Sprintf(":%d", 8080)
		if err := g.Run(addr); err != nil {
			log.Fatal(err)
		}
	}
}