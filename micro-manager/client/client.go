package client

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3"
	manager "github.com/pingcap/ticp/micro-manager/proto"
	"github.com/pingcap/ticp/micro-manager/service"
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

func init() {
	appendToInitFpArray(initManagerClient)
}

var ManagerClient manager.TiCPManagerService

func initManagerClient(srv micro.Service) {
	ManagerClient = manager.NewTiCPManagerService(service.TiCPManagerServiceName, srv.Client())
}
