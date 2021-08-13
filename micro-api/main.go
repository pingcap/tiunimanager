package main

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	_ "github.com/pingcap/tiem/docs"
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
func main()  {
	initConfig()
	//initService()
	initClient()
	//initPrometheus()
	initKnowledge()
	initGinEngine()

}