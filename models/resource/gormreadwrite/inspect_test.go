/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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

package gormreadwrite

import (
	"context"
	"testing"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/structs"
	resource_structs "github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/management/structs"
	"github.com/stretchr/testify/assert"
)

func Test_CheckCpuAllocated(t *testing.T) {
	id1, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "474.111.111.117",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)
	id2, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host2", "474.111.111.118",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 3)
	id3, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host3", "474.111.111.119",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 3)

	loc := structs.Location{}
	loc.Region = "Test_Region1"
	loc.Zone = "Test_Zone1"
	filter1 := resource_structs.Filter{}
	filter1.DiskType = string(constants.SSD)
	filter1.Purpose = string(constants.PurposeCompute)
	filter1.Arch = string(constants.ArchX8664)
	require := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   resource_structs.RandomRack,
		Require:    *require,
		Count:      3,
	})

	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	_, err := GormRW.AllocResources(context.TODO(), &batchReq)
	assert.Nil(t, err)

	resultFromHostTable, resultFromUsedTable, resultFromInstTable, err := GormRW.GetUsedCpuCores(context.TODO(), []string{id1[0], id2[0], id3[0]})
	assert.Nil(t, err)
	assert.Equal(t, 12, resultFromHostTable[id1[0]])
	assert.Equal(t, 12, resultFromUsedTable[id2[0]])
	assert.Equal(t, 0, resultFromInstTable[id3[0]])

	err = recycleClusterResources("TestCluster1")
	assert.Nil(t, err)
	resultFromHostTable, resultFromUsedTable, resultFromInstTable, err = GormRW.GetUsedMemory(context.TODO(), []string{id1[0], id2[0], id3[0]})
	assert.Nil(t, err)
	assert.Equal(t, 0, resultFromHostTable[id1[0]])
	assert.Equal(t, 0, resultFromUsedTable[id2[0]])
	assert.Equal(t, 0, resultFromInstTable[id3[0]])
	err = GormRW.Delete(context.TODO(), id1)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id2)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id3)
	assert.Nil(t, err)
}

func Test_CheckMemAllocated(t *testing.T) {
	id1, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "474.111.111.117",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)
	id2, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host2", "474.111.111.118",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 3)
	id3, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host3", "474.111.111.119",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 3)

	loc := structs.Location{}
	loc.Region = "Test_Region1"
	loc.Zone = "Test_Zone1"
	filter1 := resource_structs.Filter{}
	filter1.DiskType = string(constants.SSD)
	filter1.Purpose = string(constants.PurposeCompute)
	filter1.Arch = string(constants.ArchX8664)
	require := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   resource_structs.RandomRack,
		Require:    *require,
		Count:      3,
	})

	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	_, err := GormRW.AllocResources(context.TODO(), &batchReq)
	assert.Nil(t, err)

	resultFromHostTable, resultFromUsedTable, resultFromInstTable, err := GormRW.GetUsedMemory(context.TODO(), []string{id1[0], id2[0], id3[0]})
	assert.Nil(t, err)
	assert.Equal(t, 24, resultFromHostTable[id1[0]])
	assert.Equal(t, 24, resultFromUsedTable[id2[0]])
	assert.Equal(t, 0, resultFromInstTable[id3[0]])

	err = recycleClusterResources("TestCluster1")
	assert.Nil(t, err)
	resultFromHostTable, resultFromUsedTable, resultFromInstTable, err = GormRW.GetUsedMemory(context.TODO(), []string{id1[0], id2[0], id3[0]})
	assert.Nil(t, err)
	assert.Equal(t, 0, resultFromHostTable[id1[0]])
	assert.Equal(t, 0, resultFromUsedTable[id2[0]])
	assert.Equal(t, 0, resultFromInstTable[id3[0]])
	err = GormRW.Delete(context.TODO(), id1)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id2)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id3)
	assert.Nil(t, err)
}

func Test_CheckDiskAllocated(t *testing.T) {
	id1, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "474.111.111.117",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 17, 64, 3)
	id2, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host2", "474.111.111.118",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 16, 64, 3)
	id3, _ := createTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host3", "474.111.111.119",
		string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.SSD), 15, 64, 3)

	loc := structs.Location{}
	loc.Region = "Test_Region1"
	loc.Zone = "Test_Zone1"
	filter1 := resource_structs.Filter{}
	filter1.DiskType = string(constants.SSD)
	filter1.Purpose = string(constants.PurposeCompute)
	filter1.Arch = string(constants.ArchX8664)
	require := newRequirementForRequest(4, 8, true, 256, string(constants.SSD), 10000, 10015, 5)

	var test_req resource_structs.AllocReq
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, resource_structs.AllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   resource_structs.RandomRack,
		Require:    *require,
		Count:      3,
	})

	var batchReq resource_structs.BatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, test_req)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	_, err := GormRW.AllocResources(context.TODO(), &batchReq)
	assert.Nil(t, err)

	resultFromHostTable, resultFromUsedTable, resultFromInstTable, err := GormRW.GetUsedDisks(context.TODO(), []string{id1[0], id2[0], id3[0]})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(*resultFromHostTable[id1[0]]))
	assert.Equal(t, 3, len(*resultFromUsedTable[id2[0]]))
	assert.Nil(t, resultFromInstTable[id3[0]])

	err = recycleClusterResources("TestCluster1")
	assert.Nil(t, err)
	resultFromHostTable, resultFromUsedTable, resultFromInstTable, err = GormRW.GetUsedDisks(context.TODO(), []string{id1[0], id2[0], id3[0]})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resultFromHostTable))
	assert.Equal(t, 0, len(resultFromUsedTable))
	assert.Equal(t, 0, len(resultFromInstTable))
	err = GormRW.Delete(context.TODO(), id1)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id2)
	assert.Nil(t, err)
	err = GormRW.Delete(context.TODO(), id3)
	assert.Nil(t, err)
}
