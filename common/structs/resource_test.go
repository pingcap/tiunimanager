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

package structs

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/common/errors"
)

func Test_GenDomainCodeByName(t *testing.T) {
	type want struct {
		code string
	}
	tests := []struct {
		testName string
		pre      string
		name     string
		want     want
	}{
		{"ZoneCode", "Test_Region1", "Test_Zone1", want{"Test_Region1,Test_Zone1"}},
		{"RackCode", "Test_Region1,Test_Zone1", "Test_Rack1", want{"Test_Region1,Test_Zone1,Test_Rack1"}},
		{"IpCode", "Test_Host1", "192.168.168.168", want{"Test_Host1,192.168.168.168"}},
		{"Empty", "", "", want{""}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			code := GenDomainCodeByName(tt.pre, tt.name)
			assert.Equal(t, tt.want.code, code)
		})
	}
}

func Test_GetDomainNameFromCode(t *testing.T) {
	type want struct {
		name string
	}
	tests := []struct {
		testName string
		code     string
		want     want
	}{
		{"RegionName", "Test_Region1", want{"Test_Region1"}},
		{"ZoneName", "Test_Region1,Test_Zone1", want{"Test_Zone1"}},
		{"RackName", "Test_Region1,Test_Zone1,Test_Rack1", want{"Test_Rack1"}},
		{"IP", "Test_Host1,192.168.168.168", want{"192.168.168.168"}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			name := GetDomainNameFromCode(tt.code)
			assert.Equal(t, tt.want.name, name)
		})
	}
}

func Test_GetDomainPrefixFromCode(t *testing.T) {
	type want struct {
		pre string
	}
	tests := []struct {
		testName string
		code     string
		want     want
	}{
		{"regionName", "Test_Region1", want{"Test_Region1"}},
		{"ZoneName", "Test_Region1,Test_Zone1", want{"Test_Region1"}},
		{"RackName", "Test_Region1,Test_Zone1,Test_Rack1", want{"Test_Region1,Test_Zone1"}},
		{"HostName", "Test_Host1,192.168.168.168", want{"Test_Host1"}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			pre := GetDomainPrefixFromCode(tt.code)
			assert.Equal(t, tt.want.pre, pre)
		})
	}
}

func Test_GetTraitByName(t *testing.T) {
	type want struct {
		trait int64
		err   error
	}
	tests := []struct {
		testName string
		name     string
		want     want
	}{
		{"EM", string(constants.EMProductIDEnterpriseManager), want{trait: 0x0000000000000004, err: nil}},
		{"Schedule", string(constants.PurposeSchedule), want{trait: 0x0000000000000020, err: nil}},
		{"SSD", string(constants.SSD), want{trait: 0x0000000000000080, err: nil}},
		{"ERROR", "InvalidName", want{trait: 0, err: errors.NewErrorf(errors.TIEM_RESOURCE_TRAIT_NOT_FOUND,
			"label type %v not found in system default labels", "InvalidName")}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			trait, err := GetTraitByName(tt.name)
			if err == nil {
				assert.Equal(t, tt.want.trait, trait)
			} else {
				assert.Equal(t, err.(errors.EMError).GetCode(), errors.TIEM_RESOURCE_TRAIT_NOT_FOUND)
			}

		})
	}
}

func Test_GetLabelNamesByTraits(t *testing.T) {
	type want struct {
		labels []string
	}
	tests := []struct {
		testName string
		traits   int64
		want     want
	}{
		{"DM", 0x0000000000000002, want{labels: []string{string(constants.EMProductIDDataMigration)}}},
		{"Mix", 0x0000000000000099, want{labels: []string{string(constants.EMProductIDTiDB), string(constants.PurposeCompute), string(constants.PurposeStorage), string(constants.SSD)}}},
		{"Nil", 0x0000000000032000, want{labels: nil}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			labels := GetLabelNamesByTraits(tt.traits)
			assert.Equal(t, len(tt.want.labels), len(labels))
		})
	}
}

func TestHost_IsExhaust(t *testing.T) {
	type want struct {
		stat      constants.HostLoadStatus
		isExhaust bool
	}
	tests := []struct {
		name string
		host HostInfo
		want want
	}{
		{"normal", HostInfo{UsedCpuCores: 4, UsedMemory: 8, CpuCores: 8, Memory: 16, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{constants.HostLoadStatusWhatever, false}},
		{"exhaust1", HostInfo{UsedCpuCores: 8, UsedMemory: 16, CpuCores: 8, Memory: 16, Disks: []DiskInfo{{Status: string(constants.DiskInUsed)}}}, want{constants.HostLoadExhaust, true}},
		{"exhaust2", HostInfo{UsedCpuCores: 8, UsedMemory: 16, CpuCores: 8, Memory: 16, Disks: []DiskInfo{{Status: string(constants.DiskExhaust)}}}, want{constants.HostLoadExhaust, true}},
		{"exhaust3", HostInfo{UsedCpuCores: 8, UsedMemory: 16, CpuCores: 8, Memory: 16, Disks: []DiskInfo{{Status: string(constants.DiskExhaust)}, {Status: string(constants.DiskInUsed)}}}, want{constants.HostLoadExhaust, true}},
		{"with_disk", HostInfo{UsedCpuCores: 8, UsedMemory: 16, CpuCores: 8, Memory: 16, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{constants.HostLoadComputeExhaust, true}},
		{"without_disk", HostInfo{UsedCpuCores: 4, UsedMemory: 8, CpuCores: 8, Memory: 16}, want{constants.HostLoadDiskExhaust, true}},
		{"without_cpu", HostInfo{UsedCpuCores: 8, UsedMemory: 8, CpuCores: 8, Memory: 16, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{constants.HostLoadComputeExhaust, true}},
		{"without_momery", HostInfo{UsedCpuCores: 4, UsedMemory: 16, CpuCores: 8, Memory: 16, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{constants.HostLoadComputeExhaust, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if stat, isExaust := tt.host.IsExhaust(); stat != tt.want.stat || isExaust != tt.want.isExhaust {
				t.Errorf("IsExhaust() = %v, want %v", stat, tt.want)
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
		host HostInfo
		want want
	}{
		{"normal", HostInfo{CpuCores: 4, Memory: 8, UsedCpuCores: 0, UsedMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{true}},
		{"diskused", HostInfo{CpuCores: 16, Memory: 64, UsedCpuCores: 0, UsedMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskInUsed)}}}, want{false}},
		{"diskused2", HostInfo{CpuCores: 16, Memory: 64, UsedCpuCores: 0, UsedMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskExhaust)}}}, want{false}},
		{"diskused3", HostInfo{CpuCores: 16, Memory: 64, UsedCpuCores: 0, UsedMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskExhaust)}, {Status: string(constants.DiskAvailable)}}}, want{false}},
		{"cpuused", HostInfo{CpuCores: 16, Memory: 64, UsedCpuCores: 4, UsedMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{false}},
		{"memoryused", HostInfo{CpuCores: 16, Memory: 64, UsedCpuCores: 0, UsedMemory: 56}, want{false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsLoadless := tt.host.IsLoadless(); IsLoadless != tt.want.IsLoadless {
				t.Errorf("IsLoadless() = %v, want %v", IsLoadless, tt.want)
			}
		})
	}
}

func Test_GetPurposes(t *testing.T) {
	type want struct {
		purposes []string
	}
	tests := []struct {
		testName string
		host     HostInfo
		want     want
	}{
		{"Single", HostInfo{Purpose: "Compute"}, want{purposes: []string{"Compute"}}},
		{"Mix", HostInfo{Purpose: "Compute,Storage,Schedule"}, want{purposes: []string{"Compute", "Storage", "Schedule"}}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			purposes := (tt.host.GetPurposes())
			assert.Equal(t, len(tt.want.purposes), len(purposes))
		})
	}
}

func Test_AddTraits(t *testing.T) {
	type want struct {
		traits int64
	}
	tests := []struct {
		testName string
		host     HostInfo
		labels   []string
		want     want
	}{
		{"Single", HostInfo{Traits: 0}, []string{string(constants.EMProductIDEnterpriseManager)}, want{0x0000000000000004}},
		{"Mix", HostInfo{Traits: 1}, []string{string(constants.PurposeCompute), string(constants.PurposeSchedule), string(constants.NVMeSSD)}, want{0x0000000000000069}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			for _, label := range tt.labels {
				tt.host.AddTraits(label)
			}
			assert.Equal(t, tt.want.traits, tt.host.Traits)
		})
	}
}

func TestGenSpecCode(t *testing.T) {
	type args struct {
		cpuCores int32
		mem      int32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{3, 6}, "3C6G"},
		{"normal", args{999, 999}, "999C999G"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenSpecCode(tt.args.cpuCores, tt.args.mem); got != tt.want {
				t.Errorf("GenSpecCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCpu(t *testing.T) {
	type args struct {
		specCode string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"normal", args{"3C6G"}, 3},
		{"normal", args{"999C6G"}, 999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCpu(tt.args.specCode); got != tt.want {
				t.Errorf("ParseCpu() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMemory(t *testing.T) {
	type args struct {
		specCode string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"normal", args{"3C6G"}, 6},
		{"normal", args{"3C999G"}, 999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseMemory(tt.args.specCode); got != tt.want {
				t.Errorf("ParseMemory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ValidateDisk(t *testing.T) {
	fakeHostId := "fake-host-id"
	diskName := "sdb"
	diskPath := "/mnt/sdb"
	var capacity int32 = 256
	type want struct {
		err error
	}
	tests := []struct {
		name     string
		hostId   string
		diskType constants.DiskType
		disk     DiskInfo
		want     want
	}{
		{"normal", fakeHostId, constants.NVMeSSD, DiskInfo{Name: diskName, Path: diskPath, Capacity: capacity, Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{nil}},
		{"normal_without_type", fakeHostId, constants.NVMeSSD, DiskInfo{Name: diskName, Path: diskPath, Capacity: capacity, Status: string(constants.DiskAvailable)}, want{nil}},
		{"normal_without_status", fakeHostId, constants.NVMeSSD, DiskInfo{Name: diskName, Path: diskPath, Capacity: capacity}, want{nil}},
		{"no_name", fakeHostId, constants.NVMeSSD, DiskInfo{Path: diskPath, Capacity: capacity, Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
				fakeHostId, "", diskPath, capacity),
		}},
		{"no_path", fakeHostId, constants.NVMeSSD, DiskInfo{Name: diskName, Capacity: capacity, Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
				fakeHostId, diskName, "", capacity),
		}},
		{"no_capacity", fakeHostId, constants.NVMeSSD, DiskInfo{Name: diskName, Path: diskPath, Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
				fakeHostId, diskName, diskPath, 0),
		}},
		{"hostId_mismatch", fakeHostId, constants.NVMeSSD, DiskInfo{Name: diskName, Path: diskPath, Capacity: capacity, HostId: "fake-host2", Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s failed, host id conflict %s vs %s",
				diskName, diskPath, "fake-host2", fakeHostId),
		}},
		{"diskType_mismatch", fakeHostId, constants.NVMeSSD, DiskInfo{Name: diskName, Path: diskPath, Capacity: capacity, Type: string(constants.SSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s for host %s failed, disk type conflict %s vs %s",
				diskName, diskPath, fakeHostId, string(constants.SSD), string(constants.NVMeSSD)),
		}},
		{"status_invalid", fakeHostId, constants.NVMeSSD, DiskInfo{Name: diskName, Path: diskPath, Capacity: capacity, Type: string(constants.NVMeSSD), Status: "bad_status"}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s for host %s specified a invalid status %s, [Available|Reserved]",
				diskName, diskPath, fakeHostId, "bad_status"),
		}},
		{"type_invalid", fakeHostId, constants.NVMeSSD, DiskInfo{Name: diskName, Path: diskPath, Capacity: capacity, Type: "bad_disk_type"}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "validate disk %s %s for host %s failed, %v",
				diskName, diskPath, fakeHostId, constants.ValidDiskType("bad_disk_type")),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.disk.ValidateDisk(tt.hostId, string(tt.diskType))
			if err == nil {
				assert.Equal(t, err, tt.want.err)
			} else {
				assert.Equal(t, tt.want.err.(errors.EMError).GetCode(), err.(errors.EMError).GetCode())
				assert.Equal(t, tt.want.err.(errors.EMError).GetMsg(), err.(errors.EMError).GetMsg())
			}
		})
	}
}
