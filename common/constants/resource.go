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

/*******************************************************************************
 * @File: resource.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package constants

import (
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
)

type ArchType string

//Definition of host architecture names
const (
	ArchX86   ArchType = "X86"
	ArchX8664 ArchType = "X86_64"
	ArchArm   ArchType = "ARM"
	ArchArm64 ArchType = "ARM64"
)

func ValidArchType(arch string) error {
	if arch == string(ArchX86) || arch == string(ArchX8664) || arch == string(ArchArm) || arch == string(ArchArm64) {
		return nil
	}
	return framework.NewTiEMErrorf(common.TIEM_RESOURCE_INVALID_ARCH, "valid arch type: [%s|%s|%s|%s]",
		string(ArchX86), string(ArchX8664), string(ArchArm), string(ArchArm64))
}

type HostStatus string

//Definition of host status
const (
	HostWhatever HostStatus = "Whatever"
	HostOnline   HostStatus = "Online"
	HostOffline  HostStatus = "Offline"
	HostDeleted  HostStatus = "Deleted"
)

func (s HostStatus) IsValidStatus() bool {
	return (s == HostOnline ||
		s == HostOffline ||
		s == HostDeleted)
}

type HostLoadStatus string

//Definition of host load status
const (
	HostLoadStatusWhatever HostLoadStatus = "Whatever"
	HostLoadLoadLess       HostLoadStatus = "LoadLess"
	HostLoadInUsed         HostLoadStatus = "InUsed"
	HostLoadExhaust        HostLoadStatus = "Exhaust"
	HostLoadComputeExhaust HostLoadStatus = "ComputeExhaust"
	HostLoadDiskExhaust    HostLoadStatus = "DiskExhaust"
	HostLoadExclusive      HostLoadStatus = "Exclusive"
)

func (s HostLoadStatus) IsValidLoadStatus() bool {
	return (s == HostLoadLoadLess ||
		s == HostLoadInUsed ||
		s == HostLoadExhaust ||
		s == HostLoadComputeExhaust ||
		s == HostLoadDiskExhaust ||
		s == HostLoadExclusive)
}

type DiskType string

//Definition of disk type
const (
	NVMeSSD DiskType = "NVMeSSD"
	SSD     DiskType = "SSD"
	SATA    DiskType = "SATA"
)

func ValidDiskType(diskType string) error {
	if diskType == string(NVMeSSD) || diskType == string(SSD) || diskType == string(SATA) {
		return nil
	}
	return framework.NewTiEMErrorf(common.TIEM_RESOURCE_INVALID_PURPOSE, "valid disk type: [%s|%s|%s]",
		string(NVMeSSD), string(SSD), string(SATA))
}

type DiskStatus string

//Definition of disk status
const (
	DiskStatusWhatever DiskStatus = "Whatever"
	DiskAvailable      DiskStatus = "Available"
	DiskReserved       DiskStatus = "Reserved"
	DiskInUsed         DiskStatus = "InUsed"
	DiskExhaust        DiskStatus = "Exhaust"
	DiskError          DiskStatus = "Error"
)

type PurposeType string

//Types of purpose available
const (
	PurposeCompute  PurposeType = "Compute"
	PurposeStorage  PurposeType = "Storage"
	PurposeSchedule PurposeType = "Schedule"
)

func ValidPurposeType(p string) error {
	if p == string(PurposeCompute) || p == string(PurposeStorage) || p == string(PurposeSchedule) {
		return nil
	}
	return framework.NewTiEMErrorf(common.TIEM_RESOURCE_INVALID_PURPOSE, "valid purpose name: [%s|%s|%s]",
		string(PurposeCompute), string(PurposeStorage), string(PurposeSchedule))
}

type ResourceLabelCategory int8

//TODO It is recommended to move to the resource module in cluster-server
//Definition of Resource label category
const (
	UserSpecify ResourceLabelCategory = 0
	Cluster     ResourceLabelCategory = 1
	Component   ResourceLabelCategory = 2
	DiskPerf    ResourceLabelCategory = 3
)

type HierarchyTreeNodeLevel int32

//TODO It is recommended to move to the resource module in cluster-server

const (
	ROOT HierarchyTreeNodeLevel = iota
	REGION
	ZONE
	RACK
	HOST
	DISK
)
