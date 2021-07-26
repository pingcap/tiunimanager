package domain

import (
	"github.com/pingcap/ticp/knowledge"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"time"
)

type Cluster struct {
	Id         			string
	Code 				string
	TenantId   			string

	ClusterName    		string
	DbPassword     		string
	ClusterType    		knowledge.ClusterType
	ClusterVersion 		knowledge.ClusterVersion
	Tags           		[]string
	Tls            		bool

	Status 				ClusterStatus

	Demands 			[]*ClusterComponentDemand

	WorkFlowId 			uint

	OwnerId 			string

	CreateTime 			time.Time
	UpdateTime 			time.Time
	DeleteTime 			time.Time
}

func (c *Cluster) Delete()  {
	c.Status = ClusterStatusDeleted
}

type ClusterComponentDemand struct {
	ComponentType     *knowledge.ClusterComponent
	TotalNodeCount    int
	DistributionItems []*ClusterNodeDistributionItem
}

type ClusterDemandRecord struct {
	Id 					uint
	TenantId 			string
	ClusterId 			string
	Content 			[]*ClusterComponentDemand
	CreateTime  		time.Time
}

type ClusterNodeDistributionItem struct {
	ZoneCode 		string
	SpecCode		string
	Count  			int
}

type TiUPConfigRecord struct {
	Id 					uint
	TenantId 			string
	ClusterId 			string
	ConfigModel 		*spec.Specification
	CreateTime 			time.Time
}