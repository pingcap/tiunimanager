package domain

import (
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"time"
)

type TiUPConfigRecord struct {
	Id 					uint
	TenantId 			string
	ClusterId 			string
	ConfigModel 		*spec.Specification
	CreateTime 			time.Time
}

type ClusterInstanceProxy struct {
	ClusterId         	string
	TenantId   			string
	CurrentTiUPConfig	*TiUPConfigRecord
}
