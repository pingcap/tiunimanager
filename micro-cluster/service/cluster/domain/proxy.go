package domain

import (
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"time"
)

type TiUPConfigRecord struct {
	Id 					uint
	TenantId 			string
	ClusterId 			string
	FlowWorkId 			string
	ConfigModel 		spec.Specification
	CreateTime 			time.Time
}

type ClusterInstanceProxy struct {
	ClusterId         	string
	TenantId   			string
}

func (proxy ClusterInstanceProxy) Topology(clusterId string) Topology {
	return Topology{}
}

type Topology struct {

}