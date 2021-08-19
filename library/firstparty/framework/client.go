package framework

import (
	"github.com/pingcap/tiem/library/firstparty/util"
	cluster "github.com/pingcap/tiem/micro-cluster/proto"
	db "github.com/pingcap/tiem/micro-metadb/proto"
)

type Client interface {
	GetManagerClient() cluster.TiEMManagerService
	GetClusterClient() cluster.TiEMManagerService
	GetDBClient() db.TiEMDBService
}

func initClient() (Client, error) {
	util.AssertWithInfo(false, "Not Implemented Yet")
	return nil, nil
}
