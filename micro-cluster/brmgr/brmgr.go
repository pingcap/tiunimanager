package main

import "github.com/pingcap/ticp/micro-cluster/service/clusteroperate/libbr"

func init() {
	libbr.BrMgrInit()
}

func main() {
	libbr.BrMgrRoutine()
}
