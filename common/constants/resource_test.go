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
 ******************************************************************************/

package constants

import (
	"testing"

	"github.com/bmizerany/assert"
	"github.com/pingcap/tiunimanager/common/errors"
)

func Test_ValidArchType(t *testing.T) {
	type want struct {
		err error
	}
	tests := []struct {
		testName string
		archName string
		want     want
	}{
		{"Test_X86", "X86", want{nil}},
		{"Test_X86_64", "X86_64", want{nil}},
		{"Test_ARM", "ARM", want{nil}},
		{"Test_ARM64", "ARM64", want{nil}},
		{"Test_Fail", "x86", want{errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_INVALID_ARCH, "valid arch type: [%s|%s|%s|%s]",
			string(ArchX86), string(ArchX8664), string(ArchArm), string(ArchArm64))}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := ValidArchType(tt.archName)
			assert.Equal(t, tt.want.err, err)
		})
	}
}

func Test_GetArchAlias(t *testing.T) {
	type want struct {
		alias string
	}
	tests := []struct {
		testName string
		archName ArchType
		want     want
	}{
		{"Test_X86", ArchX86, want{"amd64"}},
		{"Test_X86_64", ArchX8664, want{"amd64"}},
		{"Test_ARM", ArchArm, want{"arm64"}},
		{"Test_ARM64", ArchArm64, want{"arm64"}},
		{"Test_Fail", ArchType("xx"), want{""}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			alias := GetArchAlias(tt.archName)
			assert.Equal(t, tt.want.alias, alias)
		})
	}
}

func Test_ValidStatus(t *testing.T) {
	type want struct {
		valid bool
	}
	tests := []struct {
		testName string
		status   HostStatus
		want     want
	}{
		{"Test_Online", HostOnline, want{true}},
		{"Test_Offline", HostOffline, want{true}},
		{"Test_Deleted", HostDeleted, want{true}},
		{"Test_Init", HostInit, want{true}},
		{"Test_Failed", HostFailed, want{true}},
		{"Test_Deleting", HostDeleting, want{true}},
		{"Test_OtherStatus", HostWhatever, want{false}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			valid := tt.status.IsValidStatus()
			assert.Equal(t, tt.want.valid, valid)
		})
	}
}

func Test_ValidLoadStatus(t *testing.T) {
	type want struct {
		valid bool
	}
	tests := []struct {
		testName string
		status   HostLoadStatus
		want     want
	}{
		{"Test_Loadless", HostLoadLoadLess, want{true}},
		{"Test_Inused", HostLoadInUsed, want{true}},
		{"Test_Exhuast", HostLoadExhaust, want{true}},
		{"Test_Com_Exhuast", HostLoadComputeExhaust, want{true}},
		{"Test_Disk_Exhuast", HostLoadDiskExhaust, want{true}},
		{"Test_Excluded", HostLoadExclusive, want{true}},
		{"Test_OtherStatus", HostLoadStatusWhatever, want{false}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			valid := tt.status.IsValidLoadStatus()
			assert.Equal(t, tt.want.valid, valid)
		})
	}
}

func Test_ValidDiskStatus(t *testing.T) {
	type want struct {
		valid bool
	}
	tests := []struct {
		testName string
		status   DiskStatus
		want     want
	}{
		{"Test_Available", DiskAvailable, want{true}},
		{"Test_Inused", DiskInUsed, want{true}},
		{"Test_Reserved", DiskReserved, want{true}},
		{"Test_Exhuast", DiskExhaust, want{true}},
		{"Test_Error", DiskError, want{true}},
		{"Test_OtherStatus", DiskStatusWhatever, want{false}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			valid := tt.status.IsValidStatus()
			assert.Equal(t, tt.want.valid, valid)
		})
	}
}

func Test_ValidDiskType(t *testing.T) {
	type want struct {
		err error
	}
	tests := []struct {
		testName string
		disktype string
		want     want
	}{
		{"Test_NVMeSSD", "NVMeSSD", want{nil}},
		{"Test_SSD", "SSD", want{nil}},
		{"Test_SATA", "SATA", want{nil}},
		{"Test_Fail", "sata", want{errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_INVALID_PURPOSE, "valid disk type: [%s|%s|%s]",
			string(NVMeSSD), string(SSD), string(SATA))}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := ValidDiskType(tt.disktype)
			assert.Equal(t, tt.want.err, err)
		})
	}
}

func Test_ValidPurposeType(t *testing.T) {
	type want struct {
		err error
	}
	tests := []struct {
		testName string
		purpose  string
		want     want
	}{
		{"Test_Compute", "Compute", want{nil}},
		{"Test_Storage", "Storage", want{nil}},
		{"Test_Schedule", "Schedule", want{nil}},
		{"Test_Fail", "sata", want{errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_INVALID_PURPOSE, "valid purpose name: [%s|%s|%s]",
			string(PurposeCompute), string(PurposeStorage), string(PurposeSchedule))}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := ValidPurposeType(tt.purpose)
			assert.Equal(t, tt.want.err, err)
		})
	}
}
