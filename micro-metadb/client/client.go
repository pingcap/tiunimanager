package client

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3"
	db "github.com/pingcap/ticp/micro-db/proto"
	"github.com/pingcap/ticp/micro-db/service"
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

var DBClient db.TiCPDBService

func init() {
	appendToInitFpArray(initDBClient)
}

func initDBClient(srv micro.Service) {
	DBClient = db.NewTiCPDBService(service.TiCPDbServiceName, srv.Client())
}
