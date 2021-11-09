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
	"testing"
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
		{"valid_X86", "X86", want{nil}},
		{"valid_ARM64", "ARM64", want{nil}},
		{"invalid_x86", "x86", want{errors.New("valid arch type: [ARM64 | X86]")}},
		{"invalid_arm64", "arm64", want{errors.New("valid arch type: [ARM64 | X86]")}},
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
		{"valid_General", "General", want{nil}},
		{"valid_Compute", "Compute", want{nil}},
		{"valid_Storage", "Storage", want{nil}},
		{"invalid_UnDefine", "UnDefine", want{errors.New("valid purpose: [Compute | Storage | General]")}},
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
