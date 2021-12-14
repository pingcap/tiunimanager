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

package gormreadwrite

import (
	"context"
	"fmt"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"github.com/stretchr/testify/assert"
)

func CreateTestHost(region, zone, rack, hostName, ip, clusterType, purpose, diskType string, freeCpuCores, freeMemory, availableDiskCount int32) (id []string, err error) {
	h := resourcepool.Host{
		HostName:     hostName,
		IP:           ip,
		UserName:     "root",
		Passwd:       "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:       string(constants.HostOnline),
		Stat:         string(constants.HostLoadLoadLess),
		Arch:         string(constants.ArchX8664),
		OS:           "CentOS",
		Kernel:       "5.0.0",
		CpuCores:     freeCpuCores,
		Memory:       freeMemory,
		FreeCpuCores: freeCpuCores,
		FreeMemory:   freeMemory,
		Nic:          "1GE",
		Region:       region,
		AZ:           zone,
		Rack:         rack,
		ClusterType:  clusterType,
		Purpose:      purpose,
		DiskType:     diskType,
		Disks: []resourcepool.Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: string(constants.DiskReserved), Type: diskType},
		},
	}
	h.BuildDefaultTraits()
	for i := 0; i < int(availableDiskCount); i++ {
		deviceName := fmt.Sprintf("sd%c", 'b'+i)
		path := fmt.Sprintf("/mnt%d", i+1)
		h.Disks = append(h.Disks, resourcepool.Disk{
			Name:     deviceName,
			Path:     path,
			Capacity: 256,
			Status:   string(constants.DiskAvailable),
			Type:     diskType,
		})
	}
	return GormRW.Create(context.TODO(), []resourcepool.Host{h})
}

func Test_Create_Delete_Host_Succeed(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
}

func Test_Create_Dup_Host(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	assert.NotNil(t, err)
	assert.Nil(t, id2)
}

func Test_Create_Query_Host_Succeed(t *testing.T) {
	hostIp := "192.168.999.999"
	hostName := "Test_Host2"
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", hostName, hostIp,
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	hosts, err := GormRW.Query(context.TODO(), &structs.HostFilter{HostID: id1[0]}, 0, 3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hosts))
	assert.Equal(t, id1[0], hosts[0].ID)
	assert.Equal(t, hostName, hosts[0].HostName)
	assert.Equal(t, hostIp, hosts[0].IP)
	assert.Equal(t, 3, len(hosts[0].Disks))
}

func Test_UpdateHostReserved_Succeed(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	hosts, err := GormRW.Query(context.TODO(), &structs.HostFilter{Arch: string(constants.ArchX8664), Status: string(constants.HostOnline)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.False(t, hosts[0].Reserved)
	GormRW.UpdateHostReserved(context.TODO(), id1, true)
	hosts, err = GormRW.Query(context.TODO(), &structs.HostFilter{Stat: string(constants.HostLoadLoadLess)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.True(t, hosts[0].Reserved)
}

func Test_UpdateHostStatus_Succeed(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	hosts, err := GormRW.Query(context.TODO(), &structs.HostFilter{Arch: string(constants.ArchX8664), Status: string(constants.HostOnline)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.Equal(t, string(constants.HostOnline), hosts[0].Status)
	GormRW.UpdateHostStatus(context.TODO(), id1, string(constants.HostOffline))
	hosts, err = GormRW.Query(context.TODO(), &structs.HostFilter{Purpose: string(constants.PurposeCompute)}, 0, 3)
	assert.True(t, err == nil && len(hosts) == 1)
	assert.Equal(t, string(constants.HostOffline), hosts[0].Status)
}

func Test_GetHostItems(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack2", "Test_Host2", "192.168.192.169",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id2) }()
	assert.Nil(t, err)
	id3, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone2,Test_Rack1", "Test_Host3", "192.168.192.170",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id3) }()
	assert.Nil(t, err)
	items, err := GormRW.GetHostItems(context.TODO(), &structs.HostFilter{Arch: string(constants.ArchX8664)}, 1, 3)
	assert.True(t, err == nil && len(items) == 3)
	assert.Equal(t, "192.168.192.168", items[0].Ip)
	assert.Equal(t, "192.168.192.169", items[1].Ip)
	assert.Equal(t, "192.168.192.170", items[2].Ip)
}

func Test_GetHostStocks(t *testing.T) {
	id1, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack1", "Test_Host1", "192.168.192.168",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id1) }()
	assert.Nil(t, err)
	id2, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone1", "Test_Region1,Test_Zone1,Test_Rack2", "Test_Host2", "192.168.192.169",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 8, 16, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id2) }()
	assert.Nil(t, err)
	id3, err := CreateTestHost("Test_Region1", "Test_Region1,Test_Zone2", "Test_Region1,Test_Zone2,Test_Rack1", "Test_Host3", "192.168.192.170",
		string(constants.EMProductNameTiDB), string(constants.PurposeCompute), string(constants.SSD), 32, 64, 2)
	defer func() { _ = GormRW.Delete(context.TODO(), id3) }()
	assert.Nil(t, err)
	stocks, err := GormRW.GetHostStocks(context.TODO(), &structs.Location{Region: "Test_Region1", Zone: "Test_Zone2"}, &structs.HostFilter{}, &structs.DiskFilter{})
	t.Log(stocks)
	assert.True(t, err == nil)
	assert.Equal(t, 1, len(stocks))
	assert.Equal(t, int32(32), stocks[0].FreeCpuCores)
	assert.Equal(t, int32(64), stocks[0].FreeMemory)
	assert.Equal(t, int32(3), stocks[0].FreeDiskCount)
	assert.Equal(t, int32(3*256), stocks[0].FreeDiskCapacity)
}
