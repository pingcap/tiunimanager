package domain

import (
	"github.com/pingcap-inc/tiem/library/knowledge"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClusterAggregation_ExtractInstancesDTO(t *testing.T) {
	got := buildAggregation().ExtractInstancesDTO()
	assert.Equal(t, "127.0.0.1",got.ExtranetConnectAddresses[0])
}

func buildAggregation() *ClusterAggregation {
	aggregation := &ClusterAggregation{
		Cluster: &Cluster{
			Id: "111",
			TenantId: "222",
			ClusterType: *knowledge.ClusterTypeFromCode("TiDB"),
			ClusterVersion: *knowledge.ClusterVersionFromCode("v5.0.0"),
		},
		AvailableResources: &proto.AllocHostResponse{
			TidbHosts: []*proto.AllocHost{
				{Ip: "127.0.0.1", Disk: &proto.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &proto.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &proto.Disk{Path: "/"}},
			},
			TikvHosts: []*proto.AllocHost{
				{Ip: "127.0.0.1", Disk: &proto.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &proto.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &proto.Disk{Path: "/"}},
			},
			PdHosts: []*proto.AllocHost{
				{Ip: "127.0.0.1", Disk: &proto.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &proto.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &proto.Disk{Path: "/"}},
			},
		},
	}
	aggregation.CurrentTopologyConfigRecord = &TopologyConfigRecord{
		TenantId:    aggregation.Cluster.TenantId,
		ClusterId:   aggregation.Cluster.Id,
		ConfigModel: convertConfig(aggregation.AvailableResources, aggregation.Cluster),
	}

	return aggregation
}

func TestClusterAggregation_ExtractComponentDTOs(t *testing.T) {
	got := buildAggregation().ExtractComponentDTOs()
	assert.Equal(t, "127.0.0.1", got[0].Nodes[1].Instance.HostId)
}