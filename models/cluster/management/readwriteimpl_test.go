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
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
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
	t.Run("duplicated name", func(t *testing.T) {
		got1, _ := testRW.Create(context.TODO(), &Cluster{
			Name: "test duplicated name",
			Entity: common.Entity{
				TenantId: "111",
				Status:   string(constants.ClusterRunning),
			},
			Tags: []string{"tag1", "tag2"},
		})
		defer testRW.Delete(context.TODO(), got1.ID)

		_, err := testRW.Create(context.TODO(), &Cluster{
			Name: "test duplicated name",
			Entity: common.Entity{
				TenantId: "111",
				Status:   string(constants.ClusterRunning),
			},
			Tags: []string{"tag1", "tag2"},
		})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_DUPLICATED_NAME, err.(errors.EMError).GetCode())
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
		assert.Equal(t, errors.TIEM_CLUSTER_NOT_FOUND, err.(errors.EMError).GetCode())

		err = testRW.Delete(context.TODO(), "")
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
	})
}

func TestGormClusterReadWrite_DeleteInstance(t *testing.T) {
	instance := &ClusterInstance{
		Entity: common.Entity{
			TenantId: "abc",
		},
		Type:      "TiDB",
		Version:   "v5.0.0",
		ClusterID: "testCluster",
	}
	t.Run("normal", func(t *testing.T) {
		err := testRW.DB(context.TODO()).Create(instance).Error
		assert.NoError(t, err)
		err = testRW.DeleteInstance(context.TODO(), instance.ID)
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		err := testRW.DeleteInstance(context.TODO(), "testInstance")
		assert.Error(t, err)
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
			Entity: common.Entity{
				ID: "whatever",
			},
		})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_CLUSTER_NOT_FOUND, err.(errors.EMError).GetCode())

		err = testRW.UpdateClusterInfo(context.TODO(), &Cluster{})
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())

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
	assert.NoError(t, err)
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
		assert.Equal(t, errors.TIEM_CLUSTER_NOT_FOUND, err.(errors.EMError).GetCode())

		err = testRW.UpdateStatus(context.TODO(), "", constants.ClusterRunning)
		assert.Error(t, err)
		assert.Equal(t, errors.TIEM_PARAMETER_INVALID, err.(errors.EMError).GetCode())
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

func TestClusterReadWrite_QueryMetas(t *testing.T) {
	cluster1 := mockCluster("QueryMetas_test1", "TiDB", constants.ClusterRunning, []string{"tag1", "tag2"})
	cluster2 := mockCluster("QueryMetas_test2", "TiDB", constants.ClusterInitializing, []string{"tag2", "tag1"})
	cluster3 := mockCluster("3test_QueryMetas", "TiDB", constants.ClusterInitializing, []string{"tag1", "tag2"})

	cluster4 := mockCluster("tes_QueryMetas", "Other", constants.ClusterRunning, []string{"tag1", "tag2"})
	cluster5 := mockCluster("QueryMetas_test5", "TiDB", constants.ClusterRunning, []string{"tag121"})
	cluster6 := mockCluster("QueryMetas_test6", "TiDB", constants.ClusterRunning, []string{""})
	cluster7 := mockCluster("QueryMetas_test7", "TiDB", constants.ClusterStopped, []string{"tag1"})

	defer testRW.Delete(context.TODO(), cluster1)
	defer testRW.Delete(context.TODO(), cluster2)
	defer testRW.Delete(context.TODO(), cluster3)
	defer testRW.Delete(context.TODO(), cluster4)
	defer testRW.Delete(context.TODO(), cluster5)
	defer testRW.Delete(context.TODO(), cluster6)
	defer testRW.Delete(context.TODO(), cluster7)

	t.Run("normal", func(t *testing.T) {
		results, page, err := testRW.QueryMetas(context.TODO(), Filters{
			TenantId: "1919",
			NameLike: "test",
			Tag:      "tag1",
			Type:     "TiDB",
			StatusFilters: []constants.ClusterRunningStatus{
				constants.ClusterRunning,
				constants.ClusterInitializing,
			},
		}, structs.PageRequest{
			Page:     1,
			PageSize: 2,
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, page.Total)
		assert.Equal(t, cluster3, results[0].Cluster.ID)
		assert.Equal(t, 2, len(results))
		assert.Equal(t, 2, len(results[0].Instances))
	})

	t.Run("no result", func(t *testing.T) {
		_, page, err := testRW.QueryMetas(context.TODO(), Filters{
			TenantId: "1919",
			NameLike: "whatever",
			Tag:      "tag1",
			Type:     "TiDB",
			StatusFilters: []constants.ClusterRunningStatus{
				constants.ClusterRunning,
				constants.ClusterInitializing,
			},
		}, structs.PageRequest{
			Page:     0,
			PageSize: 5,
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, page.Total)
	})

	t.Run("error", func(t *testing.T) {
		_, _, err := testRW.QueryMetas(context.TODO(), Filters{
			NameLike: "whatever",
			Tag:      "tag1",
			Type:     "TiDB",
			StatusFilters: []constants.ClusterRunningStatus{
				constants.ClusterRunning,
				constants.ClusterInitializing,
			},
		}, structs.PageRequest{
			Page:     0,
			PageSize: 5,
		})
		assert.Error(t, err)
	})
	t.Run("empty filter", func(t *testing.T) {
		_, page, err := testRW.QueryMetas(context.TODO(), Filters{
			TenantId:      "1919",
			NameLike:      "",
			Tag:           "",
			Type:          "",
			StatusFilters: []constants.ClusterRunningStatus{},
		}, structs.PageRequest{
			Page:     0,
			PageSize: 5,
		})
		assert.NoError(t, err)
		assert.Equal(t, 7, page.Total)
	})

	t.Run("page", func(t *testing.T) {
		result, page, err := testRW.QueryMetas(context.TODO(), Filters{
			TenantId:      "1919",
			NameLike:      "",
			Tag:           "",
			Type:          "",
			StatusFilters: []constants.ClusterRunningStatus{},
		}, structs.PageRequest{
			Page:     5,
			PageSize: 5,
		})
		assert.NoError(t, err)
		assert.Equal(t, 7, page.Total)
		assert.Equal(t, 0, len(result))
	})
}

func mockCluster(name string, clusterType string, status constants.ClusterRunningStatus, tags []string) string {
	got, _ := testRW.Create(context.TODO(), &Cluster{
		Name: name,
		Entity: common.Entity{
			TenantId: "1919",
			Status:   string(status),
		},
		Tags: tags,
		Type: clusterType,
	})

	instances := []*ClusterInstance{
		{Entity: common.Entity{TenantId: "1919"}, ClusterID: got.ID, Type: "TiKV", Version: "v5.0.0"},
		{Entity: common.Entity{TenantId: "1919"}, ClusterID: got.ID, Type: "PD", Version: "v5.0.0"},
	}
	testRW.UpdateInstance(context.TODO(), instances...)
	return got.ID
}

func TestClusterReadWrite_Relations(t *testing.T) {
	relation1 := &ClusterRelation{
		ObjectClusterID:  "test_relation",
		SubjectClusterID: "222",
		RelationType:     constants.ClusterRelationCloneFrom,
	}
	err := testRW.CreateRelation(context.TODO(), relation1)
	assert.NoError(t, err)

	err = testRW.CreateRelation(context.TODO(), &ClusterRelation{
		ObjectClusterID:  "test_relation",
		SubjectClusterID: "222",
		RelationType:     constants.ClusterRelationSlaveTo,
	})
	assert.NoError(t, err)

	err = testRW.CreateRelation(context.TODO(), &ClusterRelation{
		ObjectClusterID:  "test_relation",
		SubjectClusterID: "333",
		RelationType:     constants.ClusterRelationSlaveTo,
	})
	assert.NoError(t, err)

	err = testRW.CreateRelation(context.TODO(), &ClusterRelation{
		ObjectClusterID:  "333",
		SubjectClusterID: "test_relation",
		RelationType:     constants.ClusterRelationSlaveTo,
	})
	assert.NoError(t, err)

	r, err := testRW.GetRelations(context.TODO(), "test_relation")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(r))

	r, err = testRW.GetRelations(context.TODO(), "222")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(r))

	err = testRW.DeleteRelation(context.TODO(), relation1.ID)
	assert.NoError(t, err)

	r, err = testRW.GetRelations(context.TODO(), "test_relation")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(r))

}

func TestClusterReadWrite_QueryInstancesByHost(t *testing.T) {
	got, _ := testRW.Create(context.TODO(), &Cluster{
		Name: "testQueryInstance",
		Entity: common.Entity{
			TenantId: "testQueryInstance",
		},
		Tags: []string{"tag1", "tag2"},
	})
	defer testRW.Delete(context.TODO(), got.ID)

	instances := []*ClusterInstance{
		{HostID: "testHostId", Entity: common.Entity{TenantId: "testQueryInstance", Status: string(constants.ClusterInstanceRunning)}, ClusterID: got.ID, Type: "TiKV", Version: "v5.0.0"},
		{Entity: common.Entity{TenantId: "testQueryInstance", Status: string(constants.ClusterInstanceInitializing)}, ClusterID: got.ID, Type: "PD", Version: "v5.0.0"},
		{HostID: "testHostId", Entity: common.Entity{TenantId: "testQueryInstance", Status: string(constants.ClusterInstanceFailure)}, ClusterID: got.ID, Type: "CDC", Version: "v5.0.0"},
	}
	testRW.UpdateInstance(context.TODO(), instances...)

	got2, _ := testRW.Create(context.TODO(), &Cluster{
		Name: "another",
		Entity: common.Entity{
			TenantId: "testQueryInstance",
		},
		Tags: []string{"tag1", "tag2"},
	})
	defer testRW.Delete(context.TODO(), got2.ID)

	instances2 := []*ClusterInstance{
		{HostID: "testHostId", Entity: common.Entity{TenantId: "testQueryInstance", Status: string(constants.ClusterInstanceRecovering)}, ClusterID: got2.ID, Type: "TiKV", Version: "v5.0.0"},
		{HostID: "testHostId", Entity: common.Entity{TenantId: "testQueryInstance", Status: string(constants.ClusterInstanceInitializing)}, ClusterID: got2.ID, Type: "PD", Version: "v5.0.0"},
		{Entity: common.Entity{TenantId: "testQueryInstance", Status: string(constants.ClusterInstanceRecovering)}, ClusterID: got2.ID, Type: "CDC", Version: "v5.0.0"},
	}
	testRW.UpdateInstance(context.TODO(), instances2...)

	t.Run("normal", func(t *testing.T) {
		result, err := testRW.QueryInstancesByHost(context.TODO(), "testHostId", []string{}, []string{})
		assert.NoError(t, err)
		assert.Equal(t, 4, len(result))
	})
	t.Run("without host", func(t *testing.T) {
		_, err := testRW.QueryInstancesByHost(context.TODO(), "", []string{}, []string{})
		assert.Error(t, err)
	})
	t.Run("types", func(t *testing.T) {
		result, err := testRW.QueryInstancesByHost(context.TODO(), "testHostId", []string{"CDC", "PD"}, []string{})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
	t.Run("status", func(t *testing.T) {
		result, err := testRW.QueryInstancesByHost(context.TODO(), "testHostId", []string{"TiKV"}, []string{string(constants.ClusterInstanceRunning)})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result))
	})
}