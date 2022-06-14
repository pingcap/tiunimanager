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

package structs

import "github.com/pingcap/tiunimanager/common/structs"

///////////////////////////////
// Alloc Request
//////////////////////////////

type BatchAllocRequest struct {
	BatchRequests []AllocReq
}

type AllocReq struct {
	Applicant Applicant
	Requires  []AllocRequirement
}

type Applicant struct {
	HolderId          string
	RequestId         string
	TakeoverOperation bool
}

type AllocRequirement struct {
	Location     structs.Location
	HostExcluded Excluded
	HostFilter   Filter
	Require      Requirement
	Strategy     AllocStrategy
	Count        int32
}

type Excluded struct {
	Hosts []string
}

type Filter struct {
	Arch       string
	Purpose    string
	DiskType   string
	HostTraits int64
}

type Requirement struct {
	Exclusive  bool // The Resource meets the Requirement will be used by exclusive
	DiskReq    DiskRequirement
	ComputeReq ComputeRequirement
	PortReq    []PortRequirement
}

type ComputeRequirement struct {
	ComputeResource
}

type DiskRequirement struct {
	NeedDisk    bool
	Capacity    int32  // Reserved, not used by now
	DiskType    string // Reserved, not used by now
	DiskSpecify string // specify diskId
}

type PortRequirement struct {
	Start   int32
	End     int32
	PortCnt int32
}

type AllocStrategy int32

const (
	RandomRack         AllocStrategy = iota // Require 'Region' and 'Zone', return diff Host
	DiffRackBestEffort                      // Require 'Region' and 'Zone', try best effort to alloc host in diff rack
	UserSpecifyRack                         // Require 'Region' 'Zone' and 'Rack', return diff hosts in Rack
	UserSpecifyHost                         // Return Resource in the Host Specified
	ClusterPorts                            // Returns port range in every host within a region for cluster
)

///////////////////////////////
// Recycle Request
//////////////////////////////

type RecycleType int32

const (
	RecycleHolder  RecycleType = iota // Recycle the resources owned by HolderID
	RecycleOperate                    // Recycle the resources operated in RequestID
	RecycleHost                       // Recycle resources on specified host
)

type RecycleRequire struct {
	RecycleType RecycleType
	HolderID    string
	RequestID   string
	HostID      string
	HostIP      string
	ComputeReq  ComputeRequirement
	PortReq     []PortResource
	DiskReq     []DiskResource
}

type RecycleRequest struct {
	RecycleReqs []RecycleRequire
}
