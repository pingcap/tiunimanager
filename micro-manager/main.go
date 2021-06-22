package main

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
)

func main() {
	initPort()
	initConfig()
	initLogger()
	initService()
	//initPrometheus()
	initClient()
}
