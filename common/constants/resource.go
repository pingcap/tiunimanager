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

type ArchType string

//Definition of host architecture names
const (
	ArchX86   ArchType = "X86"
	ArchX8664 ArchType = "X86_64"
	ArchArm   ArchType = "ARM"
	ArchArm64 ArchType = "ARM64"
)

//Constants for importing host information
//TODO It is recommended to move to the resource module in api-server
const (
	ImportHostTemplateFileName string = "hostInfo_template.xlsx"
	ImportHostTemplateFilePath string = "./etc"
)

type HostStatus string

//Definition of host status
const (
	HostWhatever HostStatus = "Whatever"
	HostOnline              = "Online"
	HostOffline             = "Offline"
	HostDeleted             = "Deleted"
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

//TODO It is recommended to move to the resource module in cluster-server
//Definition of Resource label category
const (
	UserSpecify ResourceLabelCategory = 0
	Cluster     ResourceLabelCategory = 1
	Component   ResourceLabelCategory = 2
	DiskPerf    ResourceLabelCategory = 3
)

type FailureDomain int32

//TODO It is recommended to move to the resource module in cluster-server
const (
	ROOT             FailureDomain = iota
	stockRegionLevel               = 1
	ZONE
	RACK
	HOST
	DISK
)
