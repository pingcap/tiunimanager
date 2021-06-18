package client

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
	"github.com/pingcap/ticp/micro-cluster/service"
)

var initFpArray []func(srv micro.Service)

func InitClient(srv micro.Service) {
	for _, fp := range initFpArray {
		fp(srv)
	}
}

func appendToInitFpArray(fp func(srv micro.Service)) {
	initFpArray = append(initFpArray, fp)
}

var ClusterClient cluster.TiCPClusterService

func init() {
	appendToInitFpArray(initClusterClient)
}

func initClusterClient(srv micro.Service) {
	ClusterClient = cluster.NewTiCPClusterService(service.TiCPClusterServiceName, srv.Client())
}
