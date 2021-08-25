package main

import "github.com/pingcap-inc/tiem/library/secondparty/libbr"

func init() {
	libbr.BrMgrInit()
}

func main() {
	libbr.BrMgrRoutine()
}
