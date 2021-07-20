package clusterapi

import (
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-api/controller/hostapi"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
	"time"
)

func ParseClusterBaseInfoFromDTO(dto *cluster.ClusterBaseInfoDTO) (baseInfo *ClusterBaseInfo) {
	baseInfo = &ClusterBaseInfo{
		ClusterName: dto.ClusterName,
		DbPassword: dto.DbPassword,
		ClusterType: dto.ClusterType.Code,
		ClusterVersion: dto.ClusterVersion.Code,
		Tags: dto.Tags,
		Tls: dto.Tls,
	}

	return
}

func ParseStatusFromDTO(dto *cluster.DisplayStatusDTO) (statusInfo *controller.StatusInfo) {
	statusInfo = &controller.StatusInfo{
		StatusCode:      dto.StatusCode,
		StatusName:      dto.StatusName,
		InProcessFlowId: int(dto.InProcessFlowId),
		CreateTime:      time.Unix(dto.CreateTime, 0),
		UpdateTime:      time.Unix(dto.UpdateTime, 0),
		DeleteTime:      time.Unix(dto.DeleteTime, 0),
	}

	return
}

func ParseDisplayInfoFromDTO(dto *cluster.ClusterDisplayDTO) (displayInfo *ClusterDisplayInfo) {
	displayInfo = &ClusterDisplayInfo{
		ClusterId: dto.ClusterId,
		ClusterBaseInfo: *ParseClusterBaseInfoFromDTO(dto.BaseInfo),
		StatusInfo: *ParseStatusFromDTO(dto.Status),
		ClusterInstanceInfo: *ParseInstanceInfoFromDTO(dto.Instances),
	}
	return displayInfo
}

func ParseMaintenanceInfoFromDTO(dto *cluster.ClusterMaintenanceDTO) (maintenance *ClusterMaintenanceInfo) {
	maintenance = &ClusterMaintenanceInfo{
		MaintainTaskCron: dto.MaintainTaskCron,
	}
	return
}

func ParseComponentInfoFromDTO(dto *cluster.ComponentInstanceDTO) (instance *ComponentInstance) {
	nodes := make([]ComponentNodeDisplayInfo, len(dto.Nodes), len(dto.Nodes))

	for i,v := range dto.Nodes {
		nodes[i] = *ParseComponentNodeFromDTO(v)
	}

	instance = &ComponentInstance{
		ComponentBaseInfo: *ParseComponentBaseInfoFromDTO(dto.GetBaseInfo()),
		Nodes: nodes,
	}

	return
}

func ParseComponentNodeFromDTO(dto *cluster.ComponentNodeDisplayInfoDTO) (node *ComponentNodeDisplayInfo) {
	node = &ComponentNodeDisplayInfo{
		NodeId: dto.NodeId,
		Version: dto.Version,
		Status: dto.Status,
		ComponentNodeInstanceInfo: *ParseComponentNodeInstanceFromDTO(dto.Instance),
		ComponentNodeUsageInfo: *ParseComponentNodeUsageFromDTO(dto.Usages),
	}
	return
}

func ParseComponentNodeUsageFromDTO(dto *cluster.ComponentNodeUsageDTO) (usages *ComponentNodeUsageInfo) {
	usages = &ComponentNodeUsageInfo{
		IoUtil: dto.IoUtil,
		Iops: dto.Iops,
		CpuUsage: *controller.ParseUsageFromDTO(dto.CpuUsage),
		MemoryUsage: *controller.ParseUsageFromDTO(dto.MemoryUsage),
		StorageUsage: *controller.ParseUsageFromDTO(dto.StoregeUsage),
	}

	return
}

func ParseComponentNodeInstanceFromDTO(dto *cluster.ComponentNodeInstanceDTO) (instance *ComponentNodeInstanceInfo) {
	instance = &ComponentNodeInstanceInfo{
		HostId: dto.HostId,
		Port: int(dto.Port),
		Role:   ComponentNodeRole{dto.Role.RoleCode, dto.Role.RoleName},
		Spec:   hostapi.SpecBaseInfo{SpecCode: dto.Spec.SpecCode, SpecName: dto.Spec.SpecName},
		Zone:   hostapi.ZoneBaseInfo{ZoneCode: dto.Zone.ZoneCode, ZoneName: dto.Zone.ZoneName},
	}

	return
}

func ParseComponentBaseInfoFromDTO(dto *cluster.ComponentBaseInfoDTO) (baseInfo *ComponentBaseInfo) {
	baseInfo = &ComponentBaseInfo{
		ComponentType: dto.ComponentType,
		ComponentName: dto.ComponentName,
	}
	return
}

func ParseInstanceInfoFromDTO(dto *cluster.ClusterInstanceDTO) (instance *ClusterInstanceInfo) {
	instance = &ClusterInstanceInfo{
		IntranetConnectAddresses: dto.IntranetConnectAddresses,
		ExtranetConnectAddresses: dto.ExtranetConnectAddresses,
		Whitelist:                dto.Whitelist,
		Port: int(dto.Port),
		DiskUsage: *controller.ParseUsageFromDTO(dto.DiskUsage),
		CpuUsage: *controller.ParseUsageFromDTO(dto.CpuUsage),
		MemoryUsage: *controller.ParseUsageFromDTO(dto.MemoryUsage),
		StorageUsage: *controller.ParseUsageFromDTO(dto.StorageUsage),
		BackupFileUsage: *controller.ParseUsageFromDTO(dto.BackupFileUsage),

	}

	return
}

