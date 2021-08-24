package client

import (
	cluster "github.com/pingcap/tiem/micro-cluster/proto"
	db "github.com/pingcap/tiem/micro-metadb/proto"
)

var DBClient db.TiEMDBService

var ClusterClient cluster.ClusterService

