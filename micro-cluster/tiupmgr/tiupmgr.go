package main

import (
	"github.com/pingcap/tiem/library/secondparty/libtiup"
)

func init() {
	libtiup.TiupMgrInit()
}

func main() {
	libtiup.TiupMgrRoutine()
}
