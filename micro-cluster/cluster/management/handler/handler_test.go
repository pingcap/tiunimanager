/*******************************************************************************
 * @File: handler_test
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/9
*******************************************************************************/

package handler

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClusterMeta_BuildCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	params := structs.CreateClusterParameter {
		Name: "aaa",
		DBUser: "user",
		DBPassword: "password",
		Type: "type",
		Version: "version",
		TLS: false,
		Tags: []string{"t1", "t2"},
	}
	meta := &ClusterMeta{}

	rw.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&management.Cluster{}, nil)

	err := meta.BuildCluster(context.TODO(), params)
	assert.NoError(t, err)
	assert.Equal(t, meta.Cluster.Name, params.Name)
	assert.Equal(t, meta.Cluster.DBUser, params.DBUser)
	assert.Equal(t, meta.Cluster.DBPassword, params.DBPassword)
	assert.Equal(t, meta.Cluster.Type, params.Type)
	assert.Equal(t, meta.Cluster.Version, params.Version)
	assert.Equal(t, meta.Cluster.TLS, params.TLS)
	assert.Equal(t, meta.Cluster.Tags, params.Tags)
}

func TestClusterMeta_AddInstances(t *testing.T) {
	meta := &ClusterMeta {
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
						Status: string(constants.ClusterRunning),
					},
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		err := meta.AddInstances(context.TODO(), []structs.ClusterResourceParameterCompute{
			{"TiDB", 2, []structs.ClusterResourceParameterComputeResource{
				{Zone: "zone1", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1 },
				{Zone: "zone2", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1 },
			}},
			{"TiKV", 2, []structs.ClusterResourceParameterComputeResource{
				{Zone: "zone1", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1 },
				{Zone: "zone2", Spec: "4C8G", DiskType: "SSD", DiskCapacity: 3, Count: 1 },
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

func TestClusterMeta_GenerateTopologyConfig(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		meta := &ClusterMeta {
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
							Status: string(constants.ClusterRunning),
						},
						HostIP: []string{"127.0.0.1"},
						Ports: []int32{
							1,2,3,4,5,6,
						},
						// todo render
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
	rw.EXPECT().UpdateStatus(gomock.Any(), "111", gomock.Any()).Return(errors.New("not existed"))
	rw.EXPECT().UpdateStatus(gomock.Any(), "222", gomock.Any()).Return(nil)
	rw.EXPECT().UpdateStatus(gomock.Any(), "", gomock.Any()).Return(errors.New("empty"))

	meta := &ClusterMeta {
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
	meta := &ClusterMeta {
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
						ID: "tidb1111",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports: []int32{111},
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		instance, err := meta.GetInstance(context.TODO(), "127.0.0.1:111")
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
	meta := &ClusterMeta{
		Cluster: &management.Cluster{
			Type: "TiDB",
			Version: "v4.0.12",
		},
	}
	assert.True(t, meta.IsComponentRequired(context.TODO(), "TiKV"))
	assert.False(t, meta.IsComponentRequired(context.TODO(), "TiFlash"))
}

func TestClusterMeta_DeleteInstance(t *testing.T)  {
	meta := &ClusterMeta {
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
						ID: "tidb1111",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports: []int32{111},
				},
			},
		},
	}

	t.Run("normal", func(t *testing.T) {
		err := meta.DeleteInstance(context.TODO(), "127.0.0.1:111")
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		err := meta.DeleteInstance(context.TODO(), "127.0.0.1:222")
		assert.Error(t, err)
		err = meta.DeleteInstance(context.TODO(), "127.0.0.2:111")
		assert.Error(t, err)
		err = meta.DeleteInstance(context.TODO(), "127.0.0.1:ss")
		assert.Error(t, err)
		err = meta.DeleteInstance(context.TODO(), "127.0.0.:111")
		assert.Error(t, err)
	})

}

func TestClusterMeta_StartMaintenance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rw := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)

	rw.EXPECT().SetMaintenanceStatus(gomock.Any(), "111", constants.ClusterMaintenanceScaleIn).Return(nil)
	rw.EXPECT().SetMaintenanceStatus(gomock.Any(), "111", constants.ClusterMaintenanceStopping).Return(errors.New("conflicted"))

	meta := &ClusterMeta {
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
	rw.EXPECT().ClearMaintenanceStatus(gomock.Any(), "111", constants.ClusterMaintenanceStopping).Return(errors.New("conflicted"))

	meta := &ClusterMeta {
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

	meta := &ClusterMeta {
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
						ID: "tidb1111",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports: []int32{111},
				},
				{
					Entity: common.Entity{
						ID: "tidb2222",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports: []int32{111},
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						ID: "tikv",
						Status: string(constants.ClusterRunning),
					},
					HostIP: []string{"127.0.0.1"},
					Ports: []int32{111},
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
	rw.EXPECT().Delete(gomock.Any(), "").Return(errors.New("empty"))

	meta := &ClusterMeta {
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
				ID: "111111",
				Status: string(constants.ClusterRunning),
			},
			Type: "TiDB",
			HostIP: []string{"127.0.0.1"},
			Ports: []int32{111},
		},
		{
			Entity: common.Entity{
				ID: "222222",
				Status: string(constants.ClusterRunning),
			},
			Type: "TiDB",
			HostIP: []string{"127.0.0.1"},
			Ports: []int32{111},
		},
		{
			Entity: common.Entity{
				ID: "333333",
				Status: string(constants.ClusterRunning),
			},
			Type: "TiKV",
			HostIP: []string{"127.0.0.1"},
			Ports: []int32{111},
		},
	}, nil)

	rw.EXPECT().GetMeta(gomock.Any(), "222").Return(nil, nil, errors.New("empty"))

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
