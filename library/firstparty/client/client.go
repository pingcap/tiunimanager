package client

import (
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
)

var DBClient db.TiEMDBService

var ClusterClient cluster.ClusterService

