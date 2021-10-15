
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package models

import (
	"github.com/pingcap-inc/tiem/library/util/slice"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDAOClusterManager_AddClusterComponentInstance(t *testing.T) {
	dao := Dao.ClusterManager()
	t.Run("normal", func(t *testing.T) {
		data := []*ComponentInstance{
			{
				Entity: Entity {
					TenantId: "1111",
				},
				ClusterId: "cluster11",
				ComponentType: "TiKV",
				Role: "Leader",
				Spec: "4C8G",
				PortInfo: "{}",
				DiskId: "2222",
				Version: "v5.1.1",
				HostId: "Host11",
				AllocRequestId: "121212",
			},
			{
				Entity: Entity {
					TenantId: "1111",
				},
				ClusterId: "cluster11",
				ComponentType: "TiKV",
				Role: "Flower",
				Spec: "8C16G",
				PortInfo: "{}",
				DiskId: "3333",
				Version: "v5.1.1",
				HostId: "Host11",
				AllocRequestId: "121212",
			},
		}
		dataWithId, err := dao.AddClusterComponentInstance("cluster11", data)
		assert.NoError(t, err)
		assert.NotNil(t, dataWithId)
		assert.NotEmpty(t, dataWithId[1].ID)
		dao.Db().Delete(dataWithId)
	})

	t.Run("empty", func(t *testing.T) {
		data := []*ComponentInstance{}
		_ , err := dao.AddClusterComponentInstance("whatever", data)
		assert.Error(t, err)
	})

	t.Run("without clusterId", func(t *testing.T) {
		data := []*ComponentInstance{
			{
				Entity: Entity {
					TenantId: "1111",
				},
				ClusterId: "whatever",
				ComponentType: "TiKV",
				Role: "Leader",
				Spec: "4C8G",
				PortInfo: "{}",
				DiskId: "2222",
				Version: "v5.1.1",
				HostId: "111111",
				AllocRequestId: "121212",
			},
			{
				Entity: Entity {
					TenantId: "1111",
				},
				ClusterId: "whatever",
				ComponentType: "TiKV",
				Role: "Flower",
				Spec: "8C16G",
				PortInfo: "{}",
				DiskId: "3333",
				Version: "v5.1.1",
				HostId: "111111",
				AllocRequestId: "121212",
			},
		}
		_ , err := dao.AddClusterComponentInstance("", data)
		assert.Error(t, err)
	})
}

func TestDAOClusterManager_ListComponentInstances(t *testing.T) {
	dao := Dao.ClusterManager()
	data1 := []*ComponentInstance{
		{
			Entity: Entity {
				TenantId: "1111",
			},
			ClusterId: "cluster21",
			ComponentType: "TiKV",
			Role: "Leader",
			Spec: "4C8G",
			PortInfo: "{}",
			DiskId: "1111",
			Version: "v5.1.1",
			HostId: "HostId21",
			AllocRequestId: "121212",
		},
		{
			Entity: Entity {
				TenantId: "1111",
			},
			ClusterId: "cluster21",
			ComponentType: "TiKV",
			Role: "Flower",
			Spec: "8C16G",
			PortInfo: "{}",
			DiskId: "2222",
			Version: "v5.1.1",
			HostId: "HostId22",
			AllocRequestId: "121212",
		},
	}
	ids , err := dao.AddClusterComponentInstance("cluster21", data1)
	assert.NoError(t, err)
	defer dao.Db().Delete(ids)

	data2 := []*ComponentInstance{
		{
			Entity: Entity {
				TenantId: "1111",
			},
			ClusterId: "cluster22",
			ComponentType: "TiKV",
			Role: "Flower",
			Spec: "8C16G",
			PortInfo: "{}",
			DiskId: "3333",
			Version: "v5.1.1",
			HostId: "HostId21",
			AllocRequestId: "121212",
		},
	}
	ids2 , err := dao.AddClusterComponentInstance("cluster22", data2)
	assert.NoError(t, err)
	defer dao.Db().Delete(ids2)

	t.Run("normal", func(t *testing.T) {
		gotComponentInstances, err := dao.ListComponentInstances("cluster21")
		assert.NoError(t, err)
		assert.NotNil(t, gotComponentInstances)
		assert.Equal(t, 2, len(gotComponentInstances))
		assert.True(t, slice.AllOf(gotComponentInstances, func(i int) bool {
			return gotComponentInstances[i].ClusterId == "cluster21"
		}))
	})
	t.Run("without cluster id", func(t *testing.T) {
		_, err := dao.ListComponentInstances("")
		assert.Error(t, err)
	})
}

func TestDAOClusterManager_ListComponentInstancesByHost(t *testing.T) {
	dao := Dao.ClusterManager()
	data1 := []*ComponentInstance{
		{
			Entity: Entity {
				TenantId: "1111",
			},
			ClusterId: "cluster31",
			ComponentType: "TiKV",
			Role: "Leader",
			Spec: "4C8G",
			PortInfo: "{}",
			DiskId: "1111",
			Version: "v5.1.1",
			HostId: "HostId31",
			AllocRequestId: "121212",
		},
		{
			Entity: Entity {
				TenantId: "1111",
			},
			ClusterId: "cluster31",
			ComponentType: "TiKV",
			Role: "Flower",
			Spec: "8C16G",
			PortInfo: "{}",
			DiskId: "2222",
			Version: "v5.1.1",
			HostId: "HostId32",
			AllocRequestId: "121212",
		},
	}
	ids , err := dao.AddClusterComponentInstance("cluster31", data1)
	assert.NoError(t, err)
	defer dao.Db().Delete(ids)

	data2 := []*ComponentInstance{
		{
			Entity: Entity {
				TenantId: "1111",
			},
			ClusterId: "cluster32",
			ComponentType: "TiKV",
			Role: "Flower",
			Spec: "8C16G",
			PortInfo: "{}",
			DiskId: "3333",
			Version: "v5.1.1",
			HostId: "HostId31",
			AllocRequestId: "121212",
		},
	}
	ids2 , err := dao.AddClusterComponentInstance("cluster32", data2)
	assert.NoError(t, err)
	defer dao.Db().Delete(ids2)

	t.Run("normal", func(t *testing.T) {
		gotComponentInstances, err := dao.ListComponentInstancesByHost("HostId31")
		assert.NoError(t, err)
		assert.NotNil(t, gotComponentInstances)
		assert.Equal(t, 2, len(gotComponentInstances))
		assert.True(t, slice.AllOf(gotComponentInstances, func(i int) bool {
			return gotComponentInstances[i].HostId == "HostId31"
		}))
	})
	t.Run("without host id", func(t *testing.T) {
		_, err := dao.ListComponentInstancesByHost("")
		assert.Error(t, err)
	})
}