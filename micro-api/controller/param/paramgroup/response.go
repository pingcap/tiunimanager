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

/*******************************************************************************
 * @File: response
 * @Description: wrapping response structures
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/17 10:26
*******************************************************************************/

package paramgroup

// QueryParamGroupResp
// @Description: query param group
type QueryParamGroupResp struct {
	ID         string        `json:"id" example:"abc"`
	Name       string        `json:"name" example:"default"`
	DbType     int           `json:"dbType" example:"0"`
	HasDefault bool          `json:"hasDefault" example:"true"`
	Version    string        `json:"version" example:"v5.0"`
	Spec       string        `json:"spec" example:"8C16G"`
	GroupType  int           `json:"groupType" example:"0"`
	Note       string        `json:"note" example:"default param group"`
	CreatedAt  int64         `json:"createTime" example:"1636698675"`
	UpdatedAt  int64         `json:"updateTime" example:"1636698675"`
	Params     []ParamDetail `json:"params"`
}

type CommonParamGroupResp struct {
	ID string `json:"id" example:"abc"`
}

type ApplyParamGroupResp struct {
	ID        string `json:"id" example:"abc"`
	ClusterId string `json:"clusterId" example:"123"`
}

// ParamDetail
// @Description: param detail struct
type ParamDetail struct {
	ID            string   `json:"id" example:"123"`
	Name          string   `json:"name" example:"binlog_size"`
	ComponentType string   `json:"componentType" example:"tidb"`
	Type          int      `json:"type" example:"0"`
	Unit          string   `json:"unit" example:"mb"`
	Range         []string `json:"range" example:"1, 1000"`
	HasReboot     int      `json:"hasReboot" example:"0"`
	DefaultValue  string   `json:"defaultValue" example:"1"`
	Description   string   `json:"description" example:"binlog cache size"`
	Note          string   `json:"note" example:"binlog cache size"`
	CreatedAt     int64    `json:"createTime" example:"1636698675"`
	UpdatedAt     int64    `json:"updateTime" example:"1636698675"`
}
