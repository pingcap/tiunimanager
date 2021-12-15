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

type ZoneHostStockRsp struct {
	AvailableStocks map[string][]ZoneHostStock
}

type DomainResource struct {
	ZoneName string `json:"zoneName"`
	ZoneCode string `json:"zoneCode"`
	Purpose  string `json:"purpose"`
	SpecName string `json:"specName"`
	SpecCode string `json:"specCode"`
	Count    int32  `json:"count"`
}

type DomainResourceRsp struct {
	Resources []DomainResource `json:"resources"`
}

type Node struct {
	Code     string `json:"Code"`
	Prefix   string `json:"Prefix"`
	Name     string `json:"Name"`
	SubNodes []Node `json:"SubNodes"`
}

type GetHierarchyRsp struct {
	Root Node `json:"Root"`
}

type Stocks struct {
	FreeHostCount    int32 `json:"freeHostCount"`
	FreeCpuCores     int32 `json:"freeCpuCores"`
	FreeMemory       int32 `json:"freeMemory"`
	FreeDiskCount    int32 `json:"freeDiskCount"`
	FreeDiskCapacity int32 `json:"freeDiskCapacity"`
}

type GetStocksRsp struct {
	Stocks Stocks `json:"stocks"`
}