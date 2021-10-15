package domain

import (
	"github.com/pingcap-inc/tiem/library/knowledge"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func defaultOperator() *Operator{
	return  &Operator{
		Id: "testoperator",
		Name: "testoperator",
		TenantId: "testtenant",
	}
}

func defaultFlow() *FlowWorkEntity{
	return &FlowWorkEntity{
		Id: 1,
		FlowName: "testflow",
		Status: TaskStatusInit,
		StatusAlias: "创建中",
		BizId: "testbizid",
	}
}

func defaultCluster() *ClusterAggregation {
	br := &BackupRecord{
		Id: 222,
		ClusterId: "111",
		BackupType: BackupTypeLogic,
		BackupMode: BackupModeAuto,
		Range: BackupRangeFull,
		StartTime: time.Now().Unix(),
		EndTime: time.Now().Unix(),
	}
	return &ClusterAggregation{
		Cluster: &Cluster{
			Id: "testCluster",
			ClusterName: "testCluster",
			Status: ClusterStatusOnline,
		},
		LastBackupRecord: br,
		LastRecoverRecord: &RecoverRecord{
			Id: 333,
			ClusterId: "111",
			BackupRecord: *br,
		},

		CurrentWorkFlow: defaultFlow(),
		CurrentOperator: defaultOperator(),
		MaintainCronTask: &CronTaskEntity{
			ID: 222,
			Cron: "testcron",
		},
	}
}

func TestClusterAggregation_ExtractBackupRecordDTO(t *testing.T) {
	aggregation := defaultCluster()
	dto := aggregation.ExtractBackupRecordDTO()
	assert.Equal(t, aggregation.LastBackupRecord.Id, dto.Id)
}

func TestClusterAggregation_ExtractBaseInfoDTO(t *testing.T) {
	aggregation := defaultCluster()
	dto := aggregation.ExtractBaseInfoDTO()
	assert.Equal(t, aggregation.Cluster.ClusterName, dto.ClusterName)
}

func TestClusterAggregation_ExtractDisplayDTO(t *testing.T) {
	aggregation := defaultCluster()
	dto := aggregation.ExtractDisplayDTO()
	assert.Equal(t, strconv.Itoa(int(aggregation.Cluster.Status)), dto.Status.StatusCode)
	assert.Equal(t, aggregation.Cluster.ClusterName, dto.BaseInfo.ClusterName)
	assert.Equal(t, int64(4000), dto.Instances.PortList[0])
}

func TestClusterAggregation_ExtractMaintenanceDTO(t *testing.T) {
	aggregation := defaultCluster()
	dto := aggregation.ExtractMaintenanceDTO()
	assert.Equal(t, aggregation.MaintainCronTask.Cron, dto.MaintainTaskCron)
}

func TestClusterAggregation_ExtractRecoverRecordDTO(t *testing.T) {
	aggregation := defaultCluster()
	dto := aggregation.ExtractRecoverRecordDTO()
	assert.Equal(t, int64(aggregation.LastRecoverRecord.Id), dto.Id)
	assert.Equal(t, aggregation.LastRecoverRecord.BackupRecord.Id, dto.BackupRecordId)
}

func Test_convertAllocHostsRequest(t *testing.T) {
	demands := []*ClusterComponentDemand{
		{
			ComponentType: &knowledge.ClusterComponent{
				"TiDB",
				"TiDB",
			},
			TotalNodeCount: 999,
			DistributionItems: []*ClusterNodeDistributionItem{
				{
					"zone1",
					"4C8G",
					999,
				},
			},
		},
		{
			ComponentType: &knowledge.ClusterComponent{
				"TiKV",
				"TiKV",
			},
			TotalNodeCount: 3,
			DistributionItems: []*ClusterNodeDistributionItem{
				{
					"zone1",
					"4C8G",
					2,
				},
			},
		},
		{
			ComponentType: &knowledge.ClusterComponent{
				"PD",
				"PD",
			},
			TotalNodeCount: 3,
			DistributionItems: []*ClusterNodeDistributionItem{
				{
					"zone1",
					"4C8G",
					2,
				},
			},
		},
	}
	req := convertAllocHostsRequest(demands)
	assert.Equal(t, req.TidbReq[0].Count, int32(999))
	assert.Equal(t, len(req.PdReq), 1)
	assert.Equal(t, req.TikvReq[0].FailureDomain, "zone1")

}

func Test_convertAllocationReq(t *testing.T) {
	item := &ClusterNodeDistributionItem{
		SpecCode: "111C8G",
	}
	req := convertAllocationReq(item)
	assert.Equal(t, int32(111), req.CpuCores)
}

func Test_convertConfig(t *testing.T) {
	config := convertConfig(&proto.AllocHostResponse{
		PdHosts: []*proto.AllocHost{
			{Ip: "127.0.0.1",Disk: &proto.Disk{Path: "/1/a"}},
			{Ip: "127.0.0.1",Disk: &proto.Disk{Path: "/1/b"}},
			{Ip: "127.0.0.1",Disk: &proto.Disk{Path: "/1/c"}},
		},
		TidbHosts: []*proto.AllocHost{
			{Ip: "127.0.0.1",Disk: &proto.Disk{Path: "/2/a"}},
			{Ip: "127.0.0.1",Disk: &proto.Disk{Path: "/2/b"}},
			{Ip: "127.0.0.1",Disk: &proto.Disk{Path: "/2/c"}},
		},
		TikvHosts: []*proto.AllocHost{
			{Ip: "127.0.0.1",Disk: &proto.Disk{Path: "/3/a"}},
			{Ip: "127.0.0.1",Disk: &proto.Disk{Path: "/3/b"}},
			{Ip: "127.0.0.1",Disk: &proto.Disk{Path: "/3/c"}},
		},
	},
	&Cluster{
		Id: "111",
	})
	assert.Equal(t, 3, len(config.TiKVServers))
	assert.Equal(t, "111/tidb-data", config.GlobalOptions.DataDir)
	assert.Equal(t, "127.0.0.1", config.PDServers[2].Host)

}

func Test_parseDistributionItemFromDTO(t *testing.T) {
	r := parseDistributionItemFromDTO(&proto.DistributionItemDTO{
		SpecCode: "5C9G",
		ZoneCode: "zone1",
		Count: 999,
	})
	assert.Equal(t, 999, r.Count)
	assert.Equal(t, "5C9G", r.SpecCode)
	assert.Equal(t, "zone1", r.ZoneCode)

}

func Test_parseNodeDemandFromDTO(t *testing.T) {
	r := parseNodeDemandFromDTO(&proto.ClusterNodeDemandDTO{
		ComponentType: "TiDB",
		TotalNodeCount: 3,
		Items: []*proto.DistributionItemDTO{
			{
				SpecCode: "5C9G",
				ZoneCode: "zone1",
				Count:    1,
			},
			{
				SpecCode: "1C2G",
				ZoneCode: "ZONE2",
				Count:    2,
			},
		},
	})
	assert.Equal(t, 3, r.TotalNodeCount)
	assert.Equal(t, 1, r.DistributionItems[0].Count)

}

func Test_parseOperatorFromDTO(t *testing.T) {
	gotOperator := parseOperatorFromDTO(&proto.OperatorDTO{
		Id: "111",
		Name: "ope",
		TenantId: "222",
	})
	assert.Equal(t, "111", gotOperator.Id)
	assert.Equal(t, "ope", gotOperator.Name)
	assert.Equal(t, "222", gotOperator.TenantId)

}

func Test_parseRecoverInFoFromDTO(t *testing.T) {
	gotInfo := parseRecoverInFoFromDTO(&proto.RecoverInfoDTO{
		SourceClusterId: "111",
		BackupRecordId: 222,
	})
	assert.Equal(t, "111", gotInfo.SourceClusterId)
	assert.Equal(t, int64(222), gotInfo.BackupRecordId)

}

func TestCreateCluster(t *testing.T) {
	got, err := CreateCluster(&proto.OperatorDTO {
		Id: "testoperator",
		Name: "testoperator",
		TenantId: "testoperator",
	},
	&proto.ClusterBaseInfoDTO{
		ClusterName: "testCluster",
		ClusterType: &proto.ClusterTypeDTO{
			Code: "TiDB",
			Name: "TiDB",
		},
		ClusterVersion: &proto.ClusterVersionDTO{
			Code: "v5.0.0",
			Name: "v5.0.0",
		},
	},
	[]*proto.ClusterNodeDemandDTO{
		{ComponentType: "TiDB", TotalNodeCount: 3, Items: []*proto.DistributionItemDTO{
			{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
			{SpecCode: "4C8G", ZoneCode: "zone2", Count: 1},
			{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
		}},
		{ComponentType: "TiKV", TotalNodeCount: 3, Items: []*proto.DistributionItemDTO{
			{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
			{SpecCode: "4C8G", ZoneCode: "zone2", Count: 2},
			{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
		}},
		{ComponentType: "PD", TotalNodeCount: 3, Items: []*proto.DistributionItemDTO{
			{SpecCode: "4C8G", ZoneCode: "zone1", Count: 3},
		}},
	})

	assert.NoError(t, err)
	assert.Equal(t, "testCluster", got.Cluster.ClusterName)
}

func TestDeleteCluster(t *testing.T) {
	got, err := DeleteCluster(&proto.OperatorDTO {
		Id: "testoperator",
		Name: "testoperator",
		TenantId: "testoperator",
	}, "testCluster")

	assert.NoError(t, err)
	assert.Equal(t, "testCluster", got.Cluster.ClusterName)

}

func TestModifyParameters(t *testing.T) {
	got, err := ModifyParameters(&proto.OperatorDTO {
		Id: "testoperator",
		Name: "testoperator",
		TenantId: "testoperator",
	}, "testCluster", "content")
	assert.NoError(t, err)
	assert.Equal(t, "testCluster", got.Cluster.ClusterName)
}

func Test_destroyCluster(t *testing.T) {
	assert.True(t, destroyCluster(&TaskEntity{}, nil))
}

func Test_destroyTasks(t *testing.T) {
	assert.True(t, destroyTasks(&TaskEntity{}, nil))
}

func Test_freedResource(t *testing.T) {
	assert.True(t, freedResource(&TaskEntity{}, nil))
}

func Test_modifyParameters(t *testing.T) {
	assert.True(t, modifyParameters(&TaskEntity{}, nil))
}
