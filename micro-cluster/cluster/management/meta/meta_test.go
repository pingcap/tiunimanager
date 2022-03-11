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
 ******************************************************************************/

package meta

import (
	"context"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models/platform/config"
	mock_product "github.com/pingcap-inc/tiem/test/mockmodels"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockconfig"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/stretchr/testify/assert"
)

func TestClusterMeta_BuildCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	params := structs.CreateClusterParameter{
		Name:       "aaa",
		DBPassword: "password",
		Type:       "type",
		Version:    "version",
		TLS:        false,
		Tags:       []string{"t1", "t2"},
	}
	meta := &ClusterMeta{}

	rw.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "1234"}}, nil)

	err := meta.BuildCluster(context.TODO(), params)
	assert.NoError(t, err)
	assert.Equal(t, meta.Cluster.Name, params.Name)
	assert.Equal(t, meta.Cluster.Type, params.Type)
	assert.Equal(t, meta.Cluster.Version, params.Version)
	assert.Equal(t, meta.Cluster.TLS, params.TLS)
	assert.Equal(t, meta.Cluster.Tags, params.Tags)
	assert.Equal(t, meta.DBUsers[string(constants.Root)].Password.Val, params.DBPassword)
	assert.NotEmpty(t, meta.DBUsers[string(constants.Root)].ClusterID)
	assert.Equal(t, meta.DBUsers[string(constants.Root)].ClusterID, "1234")
}

func TestClusterMeta_AddInstances(t *testing.T) {
	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			Version: "v4.1.1",
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		err := meta.AddInstances(context.TODO(), []structs.ClusterResourceParameterCompute{
			{"TiDB", 2, []structs.ClusterResourceParameterComputeResource{
				{Zone: "zone1", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
				{Zone: "zone2", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
			}},
			{"TiKV", 2, []structs.ClusterResourceParameterComputeResource{
				{Zone: "zone1", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
				{Zone: "zone2", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
			}},
		})
		assert.NoError(t, err)

		assert.Equal(t, 2, len(meta.Instances["TiKV"]))
		assert.Equal(t, 3, len(meta.Instances["TiDB"]))
		assert.Equal(t, "111", meta.Instances["TiKV"][0].ClusterID)
		assert.Equal(t, int32(3), meta.Instances["TiDB"][1].DiskCapacity)
		assert.Equal(t, "SSD", meta.Instances["TiKV"][1].DiskType)
		assert.Equal(t, "zone2", meta.Instances["TiDB"][2].Zone)
	})
	t.Run("empty", func(t *testing.T) {
		err := meta.AddInstances(context.TODO(), []structs.ClusterResourceParameterCompute{})
		assert.Error(t, err)
	})

}

func TestClusterMeta_AddDefaultInstances(t *testing.T) {
	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			Version: "v4.1.1",
			Type:    string(constants.EMProductIDTiDB),
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		err := meta.AddDefaultInstances(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 4, len(meta.Instances))
	})
}

func TestClusterMeta_GenerateInstanceResourceRequirements(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)
	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:        "2145635758",
				TenantId:  "324567",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:              "koojdafij",
			Type:              "TiDB",
			Version:           "v5.2.2",
			Tags:              []string{"111", "333"},
			OwnerId:           "436534636u",
			ParameterGroupID:  "352467890",
			Copies:            4,
			Region:            "Region1",
			Exclusive:         false,
			CpuArchitecture:   "x86_64",
			MaintenanceStatus: constants.ClusterMaintenanceCreating,
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "TiDB",
					Version:      "v5.0.0",
					Ports:        []int32{10001, 10002, 10003, 10004},
					HostIP:       []string{"127.0.0.1"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "TiKV",
					Version:      "v5.0.0",
					Ports:        []int32{20001, 20002, 20003, 20004},
					HostIP:       []string{"127.0.0.2"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
			"PD": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "PD",
					Version:      "v5.0.0",
					Ports:        []int32{30001, 30002, 30003, 30004},
					HostIP:       []string{"127.0.0.3"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		mockQueryTiDBFromDB(productRW.EXPECT())

		got1, got2, err := meta.GenerateInstanceResourceRequirements(context.TODO())
		assert.Equal(t, 3, len(got1))
		assert.Equal(t, 3, len(got2))
		assert.NoError(t, err)
	})
}

func TestClusterMeta_GenerateGlobalPortRequirements(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rw := mockconfig.NewMockReaderWriter(ctrl)
	models.SetConfigReaderWriter(rw)

	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:        "2145635758",
				TenantId:  "324567",
				Status:    string(constants.ClusterInitializing),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:              "koojdafij",
			Type:              "TiDB",
			Version:           "v5.0.0",
			Tags:              []string{"111", "333"},
			OwnerId:           "436534636u",
			ParameterGroupID:  "352467890",
			Copies:            4,
			Region:            "Region1",
			Exclusive:         false,
			CpuArchitecture:   "x86_64",
			MaintenanceStatus: constants.ClusterMaintenanceCreating,
		},
	}

	t.Run("normal", func(t *testing.T) {
		rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{
			ConfigValue: "[10,11]",
		}, nil).Times(1)
		got, err := meta.GenerateGlobalPortRequirements(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(got))
	})
	t.Run("error", func(t *testing.T) {
		rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{
			ConfigValue: "",
		}, nil).Times(1)
		_, err := meta.GenerateGlobalPortRequirements(context.TODO())
		assert.Error(t, err)
	})
}

func TestClusterMeta_ApplyGlobalPortResource(t *testing.T) {
	meta := &ClusterMeta{}
	t.Run("normal", func(t *testing.T) {
		meta.ApplyGlobalPortResource(8001, 8002)
		assert.Equal(t, 8001, int(meta.NodeExporterPort))
		assert.Equal(t, 8002, int(meta.BlackboxExporterPort))
	})
}

func TestClusterMeta_GetInstanceByStatus(t *testing.T) {
	meta := &ClusterMeta{
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
				},
			},
		},
	}

	got := meta.GetInstanceByStatus(context.TODO(), constants.ClusterInstanceRunning)
	assert.Equal(t, 2, len(got))
	got = meta.GetInstanceByStatus(context.TODO(), constants.ClusterInstanceInitializing)
	assert.Equal(t, 1, len(got))
}

func TestClusterMeta_GenerateTopologyConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rw := mockconfig.NewMockReaderWriter(ctrl)
	models.SetConfigReaderWriter(rw)
	rw.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT))
	t.Run("normal", func(t *testing.T) {
		meta := &ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID:     "111",
					Status: string(constants.ClusterRunning),
				},
				Version: "v4.1.1",
			},
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {
					{
						Entity: common.Entity{
							Status: string(constants.ClusterInstanceInitializing),
						},
						HostIP: []string{"127.0.0.1"},
						Ports: []int32{
							1, 2, 3, 4, 5, 6,
						},
					},
					{
						Entity: common.Entity{
							Status: string(constants.ClusterInstanceRunning),
						},
						HostIP: []string{"127.0.0.1"},
						Ports: []int32{
							1, 2, 3, 4, 5, 6,
						},
					},
				},
			},
		}
		config, err := meta.GenerateTopologyConfig(context.TODO())
		assert.NoError(t, err)
		assert.NotEmpty(t, config)
	})
	t.Run("error", func(t *testing.T) {
		empty := &ClusterMeta{
			Cluster: &management.Cluster{},
		}
		_, err := empty.GenerateTopologyConfig(context.TODO())
		assert.Error(t, err)

		empty = &ClusterMeta{
			Instances: map[string][]*management.ClusterInstance{
				"TiDB": {},
			},
		}
		_, err = empty.GenerateTopologyConfig(context.TODO())
		assert.Error(t, err)

	})
}

func TestClusterMeta_UpdateClusterStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)
	rw.EXPECT().UpdateStatus(gomock.Any(), "111", gomock.Any()).Return(errors.Error(errors.TIEM_CLUSTER_NOT_FOUND))
	rw.EXPECT().UpdateStatus(gomock.Any(), "222", gomock.Any()).Return(nil)
	rw.EXPECT().UpdateStatus(gomock.Any(), "", gomock.Any()).Return(errors.Error(errors.TIEM_CLUSTER_NOT_FOUND))

	meta := &ClusterMeta{
		Cluster: &management.Cluster{},
	}
	t.Run("normal", func(t *testing.T) {
		err := meta.UpdateClusterStatus(context.TODO(), constants.ClusterRunning)
		assert.Error(t, err)

		meta.Cluster.ID = "111"
		err = meta.UpdateClusterStatus(context.TODO(), constants.ClusterRunning)
		assert.Error(t, err)

		meta.Cluster.ID = "222"
		err = meta.UpdateClusterStatus(context.TODO(), constants.ClusterRunning)
		assert.NoError(t, err)
	})
}

func TestClusterMeta_GetInstance(t *testing.T) {
	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			Version: "v4.1.1",
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						ID:     "tidb1111",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{111},
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		instance, err := meta.GetInstance(context.TODO(), "tidb1111")
		assert.NoError(t, err)
		assert.NotEmpty(t, instance)
		assert.Equal(t, "tidb1111", instance.ID)
	})

	t.Run("error", func(t *testing.T) {
		_, err := meta.GetInstance(context.TODO(), "127.0.0.1:222")
		assert.Error(t, err)
		_, err = meta.GetInstance(context.TODO(), "127.0.0.2:111")
		assert.Error(t, err)
		_, err = meta.GetInstance(context.TODO(), "127.0.0.1:ss")
		assert.Error(t, err)
		_, err = meta.GetInstance(context.TODO(), "127.0.0.:111")
		assert.Error(t, err)
	})
}

func TestClusterMeta_IsComponentRequired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productRW := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(productRW)
	t.Run("error", func(t *testing.T) {
		mockQueryTiDBFromDB(productRW.EXPECT())
		meta := &ClusterMeta{
			Cluster: &management.Cluster{
				Type:            "TiDB",
				Version:         "v5.0.0",
				CpuArchitecture: "x86_64",
			},
		}
		assert.False(t, meta.IsComponentRequired(context.TODO(), "TTT"))
	})

	t.Run("normal", func(t *testing.T) {
		mockQueryTiDBFromDBAnyTimes(productRW.EXPECT())
		meta := &ClusterMeta{
			Cluster: &management.Cluster{
				Type:            "TiDB",
				Version:         "v5.2.2",
				CpuArchitecture: "x86_64",
			},
		}
		assert.True(t, meta.IsComponentRequired(context.TODO(), "TiKV"))
		assert.False(t, meta.IsComponentRequired(context.TODO(), "CDC"))
	})
}

func TestClusterMeta_DeleteInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().DeleteInstance(gomock.Any(), gomock.Any()).Return(nil)

	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			Version: "v4.1.1",
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						ID:     "tidb1111",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{111},
				},
				{
					Entity: common.Entity{
						ID:     "tidb1112",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{112},
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		_, err := meta.DeleteInstance(context.TODO(), "tidb1111")
		assert.NoError(t, err)
		assert.Equal(t, len(meta.Instances["TiDB"]), 1)
		assert.Equal(t, meta.Instances["TiDB"][0].ID, "tidb1112")
	})

	t.Run("error", func(t *testing.T) {
		_, err := meta.DeleteInstance(context.TODO(), "127.0.0.1:222")
		assert.Error(t, err)
		_, err = meta.DeleteInstance(context.TODO(), "127.0.0.2:111")
		assert.Error(t, err)
		_, err = meta.DeleteInstance(context.TODO(), "127.0.0.1:ss")
		assert.Error(t, err)
		_, err = meta.DeleteInstance(context.TODO(), "127.0.0.:111")
		assert.Error(t, err)
	})

}

func TestClusterMeta_CloneMeta(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "cluster01"}}, nil)
	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:       "testCluster",
				TenantId: "tenant01",
			},
			Type:             "TiDB",
			Version:          "v5.0.0",
			Tags:             []string{"tag1"},
			TLS:              false,
			ParameterGroupID: "param1",
			Copies:           4,
			Exclusive:        false,
			CpuArchitecture:  constants.ArchX86,
			MaintainWindow:   "window1",
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Type:         "TiDB",
					Zone:         "Zone1",
					CpuCores:     4,
					Memory:       8,
					DiskType:     "ssd",
					DiskCapacity: 32,
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		got, err := meta.CloneMeta(context.TODO(), structs.CreateClusterParameter{
			Name:       "cluster01",
			DBPassword: "1234",
			Region:     "Region01",
		}, []structs.ClusterResourceParameterCompute{
			{"TiDB", 2, []structs.ClusterResourceParameterComputeResource{
				{Zone: "zone1", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
				{Zone: "zone2", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
			}},
			{"TiKV", 2, []structs.ClusterResourceParameterComputeResource{
				{Zone: "zone1", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
				{Zone: "zone2", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
			}},
		})
		assert.NoError(t, err)
		assert.Equal(t, 5, len(got.Instances))
	})

	t.Run("version error", func(t *testing.T) {
		_, err := meta.CloneMeta(context.TODO(), structs.CreateClusterParameter{
			Name:       "cluster01",
			DBUser:     "user01",
			DBPassword: "1234",
			Region:     "Region01",
			Version:    "v3.0.0",
		}, []structs.ClusterResourceParameterCompute{
			{"TiDB", 2, []structs.ClusterResourceParameterComputeResource{
				{Zone: "zone1", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
				{Zone: "zone2", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
			}},
			{"TiKV", 2, []structs.ClusterResourceParameterComputeResource{
				{Zone: "zone1", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
				{Zone: "zone2", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1},
			}},
		})
		assert.Error(t, err)
	})
}

func TestClusterMeta_GetRelations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().GetRelations(gomock.Any(), "111").Return([]*management.ClusterRelation{
		{
			RelationType:         constants.ClusterRelationStandBy,
			SubjectClusterID:     "111",
			ObjectClusterID:      "222",
			SyncChangeFeedTaskID: "task01",
		},
	}, nil)
	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
		},
	}
	t.Run("normal", func(t *testing.T) {
		got, err := meta.GetRelations(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(got))
	})
}

func TestClusterMeta_StartMaintenance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().SetMaintenanceStatus(gomock.Any(), "111", constants.ClusterMaintenanceScaleIn).Return(nil)
	rw.EXPECT().SetMaintenanceStatus(gomock.Any(), "111", constants.ClusterMaintenanceStopping).Return(errors.Error(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT))

	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			MaintenanceStatus: constants.ClusterMaintenanceNone,
		},
	}

	t.Run("normal", func(t *testing.T) {
		err := meta.StartMaintenance(context.TODO(), constants.ClusterMaintenanceScaleIn)
		assert.NoError(t, err)
		assert.Equal(t, constants.ClusterMaintenanceScaleIn, meta.Cluster.MaintenanceStatus)
	})
	t.Run("err", func(t *testing.T) {
		err := meta.StartMaintenance(context.TODO(), constants.ClusterMaintenanceStopping)
		assert.Error(t, err)
	})
}

func TestClusterMeta_EndMaintenance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().ClearMaintenanceStatus(gomock.Any(), "111", constants.ClusterMaintenanceScaleIn).Return(nil)
	rw.EXPECT().ClearMaintenanceStatus(gomock.Any(), "111", constants.ClusterMaintenanceStopping).Return(errors.Error(errors.TIEM_CLUSTER_MAINTENANCE_CONFLICT))

	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			MaintenanceStatus: constants.ClusterMaintenanceNone,
		},
	}

	t.Run("normal", func(t *testing.T) {
		err := meta.EndMaintenance(context.TODO(), constants.ClusterMaintenanceScaleIn)
		assert.NoError(t, err)
		assert.Equal(t, constants.ClusterMaintenanceNone, meta.Cluster.MaintenanceStatus)
	})
	t.Run("err", func(t *testing.T) {
		err := meta.EndMaintenance(context.TODO(), constants.ClusterMaintenanceStopping)
		assert.Error(t, err)
	})
}

func TestClusterMeta_UpdateMeta(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().UpdateMeta(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			MaintenanceStatus: constants.ClusterMaintenanceNone,
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						ID:     "tidb1111",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{111},
				},
				{
					Entity: common.Entity{
						ID:     "tidb2222",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{111},
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						ID:     "tikv",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports:  []int32{111},
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		err := meta.UpdateMeta(context.TODO())
		assert.NoError(t, err)
	})

}

func TestClusterMeta_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().Delete(gomock.Any(), "111").Return(nil)
	rw.EXPECT().Delete(gomock.Any(), "").Return(errors.Error(errors.TIEM_CLUSTER_NOT_FOUND))

	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			MaintenanceStatus: constants.ClusterMaintenanceNone,
		},
	}

	t.Run("normal", func(t *testing.T) {
		meta.Cluster.ID = "111"
		err := meta.Delete(context.TODO())
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		meta.Cluster.ID = ""
		err := meta.Delete(context.TODO())
		assert.Error(t, err)
	})
}

func TestClusterMeta_ClearClusterPhysically(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().ClearClusterPhysically(gomock.Any(), "111").Return(nil)
	rw.EXPECT().ClearClusterPhysically(gomock.Any(), "").Return(errors.Error(errors.TIEM_UNSUPPORT_PRODUCT))

	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID: "111",
			},
			MaintenanceStatus: constants.ClusterMaintenanceNone,
		},
	}

	t.Run("normal", func(t *testing.T) {
		meta.Cluster.ID = "111"
		err := meta.ClearClusterPhysically(context.TODO())
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		meta.Cluster.ID = ""
		err := meta.ClearClusterPhysically(context.TODO())
		assert.Error(t, err)
	})
}

func TestClusterMeta_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{
		Entity: common.Entity{
			ID: "111",
		},
		MaintenanceStatus: constants.ClusterMaintenanceNone,
	}, []*management.ClusterInstance{
		{
			Entity: common.Entity{
				ID:     "111111",
				Status: string(constants.ClusterRunning),
			},
			Type:   "TiDB",
			HostIP: []string{"127.0.0.1"},
			Ports:  []int32{111},
		},
		{
			Entity: common.Entity{
				ID:     "222222",
				Status: string(constants.ClusterRunning),
			},
			Type:   "TiDB",
			HostIP: []string{"127.0.0.1"},
			Ports:  []int32{111},
		},
		{
			Entity: common.Entity{
				ID:     "333333",
				Status: string(constants.ClusterRunning),
			},
			Type:   "TiKV",
			HostIP: []string{"127.0.0.1"},
			Ports:  []int32{111},
		},
	}, []*management.DBUser{
		{
			ClusterID: "111",
			Name:      constants.DBUserName[constants.Root],
			Password:  common.PasswordInExpired{Val: "12345678", UpdateTime: time.Now().AddDate(0, 0, -5)},
			RoleType:  string(constants.Root),
		}}, nil)

	rw.EXPECT().GetMeta(gomock.Any(), "222").Return(nil, nil, nil, errors.Error(errors.TIEM_UNSUPPORT_PRODUCT))

	t.Run("normal", func(t *testing.T) {
		meta, err := Get(context.TODO(), "111")
		assert.NoError(t, err)
		assert.Equal(t, "111", meta.Cluster.ID)
		assert.Equal(t, 2, len(meta.Instances["TiDB"]))
		assert.Equal(t, "333333", meta.Instances["TiKV"][0].ID)
	})

	t.Run("error", func(t *testing.T) {
		_, err := Get(context.TODO(), "222")
		assert.Error(t, err)
	})
}

func TestClusterMeta_Address(t *testing.T) {
	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:        "2145635758",
				TenantId:  "324567",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:              "koojdafij",
			Type:              "TiDB",
			Version:           "v5.0.0",
			Tags:              []string{"111", "333"},
			OwnerId:           "436534636u",
			ParameterGroupID:  "352467890",
			Copies:            4,
			Region:            "Region1",
			CpuArchitecture:   "x86_64",
			MaintenanceStatus: constants.ClusterMaintenanceCreating,
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "TiDB",
					Version:  "v5.0.0",
					Ports:    []int32{10001, 10002, 10003, 10004},
					HostIP:   []string{"127.0.0.1"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 3,
					Memory:   7,
					Type:     "TiDB",
					Version:  "v5.0.0",
					Ports:    []int32{10001, 10002, 10003, 10004},
					HostIP:   []string{"127.0.0.1"},
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "TiKV",
					Version:  "v5.0.0",
					Ports:    []int32{20001, 20002, 20003, 20004},
					HostIP:   []string{"127.0.0.2"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 3,
					Memory:   7,
					Type:     "TiKV",
					Version:  "v5.0.0",
					Ports:    []int32{20001, 20002, 20003, 20004},
					HostIP:   []string{"127.0.0.2"},
				},
			},
			"PD": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "PD",
					Version:  "v5.0.0",
					Ports:    []int32{30001, 30002, 30003, 30004},
					HostIP:   []string{"127.0.0.3"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 3,
					Memory:   7,
					Type:     "PD",
					Version:  "v5.0.0",
					Ports:    []int32{30001, 30002, 30003, 30004},
					HostIP:   []string{"127.0.0.3"},
				},
			},
			"Prometheus": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "Prometheus",
					Version:  "v5.0.0",
					Ports:    []int32{40001, 40002},
					HostIP:   []string{"127.0.0.4"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 3,
					Memory:   7,
					Type:     "Prometheus",
					Version:  "v5.0.0",
					Ports:    []int32{40001, 40002},
					HostIP:   []string{"127.0.0.4"},
				},
			},
			"Grafana": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "Grafana",
					Version:  "v5.0.0",
					Ports:    []int32{50001, 50002},
					HostIP:   []string{"127.0.0.5"},
				},
			},
			"AlertManger": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "AlertManger",
					Version:  "v5.0.0",
					Ports:    []int32{60001, 60002},
					HostIP:   []string{"127.0.0.6"},
				},
			},
		},
	}

	t.Run("GetClusterConnectAddresses", func(t *testing.T) {
		addresses := meta.GetClusterConnectAddresses()
		assert.Equal(t, 2, len(addresses))
		assert.Equal(t, "127.0.0.1", addresses[0].IP)
		assert.Equal(t, 10001, addresses[1].Port)
	})

	t.Run("GetClusterStatusAddress", func(t *testing.T) {
		addresses := meta.GetClusterStatusAddress()
		assert.Equal(t, 2, len(addresses))
		assert.Equal(t, "127.0.0.1", addresses[0].IP)
		assert.Equal(t, 10002, addresses[1].Port)
	})

	t.Run("GetTiKVStatusAddress", func(t *testing.T) {
		addresses := meta.GetTiKVStatusAddress()
		assert.Equal(t, 2, len(addresses))
		assert.Equal(t, "127.0.0.2", addresses[0].IP)
		assert.Equal(t, 20002, addresses[1].Port)
	})

	t.Run("GetPDClientAddresses", func(t *testing.T) {
		addresses := meta.GetPDClientAddresses()
		assert.Equal(t, 2, len(addresses))
		assert.Equal(t, "127.0.0.3", addresses[0].IP)
		assert.Equal(t, 30001, addresses[1].Port)
	})
	t.Run("GetMonitorAddresses", func(t *testing.T) {
		addresses := meta.GetMonitorAddresses()
		assert.Equal(t, 2, len(addresses))
		assert.Equal(t, "127.0.0.4", addresses[0].IP)
		assert.Equal(t, 40001, addresses[1].Port)
	})
	t.Run("GetGrafanaAddresses", func(t *testing.T) {
		address := meta.GetGrafanaAddresses()
		assert.Equal(t, 1, len(address))
		assert.Equal(t, "127.0.0.5", address[0].IP)
		assert.Equal(t, 50001, address[0].Port)
	})
	t.Run("GetAlertManagerAddresses", func(t *testing.T) {
		address := meta.GetAlertManagerAddresses()
		assert.Equal(t, 1, len(address))
		assert.Equal(t, "127.0.0.6", address[0].IP)
		assert.Equal(t, 60001, address[0].Port)
	})
	//t.Run("GetClusterUserNamePasswd", func(t *testing.T) {
	//	user := meta.GetClusterUserNamePasswd()
	//	assert.Equal(t, "2145635758", user.ClusterID)
	//})
}
func TestClusterMeta_Display(t *testing.T) {

	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:        "2145635758",
				TenantId:  "324567",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:              "koojdafij",
			Version:           "v5.0.0",
			Tags:              []string{"111", "333"},
			OwnerId:           "436534636u",
			ParameterGroupID:  "352467890",
			Copies:            4,
			Region:            "Region1",
			CpuArchitecture:   "x86_64",
			MaintenanceStatus: constants.ClusterMaintenanceCreating,
		},
		Instances: map[string][]*management.ClusterInstance{
			"Prometheus": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "Prometheus",
					Version:  "v5.0.0",
					HostIP:   []string{"127.0.0.1"},
				},
			},
			"AlertManger": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "AlertManger",
					Version:  "v5.0.0",
					HostIP:   []string{"127.0.0.1"},
					Ports:    []int32{999},
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "TiKV",
					Version:  "v5.0.0",
					HostIP:   []string{"127.0.0.1"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 3,
					Memory:   7,
					Type:     "TiKV",
					Version:  "v5.0.0",
					HostIP:   []string{"127.0.0.1"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone2",
					CpuCores: 4,
					Memory:   8,
					Type:     "TiKV",
					Version:  "v5.0.0",
					HostIP:   []string{"127.0.0.1"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "TiKV",
					Version:  "v5.0.0",
					HostIP:   []string{"127.0.0.1"},
					Ports:    []int32{1},
				},
			},
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "TiDB",
					Version:  "v5.0.0",
					Ports:    []int32{1},
					HostIP:   []string{"127.0.0.1"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 3,
					Memory:   7,
					Type:     "TiDB",
					Version:  "v5.0.0",
					Ports:    []int32{1},

					HostIP: []string{"127.0.0.1"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone2",
					CpuCores: 4,
					Memory:   8,
					Type:     "TiDB",
					Version:  "v5.0.0",
					Ports:    []int32{1},

					HostIP: []string{"127.0.0.1"},
				},
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "TiDB",
					Version:  "v5.0.0",
					Ports:    []int32{1},
					HostIP:   []string{"127.0.0.1"},
				},
			},
			"Grafana": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceRunning),
					},
					Zone:     "zone1",
					CpuCores: 4,
					Memory:   8,
					Type:     "Grafana",
					Version:  "v5.0.0",
					HostIP:   []string{"127.4.5.6"},
					Ports:    []int32{888},
				},
			},
		},
	}

	t.Run("cluster", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		rw := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(rw)
		rw.EXPECT().GetMasters(gomock.Any(), gomock.Any()).Return([]*management.ClusterRelation{
			{
				SubjectClusterID: "01",
				ObjectClusterID:  "02",
			},
		}, nil)
		rw.EXPECT().GetSlaves(gomock.Any(), gomock.Any()).Return([]*management.ClusterRelation{
			{
				SubjectClusterID: "01",
				ObjectClusterID:  "02",
			},
		}, nil)
		cluster := meta.DisplayClusterInfo(context.TODO())
		assert.Equal(t, meta.Cluster.ID, cluster.ID)
		assert.Equal(t, meta.Cluster.Name, cluster.Name)
		assert.Equal(t, meta.Cluster.Version, cluster.Version)
		assert.Equal(t, meta.Cluster.Type, cluster.Type)
		assert.Equal(t, meta.Cluster.OwnerId, cluster.UserID)
		assert.Equal(t, cluster.ExtranetConnectAddresses, cluster.IntranetConnectAddresses)
		assert.Equal(t, string(meta.Cluster.CpuArchitecture), cluster.CpuArchitecture)
		assert.Equal(t, string(meta.Cluster.MaintenanceStatus), cluster.MaintainStatus)
		assert.Equal(t, meta.Cluster.Copies, cluster.Copies)
		assert.Equal(t, "127.0.0.1:999", cluster.AlertUrl)
		assert.Equal(t, "127.4.5.6:888", cluster.GrafanaUrl)

	})
	t.Run("instance", func(t *testing.T) {
		topology, resource := meta.DisplayInstanceInfo(context.TODO())
		// sorted topology
		assert.Equal(t, 11, len(topology.Topology))
		assert.Equal(t, topology.Topology[1].Type, topology.Topology[0].Type)
		assert.NotEqual(t, topology.Topology[3].Type, topology.Topology[4].Type)

		// unsorted instanceResource
		assert.Equal(t, 5, len(resource.InstanceResource))
		assert.NotEqual(t, resource.InstanceResource[0].Type, resource.InstanceResource[1].Type)
	})
}

func TestClusterMeta_Query(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().QueryMetas(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResult("test"),
		structs.Page{
			Page: 1, PageSize: 1, Total: 5,
		}, nil)

	rw.EXPECT().GetMasters(gomock.Any(), gomock.Any()).Return([]*management.ClusterRelation{
		{
			SubjectClusterID: "01",
			ObjectClusterID:  "02",
		},
	}, nil)
	rw.EXPECT().GetSlaves(gomock.Any(), gomock.Any()).Return([]*management.ClusterRelation{
		{
			SubjectClusterID: "01",
			ObjectClusterID:  "02",
		},
	}, nil)

	resp, total, err := Query(context.TODO(), cluster.QueryClustersReq{
		ClusterID: "111",
		Name:      "111",
		Type:      "TiDB",
		Status:    string(constants.ClusterRunning),
		Tag:       "t",
	})
	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Equal(t, 1, len(resp.Clusters))
	assert.NotEmpty(t, resp.Clusters[0].AlertUrl)
	assert.NotEmpty(t, resp.Clusters[0].GrafanaUrl)

}

func mockResult(name string) []*management.Result {
	one := &management.Result{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:        "id",
				TenantId:  "324567",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name: name,
			//DBUser:            "kodjsfn",
			Type:              "TiDB",
			Version:           "v5.0.0",
			Tags:              []string{"111", "333"},
			OwnerId:           "436534636u",
			ParameterGroupID:  "352467890",
			Copies:            4,
			Region:            "Region1",
			CpuArchitecture:   "x86_64",
			MaintenanceStatus: constants.ClusterMaintenanceCreating,
		},
		Instances: []*management.ClusterInstance{
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "TiDB",
				Version:  "v5.0.0",
				Ports:    []int32{1},
				HostIP:   []string{"127.0.0.1"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "TiKV",
				Version:  "v5.0.0",
				HostIP:   []string{"127.0.0.1"},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "Grafana",
				Version:  "v5.0.0",
				HostIP:   []string{"127.4.5.6"},
				Ports:    []int32{888},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "AlertManger",
				Version:  "v5.0.0",
				HostIP:   []string{"127.0.0.1"},
				Ports:    []int32{999},
			},
			{
				Entity: common.Entity{
					Status: string(constants.ClusterInstanceRunning),
				},
				Zone:     "zone1",
				CpuCores: 4,
				Memory:   8,
				Type:     "Prometheus",
				Version:  "v5.0.0",
				HostIP:   []string{"127.0.0.1"},
			},
		},
		DBUsers: []*management.DBUser{
			{
				ClusterID: "id",
				Name:      constants.DBUserName[constants.Root],
				Password:  common.PasswordInExpired{Val: "12345678", UpdateTime: time.Now().AddDate(0, 0, -5)},
				RoleType:  string(constants.Root),
			},
		},
	}

	return []*management.Result{one}
}

func TestQueryInstanceLogInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().QueryInstancesByHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*management.ClusterInstance{
		{Type: string(constants.ComponentIDPD), DiskPath: "/a/a", ClusterID: "aaaa", HostIP: []string{"127.0.0.1"}},
		{Type: string(constants.ComponentIDCDC), DiskPath: "/b/b", ClusterID: "bbbb", HostIP: []string{"127.0.0.2"}},
	}, nil)

	t.Run("normal", func(t *testing.T) {
		infos, err := QueryInstanceLogInfo(context.TODO(), "host", []string{}, []string{})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(infos))
		assert.Equal(t, "/a/a/aaaa/pd-data", infos[0].DataDir)
		assert.Equal(t, "/b/b/bbbb/cdc-deploy", infos[1].DeployDir)
		assert.Equal(t, constants.ComponentIDPD, infos[0].InstanceType)
		assert.Equal(t, "127.0.0.2", infos[1].IP)
	})
}

func mockSpec() *spec.Specification {
	return &spec.Specification{
		TiDBServers: []*spec.TiDBSpec{
			{Host: "127.0.0.1", Port: 1, StatusPort: 2},
			{Host: "127.0.0.2", Port: 3, StatusPort: 4},
		},
		TiKVServers: []*spec.TiKVSpec{
			{Host: "127.0.0.4", Port: 5, StatusPort: 6},
			{Host: "127.0.0.5", Port: 7, StatusPort: 8},
		},
		TiFlashServers: []*spec.TiFlashSpec{
			{Host: "127.0.0.6", TCPPort: 9, HTTPPort: 10, FlashServicePort: 11, FlashProxyPort: 12, FlashProxyStatusPort: 13, StatusPort: 14},
		},
		CDCServers: []*spec.CDCSpec{
			{Host: "127.0.0.7", Port: 15},
			{Host: "127.0.0.8", Port: 16},
		},
		PDServers: []*spec.PDSpec{
			{Host: "127.0.0.9", ClientPort: 17, PeerPort: 18},
			{Host: "127.0.0.10", ClientPort: 19, PeerPort: 20},
		},
		Grafanas: []*spec.GrafanaSpec{
			{Host: "127.0.0.11", Port: 21},
		},
		Alertmanagers: []*spec.AlertmanagerSpec{
			{Host: "127.0.0.12", WebPort: 22, ClusterPort: 23},
		},
		Monitors: []*spec.PrometheusSpec{
			{Host: "127.0.0.13", Port: 24},
		},
	}
}

func TestClusterMeta_ParseTopologyFromConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().UpdateInstance(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		meta := &ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID:       "clusterId",
					TenantId: "tenantId",
				},
				Version: "v5.2.2",
			},
		}
		err := meta.ParseTopologyFromConfig(context.TODO(), mockSpec())
		assert.NoError(t, err)
		assert.Equal(t, 8, len(meta.Instances))
		assert.Equal(t, "v5.2.2", meta.Instances[string(constants.ComponentIDTiKV)][1].Version)
		assert.Equal(t, "tenantId", meta.Instances[string(constants.ComponentIDTiDB)][1].TenantId)
		assert.Equal(t, "clusterId", meta.Instances[string(constants.ComponentIDPD)][0].ClusterID)
		assert.NotEmpty(t, meta.Instances[string(constants.ComponentIDCDC)][0].Ports)
		assert.NotEmpty(t, meta.Instances[string(constants.ComponentIDTiFlash)][0].HostIP)
		assert.Equal(t, string(constants.ClusterInstanceRunning), meta.Instances[string(constants.ComponentIDGrafana)][0].Status)
		assert.Equal(t, string(constants.ClusterInstanceRunning), meta.Instances[string(constants.ComponentIDGrafana)][0].Status)
		assert.Equal(t, string(constants.ClusterInstanceRunning), meta.Instances[string(constants.ComponentIDGrafana)][0].Status)
	})
}

func TestClusterMeta_GenerateTakeoverResourceRequirements(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().UpdateInstance(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		meta := &ClusterMeta{
			Cluster: &management.Cluster{
				Entity: common.Entity{
					ID:       "clusterId",
					TenantId: "tenantId",
				},
				Version: "v5.2.2",
			},
		}
		err := meta.ParseTopologyFromConfig(context.TODO(), mockSpec())
		assert.NoError(t, err)
		requirements, instances := meta.GenerateTakeoverResourceRequirements(context.TODO())
		assert.Equal(t, 12, len(requirements))
		assert.Equal(t, len(requirements), len(instances))
	})
}

func TestClusterMeta_GetVersion(t *testing.T) {
	meta := ClusterMeta{
		Cluster: &management.Cluster{
			Version: "v5.2.2",
		},
	}

	assert.Equal(t, "v5", meta.GetMajorVersion())
	assert.Equal(t, "v5.2", meta.GetMinorVersion())
	assert.Equal(t, "v5.2.2", meta.GetRevision())

}

func TestClusterMeta_BuildForTakeover(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	meta := &ClusterMeta{}

	rw.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{Entity: common.Entity{ID: "1234"}}, nil)
	rw.EXPECT().CreateDBUser(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	err := meta.BuildForTakeover(context.TODO(), "test", "1234")
	assert.NoError(t, err)
	//assert.Equal(t, meta.Cluster.Name, params.Name)
	//assert.Equal(t, meta.Cluster.Type, params.Type)
	//assert.Equal(t, meta.Cluster.Version, params.Version)
	//assert.Equal(t, meta.Cluster.TLS, params.TLS)
	//assert.Equal(t, meta.Cluster.Tags, params.Tags)
	//assert.Equal(t, meta.DBUsers[string(constants.Root)].Password, params.DBPassword)
	assert.NotEmpty(t, meta.DBUsers[string(constants.Root)].ClusterID)
	assert.Equal(t, meta.DBUsers[string(constants.Root)].ClusterID, "1234")

	//cluster := &management.Cluster{
	//	Entity: common.Entity{
	//		ID: "123",
	//		TenantId: "tttt",
	//	},
	//	Name: "testName",
	//	OwnerId: "oooo",
	//}
	//instance := make(map[string][]*management.ClusterInstance)
	//user := make(map[string]*management.DBUser)
	//type args struct {
	//	ctx        context.Context
	//	name       string
	//	dbUser     string
	//	dbPassword string
	//}
	//tests := []struct {
	//	name    string
	//	args    args
	//	wantErr bool
	//}{
	//	// TODO: Add test cases.
	//	{"normal", args{context.TODO(), "test", "root", "1234567"},false},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		p := &ClusterMeta{
	//			Cluster:              cluster,
	//			Instances:            instance,
	//			DBUsers:              user,
	//		}
	//		if err := p.BuildForTakeover(tt.args.ctx, tt.args.name, tt.args.dbUser, tt.args.dbPassword); (err != nil) != tt.wantErr {
	//			t.Errorf("BuildForTakeover() error = %v, wantErr %v", err, tt.wantErr)
	//		} else {
	//			fmt.Println(p.Cluster.ID, p.DBUsers[string(constants.Root)].ID)
	//
	//			assert.Equal(t, p.Cluster.ID, p.DBUsers[string(constants.Root)].ID)
	//			assert.NotEmpty(t, p.DBUsers[string(constants.Root)].ID)
	//		}
	//	})
	//}
}

func TestClusterMeta_IsTakenOver(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		meta := ClusterMeta{
			Cluster: &management.Cluster{},
		}
		assert.False(t, meta.IsTakenOver())
	})
	t.Run("true", func(t *testing.T) {
		meta := ClusterMeta{
			Cluster: &management.Cluster{
				Tags: []string{"aaa", TagTakeover, "bbb"},
			},
		}
		assert.True(t, meta.IsTakenOver())
	})
	t.Run("false", func(t *testing.T) {
		meta := ClusterMeta{
			Cluster: &management.Cluster{
				Tags: []string{"aaa", "bbb"},
			},
		}
		assert.False(t, meta.IsTakenOver())
	})
}
