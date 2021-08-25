package main

import (
	"github.com/pingcap-inc/tiem/library/secondparty/libtiup"
)

func init() {
	libtiup.TiupMgrInit()
}

func main() {
	libtiup.TiupMgrRoutine()
}
