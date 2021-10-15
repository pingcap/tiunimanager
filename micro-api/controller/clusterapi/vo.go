package clusterapi

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/hostapi"
)

type ClusterBaseInfo struct {
	ClusterName 		string   	`json:"clusterName"`
	DbPassword 			string   	`json:"dbPassword"`
	ClusterType 		string   	`json:"clusterType"`
	ClusterVersion 		string		`json:"clusterVersion"`
	Tags 				[]string		`json:"tags"`
	Tls 				bool		`json:"tls"`
}

type ClusterInstanceInfo struct {
	IntranetConnectAddresses	[]string	`json:"intranetConnectAddresses"`
	ExtranetConnectAddresses	[]string	`json:"extranetConnectAddresses"`
	Whitelist					[]string	`json:"whitelist"`
	PortList 					[]int			`json:"portList"`
	DiskUsage           		controller.Usage	`json:"diskUsage"`
	CpuUsage    				controller.Usage	`json:"cpuUsage"`
	MemoryUsage 				controller.Usage	`json:"memoryUsage"`
	StorageUsage				controller.Usage	`json:"storageUsage"`
	BackupFileUsage				controller.Usage	`json:"backupFileUsage"`
}

type ClusterMaintenanceInfo struct {
	MaintainTaskCron  			string	`json:"maintainTaskCron"`
}

type ClusterDisplayInfo struct {
	ClusterId 			string	`json:"clusterId"`
	ClusterBaseInfo
	controller.StatusInfo
	ClusterInstanceInfo
}

type ClusterNodeDemand struct {
	ComponentType 		string	`json:"componentType"`
	TotalNodeCount  	int		`json:"totalNodeCount"`
	DistributionItems  	[]DistributionItem	`json:"distributionItems"`
}

type DistributionItem struct {
	ZoneCode 		string	`json:"zoneCode"`
	SpecCode		string	`json:"specCode"`
	Count  			int		`json:"count"`
}

type ComponentInstance struct {
	ComponentBaseInfo
	Nodes []ComponentNodeDisplayInfo	`json:"nodes"`
}

type ComponentNodeDisplayInfo struct {
	NodeId 				string	`json:"nodeId"`
	Version 			string	`json:"version"`
	Status  			string	`json:"status"`
	ComponentNodeInstanceInfo
	ComponentNodeUsageInfo
}

type ComponentNodeInstanceInfo struct {
	HostId 				string	`json:"hostId"`
	Port 				int		`json:"port"`
	Role 				ComponentNodeRole		`json:"role"`
	Spec 				hostapi.SpecBaseInfo	`json:"spec"`
	Zone 				hostapi.ZoneBaseInfo	`json:"zone"`
}

type ComponentNodeUsageInfo struct {
	IoUtil 				float32			`json:"ioUtil"`
	Iops 				[]float32		`json:"iops"`
	CpuUsage 			controller.Usage	`json:"cpuUsage"`
	MemoryUsage 		controller.Usage	`json:"memoryUsage"`
	StorageUsage 		controller.Usage	`json:"storageUsage"`
}

type ComponentNodeRole struct {
	RoleCode string	`json:"roleCode"`
	RoleName string	`json:"roleName"`
}

type ComponentBaseInfo struct {
	ComponentType 			string	`json:"componentType"`
	ComponentName			string	`json:"componentName"`
}







