package main

import (
	"fmt"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
)

func main() {
	initPort()
	initConfig()
	initLogger()
	initClient()

	if err := initHostManager(); err != nil {
		fmt.Println("Init Host Manager Failed, err:", err)
		return
	}

	initService()
	//initPrometheus()
}
