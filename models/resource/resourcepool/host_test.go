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

package resourcepool

import (
	"github.com/pingcap-inc/tiem/models/common"
	"os"
	"testing"

	"github.com/pingcap-inc/tiem/util/uuidutil"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_BuildDefaultTraits(t *testing.T) {
	type want struct {
		errcode errors.EM_ERROR_CODE
		Traits  int64
	}
	tests := []struct {
		name string
		host Host
		want want
	}{
		{"test1", Host{ClusterType: string(constants.EMProductIDTiDB), Purpose: string(constants.PurposeCompute), DiskType: string(constants.NVMeSSD)}, want{errors.TIEM_SUCCESS, 73}},
		{"test2", Host{ClusterType: string(constants.EMProductIDTiDB), Purpose: string(constants.PurposeCompute) + "," + string(constants.PurposeSchedule), DiskType: string(constants.SSD)}, want{errors.TIEM_SUCCESS, 169}},
		{"test3", Host{ClusterType: string(constants.EMProductIDTiDB), Purpose: "General", DiskType: string(constants.NVMeSSD)}, want{errors.TIEM_RESOURCE_TRAIT_NOT_FOUND, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.host.BuildDefaultTraits()
			if err == nil {
				assert.Equal(t, tt.want.Traits, tt.host.Traits)
			} else {
				te, ok := err.(errors.EMError)
				assert.Equal(t, true, ok)
				assert.True(t, tt.want.errcode.Equal(int32(te.GetCode())))
			}
		})
	}
}

func TestHost_IsLoadLess(t *testing.T) {
	type want struct {
		IsLoadless bool
	}
	tests := []struct {
		name string
		host Host
		want want
	}{
		{"normal", Host{CpuCores: 4, Memory: 8, FreeCpuCores: 4, FreeMemory: 8, Disks: []Disk{{Status: string(constants.DiskAvailable)}}}, want{true}},
		{"diskused", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Disks: []Disk{{Status: string(constants.DiskInUsed)}}}, want{false}},
		{"diskused2", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Disks: []Disk{{Status: string(constants.DiskExhaust)}}}, want{false}},
		{"diskused3", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Disks: []Disk{{Status: string(constants.DiskExhaust)}, {Status: string(constants.DiskInUsed)}}}, want{false}},
		{"cpuused", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 12, FreeMemory: 64, Disks: []Disk{{Status: string(constants.DiskAvailable)}}}, want{false}},
		{"memoryused", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 8}, want{false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsLoadless := tt.host.IsLoadless(); IsLoadless != tt.want.IsLoadless {
				t.Errorf("IsLoadless() = %v, want %v", IsLoadless, tt.want)
			}
		})
	}
}

func TestHost_IsInused(t *testing.T) {
	type want struct {
		IsLoadless bool
	}
	tests := []struct {
		name string
		host Host
		want want
	}{
		{"normal", Host{CpuCores: 4, Memory: 8, FreeCpuCores: 4, FreeMemory: 8, Stat: string(constants.HostLoadLoadLess), Disks: []Disk{{Status: string(constants.DiskAvailable)}}}, want{false}},
		{"inused", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Stat: string(constants.HostLoadInUsed), Disks: []Disk{{Status: string(constants.DiskInUsed)}}}, want{true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsInused := tt.host.IsInused(); IsInused != tt.want.IsLoadless {
				t.Errorf("IsInused() = %v, want %v", IsInused, tt.want)
			}
		})
	}
}

func createDB(path string) (db *gorm.DB, err error) {
	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	} else {
		return db, nil
	}
}

func genFakeHost(region, zone, rack, hostName, ip string, freeCpuCores, freeMemory int32, clusterType, purpose, diskType string) (h *Host) {
	host := Host{
		ID:           "xxxx-aaaa-bbbb-cccc",
		HostName:     hostName,
		IP:           ip,
		UserName:     "root",
		Passwd:       common.Password{Val: "admin2"},
		Status:       string(constants.HostOnline),
		Stat:         string(constants.HostLoadLoadLess),
		Arch:         "X86",
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
		Disks: []Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: string(constants.DiskReserved), Type: diskType},
			{Name: "sdb", Path: "/", Capacity: 256, Status: string(constants.DiskAvailable), Type: diskType},
		},
	}
	host.BuildDefaultTraits()
	return &host
}

func Test_Host_Hooks(t *testing.T) {
	dbPath := "./test_resource_" + uuidutil.ShortId() + ".db"
	db, err := createDB(dbPath)
	assert.Nil(t, err)
	defer func() { _ = os.Remove(dbPath) }()

	host := genFakeHost("Region1", "Region1,Zone1", "Region1,Zone1,Rack1", "TEST_HOST1", "192.168.999.999", 32, 64,
		string(constants.EMProductIDDataMigration), string(constants.PurposeSchedule), string(constants.NVMeSSD))
	db.AutoMigrate(&Host{})
	db.AutoMigrate(&Disk{})
	err = db.Model(&Host{}).Create(host).Error
	assert.Nil(t, err)
	hostId := host.ID
	assert.NotNil(t, hostId)
	for _, v := range host.Disks {
		assert.NotNil(t, v)
	}
	err = db.Delete(&Host{ID: host.ID}).Error
	assert.Nil(t, err)
	var queryHost Host
	err = db.Unscoped().Where("id = ?", host.ID).Find(&queryHost).Error
	assert.Nil(t, err)
	assert.Equal(t, string(constants.HostDeleted), queryHost.Status)
}

func Test_ConstructFromHostInfo(t *testing.T) {
	src := structs.HostInfo{
		IP:           "192.168.999.999",
		UserName:     "root",
		Region:       "TEST_Region1",
		AZ:           "TEST_Zone1",
		Rack:         "TEST_Rack1",
		CpuCores:     8,
		Memory:       16,
		UsedCpuCores: 4,
		UsedMemory:   10,
		Disks: []structs.DiskInfo{
			{Name: "sda", Path: "/", Capacity: 256, Status: string(constants.DiskReserved), Type: string(constants.SSD)},
			{Name: "sdb", Path: "/", Capacity: 256, Status: string(constants.DiskAvailable), Type: string(constants.NVMeSSD)},
		},
	}
	dst := &Host{}
	dst.ConstructFromHostInfo(&src)
	assert.Equal(t, "192.168.999.999", dst.IP)
	assert.Equal(t, "root", dst.UserName)
	assert.Equal(t, "TEST_Region1", dst.Region)
	assert.Equal(t, "TEST_Region1,TEST_Zone1", dst.AZ)
	assert.Equal(t, "TEST_Region1,TEST_Zone1,TEST_Rack1", dst.Rack)
	assert.Equal(t, int32(4), dst.FreeCpuCores)
	assert.Equal(t, int32(6), dst.FreeMemory)
	assert.Equal(t, 2, len(dst.Disks))
}

func Test_ToHostInfo(t *testing.T) {
	src := genFakeHost("Region1", "Region1,Zone1", "Region1,Zone1,Rack1", "TEST_HOST1", "192.168.999.999", 32, 64,
		string(constants.EMProductIDDataMigration), string(constants.PurposeSchedule), string(constants.NVMeSSD))

	src.FreeCpuCores = 0
	src.FreeMemory = 28
	src.Stat = string(constants.HostLoadInUsed)

	var dst structs.HostInfo
	src.ToHostInfo(&dst)
	assert.Equal(t, "Region1", dst.Region)
	assert.Equal(t, "Zone1", dst.AZ)
	assert.Equal(t, "Rack1", dst.Rack)
	assert.Equal(t, 2, len(dst.Disks))
	assert.Equal(t, int32(32), dst.UsedCpuCores)
	assert.Equal(t, int32(36), dst.UsedMemory)
	assert.Equal(t, string(constants.HostLoadComputeExhaust), dst.Stat)
}
