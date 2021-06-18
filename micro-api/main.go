package main

import (
	_ "github.com/pingcap/ticp/docs"
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
	initConfig()
	initService()
	initClient()
	initPrometheus()
	initGinEngine()
}