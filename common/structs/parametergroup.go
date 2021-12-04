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
 * @File: parametergroup.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package structs

type ParameterGroupParameterInfo struct {
	ID            string   `json:"paramId" example:"1"`
	Name          string   `json:"name" example:"binlog_size"`
	ComponentType string   `json:"componentType" example:"tidb"`
	Type          int32    `json:"type" example:"0"`
	Unit          string   `json:"unit" example:"mb"`
	Range         []string `json:"range" example:"1, 1000"`
	HasReboot     int32    `json:"hasReboot" example:"0"`
	DefaultValue  string   `json:"defaultValue" example:"1"`
	Description   string   `json:"description" example:"binlog cache size"`
	Note          string   `json:"note" example:"binlog cache size"`
	CreatedAt     int64    `json:"createTime" example:"1636698675"`
	UpdatedAt     int64    `json:"updateTime" example:"1636698675"`
}

type ParameterGroupParameterSampleInfo struct {
	ID           string `json:"paramId" example:"123"`
	DefaultValue string `json:"defaultValue" example:"1"`
	Note         string `json:"description" example:"binlog cache size"`
}
