/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package importexport

import (
	"context"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
	"github.com/pingcap/tiunimanager/models/cluster/management"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDataImportConfig_case1(t *testing.T) {
	instanceMap := make(map[string][]*management.ClusterInstance)
	tidb := make([]*management.ClusterInstance, 1)
	tidb[0] = &management.ClusterInstance{}
	tidb[0].Status = string(constants.ClusterInstanceRunning)
	tidb[0].HostIP = []string{"127.0.0.1", "127.0.0.2"}
	tidb[0].Ports = []int32{4000, 4001}
	instanceMap[string(constants.ComponentIDTiDB)] = tidb

	pd := make([]*management.ClusterInstance, 1)
	pd[0] = &management.ClusterInstance{}
	pd[0].Status = string(constants.ClusterInstanceRunning)
	pd[0].HostIP = []string{"127.0.0.1", "127.0.0.2"}
	pd[0].Ports = []int32{4000, 4001}
	instanceMap[string(constants.ComponentIDPD)] = pd

	meta := &meta.ClusterMeta{
		Instances: instanceMap,
		Cluster:   &management.Cluster{},
	}
	info := &importInfo{
		ClusterId:   "test-cls",
		UserName:    "root",
		Password:    "root",
		FilePath:    "/home/em/import",
		RecordId:    "",
		StorageType: "s3",
		ConfigPath:  "/home/em/import",
	}
	config := NewDataImportConfig(context.TODO(), meta, info)
	assert.NotNil(t, config)
}

func TestNewDataImportConfig_case2(t *testing.T) {
	instanceMap := make(map[string][]*management.ClusterInstance)
	tidb := make([]*management.ClusterInstance, 1)
	tidb[0] = &management.ClusterInstance{}
	tidb[0].Status = string(constants.ClusterInstanceRunning)
	tidb[0].HostIP = []string{"127.0.0.1", "127.0.0.2"}
	tidb[0].Ports = []int32{4000, 4001}
	instanceMap[string(constants.ComponentIDTiDB)] = tidb

	pd := make([]*management.ClusterInstance, 1)
	pd[0] = &management.ClusterInstance{}
	pd[0].Status = string(constants.ClusterInstanceRunning)
	pd[0].HostIP = []string{"127.0.0.1", "127.0.0.2"}
	pd[0].Ports = []int32{4000, 4001}
	instanceMap[string(constants.ComponentIDPD)] = pd

	meta := &meta.ClusterMeta{
		Instances: instanceMap,
		Cluster:   nil,
	}
	info := &importInfo{
		ClusterId:   "test-cls",
		UserName:    "root",
		Password:    "root",
		FilePath:    "/home/em/import",
		RecordId:    "",
		StorageType: "s3",
		ConfigPath:  "/home/em/import",
	}
	config := NewDataImportConfig(context.TODO(), meta, info)
	assert.Nil(t, config)
}

func TestNewDataImportConfig_case3(t *testing.T) {
	instanceMap := make(map[string][]*management.ClusterInstance)
	tidb := make([]*management.ClusterInstance, 0)
	instanceMap[string(constants.ComponentIDTiDB)] = tidb

	pd := make([]*management.ClusterInstance, 1)
	pd[0] = &management.ClusterInstance{}
	pd[0].Status = string(constants.ClusterInstanceRunning)
	pd[0].HostIP = []string{"127.0.0.1", "127.0.0.2"}
	pd[0].Ports = []int32{4000, 4001}
	instanceMap[string(constants.ComponentIDPD)] = pd

	meta := &meta.ClusterMeta{
		Instances: instanceMap,
		Cluster:   &management.Cluster{},
	}
	info := &importInfo{
		ClusterId:   "test-cls",
		UserName:    "root",
		Password:    "root",
		FilePath:    "/home/em/import",
		RecordId:    "",
		StorageType: "s3",
		ConfigPath:  "/home/em/import",
	}
	config := NewDataImportConfig(context.TODO(), meta, info)
	assert.Nil(t, config)
}

func TestNewDataImportConfig_case4(t *testing.T) {
	instanceMap := make(map[string][]*management.ClusterInstance)
	tidb := make([]*management.ClusterInstance, 1)
	tidb[0] = &management.ClusterInstance{}
	tidb[0].Status = string(constants.ClusterInstanceRunning)
	tidb[0].HostIP = []string{"127.0.0.1", "127.0.0.2"}
	tidb[0].Ports = []int32{4000, 4001}
	instanceMap[string(constants.ComponentIDTiDB)] = tidb

	pd := make([]*management.ClusterInstance, 0)
	instanceMap[string(constants.ComponentIDPD)] = pd

	meta := &meta.ClusterMeta{
		Instances: instanceMap,
		Cluster:   &management.Cluster{},
	}
	info := &importInfo{
		ClusterId:   "test-cls",
		UserName:    "root",
		Password:    "root",
		FilePath:    "/home/em/import",
		RecordId:    "",
		StorageType: "s3",
		ConfigPath:  "/home/em/import",
	}
	config := NewDataImportConfig(context.TODO(), meta, info)
	assert.Nil(t, config)
}
