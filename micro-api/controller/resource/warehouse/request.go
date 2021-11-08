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

package warehouse

type HostFilter struct {
	Arch string `json:"Arch"`
}

type GetHierarchyReq struct {
	Filter HostFilter `json:"Filter"`
	Level  int32      `json:"Level"`
	Depth  int32      `json:"Depth"`
}

type StockLocation struct {
	Region string `json:"Region"`
	Zone   string `json:"Zone"`
	Rack   string `json:"Rack"`
	HostIp string `json:"HostIp"`
}

type StockHostCondition struct {
	Arch       string `json:"Arch"`
	HostStatus int32  `json:"HostStatus"`
	LoadStat   int32  `json:"LoadStat"`
}

type StockDiskCondition struct {
	DiskType   string `json:"DiskType"`
	DiskStatus int32  `json:"DiskStatus"`
	Capacity   int32  `json:"Capacity"`
}

type StockCondition struct {
	StockLocation
	StockHostCondition
	StockDiskCondition
}

type GetStocksReq struct {
	Condition StockCondition `json:"stockCondition"`
}
