package main

import (
	"github.com/pingcap/ticp/micro-cluster/service/clusteroperate/libtiup"
)

func init() {
	libtiup.TiupMgrInit()
}

func main() {
	libtiup.TiupMgrRoutine()
}
