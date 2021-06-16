package temp

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3"
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
