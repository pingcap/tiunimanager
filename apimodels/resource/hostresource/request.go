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

package hostresource

import (
	"github.com/pingcap-inc/tiem/common"
	"github.com/pingcap-inc/tiem/common/resource"
)

type ImportHostsReq struct {
}

type ImportHostsRsp struct {
}

type DeleteHostsReq struct {
	HostID []string `json:"hostIds"`
}

type DeleteHostsRsp struct {
	Hosts struct {
		HostID string `json:"hostId"`
		Status string `json:"status"`
	} `json:"hosts"`
}

type UpdateHostReservedReq struct {
	HostIDs  []string `json:"hostIds"`
	Reserved *bool    `json:"reserved"`
}

type UpdateHostReservedRsp struct {
}

type UpdateHostStatusReq struct {
	HostIDs []string `json:"hostIds"`
	Status  *int32   `json:"status"`
}

type UpdateHostStatusRsp struct {
}

type QueryHostsReq struct {
	common.PageRequest
	Filter resource.HostFilter `json:"Filter"`
}

type QueryHostsRsp struct {
	Hosts []resource.HostInfo `json:"hosts"`
}

type DownloadHostTemplateFileReq struct {
}

type DownloadHostTemplateFileRsp struct {
}
