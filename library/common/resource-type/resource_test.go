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

package resource

import (
	"errors"
	em_errors "github.com/pingcap-inc/tiem/common/errors"
	"os"
	"testing"

	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_ValidArchType(t *testing.T) {
	type want struct {
		err error
	}
	tests := []struct {
		name string
		arch string
		want want
	}{
		{"valid_X86", string(X86_64), want{nil}},
		{"valid_ARM64", string(Arm64), want{nil}},
		{"invalid_x86", "x86", want{errors.New("valid arch type: [ARM64 | X86_64]")}},
		{"invalid_arm64", "arm64", want{errors.New("valid arch type: [ARM64 | X86_64]")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidArch(tt.arch)
			if err != nil && tt.want.err == nil {
				t.Errorf("Expect no err, but get error = %v", err.Error())
			}
			if err == nil && tt.want.err != nil {
				t.Errorf("Expect err: %v but no error", tt.want.err)
			}
		})
	}
}

func Test_ValidPurposeType(t *testing.T) {
	type want struct {
		err error
	}
	tests := []struct {
		name    string
		purpose string
		want    want
	}{
		{"valid_Dispatch", "Dispatch", want{nil}},
		{"valid_Compute", "Compute", want{nil}},
		{"valid_Storage", "Storage", want{nil}},
		{"invalid_UnDefine", "UnDefine", want{errors.New("valid purpose: [Compute | Storage | Dispatch]")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidPurposeType(tt.purpose)
			if err != nil && tt.want.err == nil {
				t.Errorf("Expect no err, but get error = %v", err.Error())
			}
			if err == nil && tt.want.err != nil {
				t.Errorf("Expect err: %v but no error", tt.want.err)
			}
		})
	}
}

func TestHost_IsExhaust(t *testing.T) {
	type want struct {
		stat      HostStat
		isExhaust bool
	}
	tests := []struct {
		name string
		host Host
		want want
	}{
		{"normal", Host{FreeCpuCores: 4, FreeMemory: 8, Disks: []Disk{{Status: int32(DISK_AVAILABLE)}}}, want{HOST_STAT_WHATEVER, false}},
		{"exhaust1", Host{FreeCpuCores: 0, FreeMemory: 0, Disks: []Disk{{Status: int32(DISK_INUSED)}}}, want{HOST_EXHAUST, true}},
		{"exhaust2", Host{FreeCpuCores: 0, FreeMemory: 0, Disks: []Disk{{Status: int32(DISK_EXHAUST)}}}, want{HOST_EXHAUST, true}},
		{"exhaust3", Host{FreeCpuCores: 0, FreeMemory: 0, Disks: []Disk{{Status: int32(DISK_EXHAUST)}, {Status: int32(DISK_INUSED)}}}, want{HOST_EXHAUST, true}},
		{"with_disk", Host{FreeCpuCores: 0, FreeMemory: 0, Disks: []Disk{{Status: int32(DISK_AVAILABLE)}}}, want{HOST_COMPUTE_EXHAUST, true}},
		{"without_disk", Host{FreeCpuCores: 4, FreeMemory: 8}, want{HOST_DISK_EXHAUST, true}},
		{"without_cpu", Host{FreeCpuCores: 0, FreeMemory: 8, Disks: []Disk{{Status: int32(DISK_AVAILABLE)}}}, want{HOST_COMPUTE_EXHAUST, true}},
		{"without_momery", Host{FreeCpuCores: 4, FreeMemory: 0, Disks: []Disk{{Status: int32(DISK_AVAILABLE)}}}, want{HOST_COMPUTE_EXHAUST, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if stat, isExaust := tt.host.IsExhaust(); stat != tt.want.stat || isExaust != tt.want.isExhaust {
				t.Errorf("IsExhaust() = %v, want %v", stat, tt.want)
			}
		})
	}
}

func TestHost_IsAvailable(t *testing.T) {
	type want struct {
		isAvailable bool
	}
	tests := []struct {
		name string
		host Host
		want want
	}{
		{"available1", Host{Status: int32(HOST_ONLINE), Stat: int32(HOST_LOADLESS)}, want{true}},
		{"available1", Host{Status: int32(HOST_ONLINE), Stat: int32(HOST_INUSED)}, want{true}},
		{"offline", Host{Status: int32(HOST_OFFLINE), Stat: int32(HOST_LOADLESS)}, want{false}},
		{"Exhaust", Host{Status: int32(HOST_ONLINE), Stat: int32(HOST_EXHAUST)}, want{false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isAvailable := tt.host.IsAvailable(); isAvailable != tt.want.isAvailable {
				t.Errorf("IsAvailable() = %v, want %v", isAvailable, tt.want)
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
		{"normal", Host{CpuCores: 4, Memory: 8, FreeCpuCores: 4, FreeMemory: 8, Disks: []Disk{{Status: int32(DISK_AVAILABLE)}}}, want{true}},
		{"diskused", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Disks: []Disk{{Status: int32(DISK_INUSED)}}}, want{false}},
		{"diskused2", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Disks: []Disk{{Status: int32(DISK_EXHAUST)}}}, want{false}},
		{"diskused3", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Disks: []Disk{{Status: int32(DISK_EXHAUST)}, {Status: int32(DISK_INUSED)}}}, want{false}},
		{"cpuused", Host{CpuCores: 16, Memory: 64, FreeCpuCores: 12, FreeMemory: 64, Disks: []Disk{{Status: int32(DISK_AVAILABLE)}}}, want{false}},
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

func Test_BuildDefaultTraits(t *testing.T) {
	type want struct {
		errcode em_errors.EM_ERROR_CODE
		Traits  int64
	}
	tests := []struct {
		name string
		host Host
		want want
	}{
		{"test1", Host{ClusterType: string(Database), Purpose: string(Compute), DiskType: string(NvmeSSD)}, want{em_errors.TIEM_SUCCESS, 37}},
		{"test2", Host{ClusterType: string(Database), Purpose: string(Compute) + "," + string(Dispatch), DiskType: string(SSD)}, want{em_errors.TIEM_SUCCESS, 85}},
		{"test3", Host{ClusterType: string(Database), Purpose: "General", DiskType: string(NvmeSSD)}, want{em_errors.TIEM_RESOURCE_TRAIT_NOT_FOUND, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.host.BuildDefaultTraits()
			if err == nil {
				assert.Equal(t, tt.want.Traits, tt.host.Traits)
			} else {
				te, ok := err.(em_errors.EMError)
				assert.Equal(t, true, ok)
				assert.True(t, tt.want.errcode.Equal(int32(te.GetCode())))
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
		HostName:     hostName,
		IP:           ip,
		UserName:     "root",
		Passwd:       "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:       0,
		Stat:         0,
		Arch:         "X86",
		OS:           "CentOS",
		Kernel:       "5.0.0",
		CpuCores:     freeCpuCores,
		Memory:       freeCpuCores,
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
			{Name: "sda", Path: "/", Capacity: 256, Status: 1, Type: diskType},
			{Name: "sdb", Path: "/", Capacity: 256, Status: 0, Type: diskType},
		},
	}
	host.BuildDefaultTraits()
	return &host
}

func Test_Used_Resources_Hooks(t *testing.T) {
	dbPath := "./test_used_resource_" + uuidutil.ShortId() + ".db"
	db, err := createDB(dbPath)
	assert.Nil(t, err)
	defer func() { _ = os.Remove(dbPath) }()

	fakeHostId := "TEST_HOST_ID1"
	fakeHolderId := "TEST_Cluster_ID1"
	fakeRequestId := "TEST_Request_ID1"
	fakeDiskId := "TEST_Disk_ID1"
	usedCompute := UsedCompute{
		CpuCores: 4,
		Memory:   8,
		Holder: Holder{
			HolderId:  fakeHolderId,
			RequestId: fakeRequestId,
		},
		HostId: fakeHostId,
	}
	usedDisk := UsedDisk{
		DiskId: fakeDiskId,
		HostId: fakeHostId,
		Holder: Holder{
			HolderId:  fakeHolderId,
			RequestId: fakeRequestId,
		},
	}
	usedPort := UsedPort{
		Port:   10000,
		HostId: fakeHostId,
		Holder: Holder{
			HolderId:  fakeHolderId,
			RequestId: fakeRequestId,
		},
	}
	db.AutoMigrate(&UsedCompute{})
	db.AutoMigrate(&UsedDisk{})
	db.AutoMigrate(&UsedPort{})

	err = db.Model(&UsedCompute{}).Create(&usedCompute).Error
	assert.Nil(t, err)
	assert.NotNil(t, usedCompute.ID)
	err = db.Model(&UsedDisk{}).Create(&usedDisk).Error
	assert.Nil(t, err)
	assert.NotNil(t, usedDisk.ID)
	err = db.Model(&UsedPort{}).Create(&usedPort).Error
	assert.Nil(t, err)
	assert.NotNil(t, usedPort.ID)
}

func Test_Host_Hooks(t *testing.T) {
	dbPath := "./test_resource_" + uuidutil.ShortId() + ".db"
	db, err := createDB(dbPath)
	assert.Nil(t, err)
	defer func() { _ = os.Remove(dbPath) }()

	host := genFakeHost("Region1", "Region1,Zone1", "Region1,Zone1,Rack1", "TEST_HOST1", "192.168.999.999", 32, 64, string(DataMigration), string(Dispatch), string(NvmeSSD))
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
	assert.Equal(t, int32(HOST_DELETED), queryHost.Status)
}

func Test_GetTraitByName(t *testing.T) {
	type want struct {
		errcode em_errors.EM_ERROR_CODE
		Traits  int64
	}
	tests := []struct {
		name      string
		labelName string
		want      want
	}{
		{"Test1", string(Storage), want{em_errors.TIEM_SUCCESS, 8}},
		{"Test2", string(DataMigration), want{em_errors.TIEM_SUCCESS, 2}},
		{"Test3", string(Sata), want{em_errors.TIEM_SUCCESS, 128}},
		{"Test1", "BadLabelName", want{em_errors.TIEM_RESOURCE_TRAIT_NOT_FOUND, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trait, err := GetTraitByName(tt.labelName)
			if err == nil {
				assert.Equal(t, tt.want.Traits, trait)
			} else {
				te, ok := err.(em_errors.EMError)
				assert.Equal(t, true, ok)
				assert.True(t, tt.want.errcode.Equal(int32(te.GetCode())))
			}
		})
	}
}
