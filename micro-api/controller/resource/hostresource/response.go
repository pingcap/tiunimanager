
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

package hostresource

type DemoHostInfo struct {
	HostId   string `json:"hostId"`
	HostName string `json:"hostName"`
	HostIp   string `json:"hostIp"`
}

type ImportHostRsp struct {
	HostId string `json:"hostId"`
}

type ImportHostsRsp struct {
	HostIds []string `json:"hostIds"`
}
type ListHostRsp struct {
	Hosts []HostInfo `json:"hosts"`
}

type HostDetailsRsp struct {
	Host HostInfo `json:"host"`
}

type AllocateRsp struct {
	HostName string   `json:"hostName"`
	Ip       string   `json:"ip"`
	UserName string   `json:"userName"`
	Passwd   string   `json:"passwd"`
	CpuCores int32    `json:"cpuCore"`
	Memory   int32    `json:"memory"`
	Disk     DiskInfo `json:"disk"`
}

type AllocHostsRsp struct {
	PdHosts   []AllocateRsp `json:"pdHosts"`
	TidbHosts []AllocateRsp `json:"tidbHosts"`
	TikvHosts []AllocateRsp `json:"tikvHosts"`
}
