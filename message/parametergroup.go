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

package message

import (
	"github.com/pingcap-inc/tiem/common/structs"
)

type QueryParameterGroupReq struct {
	structs.PageRequest
	ID         string `json:"id"`
	Name       string `json:"name" form:"name" example:"default"`
	DBType     int32  `json:"dbType" form:"dbType" example:"0"`
	HasDefault int32  `json:"hasDefault" form:"hasDefault" example:"0"`
	Version    string `json:"version" form:"version" example:"v5.0"`
	Spec       string `json:"spec" form:"spec" example:"8C16G"`
	HasDetail  bool   `json:"hasDetail" form:"hasDetail" example:"false"`
}

type QueryParameterGroupResp struct {
	ParamGroupID string                                `json:"paramGroupId" example:"1"`
	Name         string                                `json:"name" example:"default"`
	DBType       int32                                 `json:"dbType" example:"0"`
	HasDefault   int32                                 `json:"hasDefault" example:"1"`
	Version      string                                `json:"version" example:"v5.0"`
	Spec         string                                `json:"spec" example:"8C16G"`
	GroupType    int32                                 `json:"groupType" example:"0"`
	Note         string                                `json:"note" example:"default param group"`
	CreatedAt    int64                                 `json:"createTime" example:"1636698675"`
	UpdatedAt    int64                                 `json:"updateTime" example:"1636698675"`
	Params       []structs.ParameterGroupParameterInfo `json:"params"`
}

type CreateParameterGroupReq struct {
	Name       string                                      `json:"name" example:"8C16GV4_default"`
	DBType     int32                                       `json:"dbType" example:"1"`
	HasDefault int32                                       `json:"hasDefault" example:"1"`
	Version    string                                      `json:"version" example:"v5.0"`
	Spec       string                                      `json:"spec" example:"8C16G"`
	GroupType  int32                                       `json:"groupType" example:"1"`
	Note       string                                      `json:"note" example:"default param group"`
	Params     []structs.ParameterGroupParameterSampleInfo `json:"params"`
}

type CreateParameterGroupResp struct {
	ParamGroupId string `json:"paramGroupId" example:"1"`
}

type DeleteParameterGroupReq struct {
	ParamGroupId string `json:"paramGroupId" example:"1"`
}

type DeleteParameterGroupResp struct {
	ParamGroupId string `json:"paramGroupId" example:"1"`
}

type UpdateParameterGroupReq struct {
	ID      string                                      `json:"id"`
	Name    string                                      `json:"name" example:"8C16GV4_default"`
	Version string                                      `json:"version" example:"v5.0"`
	Spec    string                                      `json:"spec" example:"8C16G"`
	Note    string                                      `json:"note" example:"default param group"`
	Params  []structs.ParameterGroupParameterSampleInfo `json:"params"`
}

type UpdateParameterGroupResp struct {
	ParamGroupId string `json:"paramGroupId" example:"1"`
}

type CopyParameterGroupReq struct {
	ID   string `json:"id"`
	Name string `json:"name" example:"8C16GV4_copy"`
	Note string `json:"note" example:"copy param group"`
}

type CopyParameterGroupResp struct {
	ParamGroupId string `json:"paramGroupId" example:"1"`
}

type ApplyParameterGroupReq struct {
	ClusterID    string `json:"clusterId" example:"123"`
	ParamGroupID string `json:"paramGroupID" example:"123"`
}

type ApplyParameterGroupResp struct {
	ParamGroupId                  string `json:"paramGroupId" example:"1"`
	structs.AsyncTaskWorkFlowInfo `json:"workFlowID"`
}
