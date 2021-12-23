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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
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
		{"ERROR", "InvalidName", want{trait: 0, err: errors.NewEMErrorf(errors.TIEM_RESOURCE_TRAIT_NOT_FOUND,
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
		{"normal", HostInfo{FreeCpuCores: 4, FreeMemory: 8, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{constants.HostLoadStatusWhatever, false}},
		{"exhaust1", HostInfo{FreeCpuCores: 0, FreeMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskInUsed)}}}, want{constants.HostLoadExhaust, true}},
		{"exhaust2", HostInfo{FreeCpuCores: 0, FreeMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskExhaust)}}}, want{constants.HostLoadExhaust, true}},
		{"exhaust3", HostInfo{FreeCpuCores: 0, FreeMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskExhaust)}, {Status: string(constants.DiskInUsed)}}}, want{constants.HostLoadExhaust, true}},
		{"with_disk", HostInfo{FreeCpuCores: 0, FreeMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{constants.HostLoadComputeExhaust, true}},
		{"without_disk", HostInfo{FreeCpuCores: 4, FreeMemory: 8}, want{constants.HostLoadDiskExhaust, true}},
		{"without_cpu", HostInfo{FreeCpuCores: 0, FreeMemory: 8, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{constants.HostLoadComputeExhaust, true}},
		{"without_momery", HostInfo{FreeCpuCores: 4, FreeMemory: 0, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{constants.HostLoadComputeExhaust, true}},
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
		{"normal", HostInfo{CpuCores: 4, Memory: 8, FreeCpuCores: 4, FreeMemory: 8, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{true}},
		{"diskused", HostInfo{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Disks: []DiskInfo{{Status: string(constants.DiskInUsed)}}}, want{false}},
		{"diskused2", HostInfo{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Disks: []DiskInfo{{Status: string(constants.DiskExhaust)}}}, want{false}},
		{"diskused3", HostInfo{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 64, Disks: []DiskInfo{{Status: string(constants.DiskExhaust)}, {Status: string(constants.DiskAvailable)}}}, want{false}},
		{"cpuused", HostInfo{CpuCores: 16, Memory: 64, FreeCpuCores: 12, FreeMemory: 64, Disks: []DiskInfo{{Status: string(constants.DiskAvailable)}}}, want{false}},
		{"memoryused", HostInfo{CpuCores: 16, Memory: 64, FreeCpuCores: 16, FreeMemory: 8}, want{false}},
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
