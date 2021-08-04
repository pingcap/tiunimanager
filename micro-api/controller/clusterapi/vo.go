package clusterapi

import (
	"github.com/pingcap/tiem/micro-api/controller"
	"github.com/pingcap/tiem/micro-api/controller/hostapi"
)

type ClusterBaseInfo struct {
	ClusterName 		string
	DbPassword 			string
	ClusterType 		string
	ClusterVersion 		string
	Tags 				[]string
	Tls 				bool
}

type ClusterInstanceInfo struct {
	IntranetConnectAddresses	[]string
	ExtranetConnectAddresses	[]string
	Whitelist					[]string
	Port 						int
	DiskUsage           		controller.Usage
	CpuUsage    				controller.Usage
	MemoryUsage 				controller.Usage
	StorageUsage				controller.Usage
	BackupFileUsage				controller.Usage
}

type ClusterMaintenanceInfo struct {
	MaintainTaskCron  			string
}

type ClusterDisplayInfo struct {
	ClusterId 			string
	ClusterBaseInfo
	controller.StatusInfo
	ClusterInstanceInfo
}

type ClusterNodeDemand struct {
	ComponentType 			string
	TotalNodeCount  	int
	DistributionItems  	[]DistributionItem
}

type DistributionItem struct {
	ZoneCode 		string
	SpecCode		string
	Count  			int
}

type ComponentInstance struct {
	ComponentBaseInfo
	Nodes []ComponentNodeDisplayInfo
}

type ComponentNodeDisplayInfo struct {
	NodeId 				string
	Version 			string
	Status  			string
	ComponentNodeInstanceInfo
	ComponentNodeUsageInfo
}

type ComponentNodeInstanceInfo struct {
	HostId 				string
	Port 				int
	Role 				ComponentNodeRole
	Spec 				hostapi.SpecBaseInfo
	Zone 				hostapi.ZoneBaseInfo
}

type ComponentNodeUsageInfo struct {
	IoUtil 				float32
	Iops 				[]float32
	CpuUsage 			controller.Usage
	MemoryUsage 		controller.Usage
	StorageUsage 		controller.Usage
}

type ComponentNodeRole struct {
	RoleCode string
	RoleName string
}

type ComponentBaseInfo struct {
	ComponentType 			string
	ComponentName			string
}







