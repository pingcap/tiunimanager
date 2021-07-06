package domain

import (
	"time"
)

type Cluster struct {
	Id         			string
	Code 				string
	TenantId   			string

	ClusterName 		string
	DbPassword 			string
	ClusterType 		ClusterType
	ClusterVersion 		ClusterVersion
	Tls 				bool

	Status     			ClusterStatus

	Demand 				[]ClusterComponentDemand

	WorkFlowId 			uint

	CreateTime 			time.Time
	UpdateTime 			time.Time
	DeleteTime 			time.Time
}

type ClusterType struct {
	Code string
	Name string
}

type ClusterVersion struct {
	Code string
	Name string
}

type ClusterComponent struct {
	ComponentType 			string
	ComponentName			string
}

type ClusterComponentDemand struct {
	ComponentType 			string
	TotalNodeCount  		int
	DistributionItems  		[]ClusterNodeDistributionItem
}

type ClusterDemandRecord struct {
	Id 					uint
	TenantId 			string
	ClusterId 			string
	Content 			[]ClusterComponentDemand
	CreateTime  		time.Time
}

type ClusterNodeDistributionItem struct {
	ZoneCode 		string
	SpecCode		string
	Count  			int
}
