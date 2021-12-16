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

package management

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	libCommon "github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGormClusterReadWrite_MaintenanceStatus(t *testing.T) {
	got, err := testRW.Create(context.TODO(), &Cluster{
		Name: "test39907",
		Entity: common.Entity{
			TenantId: "111",
		},
		Tags: []string{"tag1", "tag2"},
	})
	defer testRW.Delete(context.TODO(), got.ID)

	assert.NoError(t, err)
	assert.Equal(t, constants.ClusterMaintenanceNone, got.MaintenanceStatus)

	// set ok
	err = testRW.SetMaintenanceStatus(context.TODO(), got.ID, constants.ClusterMaintenanceStopping)
	assert.NoError(t, err)

	// set failed
	err = testRW.SetMaintenanceStatus(context.TODO(), got.ID, constants.ClusterMaintenanceCloning)
	assert.Error(t, err)

	// check
	check, err := testRW.Get(context.TODO(), got.ID)
	assert.NoError(t, err)
	assert.Equal(t, constants.ClusterMaintenanceStopping, check.MaintenanceStatus)

	// clear failed
	err = testRW.ClearMaintenanceStatus(context.TODO(), got.ID, constants.ClusterMaintenanceCloning)
	assert.Error(t, err)

	// clear ok
	err = testRW.ClearMaintenanceStatus(context.TODO(), got.ID, constants.ClusterMaintenanceStopping)
	assert.NoError(t, err)

	// check
	check, err = testRW.Get(context.TODO(), got.ID)
	assert.NoError(t, err)
	assert.Equal(t, constants.ClusterMaintenanceNone, check.MaintenanceStatus)
}

func TestGormClusterReadWrite_Create(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got, err := testRW.Create(context.TODO(), &Cluster{
			Name: "test1",
			Entity: common.Entity{
				TenantId: "111",
				Status:   string(constants.ClusterRunning),
			},
			Tags: []string{"tag1", "tag2"},
		})
		defer testRW.Delete(context.TODO(), got.ID)

		assert.NoError(t, err)
		assert.NotEmpty(t, got.ID)
		assert.NotEmpty(t, got.UpdatedAt)
		assert.Equal(t, 2, len(got.Tags))
		assert.Equal(t, string(constants.ClusterRunning), got.Status)
	})
	t.Run("default", func(t *testing.T) {
		got, err := testRW.Create(context.TODO(), &Cluster{
			Name: "test32413",
			Entity: common.Entity{
				TenantId: "111",
			},
		})
		defer testRW.Delete(context.TODO(), got.ID)

		assert.NoError(t, err)
		assert.NotEmpty(t, got.ID)
		assert.NotEmpty(t, got.UpdatedAt)
		assert.Equal(t, 0, len(got.Tags))
		assert.Equal(t, string(constants.ClusterInitializing), got.Status)
	})
}

func TestGormClusterReadWrite_CreateRelation(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		relation := &ClusterRelation{
			ObjectClusterID:  "111",
			SubjectClusterID: "222",
			RelationType:     constants.ClusterRelationCloneFrom,
		}
		err := testRW.CreateRelation(context.TODO(), relation)

		assert.NoError(t, err)
		assert.NotEmpty(t, relation.ID)
		assert.NotEmpty(t, relation.UpdatedAt)
	})
}

func TestGormClusterReadWrite_Delete(t *testing.T) {
	cluster := &Cluster{
		Name: "test32431",
		Entity: common.Entity{
			TenantId: "111",
		},
		Tags: []string{"tag1", "tag2"},
	}
	got, _ := testRW.Create(context.TODO(), cluster)
	defer testRW.Delete(context.TODO(), got.ID)
	cluster2 := &Cluster{
		Name: "tesfasfdsaf",
		Entity: common.Entity{
			TenantId: "111",
		},
		Tags: []string{"tag1", "tag2"},
	}
	got2, _ := testRW.Create(context.TODO(), cluster2)
	defer testRW.Delete(context.TODO(), got2.ID)

	t.Run("normal", func(t *testing.T) {
		err := testRW.DB(context.TODO()).Where("id = ?", cluster.ID).First(cluster).Error
		assert.NoError(t, err)
		err = testRW.Delete(context.TODO(), cluster.ID)
		assert.NoError(t, err)
		err = testRW.DB(context.TODO()).Where("id = ?", cluster.ID).First(cluster).Error
		assert.Error(t, err)

		assert.NoError(t, testRW.DB(context.TODO()).Where("id = ?", cluster2.ID).First(cluster2).Error)
	})

	t.Run("not found", func(t *testing.T) {
		err := testRW.Delete(context.TODO(), "whatever")
		assert.Error(t, err)
		assert.Equal(t, libCommon.TIEM_CLUSTER_NOT_FOUND, err.(framework.TiEMError).GetCode())

		err = testRW.Delete(context.TODO(), "")
		assert.Error(t, err)
		assert.Equal(t, libCommon.TIEM_PARAMETER_INVALID, err.(framework.TiEMError).GetCode())
	})
}

func TestGormClusterReadWrite_DeleteRelation(t *testing.T) {
	relation := &ClusterRelation{
		ObjectClusterID:  "111",
		SubjectClusterID: "222",
		RelationType:     constants.ClusterRelationCloneFrom,
	}
	testRW.CreateRelation(context.TODO(), relation)

	t.Run("normal", func(t *testing.T) {
		err := testRW.DB(context.TODO()).Where("id = ?", relation.ID).First(relation).Error
		assert.NoError(t, err)
		err = testRW.DeleteRelation(context.TODO(), relation.ID)
		assert.NoError(t, err)
		err = testRW.DB(context.TODO()).Where("id = ?", relation.ID).First(relation).Error
		assert.Error(t, err)
	})

	t.Run("normal", func(t *testing.T) {
		err := testRW.DB(context.TODO()).Where("id = ?", "whatever").First(relation).Error
		assert.Error(t, err)
	})
}

func TestGormClusterReadWrite_GetMeta(t *testing.T) {
	got, _ := testRW.Create(context.TODO(), &Cluster{
		Name: "test2131243",
		Entity: common.Entity{
			TenantId: "111",
		},
		Tags: []string{"tag1", "tag2"},
	})
	defer testRW.Delete(context.TODO(), got.ID)

	instances := []*ClusterInstance{
		{Entity: common.Entity{TenantId: "111"}, ClusterID: got.ID, Type: "TiKV", Version: "v5.0.0"},
		{Entity: common.Entity{TenantId: "111"}, ClusterID: got.ID, Type: "PD", Version: "v5.0.0"},
	}
	err := testRW.UpdateInstance(context.TODO(), instances...)
	assert.NoError(t, err)

	gotCluster, gotInstances, err := testRW.GetMeta(context.TODO(), got.ID)
	assert.NoError(t, err)
	assert.Equal(t, gotCluster.Tags, gotCluster.Tags)
	assert.Equal(t, 2, len(gotInstances))
}

func TestGormClusterReadWrite_UpdateBaseInfo(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		err := testRW.UpdateClusterInfo(context.TODO(), &Cluster{
			Entity:common.Entity{
				ID: "whatever",
			},
		})
		assert.Error(t, err)
		assert.Equal(t, libCommon.TIEM_CLUSTER_NOT_FOUND, err.(framework.TiEMError).GetCode())

		err = testRW.UpdateClusterInfo(context.TODO(), &Cluster{})
		assert.Error(t, err)
		assert.Equal(t, libCommon.TIEM_PARAMETER_INVALID, err.(framework.TiEMError).GetCode())

	})
}

func TestGormClusterReadWrite_UpdateInstance(t *testing.T) {
	got, _ := testRW.Create(context.TODO(), &Cluster{
		Name: "test9845",
		Entity: common.Entity{
			TenantId: "111",
		},
		Tags: []string{"tag1", "tag2"},
	})
	defer testRW.Delete(context.TODO(), got.ID)

	instances := []*ClusterInstance{
		{Entity: common.Entity{TenantId: "111"}, ClusterID: got.ID, Type: "TiKV", Version: "v5.0.0"},
		{Entity: common.Entity{TenantId: "111"}, ClusterID: got.ID, Type: "PD", Version: "v5.0.0"},
	}
	err := testRW.UpdateInstance(context.TODO(), instances...)
	assert.NoError(t, err)
	_, gotInstances, err := testRW.GetMeta(context.TODO(), got.ID)
	gotInstances[0].Status = string(constants.ClusterRunning)
	gotInstances[0].HostIP = []string{"127.0.0.1", "127.0.0.2"}
	gotInstances[0].CpuCores = 4
	gotInstances[0].Memory = 8

	gotInstances = append(gotInstances,
		&ClusterInstance{Entity: common.Entity{TenantId: "111"}, ClusterID: got.ID, Type: "TiDB", Version: "v5.0.0"},
		&ClusterInstance{Entity: common.Entity{TenantId: "111"}, ClusterID: got.ID, Type: "TiFlash", Version: "v5.0.0"},
	)

	err = testRW.UpdateInstance(context.TODO(), gotInstances...)

	assert.NoError(t, err)

	_, gotInstances, err = testRW.GetMeta(context.TODO(), got.ID)
	assert.NoError(t, err)
	assert.Equal(t, string(constants.ClusterRunning), gotInstances[0].Status)
	assert.Equal(t, []string{"127.0.0.1", "127.0.0.2"}, gotInstances[0].HostIP)
	assert.Equal(t, int8(4), gotInstances[0].CpuCores)
	assert.Equal(t, int8(8), gotInstances[0].Memory)

	assert.Equal(t, string(constants.ClusterInitializing), gotInstances[1].Status)
	assert.Equal(t, []string{}, gotInstances[1].HostIP)
	assert.Equal(t, int8(0), gotInstances[1].CpuCores)
	assert.Equal(t, int8(0), gotInstances[1].Memory)
}

func TestGormClusterReadWrite_UpdateStatus(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got, err := testRW.Create(context.TODO(), &Cluster{
			Name: "test342546",
			Entity: common.Entity{
				TenantId: "111",
			},
			Tags: []string{"tag1", "tag2"},
		})
		defer testRW.Delete(context.TODO(), got.ID)

		assert.NoError(t, err)

		err = testRW.UpdateStatus(context.TODO(), got.ID, constants.ClusterRunning)
		assert.NoError(t, err)

		check, err := testRW.Get(context.TODO(), got.ID)
		assert.NoError(t, err)
		assert.Equal(t, string(constants.ClusterRunning), check.Status)
	})

	t.Run("not found", func(t *testing.T) {
		err := testRW.UpdateStatus(context.TODO(), "whatever", constants.ClusterRunning)
		assert.Error(t, err)
		assert.Equal(t, libCommon.TIEM_CLUSTER_NOT_FOUND, err.(framework.TiEMError).GetCode())

		err = testRW.UpdateStatus(context.TODO(), "", constants.ClusterRunning)
		assert.Error(t, err)
		assert.Equal(t, libCommon.TIEM_PARAMETER_INVALID, err.(framework.TiEMError).GetCode())
	})
}

func TestGormClusterReadWrite_ClusterTopologySnapshot(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		err := testRW.CreateClusterTopologySnapshot(context.TODO(), ClusterTopologySnapshot{
			TenantID: "111", ClusterID: "222", Config: "333",
		})
		assert.NoError(t, err)
	})
	t.Run("empty", func(t *testing.T) {
		err := testRW.CreateClusterTopologySnapshot(context.TODO(), ClusterTopologySnapshot{

			TenantID: "", ClusterID: "222", Config: "333",
		})
		assert.Error(t, err)

		err = testRW.CreateClusterTopologySnapshot(context.TODO(), ClusterTopologySnapshot{
			TenantID: "111", ClusterID: "", Config: "333",
		})
		assert.Error(t, err)

		err = testRW.CreateClusterTopologySnapshot(context.TODO(), ClusterTopologySnapshot{

			TenantID: "111", ClusterID: "222", Config: "",
		})
		assert.Error(t, err)

		_, err = testRW.GetLatestClusterTopologySnapshot(context.TODO(), "")
		assert.Error(t, err)
	})
	t.Run("latest", func(t *testing.T) {
		err := testRW.CreateClusterTopologySnapshot(context.TODO(), ClusterTopologySnapshot{
			TenantID: "tenant111", ClusterID: "cluster111", Config: "content111",
		})
		assert.NoError(t, err)

		err = testRW.CreateClusterTopologySnapshot(context.TODO(), ClusterTopologySnapshot{
			TenantID: "tenant111", ClusterID: "cluster111", Config: "content_modified",
		})
		assert.NoError(t, err)

		err = testRW.CreateClusterTopologySnapshot(context.TODO(), ClusterTopologySnapshot{
			TenantID: "tenant111", ClusterID: "cluster_whatever", Config: "content111",
		})
		assert.NoError(t, err)

		s, err := testRW.GetLatestClusterTopologySnapshot(context.TODO(), "cluster111")
		assert.NoError(t, err)
		assert.Equal(t, "content_modified", s.Config)
	})
}
