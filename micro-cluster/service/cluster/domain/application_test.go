/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package domain

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/secondparty"
	mock "github.com/pingcap-inc/tiem/test/mockdb"
	"github.com/pingcap-inc/tiem/test/mocksecondparty"

	"github.com/pingcap/tiup/pkg/cluster/spec"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/stretchr/testify/assert"
)

func defaultOperator() *Operator {
	return &Operator{
		Id:       "testoperator",
		Name:     "testoperator",
		TenantId: "testtenant",
	}
}

func defaultFlow() *FlowWorkEntity {
	return &FlowWorkEntity{
		Id:          1,
		FlowName:    "testflow",
		Status:      TaskStatusInit,
		StatusAlias: "创建中",
		BizId:       "testbizid",
	}
}

func defaultCluster() *ClusterAggregation {
	br := &BackupRecord{
		Id:           222,
		ClusterId:    "111",
		BackupType:   BackupTypeFull,
		BackupMode:   BackupModeAuto,
		BackupMethod: BackupMethodPhysics,
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix(),
	}
	return &ClusterAggregation{
		Cluster: &Cluster{
			Id:          "testCluster",
			ClusterName: "testCluster",
			Status:      ClusterStatusOnline,
		},
		LastBackupRecord: br,
		LastRecoverRecord: &RecoverRecord{
			Id:           333,
			ClusterId:    "111",
			BackupRecord: *br,
		},

		CurrentWorkFlow: defaultFlow(),
		CurrentOperator: defaultOperator(),
		MaintainCronTask: &CronTaskEntity{
			ID:   222,
			Cron: "testcron",
		},
		CurrentTopologyConfigRecord: &TopologyConfigRecord{
			Id:        0,
			TenantId:  "1",
			ClusterId: "testCluster",
			ConfigModel: &spec.Specification{
				GlobalOptions:    spec.GlobalOptions{},
				MonitoredOptions: spec.MonitoredOptions{},
				ServerConfigs:    spec.ServerConfigs{},
				TiDBServers: []*spec.TiDBSpec{
					{
						Host:      "127.0.0.1",
						DeployDir: "/mnt/sda/testCluster/tidb-deploy",
						LogDir:    "testCluster/tidb-log",
					},
				},
				PDServers: []*spec.PDSpec{
					{
						Host:      "127.0.0.1",
						DeployDir: "/mnt/sda/testCluster/pd-deploy",
						LogDir:    "testCluster/tidb-log",
					},
				},
				TiKVServers: []*spec.TiKVSpec{
					{
						Host:      "127.0.0.2",
						DeployDir: "/mnt/sda/testCluster/tikv-deploy",
						LogDir:    "testCluster/tidb-log",
					},
				},
				TiFlashServers: []*spec.TiFlashSpec{
					{
						Host:      "127.0.0.2",
						DeployDir: "/mnt/sda/testCluster/tiflash-deploy",
						LogDir:    "testCluster/tidb-log",
					},
				},
				CDCServers: []*spec.CDCSpec{
					{
						Host:      "127.0.0.2",
						DeployDir: "/mnt/sda/testCluster/ticdc-deploy",
						LogDir:    "testCluster/tidb-log",
					},
				},
			},
			CreateTime: time.Time{},
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
				ComponentType: "TiDB",
				ComponentName: "TiDB",
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
				ComponentType: "TiKV",
				ComponentName: "TiKV",
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
				ComponentType: "PD",
				ComponentName: "PD",
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
	config := convertConfig(&clusterpb.AllocHostResponse{
		PdHosts: []*clusterpb.AllocHost{
			{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/1/a"}},
			{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/1/b"}},
			{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/1/c"}},
		},
		TidbHosts: []*clusterpb.AllocHost{
			{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/2/a"}},
			{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/2/b"}},
			{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/2/c"}},
		},
		TikvHosts: []*clusterpb.AllocHost{
			{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/3/a"}},
			{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/3/b"}},
			{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/3/c"}},
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
	r := parseDistributionItemFromDTO(&clusterpb.DistributionItemDTO{
		SpecCode: "5C9G",
		ZoneCode: "zone1",
		Count:    999,
	})
	assert.Equal(t, 999, r.Count)
	assert.Equal(t, "5C9G", r.SpecCode)
	assert.Equal(t, "zone1", r.ZoneCode)

}

func Test_parseNodeDemandFromDTO(t *testing.T) {
	r := parseNodeDemandFromDTO(&clusterpb.ClusterNodeDemandDTO{
		ComponentType:  "TiDB",
		TotalNodeCount: 3,
		Items: []*clusterpb.DistributionItemDTO{
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
	gotOperator := parseOperatorFromDTO(&clusterpb.OperatorDTO{
		Id:       "111",
		Name:     "ope",
		TenantId: "222",
	})
	assert.Equal(t, "111", gotOperator.Id)
	assert.Equal(t, "ope", gotOperator.Name)
	assert.Equal(t, "222", gotOperator.TenantId)

}

func Test_parseRecoverInFoFromDTO(t *testing.T) {
	gotInfo := parseRecoverInFoFromDTO(&clusterpb.RecoverInfoDTO{
		SourceClusterId: "111",
		BackupRecordId:  222,
	})
	assert.Equal(t, "111", gotInfo.SourceClusterId)
	assert.Equal(t, int64(222), gotInfo.BackupRecordId)

}

func TestCreateCluster(t *testing.T) {
	got, err := CreateCluster(context.TODO(), &clusterpb.OperatorDTO{
		Id:       "testoperator",
		Name:     "testoperator",
		TenantId: "testoperator",
	},
		&clusterpb.ClusterBaseInfoDTO{
			ClusterName: "testCluster",
			ClusterType: &clusterpb.ClusterTypeDTO{
				Code: "TiDB",
				Name: "TiDB",
			},
			ClusterVersion: &clusterpb.ClusterVersionDTO{
				Code: "v5.0.0",
				Name: "v5.0.0",
			},
		},
		&clusterpb.ClusterCommonDemandDTO{
			Exclusive: false,
		},
		[]*clusterpb.ClusterNodeDemandDTO{
			{ComponentType: "TiDB", TotalNodeCount: 3, Items: []*clusterpb.DistributionItemDTO{
				{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
				{SpecCode: "4C8G", ZoneCode: "zone2", Count: 1},
				{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
			}},
			{ComponentType: "TiKV", TotalNodeCount: 3, Items: []*clusterpb.DistributionItemDTO{
				{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
				{SpecCode: "4C8G", ZoneCode: "zone2", Count: 2},
				{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
			}},
			{ComponentType: "PD", TotalNodeCount: 3, Items: []*clusterpb.DistributionItemDTO{
				{SpecCode: "4C8G", ZoneCode: "zone1", Count: 3},
			}},
		})

	assert.NoError(t, err)
	assert.Equal(t, "testCluster", got.Cluster.ClusterName)
}

func TestScaleOutCluster(t *testing.T) {
	got, err := ScaleOutCluster(context.TODO(),
		&clusterpb.OperatorDTO{
			Id:       "testoperator",
			Name:     "testoperator",
			TenantId: "testoperator",
		}, "testCluster",
		[]*clusterpb.ClusterNodeDemandDTO{
			{ComponentType: "TiDB", TotalNodeCount: 3, Items: []*clusterpb.DistributionItemDTO{
				{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
				{SpecCode: "4C8G", ZoneCode: "zone2", Count: 1},
				{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
			}},
			{ComponentType: "TiKV", TotalNodeCount: 3, Items: []*clusterpb.DistributionItemDTO{
				{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
				{SpecCode: "4C8G", ZoneCode: "zone2", Count: 2},
				{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
			}},
			{ComponentType: "PD", TotalNodeCount: 3, Items: []*clusterpb.DistributionItemDTO{
				{SpecCode: "4C8G", ZoneCode: "zone1", Count: 3},
			}},
		})

	assert.NoError(t, err)
	assert.Equal(t, "TiDB", got.AddedComponentDemand[0].ComponentType.ComponentType)
	assert.Equal(t, "TiKV", got.AddedComponentDemand[1].ComponentType.ComponentType)
	assert.Equal(t, "PD", got.AddedComponentDemand[2].ComponentType.ComponentType)
}

func TestScaleInCluster(t *testing.T) {
	got, err := ScaleInCluster(context.TODO(),
		&clusterpb.OperatorDTO{
			Id:       "testoperator",
			Name:     "testoperator",
			TenantId: "testoperator",
		},
		"testCluster",
		"127.0.0.1:4000")

	assert.Error(t, err)
	assert.Equal(t, "testoperator", got.CurrentOperator.TenantId)
}

func TestTakeoverClusters(t *testing.T) {
	got, err := TakeoverClusters(context.TODO(), &clusterpb.OperatorDTO{
		Id:       "testoperator",
		Name:     "testoperator",
		TenantId: "testoperator",
	}, &clusterpb.ClusterTakeoverReqDTO{
		ClusterNames: []string{"takeovercluster"},
	},
	)

	assert.NoError(t, err)
	assert.Equal(t, "takeovercluster", got[0].Cluster.ClusterName)
}

func TestDeleteCluster(t *testing.T) {
	got, err := DeleteCluster(context.TODO(), &clusterpb.OperatorDTO{
		Id:       "testoperator",
		Name:     "testoperator",
		TenantId: "testoperator",
	}, "testCluster")

	assert.NoError(t, err)
	assert.Equal(t, "testCluster", got.Cluster.ClusterName)

}

func TestGetTopology(t *testing.T) {
	got, err := GetTopology(context.TODO(), "testCluster")
	assert.NoError(t, err)
	assert.Equal(t, "testCluster", got.Cluster.ClusterName)
}

func TestRestartCluster(t *testing.T) {
	got, err := RestartCluster(context.TODO(), &clusterpb.OperatorDTO{
		Id:       "testoperator",
		Name:     "testoperator",
		TenantId: "testoperator",
	}, "testCluster")

	assert.NoError(t, err)
	assert.Equal(t, "testCluster", got.Cluster.ClusterName)

}

func TestStopCluster(t *testing.T) {
	got, err := StopCluster(context.TODO(), &clusterpb.OperatorDTO{
		Id:       "testoperator",
		Name:     "testoperator",
		TenantId: "testoperator",
	}, "testCluster")

	assert.NoError(t, err)
	assert.Equal(t, "testCluster", got.Cluster.ClusterName)

}

func TestBuildClusterLogConfig(t *testing.T) {
	err := BuildClusterLogConfig(context.TODO(), "testCluster")
	assert.NoError(t, err)
}

func TestMergeDemands(t *testing.T) {
	demand1 := []*ClusterComponentDemand{
		{
			ComponentType:  &knowledge.ClusterComponent{ComponentType: "TiDB", ComponentPurpose: "compute", ComponentName: "TiDB"},
			TotalNodeCount: 3, DistributionItems: []*ClusterNodeDistributionItem{
				{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
				{SpecCode: "4C8G", ZoneCode: "zone2", Count: 1},
				{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
			}},
		{ComponentType: &knowledge.ClusterComponent{ComponentType: "TiKV", ComponentPurpose: "storage", ComponentName: "TiKV"},
			TotalNodeCount: 4, DistributionItems: []*ClusterNodeDistributionItem{
				{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
				{SpecCode: "4C8G", ZoneCode: "zone2", Count: 2},
				{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
			}},
		{ComponentType: &knowledge.ClusterComponent{ComponentType: "PD", ComponentPurpose: "dispatch", ComponentName: "PD"},
			TotalNodeCount: 3, DistributionItems: []*ClusterNodeDistributionItem{
				{SpecCode: "4C8G", ZoneCode: "zone1", Count: 3},
			}}}
	demand2 := []*ClusterComponentDemand{
		{
			ComponentType:  &knowledge.ClusterComponent{ComponentType: "TiDB", ComponentPurpose: "compute", ComponentName: "TiDB"},
			TotalNodeCount: 3, DistributionItems: []*ClusterNodeDistributionItem{
			{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
			{SpecCode: "4C8G", ZoneCode: "zone2", Count: 1},
			{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
		}},
		{ComponentType: &knowledge.ClusterComponent{ComponentType: "TiKV", ComponentPurpose: "storage", ComponentName: "TiKV"},
			TotalNodeCount: 1, DistributionItems: []*ClusterNodeDistributionItem{
			{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
		}},
		{ComponentType: &knowledge.ClusterComponent{ComponentType: "PD", ComponentPurpose: "dispatch", ComponentName: "PD"},
			TotalNodeCount: 1, DistributionItems: []*ClusterNodeDistributionItem{
			{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
		}}}
	got := mergeDemands(demand1, demand2)
	assert.Equal(t, 3, len(got))
}

func TestDeleteDemands(t *testing.T) {
	demand := []*ClusterComponentDemand{
		{
			ComponentType:  &knowledge.ClusterComponent{ComponentType: "TiDB", ComponentPurpose: "compute", ComponentName: "TiDB"},
			TotalNodeCount: 3, DistributionItems: []*ClusterNodeDistributionItem{
			{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
			{SpecCode: "4C8G", ZoneCode: "zone2", Count: 1},
			{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
		}},
		{ComponentType: &knowledge.ClusterComponent{ComponentType: "TiKV", ComponentPurpose: "storage", ComponentName: "TiKV"},
			TotalNodeCount: 4, DistributionItems: []*ClusterNodeDistributionItem{
			{SpecCode: "4C8G", ZoneCode: "zone1", Count: 1},
			{SpecCode: "4C8G", ZoneCode: "zone2", Count: 2},
			{SpecCode: "4C8G", ZoneCode: "zone3", Count: 1},
		}},
		{ComponentType: &knowledge.ClusterComponent{ComponentType: "PD", ComponentPurpose: "dispatch", ComponentName: "PD"},
			TotalNodeCount: 3, DistributionItems: []*ClusterNodeDistributionItem{
			{SpecCode: "4C8G", ZoneCode: "zone1", Count: 3},
		}}}
	instance := &ComponentInstance{
		ComponentType: &knowledge.ClusterComponent{ComponentType: "PD", ComponentPurpose: "dispatch", ComponentName: "PD"},
		Location: &resource.Location{Region: "region", Zone: "zone1"},
		Compute: &resource.ComputeRequirement{
			CpuCores: 4,
			Memory: 8,
		},
	}
	got := deleteDemands(demand, instance)
	assert.Equal(t, 2, got[2].TotalNodeCount)
	assert.Equal(t, 2, got[2].DistributionItems[0].Count)
}

func TestModifyParameters(t *testing.T) {
	got, err := ModifyParameters(context.TODO(), &clusterpb.OperatorDTO{
		Id:       "testoperator",
		Name:     "testoperator",
		TenantId: "testoperator",
	}, "testCluster", &ModifyParam{NeedReboot: false, Params: []*ApplyParam{
		{
			ParamId:       1,
			Name:          "test_param_1",
			ComponentType: "TiDB",
			HasReboot:     1,
			Source:        0,
			RealValue:     clusterpb.ParamRealValueDTO{Cluster: "1"},
		},
		{
			ParamId:       2,
			Name:          "test_param_2",
			ComponentType: "TiKV",
			HasReboot:     1,
			Source:        0,
			RealValue:     clusterpb.ParamRealValueDTO{Cluster: "2"},
		},
	}})
	assert.NoError(t, err)
	assert.Equal(t, "testCluster", got.Cluster.ClusterName)
}

func Test_destroyCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	secondparty.SecondParty = mockTiup

	t.Run("success", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupDestroy(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
		})
		ret := destroyCluster(task, flowCtx)

		assert.Equal(t, true, ret)
	})
}

func Test_destroyTasks(t *testing.T) {
	assert.True(t, destroyTasks(&TaskEntity{}, nil))
}

func Test_modifyParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	secondparty.SecondParty = mockTiup

	t.Run("success", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupEditGlobalConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)

		task := &TaskEntity{
			Id: 123,
		}
		modifyCtx := NewFlowContext(context.TODO())
		modifyCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
		})
		modifyCtx.SetData(contextModifyParamsKey, &ModifyParam{
			NeedReboot: false,
			Params: []*ApplyParam{
				{
					ParamId:       1,
					Name:          "test_param_1",
					ComponentType: "TiDB",
					HasReboot:     1,
					Source:        0,
					RealValue:     clusterpb.ParamRealValueDTO{Cluster: "1"},
				},
				{
					ParamId:       2,
					Name:          "test_param_2",
					ComponentType: "TiKV",
					HasReboot:     1,
					Source:        0,
					RealValue:     clusterpb.ParamRealValueDTO{Cluster: "2"},
				},
			},
		})
		ret := modifyParameters(task, modifyCtx)
		assert.Equal(t, true, ret)
	})
}

func Test_refreshParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	secondparty.SecondParty = mockTiup

	t.Run("success", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupReload(gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)
		mockTiup.EXPECT().MicroSrvGetTaskStatusByBizID(gomock.Any(), gomock.Any()).Return(dbpb.TiupTaskStatus_Finished, "", nil)

		task := &TaskEntity{
			Id: 123,
		}
		refreshCtx := NewFlowContext(context.TODO())
		refreshCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
		})
		refreshCtx.SetData(contextModifyParamsKey, &ModifyParam{
			NeedReboot: true,
			Params: []*ApplyParam{
				{
					ParamId:       1,
					Name:          "test_param_1",
					ComponentType: "TiDB",
					HasReboot:     1,
					Source:        0,
					RealValue:     clusterpb.ParamRealValueDTO{Cluster: "1"},
				},
			},
		})
		ret := refreshParameter(task, refreshCtx)
		assert.Equal(t, true, ret)
	})
}

func Test_clusterRestart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	secondparty.SecondParty = mockTiup

	t.Run("success", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupRestart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
			CurrentTopologyConfigRecord: &TopologyConfigRecord{
				ConfigModel: &spec.Specification{
					TiDBServers: []*spec.TiDBSpec{
						{
							Host: "127.0.0.1",
							Port: 4000,
						},
					},
				},
			},
		})
		ret := clusterRestart(task, flowCtx)

		assert.Equal(t, true, ret)
	})

	t.Run("fail", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupRestart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), errors.New("wrong"))

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
			CurrentTopologyConfigRecord: &TopologyConfigRecord{
				ConfigModel: &spec.Specification{
					TiDBServers: []*spec.TiDBSpec{
						{
							Host: "127.0.0.1",
							Port: 4000,
						},
					},
				},
			},
		})
		ret := clusterRestart(task, flowCtx)

		assert.Equal(t, false, ret)
	})

}

func Test_clusterStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	secondparty.SecondParty = mockTiup

	t.Run("success", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupStop(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
			CurrentTopologyConfigRecord: &TopologyConfigRecord{
				ConfigModel: &spec.Specification{
					TiDBServers: []*spec.TiDBSpec{
						{
							Host: "127.0.0.1",
							Port: 4000,
						},
					},
				},
			},
		})
		ret := clusterStop(task, flowCtx)

		assert.Equal(t, true, ret)
	})
	t.Run("fail", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupStop(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), errors.New("wrong"))

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
			CurrentTopologyConfigRecord: &TopologyConfigRecord{
				ConfigModel: &spec.Specification{
					TiDBServers: []*spec.TiDBSpec{
						{
							Host: "127.0.0.1",
							Port: 4000,
						},
					},
				},
			},
		})
		ret := clusterStop(task, flowCtx)

		assert.Equal(t, false, ret)
	})

}

func Test_deleteCluster(t *testing.T) {
	task := &TaskEntity{
		Id: 123,
	}
	flowCtx := NewFlowContext(context.TODO())
	flowCtx.SetData(contextClusterKey, &ClusterAggregation{
		LastBackupRecord: &BackupRecord{
			Id:          123,
			StorageType: StorageTypeS3,
		},
		Cluster: &Cluster{
			Id:          "test-tidb123",
			ClusterName: "test-tidb",
		},
		CurrentTopologyConfigRecord: &TopologyConfigRecord{
			ConfigModel: &spec.Specification{
				TiDBServers: []*spec.TiDBSpec{
					{
						Host: "127.0.0.1",
						Port: 4000,
					},
				},
			},
		},
	})
	ret := deleteCluster(task, flowCtx)

	assert.Equal(t, true, ret)
}

func Test_setClusterOnline(t *testing.T) {
	task := &TaskEntity{
		Id: 123,
	}
	flowCtx := NewFlowContext(context.TODO())
	flowCtx.SetData(contextClusterKey, &ClusterAggregation{
		LastBackupRecord: &BackupRecord{
			Id:          123,
			StorageType: StorageTypeS3,
		},
		Cluster: &Cluster{
			Id:          "test-tidb123",
			ClusterName: "test-tidb",
		},
		CurrentTopologyConfigRecord: &TopologyConfigRecord{
			ConfigModel: &spec.Specification{
				TiDBServers: []*spec.TiDBSpec{
					{
						Host: "127.0.0.1",
						Port: 4000,
					},
				},
			},
		},
	})
	ret := setClusterOnline(task, flowCtx)

	assert.Equal(t, true, ret)
	assert.Equal(t, ClusterStatusOnline, flowCtx.GetData(contextClusterKey).(*ClusterAggregation).Cluster.Status)

}

func Test_setClusterOffline(t *testing.T) {
	task := &TaskEntity{
		Id: 123,
	}
	flowCtx := NewFlowContext(context.TODO())
	flowCtx.SetData(contextClusterKey, &ClusterAggregation{
		LastBackupRecord: &BackupRecord{
			Id:          123,
			StorageType: StorageTypeS3,
		},
		Cluster: &Cluster{
			Id:          "test-tidb123",
			ClusterName: "test-tidb",
		},
		CurrentTopologyConfigRecord: &TopologyConfigRecord{
			ConfigModel: &spec.Specification{
				TiDBServers: []*spec.TiDBSpec{
					{
						Host: "127.0.0.1",
						Port: 4000,
					},
				},
			},
		},
	})
	ret := setClusterOffline(task, flowCtx)

	assert.Equal(t, true, ret)
	assert.Equal(t, ClusterStatusOffline, flowCtx.GetData(contextClusterKey).(*ClusterAggregation).Cluster.Status)

}

func Test_startupCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	secondparty.SecondParty = mockTiup

	t.Run("success", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupStart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
			CurrentTopologyConfigRecord: &TopologyConfigRecord{
				ConfigModel: &spec.Specification{
					TiDBServers: []*spec.TiDBSpec{
						{
							Host: "127.0.0.1",
							Port: 4000,
						},
					},
				},
			},
		})
		ret := startupCluster(task, flowCtx)

		assert.Equal(t, true, ret)
	})
	t.Run("fail", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupStart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), errors.New("wrong"))

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
			CurrentTopologyConfigRecord: &TopologyConfigRecord{
				ConfigModel: &spec.Specification{
					TiDBServers: []*spec.TiDBSpec{
						{
							Host: "127.0.0.1",
							Port: 4000,
						},
					},
				},
			},
		})
		ret := startupCluster(task, flowCtx)

		assert.Equal(t, false, ret)
	})

}

func Test_deployCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	secondparty.SecondParty = mockTiup

	t.Run("success", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupDeploy(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
		})
		flowCtx.SetData(contextTopologyKey, "topology")
		ret := deployCluster(task, flowCtx)

		assert.Equal(t, true, ret)
	})
	t.Run("fail", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupDeploy(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), errors.New("wrong"))

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			LastBackupRecord: &BackupRecord{
				Id:          123,
				StorageType: StorageTypeS3,
			},
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
		})
		flowCtx.SetData(contextTopologyKey, "topology")
		ret := deployCluster(task, flowCtx)

		assert.Equal(t, false, ret)
	})

}

func Test_scaleOutCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	secondparty.SecondParty = mockTiup

	t.Run("success", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupScaleOut(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
		})
		flowCtx.SetData(contextTopologyKey, "topology")
		ret := scaleOutCluster(task, flowCtx)

		assert.Equal(t, true, ret)
	})
	t.Run("fail", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupScaleOut(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), errors.New("wrong"))

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			Cluster: &Cluster{
				Id:          "test-tidb123",
				ClusterName: "test-tidb",
			},
		})
		flowCtx.SetData(contextTopologyKey, "topology")
		ret := scaleOutCluster(task, flowCtx)

		assert.Equal(t, false, ret)
	})

}

func Test_scaleInCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	secondparty.SecondParty = mockTiup

	t.Run("success", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupScaleIn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), nil)

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			Cluster: &Cluster{
				Id:             "test-tidb123",
				ClusterName:    "test-tidb",
				ClusterVersion: knowledge.ClusterVersion{Code: "v5.0.0", Name: "v5.0.0"},
				ClusterType:    knowledge.ClusterType{Code: "TiDB", Name: "TiDB"},
			},
			CurrentComponentInstances: []*ComponentInstance{
				{
					Host:          "127.0.0.1",
					PortList:      []int{4000},
					ComponentType: &knowledge.ClusterComponent{ComponentType: "TiDB"},
				},
				{
					Host:          "127.0.0.1",
					PortList:      []int{4001},
					ComponentType: &knowledge.ClusterComponent{ComponentType: "TiDB"},
				},
			},
		})
		flowCtx.SetData(contextDeleteNodeKey, "127.0.0.1:4000")
		ret := scaleInCluster(task, flowCtx)

		assert.Equal(t, true, ret)
	})
	t.Run("fail", func(t *testing.T) {
		mockTiup.EXPECT().MicroSrvTiupScaleIn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(123), errors.New("wrong"))

		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(context.TODO())
		flowCtx.SetData(contextClusterKey, &ClusterAggregation{
			Cluster: &Cluster{
				Id:             "test-tidb123",
				ClusterName:    "test-tidb",
				ClusterVersion: knowledge.ClusterVersion{Code: "v5.0.0", Name: "v5.0.0"},
				ClusterType:    knowledge.ClusterType{Code: "TiDB", Name: "TiDB"},
			},
			CurrentComponentInstances: []*ComponentInstance{
				{
					Host:          "127.0.0.1",
					PortList:      []int{4000},
					ComponentType: &knowledge.ClusterComponent{ComponentType: "TiDB"},
				},
				{
					Host:          "127.0.0.1",
					PortList:      []int{4001},
					ComponentType: &knowledge.ClusterComponent{ComponentType: "TiDB"},
				},
			},
		})
		flowCtx.SetData(contextDeleteNodeKey, "127.0.0.1:4000")
		ret := scaleInCluster(task, flowCtx)

		assert.Equal(t, false, ret)
	})

}

func Test_freedResource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).Return(&dbpb.DBRecycleResponse{
		Rs: &dbpb.DBAllocResponseStatus{Code: 0, Message: ""},
	}, nil)
	client.DBClient = mockClient

	ctx := NewFlowContext(context.TODO())
	ctx.SetData(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
	})
	result := freedResource(&TaskEntity{}, ctx)

	assert.True(t, result)
}

func Test_freedNodeResource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().RecycleResources(gomock.Any(), gomock.Any()).Return(&dbpb.DBRecycleResponse{
		Rs: &dbpb.DBAllocResponseStatus{Code: 0, Message: ""},
	}, nil)
	client.DBClient = mockClient

	ctx := NewFlowContext(context.TODO())
	ctx.SetData(contextClusterKey, &ClusterAggregation{
		Cluster: &Cluster{
			Id: "test-abc",
		},
		CurrentComponentInstances: []*ComponentInstance{
			{
				Host:     "127.0.0.1",
				PortList: []int{4000},
				Compute:  &resource.ComputeRequirement{CpuCores: 4, Memory: 8},
			},
		},
	})
	ctx.SetData(contextDeleteNodeKey, "127.0.0.1:4000")
	result := freeNodeResource(&TaskEntity{}, ctx)
	got := ctx.GetData(contextClusterKey).(*ClusterAggregation)

	assert.Equal(t, got.CurrentComponentInstances[0].Status, ClusterStatusDeleted)
	assert.True(t, result)
}

func Test_prepareResourceSucceed(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		task1 := &TaskEntity{}
		prepareResourceSucceed(task1, &clusterpb.BatchAllocResponse{
			Rs: &clusterpb.AllocResponseStatus{
				Code: 0,
			},
			BatchResults: []*clusterpb.AllocResponse{
				{
					Rs: &clusterpb.AllocResponseStatus{
						Code: 0,
					},
					Results: []*clusterpb.HostResource{
						{
							HostIp: "127.0.0.1",
						},
						{
							HostIp: "127.0.0.2",
						},
					},
				},
				{
					Rs: &clusterpb.AllocResponseStatus{
						Code: 0,
					},
					Results: []*clusterpb.HostResource{
						{
							HostIp: "127.0.0.3",
						},
					},
				},
			},
		})

		assert.Equal(t, fmt.Sprintln("", fmt.Sprintf("alloc succeed with hosts: [127.0.0.1 127.0.0.2 127.0.0.3]")), task1.Result)

	})

	t.Run("skip", func(t *testing.T) {
		task2 := &TaskEntity{}
		prepareResourceSucceed(task2, &clusterpb.BatchAllocResponse{
			Rs: &clusterpb.AllocResponseStatus{
				Code: 0,
			},
			BatchResults: []*clusterpb.AllocResponse{
				{
					Rs: &clusterpb.AllocResponseStatus{
						Code: -1,
					},
					Results: []*clusterpb.HostResource{
						{
							HostIp: "127.0.0.1",
						},
						{
							HostIp: "127.0.0.2",
						},
					},
				},
				{
					Rs: &clusterpb.AllocResponseStatus{
						Code: 0,
					},
					Results: []*clusterpb.HostResource{
						{
							HostIp: "127.0.0.3",
						},
					},
				},
			},
		})
		assert.Equal(t, fmt.Sprintln("", fmt.Sprintf("alloc succeed with hosts: [127.0.0.3]")), task2.Result)

	})

}
