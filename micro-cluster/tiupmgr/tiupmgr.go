package main

import (
	"github.com/pingcap/ticp/library/secondparty/libtiup"
)

func init() {
	libtiup.TiupMgrInit()
}

func main() {
	libtiup.TiupMgrRoutine()
}
