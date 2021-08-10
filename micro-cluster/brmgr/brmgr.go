package main

import "github.com/pingcap/tiem/library/secondparty/libbr"

func init() {
	libbr.BrMgrInit()
}

func main() {
	libbr.BrMgrRoutine()
}
