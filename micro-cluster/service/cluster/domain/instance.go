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
		dto.IntranetConnectAddresses, dto.ExtranetConnectAddresses, dto.PortList = ConnectAddresses(record.ConfigModel)
	} else {
		dto.PortList = []int64{4000}
		dto.IntranetConnectAddresses = []string{"127.0.0.1"}
		dto.ExtranetConnectAddresses = []string{"127.0.0.1"}
	}

	return dto
}

func ConnectAddresses(spec *spec.Specification) ([]string, []string, []int64) {
	servers := spec.TiDBServers

	addressList := make([]string, len(servers), len(servers))
	portList := make([]int64, len(servers), len(servers))

	for i,v := range servers {
		addressList[i] = v.Host
		portList[i] = int64(v.Port)
	}
	return addressList, addressList, portList
}

func (aggregation *ClusterAggregation) ExtractComponentDTOs() []*proto.ComponentInstanceDTO {
	if record := aggregation.CurrentTiUPConfigRecord; aggregation.CurrentTiUPConfigRecord != nil && record.ConfigModel != nil {
		config := record.ConfigModel
		knowledge := knowledge.ClusterTypeSpecFromCode(aggregation.Cluster.ClusterType.Code)
		for _, v := range knowledge.VersionSpecs {
			if v.ClusterVersion.Code == aggregation.Cluster.ClusterVersion.Code {
				return appendAllComponentInstances(config, &v)
			}
		}
	}
	return make([]*proto.ComponentInstanceDTO, 0)
}

func appendAllComponentInstances(config *spec.Specification, knowledge *knowledge.ClusterVersionSpec) []*proto.ComponentInstanceDTO{
	components := make([]*proto.ComponentInstanceDTO, 0, len(knowledge.ComponentSpecs))

	for _, v := range knowledge.ComponentSpecs {
		code := v.ClusterComponent.ComponentType
		componentDTO := &proto.ComponentInstanceDTO {
			BaseInfo: &proto.ComponentBaseInfoDTO{
				ComponentType: code,
				ComponentName: v.ClusterComponent.ComponentName,
			},
			Nodes: ComponentAppender[code](config, knowledge.ClusterVersion.Code),
		}

		components = append(components, componentDTO)
	}
	return components
}

var ComponentAppender = map[string]func (*spec.Specification, string) []*proto.ComponentNodeDisplayInfoDTO {
	"TiDB": tiDBComponent,
	"TiKV": tiKVComponent,
	"PD": pDComponent,
	//"TiFlash": tiFlashComponent,
	//"TiCDC": tiCDCComponent,
}

func tiDBComponent(config *spec.Specification, version string) []*proto.ComponentNodeDisplayInfoDTO {
	servers := config.TiDBServers
	dto := make([]*proto.ComponentNodeDisplayInfoDTO, len(servers), len(servers))
	for i, v := range servers {
		dto[i] = &proto.ComponentNodeDisplayInfoDTO{
			NodeId: v.Host,
			Version: version, // todo
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

func tiKVComponent(config *spec.Specification, version string) []*proto.ComponentNodeDisplayInfoDTO {
	servers := config.TiKVServers
	dto := make([]*proto.ComponentNodeDisplayInfoDTO, len(servers), len(servers))
	for i, v := range servers {
		dto[i] = &proto.ComponentNodeDisplayInfoDTO{
			NodeId: v.Host,
			Version: version, // todo
			Status: "运行中", // todo
			Instance: &proto.ComponentNodeInstanceDTO{
				HostId: v.Host,
				Port: 20160,
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

func pDComponent(config *spec.Specification, version string) []*proto.ComponentNodeDisplayInfoDTO {
	servers := config.PDServers
	dto := make([]*proto.ComponentNodeDisplayInfoDTO, len(servers), len(servers))
	for i, v := range servers {
		dto[i] = &proto.ComponentNodeDisplayInfoDTO{
			NodeId: v.Host,
			Version: version, // todo
			Status: "运行中", // todo
			Instance: &proto.ComponentNodeInstanceDTO{
				HostId: v.Host,
				Port: 2379,
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

func tiCDCComponent(config *spec.Specification, version string) []*proto.ComponentNodeDisplayInfoDTO {
	dto := make([]*proto.ComponentNodeDisplayInfoDTO, 0, 0)

	return dto
}
func tiFlashComponent(config *spec.Specification, version string) []*proto.ComponentNodeDisplayInfoDTO {
	dto := make([]*proto.ComponentNodeDisplayInfoDTO, 0, 0)
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
		ZoneCode: "TEST_Zone1",
		ZoneName: "TEST_Zone1",
	}
}

func mockIops() []float32 {
	return []float32{10,20}
}

func mockIoUtil() float32 {
	return 1
}
