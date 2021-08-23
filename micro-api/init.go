package main

import (
	"fmt"
	"log"

	"github.com/asim/go-micro/v3"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/library/firstparty/config"
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/library/thirdparty/tracer"
	"github.com/pingcap/tiem/micro-api/route"
	"github.com/pingcap/tiem/micro-cluster/client"
)

var TiEMApiServiceName = "go.micro.tiem.api"

func initConfig() {
	srv := micro.NewService(
		config.GetMicroApiCliArgsOption(),
	)
	srv.Init()
	srv = nil

	config.InitForMonolith(config.MicroApiMod)
}

func initTracer() {
	tracer.InitTracer()
}

func initClient() {
	client.InitClusterClient()
	client.InitManagerClient()
}

func initGinEngine() {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()

	route.Route(g)

	port := config.GetClientArgs().RestPort
	if port <= 0 {
		port = config.DefaultRestPort
	}
	addr := fmt.Sprintf(":%d", port)
	if err := g.Run(addr); err != nil {
		log.Fatal(err)
	}

	//if err := g.RunTLS(addr, config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath()); err != nil {
	//	log.Fatal(err)
	//}
}

func initKnowledge() {
	knowledge.LoadKnowledge()
}
