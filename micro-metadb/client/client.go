package client

import (
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	db "github.com/pingcap/tiem/micro-metadb/proto"
)

var DBClient db.TiEMDBService
