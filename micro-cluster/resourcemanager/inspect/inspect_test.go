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

package hostInspector

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/common/structs"
	resource_models "github.com/pingcap-inc/tiunimanager/models/resource"
	mock_resource "github.com/pingcap-inc/tiunimanager/test/mockmodels/mockresource"
	"github.com/stretchr/testify/assert"
)

func mockHostInspector(rw resource_models.ReaderWriter) *HostInspect {
	hostInspector := new(HostInspect)
	hostInspector.SetResourceReaderWriter(rw)
	return hostInspector
}
func Test_generateCheckInt32Result_Normal_Succeed(t *testing.T) {
	hostIds := []string{"test_host1", "test_host2", "test_host3"}
	map1 := map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}
	map2 := map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}

	hostInspector := mockHostInspector(nil)
	result := hostInspector.generateCheckInt32Result(hostIds, map1, map2)
	assert.Equal(t, 3, len(result))
	for i := range result {
		assert.Equal(t, true, result[i].Valid)
	}
}

func Test_generateCheckInt32Result_Normal_Failed(t *testing.T) {
	hostIds := []string{"test_host1", "test_host2", "test_host3"}
	map1 := map[string]int{"test_host1": 8, "test_host2": 14, "test_host3": 17}
	map2 := map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}

	hostInspector := mockHostInspector(nil)
	result := hostInspector.generateCheckInt32Result(hostIds, map1, map2)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, true, result["test_host1"].Valid)
	assert.Equal(t, false, result["test_host2"].Valid)
	assert.Equal(t, int32(14), result["test_host2"].ExpectedValue)
	assert.Equal(t, int32(13), result["test_host2"].RealValue)
	assert.Equal(t, true, result["test_host3"].Valid)
}

func Test_generateCheckInt32Result_Both_Nil(t *testing.T) {
	hostIds := []string{"test_host1", "test_host2", "test_host3"}
	map1 := map[string]int{"test_host1": 8, "test_host2": 14}
	map2 := map[string]int{"test_host1": 8, "test_host2": 14}

	hostInspector := mockHostInspector(nil)
	result := hostInspector.generateCheckInt32Result(hostIds, map1, map2)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, true, result["test_host1"].Valid)
	assert.Equal(t, true, result["test_host2"].Valid)
	assert.Equal(t, true, result["test_host3"].Valid)
	assert.Equal(t, int32(0), result["test_host3"].ExpectedValue)
	assert.Equal(t, int32(0), result["test_host3"].RealValue)
}

func Test_generateCheckInt32Result_Either_Nil(t *testing.T) {
	hostIds := []string{"test_host1", "test_host2", "test_host3"}
	map1 := map[string]int{"test_host1": 8, "test_host2": 14}
	map2 := map[string]int{"test_host1": 8, "test_host2": 14, "test_host3": 17}

	hostInspector := mockHostInspector(nil)
	result := hostInspector.generateCheckInt32Result(hostIds, map1, map2)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, true, result["test_host1"].Valid)
	assert.Equal(t, true, result["test_host2"].Valid)
	assert.Equal(t, false, result["test_host3"].Valid)
	assert.Equal(t, int32(0), result["test_host3"].ExpectedValue)
	assert.Equal(t, int32(17), result["test_host3"].RealValue)
}
func Test_generateCheckStatusResult_Normal_Succeed(t *testing.T) {
	hostIds := []string{"test_host1", "test_host2", "test_host3"}
	map1 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
		"test_host2": {"disk4", "disk5", "disk6"},
		"test_host3": {"disk7", "disk8", "disk9"},
	}
	map2 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
		"test_host2": {"disk4", "disk5", "disk6"},
		"test_host3": {"disk7", "disk8", "disk9"},
	}

	hostInspector := mockHostInspector(nil)
	result := hostInspector.generateCheckStatusResult(hostIds, map1, map2)
	assert.Equal(t, 3, len(result))
	for hostId := range result {
		assert.Equal(t, 3, len(result[hostId]))
		for diskId := range result[hostId] {
			assert.Equal(t, true, result[hostId][diskId].Valid)
		}
	}
}

func Test_generateCheckStatusResult(t *testing.T) {
	hostIds := []string{"test_host1", "test_host2", "test_host3"}
	map1 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3", "disk11"},
		"test_host2": {"disk4", "disk5", "disk6"},
		"test_host3": {"disk7", "disk8", "disk9"},
	}
	map2 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
		"test_host2": {"disk4", "disk5", "disk6"},
		"test_host3": {"disk7", "disk8", "disk9", "disk12"},
	}

	hostInspector := mockHostInspector(nil)
	result := hostInspector.generateCheckStatusResult(hostIds, map1, map2)
	assert.Equal(t, 3, len(result))
	for hostId := range result {
		for diskId := range result[hostId] {
			if hostId == "test_host1" && diskId == "disk11" {
				assert.Equal(t, false, result[hostId][diskId].Valid)
				assert.Equal(t, string(constants.DiskExhaust), result[hostId][diskId].ExpectedValue)
				assert.Equal(t, "", result[hostId][diskId].RealValue)
			} else if hostId == "test_host3" && diskId == "disk12" {
				assert.Equal(t, false, result[hostId][diskId].Valid)
				assert.Equal(t, "", result[hostId][diskId].ExpectedValue)
				assert.Equal(t, string(constants.DiskExhaust), result[hostId][diskId].RealValue)
			} else {
				assert.Equal(t, true, result[hostId][diskId].Valid)
			}
		}
	}
}

func Test_CheckCpuAllocated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetUsedCpuCores(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) (map[string]int, map[string]int, map[string]int, error) {
		return map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, nil
	})

	hostInspector := mockHostInspector(mockClient)
	hosts := []structs.HostInfo{{ID: "test_host1"}, {ID: "test_host2"}, {ID: "test_host3"}}
	result, err := hostInspector.CheckCpuAllocated(context.TODO(), hosts)
	assert.Nil(t, err)
	for k := range result {
		assert.Equal(t, true, result[k].Valid)
	}
}

func Test_CheckCpuAllocated_Resource_Internal_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetUsedCpuCores(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) (map[string]int, map[string]int, map[string]int, error) {
		return map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 9, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, nil
	})

	hostInspector := mockHostInspector(mockClient)
	hosts := []structs.HostInfo{{ID: "test_host1"}, {ID: "test_host2"}, {ID: "test_host3"}}
	result, err := hostInspector.CheckCpuAllocated(context.TODO(), hosts)
	assert.Nil(t, err)
	for k := range result {
		if k == "test_host1" {
			assert.Equal(t, false, result[k].Valid)
		} else {
			assert.Equal(t, true, result[k].Valid)
		}
	}
}

func Test_CheckCpuAllocated_Mismatch_Resource_Cluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetUsedCpuCores(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) (map[string]int, map[string]int, map[string]int, error) {
		return map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 12, "test_host3": 17}, nil
	})

	hostInspector := mockHostInspector(mockClient)
	hosts := []structs.HostInfo{{ID: "test_host1"}, {ID: "test_host2"}, {ID: "test_host3"}}
	result, err := hostInspector.CheckCpuAllocated(context.TODO(), hosts)
	assert.Nil(t, err)
	for k := range result {
		if k == "test_host2" {
			assert.Equal(t, false, result[k].Valid)
		} else {
			assert.Equal(t, true, result[k].Valid)
		}
	}
}

func Test_CheckMemAllocated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetUsedMemory(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) (map[string]int, map[string]int, map[string]int, error) {
		return map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, nil
	})

	hostInspector := mockHostInspector(mockClient)
	hosts := []structs.HostInfo{{ID: "test_host1"}, {ID: "test_host2"}, {ID: "test_host3"}}
	result, err := hostInspector.CheckMemAllocated(context.TODO(), hosts)
	assert.Nil(t, err)
	for k := range result {
		assert.Equal(t, true, result[k].Valid)
	}
}

func Test_CheckMemAllocated_Resource_Internal_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetUsedMemory(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) (map[string]int, map[string]int, map[string]int, error) {
		return map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 9, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, nil
	})

	hostInspector := mockHostInspector(mockClient)
	hosts := []structs.HostInfo{{ID: "test_host1"}, {ID: "test_host2"}, {ID: "test_host3"}}
	result, err := hostInspector.CheckMemAllocated(context.TODO(), hosts)
	assert.Nil(t, err)
	for k := range result {
		if k == "test_host1" {
			assert.Equal(t, false, result[k].Valid)
		} else {
			assert.Equal(t, true, result[k].Valid)
		}
	}
}

func Test_CheckMemAllocated_Mismatch_Resource_Cluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetUsedMemory(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) (map[string]int, map[string]int, map[string]int, error) {
		return map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 13, "test_host3": 17}, map[string]int{"test_host1": 8, "test_host2": 12, "test_host3": 17}, nil
	})

	hostInspector := mockHostInspector(mockClient)
	hosts := []structs.HostInfo{{ID: "test_host1"}, {ID: "test_host2"}, {ID: "test_host3"}}
	result, err := hostInspector.CheckMemAllocated(context.TODO(), hosts)
	assert.Nil(t, err)
	for k := range result {
		if k == "test_host2" {
			assert.Equal(t, false, result[k].Valid)
		} else {
			assert.Equal(t, true, result[k].Valid)
		}
	}
}

func Test_CheckDiskAllocated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetUsedDisks(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) (map[string]*[]string, map[string]*[]string, map[string]*[]string, error) {
		disks1 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
			"test_host2": {"disk4", "disk5", "disk6"},
			"test_host3": {"disk7", "disk8", "disk9"},
		}
		disks2 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
			"test_host2": {"disk4", "disk5", "disk6"},
			"test_host3": {"disk7", "disk8", "disk9"},
		}
		disks3 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
			"test_host2": {"disk4", "disk5", "disk6"},
			"test_host3": {"disk7", "disk8", "disk9"},
		}
		return disks1, disks2, disks3, nil
	})

	hostInspector := mockHostInspector(mockClient)
	hosts := []structs.HostInfo{{ID: "test_host1"}, {ID: "test_host2"}, {ID: "test_host3"}}
	result, err := hostInspector.CheckDiskAllocated(context.TODO(), hosts)
	assert.Nil(t, err)
	for k := range result {
		for disk := range result[k] {
			assert.Equal(t, true, result[k][disk].Valid)
		}
	}
}

func Test_CheckDiskAllocated_Mismatch_Resource_Internal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetUsedDisks(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) (map[string]*[]string, map[string]*[]string, map[string]*[]string, error) {
		disks1 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
			"test_host2": {"disk4", "disk5", "disk6"},
			"test_host3": {"disk7", "disk8", "disk9"},
		}
		disks2 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
			"test_host2": {"disk4", "disk5", "disk6", "disk10"},
			"test_host3": {"disk7", "disk8", "disk9"},
		}
		disks3 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
			"test_host2": {"disk4", "disk5", "disk6"},
			"test_host3": {"disk7", "disk8", "disk9"},
		}
		return disks1, disks2, disks3, nil
	})

	hostInspector := mockHostInspector(mockClient)
	hosts := []structs.HostInfo{{ID: "test_host1"}, {ID: "test_host2"}, {ID: "test_host3"}}
	result, err := hostInspector.CheckDiskAllocated(context.TODO(), hosts)
	assert.Nil(t, err)
	for k := range result {
		for disk := range result[k] {
			if k == "test_host2" && disk == "disk10" {
				assert.Equal(t, false, result[k][disk].Valid)
			} else {
				assert.Equal(t, true, result[k][disk].Valid)
			}
		}
	}
}

func Test_CheckDiskAllocated_Mismatch_Resource_Cluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetUsedDisks(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) (map[string]*[]string, map[string]*[]string, map[string]*[]string, error) {
		disks1 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
			"test_host2": {"disk4", "disk5", "disk6"},
			"test_host3": {"disk7", "disk8", "disk9"},
		}
		disks2 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
			"test_host2": {"disk4", "disk5", "disk6"},
			"test_host3": {"disk7", "disk8", "disk9"},
		}
		disks3 := map[string]*[]string{"test_host1": {"disk1", "disk2", "disk3"},
			"test_host2": {"disk4", "disk5", "disk6"},
			"test_host3": {"disk7", "disk8", "disk9", "disk10"},
		}
		return disks1, disks2, disks3, nil
	})

	hostInspector := mockHostInspector(mockClient)
	hosts := []structs.HostInfo{{ID: "test_host1"}, {ID: "test_host2"}, {ID: "test_host3"}}
	result, err := hostInspector.CheckDiskAllocated(context.TODO(), hosts)
	assert.Nil(t, err)
	for k := range result {
		for disk := range result[k] {
			if k == "test_host3" && disk == "disk10" {
				assert.Equal(t, false, result[k][disk].Valid)
			} else {
				assert.Equal(t, true, result[k][disk].Valid)
			}
		}
	}
}
