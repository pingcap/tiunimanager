/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package hostprovider

import (
	"context"
	"fmt"
	"testing"

	"github.com/pingcap/tiunimanager/models/platform/product"

	"github.com/golang/mock/gomock"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/models"
	cluster_rw "github.com/pingcap/tiunimanager/models/cluster/management"
	resource_models "github.com/pingcap/tiunimanager/models/resource"
	resourcepool "github.com/pingcap/tiunimanager/models/resource/resourcepool"
	mock_product "github.com/pingcap/tiunimanager/test/mockmodels"
	mock_cluster "github.com/pingcap/tiunimanager/test/mockmodels/mockclustermanagement"
	mock_resource "github.com/pingcap/tiunimanager/test/mockmodels/mockresource"
	"github.com/stretchr/testify/assert"
)

func mockFileHostProvider(rw resource_models.ReaderWriter) *FileHostProvider {
	hostProvider := new(FileHostProvider)
	hostProvider.SetResourceReaderWriter(rw)
	return hostProvider
}

func genHostInfo(hostName string) *structs.HostInfo {
	host := structs.HostInfo{
		IP:       "192.168.56.11",
		HostName: hostName,
		OS:       "Centos",
		Kernel:   "3.10",
		Region:   "TEST_REGION",
		AZ:       "TEST_AZ",
		Rack:     "TEST_RACK",
		Status:   string(constants.HostOnline),
		Nic:      "10GE",
		Purpose:  "Compute",
	}
	host.Disks = append(host.Disks, structs.DiskInfo{
		Name:     "sda",
		Path:     "/",
		Status:   string(constants.DiskAvailable),
		Capacity: 512,
	})
	host.Disks = append(host.Disks, structs.DiskInfo{
		Name:     "sdb",
		Path:     "/mnt/sdb",
		Status:   string(constants.DiskAvailable),
		Capacity: 1024,
	})
	return &host
}

func genHostRspFromDB(hostId, hostName string) *resourcepool.Host {
	host := resourcepool.Host{
		ID:       hostId,
		HostName: hostName,
		IP:       "192.168.56.11",
		OS:       "Centos",
		Kernel:   "3.10",
		Region:   "TEST_REGION",
		AZ:       "TEST_REGION,TEST_AZ",
		Rack:     "TEST_REGION,TEST_AZ,TEST_RACK",
		Status:   string(constants.HostOnline),
		Nic:      "10GE",
		Purpose:  "Compute",
	}
	host.Disks = append(host.Disks, resourcepool.Disk{
		Name:     "sda",
		Path:     "/",
		Status:   string(constants.DiskAvailable),
		Capacity: 512,
	})
	host.Disks = append(host.Disks, resourcepool.Disk{
		Name:     "sdb",
		Path:     "/mnt/sdb",
		Status:   string(constants.DiskAvailable),
		Capacity: 1024,
	})
	return &host
}

func Test_ImportHosts_Succeed(t *testing.T) {
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hosts []resourcepool.Host) ([]string, error) {
		if hosts[0].HostName == "TEST_HOST1" {
			var hostIds []string
			hostIds = append(hostIds, fake_hostId)
			return hostIds, nil
		} else {
			return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "BadRequest")
		}
	})

	hostprovider := mockFileHostProvider(mockClient)

	var hosts []structs.HostInfo
	host := genHostInfo("TEST_HOST1")
	hosts = append(hosts, *host)

	hostIds, err := hostprovider.ImportHosts(context.TODO(), hosts)
	assert.Nil(t, err)

	assert.Equal(t, fake_hostId, hostIds[0])
}

func Test_ImportHosts_Failed(t *testing.T) {
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hosts []resourcepool.Host) ([]string, error) {
		if hosts[0].HostName == "TEST_HOST1" {
			var hostIds []string
			hostIds = append(hostIds, fake_hostId)
			return hostIds, nil
		} else {
			return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := mockFileHostProvider(mockClient)

	var hosts []structs.HostInfo
	host := genHostInfo("TEST_HOST2")
	hosts = append(hosts, *host)
	_, err := hostprovider.ImportHosts(context.TODO(), hosts)
	assert.NotNil(t, err)
	tiunimanagerErr, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIUNIMANAGER_PARAMETER_INVALID, tiunimanagerErr.GetCode())

}

func Test_QueryHosts_Succeed(t *testing.T) {
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fake_hostname := "fake_host_name"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, location *structs.Location, filter *structs.HostFilter, offset, limit int) (hosts []resourcepool.Host, total int64, err error) {
		assert.Equal(t, 20, offset)
		assert.Equal(t, 10, limit)
		if filter.HostID == fake_hostId {
			dbhost := genHostRspFromDB(fake_hostId, fake_hostname)
			hosts = append(hosts, *dbhost)
			return hosts, 1, nil
		} else {
			return nil, 0, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "BadRequest")
		}
	})

	models.MockDB()
	rw := mock_cluster.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(rw)
	rw.EXPECT().QueryHostInstances(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) ([]cluster_rw.HostInstanceItem, error) {
		items := []cluster_rw.HostInstanceItem{
			{HostID: fake_hostId, ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "AlertManger"},
			{HostID: fake_hostId, ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "Grafana"},
			{HostID: fake_hostId, ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "PD"},
			{HostID: fake_hostId, ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "Prometheus"},
			{HostID: fake_hostId, ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "AlertManger"},
			{HostID: fake_hostId, ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "Grafana"},
			{HostID: fake_hostId, ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "PD"},
			{HostID: fake_hostId, ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "Prometheus"},
		}
		return items, nil
	})

	hostprovider := mockFileHostProvider(mockClient)

	filter := &structs.HostFilter{
		HostID: fake_hostId,
	}
	page := &structs.PageRequest{
		Page:     3,
		PageSize: 10,
	}

	hosts, total, err := hostprovider.QueryHosts(context.TODO(), &structs.Location{}, filter, page)
	assert.Nil(t, err)
	assert.Equal(t, 1, int(total))

	assert.Equal(t, fake_hostname, hosts[0].HostName)
	assert.Equal(t, "TEST_REGION", hosts[0].Region)
	assert.Equal(t, "TEST_AZ", hosts[0].AZ)
	assert.Equal(t, "TEST_RACK", hosts[0].Rack)
	assert.Equal(t, 2, len(hosts[0].Instances))
	assert.Equal(t, 4, len(hosts[0].Instances["Jt_r1RnfT3i0O7YWtMhBzw"]))
	assert.Equal(t, 4, len(hosts[0].Instances["THwF-s3hTL-SZbZsytBMTw"]))
}

func Test_DeleteHosts_Succeed(t *testing.T) {
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "aaaa-bbbb-cccc-dddd"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().Delete(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string) error {
		if hostIds[0] == fake_hostId1 && hostIds[1] == fake_hostId2 {
			return nil
		} else {
			return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := mockFileHostProvider(mockClient)

	var hostIds []string
	hostIds = append(hostIds, fake_hostId1)
	hostIds = append(hostIds, fake_hostId2)

	err := hostprovider.DeleteHosts(context.TODO(), hostIds)
	assert.Nil(t, err)
}

func Test_UpdateHostReserved_Succeed(t *testing.T) {
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "aaaa-bbbb-cccc-dddd"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	host1 := genHostInfo("TEST_HOST1")
	host2 := genHostInfo("TEST_HOST2")
	mockClient.EXPECT().UpdateHostReserved(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string, reserved bool) error {
		if hostIds[0] == fake_hostId1 && hostIds[1] == fake_hostId2 {
			host1.Reserved = reserved
			host2.Reserved = reserved
			return nil
		} else {
			return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := mockFileHostProvider(mockClient)

	var hostIds []string
	hostIds = append(hostIds, fake_hostId1)
	hostIds = append(hostIds, fake_hostId2)

	err := hostprovider.UpdateHostReserved(context.TODO(), hostIds, true)
	assert.Nil(t, err)
	assert.Equal(t, true, host1.Reserved)
	assert.Equal(t, true, host2.Reserved)
}

func Test_UpdateHostStatus_Succeed(t *testing.T) {
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "aaaa-bbbb-cccc-dddd"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	host1 := genHostInfo("TEST_HOST1")
	host2 := genHostInfo("TEST_HOST2")
	mockClient.EXPECT().UpdateHostStatus(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostIds []string, status string) error {
		if hostIds[0] == fake_hostId1 && hostIds[1] == fake_hostId2 {
			host1.Status = status
			host2.Status = status
			return nil
		} else {
			return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := mockFileHostProvider(mockClient)

	var hostIds []string
	hostIds = append(hostIds, fake_hostId1)
	hostIds = append(hostIds, fake_hostId2)

	status := "Offline"
	err := hostprovider.UpdateHostStatus(context.TODO(), hostIds, status)
	assert.Nil(t, err)
	assert.Equal(t, status, host1.Status)
	assert.Equal(t, status, host2.Status)
}

func Test_GetHierarchy_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetHostItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, filter *structs.HostFilter, level, depth int32) (items []resource_models.HostItem, err error) {
		if filter.Arch == string(constants.ArchX8664) && level == 1 && depth == 3 {
			item1 := resource_models.HostItem{
				Region: "TEST_Region1",
				Az:     "TEST_Region1,TEST_Zone1",
				Rack:   "TEST_Region1,TEST_Zone1,TEST_Rack1",
				Ip:     "192.168.9.111",
				Name:   "HostName1",
			}
			item2 := resource_models.HostItem{
				Region: "TEST_Region1",
				Az:     "TEST_Region1,TEST_Zone1",
				Rack:   "TEST_Region1,TEST_Zone1,TEST_Rack2",
				Ip:     "192.168.9.112",
				Name:   "HostName2",
			}
			item3 := resource_models.HostItem{
				Region: "TEST_Region1",
				Az:     "TEST_Region1,TEST_Zone2",
				Rack:   "TEST_Region1,TEST_Zone2,TEST_Rack1",
				Ip:     "192.168.9.113",
				Name:   "HostName3",
			}
			items = append(items, item1)
			items = append(items, item2)
			items = append(items, item3)
			return items, nil
		} else {
			return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := mockFileHostProvider(mockClient)

	filter := structs.HostFilter{
		Arch: string(constants.ArchX8664),
	}

	root, err := hostprovider.GetHierarchy(context.TODO(), &filter, 1, 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(root.SubNodes))
	assert.Equal(t, "TEST_Region1", root.SubNodes[0].Name)
	assert.Equal(t, 2, len(root.SubNodes[0].SubNodes[0].SubNodes))
	assert.Equal(t, 1, len(root.SubNodes[0].SubNodes[1].SubNodes))
}

func Test_GetStocks_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	mockClient.EXPECT().GetHostStocks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks []structs.Stocks, err error) {
		if location.Region == "TEST_Region1" {
			stocks1 := structs.Stocks{
				Zone:             "TEST_Region1,TEST_Zone1",
				FreeCpuCores:     2,
				FreeMemory:       4,
				FreeDiskCount:    2,
				FreeDiskCapacity: 256,
			}
			stocks = append(stocks, stocks1)
			stocks2 := structs.Stocks{
				Zone:             "TEST_Region1,TEST_Zone1",
				FreeCpuCores:     1,
				FreeMemory:       1,
				FreeDiskCount:    1,
				FreeDiskCapacity: 256,
			}
			stocks = append(stocks, stocks2)
			return stocks, nil
		} else {
			return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := mockFileHostProvider(mockClient)

	location := structs.Location{Region: "TEST_Region1"}

	stocks, err := hostprovider.GetStocks(context.TODO(), &location, &structs.HostFilter{}, &structs.DiskFilter{})
	assert.Nil(t, err)
	assert.Equal(t, int32(2), stocks["TEST_Region1,TEST_Zone1"].FreeHostCount)
	assert.Equal(t, int32(3), stocks["TEST_Region1,TEST_Zone1"].FreeCpuCores)
	assert.Equal(t, int32(5), stocks["TEST_Region1,TEST_Zone1"].FreeMemory)
	assert.Equal(t, int32(3), stocks["TEST_Region1,TEST_Zone1"].FreeDiskCount)
	assert.Equal(t, int32(512), stocks["TEST_Region1,TEST_Zone1"].FreeDiskCapacity)
}

func Test_ValidateZoneInfo_Succeed(t *testing.T) {
	models.MockDB()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	prw := mock_product.NewMockReaderWriter(ctrl)
	models.SetProductReaderWriter(prw)

	hostProvider := mockFileHostProvider(nil)

	mockQueryLocalFromDB(prw.EXPECT())
	err := hostProvider.ValidateZoneInfo(context.TODO(), &structs.HostInfo{Vendor: "Local", Region: "Region1", AZ: "Zone1_1"})
	assert.Nil(t, err)

	mockQueryLocalFromDB(prw.EXPECT())
	err = hostProvider.ValidateZoneInfo(context.TODO(), &structs.HostInfo{Vendor: "Local", Region: "Fake_Region", AZ: "Fake_Zone"})
	assert.NotNil(t, err)
	assert.Equal(t, errors.TIUNIMANAGER_RESOURCE_INVALID_ZONE_INFO, err.(errors.EMError).GetCode())

	prw.EXPECT().GetVendor(gomock.Any(), gomock.Any()).Return(nil, nil, nil, errors.Error(errors.TIUNIMANAGER_PARAMETER_INVALID)).Times(1)
	err = hostProvider.ValidateZoneInfo(context.TODO(), &structs.HostInfo{Vendor: "", Region: "Fake_Region", AZ: "Fake_Zone"})
	assert.NotNil(t, err)
	assert.Equal(t, errors.TIUNIMANAGER_PARAMETER_INVALID, err.(errors.EMError).GetCode())
}

func Test_buildInstancesOnHost(t *testing.T) {
	items := []cluster_rw.HostInstanceItem{
		{HostID: "F6ejwHdcQNeHQxQDF0HzMQ", ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "TiDB"},
		{HostID: "F6ejwHdcQNeHQxQDF0HzMQ", ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "TiDB"},
		{HostID: "HWlv9r2ORayaZlUs6HUgTg", ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "TiKV"},
		{HostID: "HWlv9r2ORayaZlUs6HUgTg", ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "TiKV"},
		{HostID: "KU8QP0-uQfyqP7TvPp0deQ", ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "CDC"},
		{HostID: "ZIqy0JAuTxuglIPHNL83hg", ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "AlertManger"},
		{HostID: "ZIqy0JAuTxuglIPHNL83hg", ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "Grafana"},
		{HostID: "ZIqy0JAuTxuglIPHNL83hg", ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "PD"},
		{HostID: "ZIqy0JAuTxuglIPHNL83hg", ClusterID: "Jt_r1RnfT3i0O7YWtMhBzw", Component: "Prometheus"},
		{HostID: "ZIqy0JAuTxuglIPHNL83hg", ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "AlertManger"},
		{HostID: "ZIqy0JAuTxuglIPHNL83hg", ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "Grafana"},
		{HostID: "ZIqy0JAuTxuglIPHNL83hg", ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "PD"},
		{HostID: "ZIqy0JAuTxuglIPHNL83hg", ClusterID: "THwF-s3hTL-SZbZsytBMTw", Component: "Prometheus"},
	}
	hostProvider := mockFileHostProvider(nil)

	results := hostProvider.buildInstancesOnHost(context.TODO(), items)
	assert.Equal(t, 4, len(results))
	assert.Equal(t, 2, len(results["F6ejwHdcQNeHQxQDF0HzMQ"]))
	assert.Equal(t, 2, len(results["ZIqy0JAuTxuglIPHNL83hg"]))
	assert.Equal(t, 4, len(results["ZIqy0JAuTxuglIPHNL83hg"]["Jt_r1RnfT3i0O7YWtMhBzw"]))
	assert.Equal(t, 4, len(results["ZIqy0JAuTxuglIPHNL83hg"]["THwF-s3hTL-SZbZsytBMTw"]))
}

func mockQueryLocalFromDB(expect *mock_product.MockReaderWriterMockRecorder) {
	expect.GetVendor(gomock.Any(), gomock.Any()).Return(&product.Vendor{
		VendorID:   "Local",
		VendorName: "local",
	}, []*product.VendorZone{
		{
			VendorID:   "Local",
			RegionID:   "Region1",
			RegionName: "region1",
			ZoneID:     "Zone1_1",
			ZoneName:   "zone1_1",
			Comment:    "aa",
		},
		{
			VendorID:   "Local",
			RegionID:   "Region1",
			RegionName: "region1",
			ZoneID:     "Zone1_2",
			ZoneName:   "zone1_2",
			Comment:    "aa",
		},
		{
			VendorID:   "Local",
			RegionID:   "Region2",
			RegionName: "region2",
			ZoneID:     "Zone2_1",
			ZoneName:   "zone2_1",
			Comment:    "aa",
		},
	}, []*product.VendorSpec{
		{
			VendorID:    "Local",
			SpecID:      "c.large",
			SpecName:    "c.large",
			CPU:         4,
			Memory:      8,
			DiskType:    "SATA",
			PurposeType: "Compute",
		},
		{
			VendorID:    "Local",
			SpecID:      "s.large",
			SpecName:    "s.large",
			CPU:         4,
			Memory:      8,
			DiskType:    "SATA",
			PurposeType: "Storage",
		},
		{
			VendorID:    "Local",
			SpecID:      "s.xlarge",
			SpecName:    "s.xlarge",
			CPU:         8,
			Memory:      16,
			DiskType:    "SATA",
			PurposeType: "Storage",
		},
		{
			VendorID:    "Local",
			SpecID:      "sc.large",
			SpecName:    "sc.large",
			CPU:         4,
			Memory:      8,
			DiskType:    "SATA",
			PurposeType: "Schedule",
		},
	}, nil).Times(1)
}

func Test_UpdateHost(t *testing.T) {
	fakeHostId1 := "xxxx-xxxx-yyyy-yyyy"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	host := genHostInfo("TEST_HOST1")

	mockClient.EXPECT().UpdateHostInfo(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, host resourcepool.Host) error {
		if host.ID == fakeHostId1 {
			return nil
		} else {
			return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "BadRequest")
		}
	})
	hostprovider := mockFileHostProvider(mockClient)

	err := hostprovider.UpdateHostInfo(context.TODO(), *host)
	assert.NotNil(t, err)
	assert.Equal(t, "update host failed without host id", err.(errors.EMError).GetMsg())

	host.ID = fakeHostId1
	host.UsedCpuCores = 20
	host.UsedMemory = 20
	err = hostprovider.UpdateHostInfo(context.TODO(), *host)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("used cpu cores or used memory field should not be set while update host %s", host.ID), err.(errors.EMError).GetMsg())

	host.UsedCpuCores = 0
	host.UsedMemory = 0

	host.CpuCores = -128
	err = hostprovider.UpdateHostInfo(context.TODO(), *host)
	assert.Equal(t, fmt.Sprintf("input cpu cores(%d) or memory(%d) is invalid to update host %s", -128, 0, host.ID), err.(errors.EMError).GetMsg())
	assert.NotNil(t, err)

	host.CpuCores = 128
	err = hostprovider.UpdateHostInfo(context.TODO(), *host)
	assert.Nil(t, err)
}

func Test_CreateDisks(t *testing.T) {
	fakeHostId1 := "xxxx-xxxx-yyyy-yyyy"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)
	host := genHostInfo("TEST_HOST1")
	host.ID = fakeHostId1
	mockClient.EXPECT().CreateDisks(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostId string, disks []resourcepool.Disk) ([]string, error) {
		return nil, nil
	})
	hostprovider := mockFileHostProvider(mockClient)

	var disk structs.DiskInfo
	_, err := hostprovider.CreateDisks(context.TODO(), fakeHostId1, []structs.DiskInfo{disk})
	assert.NotNil(t, err)
	assert.Equal(t, errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_VALIDATE_DISK_ERROR, "validate disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
		fakeHostId1, disk.Name, disk.Path, disk.Capacity).GetMsg(), err.(errors.EMError).GetMsg())

	_, err = hostprovider.CreateDisks(context.TODO(), fakeHostId1, nil)
	assert.Nil(t, err)
}

func Test_UpdateDisks(t *testing.T) {
	fakeDiskId1 := "xxxx-xxxx-yyyy-yyyy"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)

	mockClient.EXPECT().UpdateDisk(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, disk resourcepool.Disk) error {
		return nil
	})
	hostprovider := mockFileHostProvider(mockClient)

	var disk structs.DiskInfo
	err := hostprovider.UpdateDisk(context.TODO(), disk)
	assert.NotNil(t, err)
	assert.Equal(t, "update disk failed without disk id", err.(errors.EMError).GetMsg())

	disk.ID = fakeDiskId1
	err = hostprovider.UpdateDisk(context.TODO(), disk)
	assert.Nil(t, err)
}

func Test_DeleteDisks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_resource.NewMockReaderWriter(ctrl)

	mockClient.EXPECT().DeleteDisks(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, diskIds []string) error {
		return nil
	})
	hostprovider := mockFileHostProvider(mockClient)

	err := hostprovider.DeleteDisks(context.TODO(), nil)
	assert.Nil(t, err)
}
