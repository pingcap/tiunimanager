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
		Passwd:       "admin2",
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
	assert.True(t, hostId != "")
	assert.Equal(t, 2, len(host.Disks))
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

func Test_prepareForUpdateSpec(t *testing.T) {
	type want struct {
		freeCpuCores int32
		freeMemory   int32
	}
	tests := []struct {
		testName   string
		originHost Host
		updateHost Host
		want       want
	}{
		{"No_update", Host{CpuCores: 64, Memory: 128, FreeCpuCores: 24, FreeMemory: 28}, Host{CpuCores: 0, Memory: 0}, want{24, 28}},
		{"No_update", Host{CpuCores: 64, Memory: 128, FreeCpuCores: 24, FreeMemory: 28}, Host{CpuCores: 64, Memory: 128}, want{24, 28}},
		{"ScaleOut_Spec_1", Host{CpuCores: 64, Memory: 128, FreeCpuCores: 24, FreeMemory: 28}, Host{CpuCores: 128, Memory: 256}, want{88, 156}},
		{"ScaleOut_Spec_2", Host{CpuCores: 30, Memory: 98, FreeCpuCores: -10, FreeMemory: -2}, Host{CpuCores: 64, Memory: 128}, want{24, 28}},
		{"ScaleOut_Spec_3", Host{CpuCores: 64, Memory: 128, FreeCpuCores: 64, FreeMemory: 128}, Host{CpuCores: 128, Memory: 256}, want{128, 256}},
		{"ScaleIn_Spec_1", Host{CpuCores: 64, Memory: 128, FreeCpuCores: 24, FreeMemory: 28}, Host{CpuCores: 50, Memory: 101}, want{10, 1}},
		{"ScaleIn_Spec_2", Host{CpuCores: 64, Memory: 128, FreeCpuCores: 24, FreeMemory: 28}, Host{CpuCores: 30, Memory: 98}, want{-10, -2}},
		{"ScaleIn_Spec_3", Host{CpuCores: 30, Memory: 98, FreeCpuCores: -10, FreeMemory: -2}, Host{CpuCores: 35, Memory: 105}, want{-5, 5}},
		{"ScaleIn_Spec_4", Host{CpuCores: 64, Memory: 128, FreeCpuCores: 64, FreeMemory: 128}, Host{CpuCores: 32, Memory: 64}, want{32, 64}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			updates := make(map[string]interface{})
			tt.originHost.prepareForUpdateSpec(tt.updateHost.CpuCores, tt.updateHost.Memory, updates)
			if tt.testName == "No_update" {
				assert.Equal(t, 0, len(updates))
			} else {
				updatedCpu, ok := updates["cpu_cores"]
				assert.True(t, ok)
				cores := updatedCpu.(int32)
				assert.Equal(t, tt.updateHost.CpuCores, cores)

				updatedMem, ok := updates["memory"]
				assert.True(t, ok)
				mem := updatedMem.(int32)
				assert.Equal(t, tt.updateHost.Memory, mem)

				updatedFreeCpu, ok := updates["free_cpu_cores"]
				assert.True(t, ok)
				freeCpuCore := updatedFreeCpu.(int32)
				assert.Equal(t, tt.want.freeCpuCores, freeCpuCore)

				updatedFreeMem, ok := updates["free_memory"]
				assert.True(t, ok)
				freeMem := updatedFreeMem.(int32)
				assert.Equal(t, tt.want.freeMemory, freeMem)
			}
		})
	}
}

func Test_UpdateHost(t *testing.T) {
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

	err = db.Model(&Host{ID: hostId}).Update("ip", "192.168.965.965").Error
	assert.NotNil(t, err)
	assert.Equal(t, err.(errors.EMError).GetMsg(), errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR, "update ip on host %s is not allowed", hostId).GetMsg())

	err = db.Model(&Host{ID: hostId}).Update("disk_type", string(constants.SSD)).Error
	assert.NotNil(t, err)
	assert.Equal(t, err.(errors.EMError).GetMsg(), errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR, "update disk type or arch type or cluster type or load stat on host %s is not allowed", hostId).GetMsg())

	err = db.Model(&Host{ID: hostId}).Update("region", "NewRegion").Error
	assert.NotNil(t, err)
	assert.Equal(t, err.(errors.EMError).GetMsg(), errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR, "update vendor/region/zone/rack info on host %s is not allowed", hostId).GetMsg())

	host2 := genFakeHost("Region1", "Region1,Zone1", "Region1,Zone1,Rack1", "TEST_HOST2", "192.168.999.999", 48, 128,
		string(constants.EMProductIDDataMigration), string(constants.PurposeStorage), string(constants.NVMeSSD))

	updates, err := host.PrepareForUpdate(host2)
	assert.Nil(t, err)
	value, ok := updates["host_name"]
	assert.True(t, ok)
	hostName := value.(string)
	assert.Equal(t, "TEST_HOST2", hostName)
	value, ok = updates["spec"]
	assert.True(t, ok)
	spec := value.(string)
	assert.Equal(t, "48C128G", spec)
	value, ok = updates["traits"]
	assert.True(t, ok)
	traits := value.(int64)
	assert.Equal(t, int64(82), traits)
	err = db.Model(&host).Updates(updates).Error
	assert.Nil(t, err)

	var queryHost Host
	err = db.Where("id = ?", host.ID).Find(&queryHost).Error
	assert.Nil(t, err)
	assert.Equal(t, "48C128G", queryHost.Spec)
	assert.Equal(t, int32(48), queryHost.FreeCpuCores)
	assert.Equal(t, int32(128), queryHost.FreeMemory)
	assert.Equal(t, string(constants.PurposeStorage), queryHost.Purpose)
	assert.Equal(t, int64(82), queryHost.Traits)
	assert.Equal(t, "TEST_HOST2", queryHost.HostName)

	host3 := genFakeHost("Region1", "Region1,Zone1", "Region1,Zone1,Rack1", "TEST_HOST2", "192.168.999.998", 48, 128,
		string(constants.EMProductIDDataMigration), string(constants.PurposeStorage), string(constants.NVMeSSD))

	updates, err = host.PrepareForUpdate(host3)
	assert.Nil(t, err)
	value, ok = updates["ip"]
	assert.True(t, ok)
	ip := value.(string)
	assert.Equal(t, "192.168.999.998", ip)
	err = db.Model(&host).Updates(updates).Error
	assert.NotNil(t, err)
	assert.Equal(t, err.(errors.EMError).GetMsg(), errors.NewErrorf(errors.TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR, "update ip on host %s is not allowed", hostId).GetMsg())

	err = db.Delete(&Host{ID: host.ID}).Error
	assert.Nil(t, err)
}

func Test_ConstructLabelRecord(t *testing.T) {
	type want struct {
		label Label
	}
	tests := []struct {
		testName string
		srcLabel structs.Label
		want     want
	}{
		{"Normal", structs.Label{Name: "Compute", Category: 0, Trait: 0x0000000000000001}, want{Label{Name: "Compute", Category: 0, Trait: 0x0000000000000001}}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var dstLabel Label
			dstLabel.ConstructLabelRecord(&tt.srcLabel)
			assert.Equal(t, tt.want.label.Name, dstLabel.Name)
			assert.Equal(t, tt.want.label.Category, dstLabel.Category)
			assert.Equal(t, tt.want.label.Trait, dstLabel.Trait)
		})
	}
}
