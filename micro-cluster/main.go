package main

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/pingcap/tiem/library/firstparty/config"
)

func main() {
	initConfig()
	initLogger(config.KEY_CLUSTER_LOG)
	initPort()
	initClient()
	initClusterOperator()
	initKnowledge()
	initService()
	//initPrometheus()

}
