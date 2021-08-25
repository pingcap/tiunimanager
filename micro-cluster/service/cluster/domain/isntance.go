package domain

import (
	"github.com/pingcap-inc/tiem/library/knowledge"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"math/rand"
)

func (aggregation *ClusterAggregation) ExtractInstancesDTO() *proto.ClusterInstanceDTO {
	dto := &proto.ClusterInstanceDTO {
		Whitelist:           	[]string{},
		DiskUsage:                MockUsage(),
		CpuUsage:                 MockUsage(),
		MemoryUsage:              MockUsage(),
		StorageUsage:             MockUsage(),
		BackupFileUsage:          MockUsage(),
	}

	if record := aggregation.CurrentTiUPConfigRecord; aggregation.CurrentTiUPConfigRecord != nil && record.ConfigModel != nil {
		dto.Port = int32(record.ConfigModel.GlobalOptions.SSHPort)
		dto.IntranetConnectAddresses = ConnectAddresses(record.ConfigModel)
		dto.ExtranetConnectAddresses = ConnectAddresses(record.ConfigModel)
	} else {
		dto.Port = 22
		dto.IntranetConnectAddresses = []string{"127.0.0.1"}
		dto.ExtranetConnectAddresses = []string{"127.0.0.1"}
	}

	return dto
}

func ConnectAddresses(spec *spec.Specification) []string {
	servers := spec.TiDBServers

	list := make([]string, len(servers), len(servers))
	for i,v := range servers {
		list[i] = v.Host
	}
	return list
}

func (aggregation *ClusterAggregation) ExtractComponentDTOs() []*proto.ComponentInstanceDTO {
	if record := aggregation.CurrentTiUPConfigRecord; aggregation.CurrentTiUPConfigRecord != nil && record.ConfigModel != nil {
		config := record.ConfigModel
		var knowledge *knowledge.ClusterVersionSpec

		return appendAllComponentInstances(config, knowledge)
	}
	return make([]*proto.ComponentInstanceDTO, 0)
}

func appendAllComponentInstances(config *spec.Specification, knowledge *knowledge.ClusterVersionSpec) []*proto.ComponentInstanceDTO{
	components := make([]*proto.ComponentInstanceDTO, len(knowledge.ComponentSpecs), len(knowledge.ComponentSpecs))

	for _, v := range knowledge.ComponentSpecs {
		code := v.ClusterComponent.ComponentType
		componentDTO := &proto.ComponentInstanceDTO {
			BaseInfo: &proto.ComponentBaseInfoDTO{
				ComponentType: code,
				ComponentName: v.ClusterComponent.ComponentName,
			},
			Nodes: ComponentAppender[code](config),
		}

		components = append(components, componentDTO)
	}
	return components
}

var ComponentAppender = map[string]func (*spec.Specification) []*proto.ComponentNodeDisplayInfoDTO {
	"tidb": tiDBComponent,
	"tikv": tiKVComponent,
	"pd": pDComponent,
}

func tiDBComponent(config *spec.Specification) []*proto.ComponentNodeDisplayInfoDTO {
	servers := config.TiDBServers
	dto := make([]*proto.ComponentNodeDisplayInfoDTO, len(servers), len(servers))
	for i, v := range servers {
		dto[i] = &proto.ComponentNodeDisplayInfoDTO{
			NodeId: v.Host,
			Version: "version", // todo
			Status: "运行中", // todo
			Instance: &proto.ComponentNodeInstanceDTO{
				HostId: v.Host,
				Port: int32(v.Port),
				Role: mockRole(),
				Spec: mockSpec(),
				Zone: mockZone(),
			},

			Usages: &proto.ComponentNodeUsageDTO{
				IoUtil:       mockIoUtil(),
				Iops:         mockIops(),
				CpuUsage:     MockUsage(),
				MemoryUsage:  MockUsage(),
				StoregeUsage: MockUsage(),
			},
		}
	}
	return dto
}

func tiKVComponent(config *spec.Specification) []*proto.ComponentNodeDisplayInfoDTO {
	servers := config.TiKVServers
	dto := make([]*proto.ComponentNodeDisplayInfoDTO, len(servers), len(servers))
	return dto
}

func pDComponent(config *spec.Specification) []*proto.ComponentNodeDisplayInfoDTO {
	servers := config.PDServers
	dto := make([]*proto.ComponentNodeDisplayInfoDTO, len(servers), len(servers))
	return dto
}

// MockUsage TODO will be replaced with monitor implement
func MockUsage() *proto.UsageDTO {
	usage := &proto.UsageDTO{
		Total:     100,
		Used: float32(rand.Intn(100)),
	}
	usage.UsageRate = usage.Used / usage.Total
	return usage
}

func mockRole() *proto.ComponentNodeRoleDTO{
	return &proto.ComponentNodeRoleDTO{
		RoleCode: "Leader",
		RoleName: "Flower",
	}
}

func mockSpec() *proto.SpecBaseInfoDTO {
	return &proto.SpecBaseInfoDTO{
		SpecCode: knowledge.GenSpecCode(4, 8),
		SpecName: knowledge.GenSpecCode(4, 8),
	}
}

func mockZone() *proto.ZoneBaseInfoDTO {
	return &proto.ZoneBaseInfoDTO{
		ZoneCode: "AZ1",
		ZoneName: "AZ1",
	}
}

func mockIops() []float32 {
	return []float32{10,20}
}

func mockIoUtil() float32 {
	return 1
}
