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

package warehouse

import "github.com/pingcap-inc/tiem/common/resource"

type GetHierarchyReq struct {
	Filter struct {
		Arch string `json:"Arch"`
	} `json:"Filter"`
	Level int32 `json:"Level"`
	Depth int32 `json:"Depth"`
}

type GetHierarchyRsp struct {
	Root resource.HierarchyTreeNode `json:"root"`
}

type GetStocksReq struct {
	StockCondition struct {
		StockLocation struct {
			Region string `json:"Region"`
			Zone   string `json:"Zone"`
			Rack   string `json:"Rack"`
			HostIp string `json:"HostIp"`
		}
		StockHostCondition struct {
			Arch       string `json:"Arch"`
			HostStatus int32  `json:"HostStatus"`
			LoadStat   int32  `json:"LoadStat"`
		}
		StockDiskCondition struct {
			DiskType   string `json:"DiskType"`
			DiskStatus int32  `json:"DiskStatus"`
			Capacity   int32  `json:"Capacity"`
		}
	} `json:"stockCondition"`
}

type GetStocksRsp struct {
	Stocks struct {
		FreeHostCount    int32 `json:"freeHostCount"`
		FreeCpuCores     int32 `json:"freeCpuCores"`
		FreeMemory       int32 `json:"freeMemory"`
		FreeDiskCount    int32 `json:"freeDiskCount"`
		FreeDiskCapacity int32 `json:"freeDiskCapacity"`
	} `json:"stocks"`
}
