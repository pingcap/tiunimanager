package backuprestore

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/management"
)

type RecoverClusterRsp struct {
	ClusterId string `json:"clusterId"`
	management.ClusterBaseInfo
	controller.StatusInfo
}
