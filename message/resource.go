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
	structs.HostFilter
}

func (req *QueryHostsReq) GetHostFilter() *structs.HostFilter {
	return &(req.HostFilter)
}
func (req *QueryHostsReq) GetPage() *structs.PageRequest {
	return &(req.PageRequest)
}

type QueryHostsResp struct {
	Hosts []structs.HostInfo `json:"hosts"`
}

type ImportHostsReq struct {
	Hosts []structs.HostInfo `json:"hosts"`
}

type ImportHostsResp struct {
	FlowInfo []structs.AsyncTaskWorkFlowInfo `json:"flowInfo"`
	HostIDS  []string                        `json:"hostIds"`
}

type DeleteHostsReq struct {
	HostIDs []string `json:"hostIds"`
}

type DeleteHostsResp struct {
	FlowInfo []structs.AsyncTaskWorkFlowInfo `json:"flowInfo"`
}

type UpdateHostReservedReq struct {
	HostIDs  []string `json:"hostIds"`
	Reserved bool     `json:"reserved"`
}

type UpdateHostReservedResp struct {
}

type UpdateHostStatusReq struct {
	HostIDs []string `json:"hostIds"`
	Status  string   `json:"status"`
}

type UpdateHostStatusResp struct {
}

type GetHierarchyReq struct {
	structs.HostFilter
	Level int `json:"Level"` // [1:Region, 2:Zone, 3:Rack, 4:Host]
	Depth int `json:"Depth"`
}

func (req *GetHierarchyReq) GetHostFilter() *structs.HostFilter {
	return &(req.HostFilter)
}

type GetHierarchyResp struct {
	Root structs.HierarchyTreeNode `json:"root"`
}

type GetStocksReq struct {
	structs.Location
	structs.HostFilter
	structs.DiskFilter
}

func (req *GetStocksReq) GetLocation() *structs.Location {
	return &(req.Location)
}
func (req *GetStocksReq) GetHostFilter() *structs.HostFilter {
	return &(req.HostFilter)
}
func (req *GetStocksReq) GetDiskFilter() *structs.DiskFilter {
	return &(req.DiskFilter)
}

type GetStocksResp struct {
	// map[zone] -> stocks
	Stocks map[string]*structs.Stocks `json:"stocks"`
}

type DownloadHostTemplateFileReq struct {
}

type DownloadHostTemplateFileResp struct {
}
