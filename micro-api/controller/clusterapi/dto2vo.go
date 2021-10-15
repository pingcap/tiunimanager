package clusterapi

import (
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/hostapi"
	"time"
)

func ParseClusterBaseInfoFromDTO(dto *clusterpb.ClusterBaseInfoDTO) (baseInfo *ClusterBaseInfo) {
	baseInfo = &ClusterBaseInfo{
		ClusterName:    dto.ClusterName,
		DbPassword:     dto.DbPassword,
		ClusterType:    dto.ClusterType.Name,
		ClusterVersion: dto.ClusterVersion.Name,
		Tags:           dto.Tags,
		Tls:            dto.Tls,
	}

	return
}

func ParseStatusFromDTO(dto *clusterpb.DisplayStatusDTO) (statusInfo *controller.StatusInfo) {
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

func ParseDisplayInfoFromDTO(dto *clusterpb.ClusterDisplayDTO) (displayInfo *ClusterDisplayInfo) {
	displayInfo = &ClusterDisplayInfo{
		ClusterId:           dto.ClusterId,
		ClusterBaseInfo:     *ParseClusterBaseInfoFromDTO(dto.BaseInfo),
		StatusInfo:          *ParseStatusFromDTO(dto.Status),
		ClusterInstanceInfo: *ParseInstanceInfoFromDTO(dto.Instances),
	}
	return displayInfo
}

func ParseMaintenanceInfoFromDTO(dto *clusterpb.ClusterMaintenanceDTO) (maintenance *ClusterMaintenanceInfo) {
	maintenance = &ClusterMaintenanceInfo{
		MaintainTaskCron: dto.MaintainTaskCron,
	}
	return
}

func ParseComponentInfoFromDTO(dto *clusterpb.ComponentInstanceDTO) (instance *ComponentInstance) {
	nodes := make([]ComponentNodeDisplayInfo, len(dto.Nodes), len(dto.Nodes))

	for i, v := range dto.Nodes {
		nodes[i] = *ParseComponentNodeFromDTO(v)
	}

	instance = &ComponentInstance{
		ComponentBaseInfo: *ParseComponentBaseInfoFromDTO(dto.GetBaseInfo()),
		Nodes:             nodes,
	}

	return
}

func ParseComponentNodeFromDTO(dto *clusterpb.ComponentNodeDisplayInfoDTO) (node *ComponentNodeDisplayInfo) {
	if dto == nil || dto.Instance == nil {
		return &ComponentNodeDisplayInfo{}
	}
	node = &ComponentNodeDisplayInfo{
		NodeId:                    dto.NodeId,
		Version:                   dto.Version,
		Status:                    dto.Status,
		ComponentNodeInstanceInfo: *ParseComponentNodeInstanceFromDTO(dto.Instance),
		ComponentNodeUsageInfo:    *ParseComponentNodeUsageFromDTO(dto.Usages),
	}
	return
}

func ParseComponentNodeUsageFromDTO(dto *clusterpb.ComponentNodeUsageDTO) (usages *ComponentNodeUsageInfo) {
	usages = &ComponentNodeUsageInfo{
		IoUtil:       dto.IoUtil,
		Iops:         dto.Iops,
		CpuUsage:     *controller.ParseUsageFromDTO(dto.CpuUsage),
		MemoryUsage:  *controller.ParseUsageFromDTO(dto.MemoryUsage),
		StorageUsage: *controller.ParseUsageFromDTO(dto.StoregeUsage),
	}

	return
}

func ParseComponentNodeInstanceFromDTO(dto *clusterpb.ComponentNodeInstanceDTO) (instance *ComponentNodeInstanceInfo) {
	instance = &ComponentNodeInstanceInfo{
		HostId: dto.HostId,
		Port:   int(dto.Port),
		Role:   ComponentNodeRole{dto.Role.RoleCode, dto.Role.RoleName},
		Spec:   hostapi.SpecBaseInfo{SpecCode: dto.Spec.SpecCode, SpecName: dto.Spec.SpecName},
		Zone:   hostapi.ZoneBaseInfo{ZoneCode: dto.Zone.ZoneCode, ZoneName: dto.Zone.ZoneName},
	}

	return
}

func ParseComponentBaseInfoFromDTO(dto *clusterpb.ComponentBaseInfoDTO) (baseInfo *ComponentBaseInfo) {
	baseInfo = &ComponentBaseInfo{
		ComponentType: dto.ComponentType,
		ComponentName: dto.ComponentName,
	}
	return
}

func ParseInstanceInfoFromDTO(dto *clusterpb.ClusterInstanceDTO) (instance *ClusterInstanceInfo) {
	portList := make([]int, len(dto.PortList), len(dto.PortList))
	for i, v := range dto.PortList {
		portList[i] = int(v)
	}
	instance = &ClusterInstanceInfo{
		IntranetConnectAddresses: dto.IntranetConnectAddresses,
		ExtranetConnectAddresses: dto.ExtranetConnectAddresses,
		Whitelist:                dto.Whitelist,
		PortList:                 portList,
		DiskUsage:                *controller.ParseUsageFromDTO(dto.DiskUsage),
		CpuUsage:                 *controller.ParseUsageFromDTO(dto.CpuUsage),
		MemoryUsage:              *controller.ParseUsageFromDTO(dto.MemoryUsage),
		StorageUsage:             *controller.ParseUsageFromDTO(dto.StorageUsage),
		BackupFileUsage:          *controller.ParseUsageFromDTO(dto.BackupFileUsage),
	}

	return
}
