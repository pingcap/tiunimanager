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

type ArchType string

//Definition of host architecture names
const (
	ArchX86   ArchType = "X86"
	ArchX8664 ArchType = "X86_64"
	ArchArm   ArchType = "ARM"
	ArchArm64 ArchType = "ARM64"
)

//Constants for importing host information
const (
	ImportHostTemplateFileName string = "hostInfo_template.xlsx"
	ImportHostTemplateFilePath string = "./etc"
)

type HostStatus string

//Definition of host status
const (
	HostWhatever HostStatus = "Whatever"
	HostOnline   HostStatus = "Online"
	HostOffline  HostStatus = "Offline"
	HostDeleted  HostStatus = "Deleted"
)

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

type DiskType string

//Definition of disk type
const (
	NVMeSSD DiskType = "NVMeSSD"
	SSD     DiskType = "SSD"
	SATA    DiskType = "SATA"
)

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

type ResourceLabelCategory int8

//Definition of Resource label category
const (
	UserSpecify ResourceLabelCategory = 0
	Cluster     ResourceLabelCategory = 1
	Component   ResourceLabelCategory = 2
	DiskPerf    ResourceLabelCategory = 3
)

type HierarchyTreeNodeLevel int32

const (
	ROOT HierarchyTreeNodeLevel = iota
	REGION
	ZONE
	RACK
	HOST
	DISK
)
