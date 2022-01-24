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
	Name           string `json:"name" form:"name" example:"default"`
	DBType         int    `json:"dbType" form:"dbType" example:"0" enums:"0,1,2"`
	HasDefault     int    `json:"hasDefault" form:"hasDefault" example:"0" enums:"0,1,2"`
	ClusterVersion string `json:"clusterVersion" form:"clusterVersion" example:"v5.0"`
	ClusterSpec    string `json:"clusterSpec" form:"clusterSpec" example:"8C16G"`
	HasDetail      bool   `json:"hasDetail" form:"hasDetail" example:"false"`
}

type QueryParameterGroupResp struct {
	ParameterGroupInfo
}

type DetailParameterGroupReq struct {
	ParamGroupID string `json:"paramGroupId" swaggerignore:"true" validate:"required,min=1,max=64"`
	ParamName    string `json:"paramName" form:"paramName"`
}

type DetailParameterGroupResp struct {
	ParameterGroupInfo
}

type CreateParameterGroupReq struct {
	Name           string                                      `json:"name" example:"8C16GV4_default"`
	DBType         int                                         `json:"dbType" example:"1" enums:"1,2"`
	ClusterVersion string                                      `json:"clusterVersion" example:"v5.0"`
	ClusterSpec    string                                      `json:"clusterSpec" example:"8C16G"`
	GroupType      int                                         `json:"groupType" example:"1" enums:"1,2"`
	Note           string                                      `json:"note" example:"default param group"`
	Params         []structs.ParameterGroupParameterSampleInfo `json:"params" validate:"required"`
	AddParams      []ParameterInfo                             `json:"addParams"`
}

type CreateParameterGroupResp struct {
	ParamGroupID string `json:"paramGroupId" example:"1"`
}

type DeleteParameterGroupReq struct {
	ParamGroupID string `json:"paramGroupId" example:"1" validate:"required,min=1,max=64"`
}

type DeleteParameterGroupResp struct {
	ParamGroupID string `json:"paramGroupId" example:"1"`
}

type UpdateParameterGroupReq struct {
	ParamGroupID   string                                      `json:"paramGroupId" swaggerignore:"true" validate:"required,min=1,max=64"`
	Name           string                                      `json:"name" example:"8C16GV4_default"`
	ClusterVersion string                                      `json:"clusterVersion" example:"v5.0"`
	ClusterSpec    string                                      `json:"clusterSpec" example:"8C16G"`
	Note           string                                      `json:"note" example:"default param group"`
	Params         []structs.ParameterGroupParameterSampleInfo `json:"params" validate:"required"`
	AddParams      []ParameterInfo                             `json:"addParams"`
	DelParams      []string                                    `json:"delParams" example:"1"`
}

type UpdateParameterGroupResp struct {
	ParamGroupID string `json:"paramGroupId" example:"1"`
}

type CopyParameterGroupReq struct {
	ParamGroupID string `json:"paramGroupId" swaggerignore:"true" validate:"required,min=1,max=64"`
	Name         string `json:"name" example:"8C16GV4_copy" validate:"required,min=1,max=64"`
	Note         string `json:"note" example:"copy param group"`
}

type CopyParameterGroupResp struct {
	ParamGroupID string `json:"paramGroupId" example:"1"`
}

type ApplyParameterGroupReq struct {
	ParamGroupId string   `json:"paramGroupId" example:"123" swaggerignore:"true" validate:"required,min=1,max=64"`
	ClusterID    string   `json:"clusterId" example:"123" validate:"required,min=8,max=64"`
	Reboot       bool     `json:"reboot"`
	Nodes        []string `json:"nodes" swaggerignore:"true"`
}

type ApplyParameterGroupResp struct {
	ClusterID    string `json:"clusterId" example:"123"`
	ParamGroupID string `json:"paramGroupId" example:"123"`
	structs.AsyncTaskWorkFlowInfo
}

type ParameterGroupInfo struct {
	ParamGroupID   string                                `json:"paramGroupId" example:"1"`
	Name           string                                `json:"name" example:"default"`
	DBType         int                                   `json:"dbType" example:"1" enums:"1,2"`
	HasDefault     int                                   `json:"hasDefault" example:"1" enums:"1,2"`
	ClusterVersion string                                `json:"clusterVersion" example:"v5.0"`
	ClusterSpec    string                                `json:"clusterSpec" example:"8C16G"`
	GroupType      int                                   `json:"groupType" example:"0" enums:"1,2"`
	Note           string                                `json:"note" example:"default param group"`
	CreatedAt      int64                                 `json:"createTime" example:"1636698675"`
	UpdatedAt      int64                                 `json:"updateTime" example:"1636698675"`
	Params         []structs.ParameterGroupParameterInfo `json:"params"`
}

type ParameterInfo struct {
	Category       string   `json:"category" example:"log"`
	Name           string   `json:"name" example:"binlog_size"`
	InstanceType   string   `json:"instanceType" example:"TiDB"`
	SystemVariable string   `json:"systemVariable" example:"log.binlog_size"`
	Type           int      `json:"type" example:"0"`
	Unit           string   `json:"unit" example:"mb"`
	Range          []string `json:"range" example:""`
	HasReboot      int      `json:"hasReboot" example:"0"`
	HasApply       int      `json:"hasApply" example:"1"`
	UpdateSource   int      `json:"updateSource" example:"0"`
	ReadOnly       int      `json:"readOnly" example:"0"`
	Description    string   `json:"description" example:"binlog size"`
	DefaultValue   string   `json:"defaultValue" example:"1024"`
	Note           string   `json:"note"`
}
