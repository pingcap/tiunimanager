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

package message

type CheckPlatformReq struct {
	DisplayMode string `json:"displayMode" form:"displayMode"`
}

type CheckPlatformRsp struct {
	Tenants map[string]TenantCheck `json:"tenants"`
	Hosts   map[string]HostCheck   `json:"hosts"`
}

type UnionString struct {
	Valid         bool   `json:"valid"`
	RealValue     string `json:"realValue"`
	ExpectedValue string `json:"expectedValue"`
}

type UnionInt struct {
	Valid         bool `json:"valid"`
	RealValue     int  `json:"realValue"`
	ExpectedValue int  `json:"expectedValue"`
}

type TenantCheck struct {
	ClusterCount int      `json:"clusterCount"`
	CPURatio     float32  `json:"cpuRatio"`
	MemoryRatio  float32  `json:"memoryRatio"`
	StorageRatio float32  `json:"storageRatio"`
	MaxCluster   UnionInt `json:"maxCluster"`
}

type HostCheck struct {
}
