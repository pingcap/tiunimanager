package client

import (
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
)

var DBClient dbpb.TiEMDBService

var ClusterClient clusterpb.ClusterService

