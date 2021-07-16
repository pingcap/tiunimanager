package domain

import (
	knowledge "github.com/pingcap/ticp/knowledge/models"
	"time"
)

type Cluster struct {
	Id         			string
	Code 				string
	TenantId   			string

	ClusterName 		string
	DbPassword 			string
	ClusterType 		knowledge.ClusterType
	ClusterVersion 		knowledge.ClusterVersion
	Tags 				[]string
	Tls 				bool

	Status 				ClusterStatus

	Demand 				[]*ClusterComponentDemand

	WorkFlowId 			uint

	CreateTime 			time.Time
	UpdateTime 			time.Time
	DeleteTime 			time.Time
}

type ClusterComponentDemand struct {
	ComponentType 			string
	TotalNodeCount  		int
	DistributionItems  		[]*ClusterNodeDistributionItem
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
