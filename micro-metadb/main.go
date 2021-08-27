package main

import (
	xlog "log"
	"time"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func main() {
	go func() {
		cfg := embed.NewConfig()
		cfg.Dir = "default.etcd"
		e, err := embed.StartEtcd(cfg)
		if err != nil {
			xlog.Fatal(err)
		}
		defer e.Close()
		select {
		case <-e.Server.ReadyNotify():
			xlog.Printf("Server is ready!")
		case <-time.After(60 * time.Second):
			e.Server.Stop() // trigger a shutdown
			xlog.Printf("Server took too long to start!")
		}
		xlog.Fatal(<-e.Err())
	}()
	initConfig()
	initLogger()
	initSqliteDB()
	initService()
}
