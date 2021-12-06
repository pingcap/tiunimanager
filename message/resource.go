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

package message

import (
	"github.com/pingcap-inc/tiem/common/structs"
)

type QueryHostsReq struct {
	structs.PageRequest
	Filter structs.HostFilter `json:"Filter"`
}

type QueryHostsResp struct {
	Hosts []structs.HostInfo `json:"hosts"`
}

type ImportHostsReq struct {
	Hosts []structs.HostInfo `json:"hosts"`
}

type ImportHostsResp struct {
	HostIDS []string `json:"hostIds"`
}

type DeleteHostsReq struct {
	HostID []string `json:"hostIds"`
}

type DeleteHostsResp struct {
	Hosts struct {
		HostID string `json:"hostId"`
		Status string `json:"status"`
	} `json:"hosts"`
}

type UpdateHostReservedReq struct {
	HostIDs  []string `json:"hostIds"`
	Reserved *bool    `json:"reserved"`
}

type UpdateHostReservedResp struct {
}

type UpdateHostStatusReq struct {
	HostIDs []string `json:"hostIds"`
	Status  *int32   `json:"status"`
}

type UpdateHostStatusResp struct {
}

type GetHierarchyReq struct {
	Filter struct {
		Arch string `json:"Arch"`
	} `json:"Filter"`
	Level int32 `json:"Level"`
	Depth int32 `json:"Depth"`
}

type GetHierarchyResp struct {
	Root structs.HierarchyTreeNode `json:"root"`
}

type GetStocksReq struct {
	StockCondition struct {
		StockLocation      structs.Location
		StockHostCondition structs.HostFilter
		StockDiskCondition structs.DiskFilter
	} `json:"stockCondition"`
}

type GetStocksResp struct {
	Stocks structs.Stocks `json:"stocks"`
}

type DownloadHostTemplateFileReq struct {
}

type DownloadHostTemplateFileResp struct {
}
