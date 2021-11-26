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

// Package classification TiEM API.
//
// the purpose of this application is to provide an application
// that is using plain go code to define an API
//
// This should demonstrate all the possible comment annotations
// that are available to turn go code into a fully compliant swagger 2.0 spec

// @title TiEM API
// @version v1
// @description This is a sample TiEM-API server.
// @BasePath /api/v1

// swagger:meta
package controller

import (
	"net/http"
	"time"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/asim/go-micro/v3/client"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
)

// DefaultTimeout
// todo adjust timeout for async flow task
var DefaultTimeout = func(o *client.CallOptions) {
	o.RequestTimeout = time.Minute * 5
	o.DialTimeout = time.Minute * 5
}

type Usage struct {
	Total     float32 `json:"total"`
	Used      float32 `json:"used"`
	UsageRate float32 `json:"usageRate"`
}

type Page struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

var DefaultPageRequest = PageRequest{
	1,
	20,
}

type PageRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"pageSize" form:"pageSize"`
}

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, Success("hello"))
}

func HelloPage(c *gin.Context) {
	c.JSON(http.StatusOK, Success("hello world"))
}

type StatusInfo struct {
	StatusCode      string    `json:"statusCode"`
	StatusName      string    `json:"statusName"`
	InProcessFlowId int       `json:"inProcessFlowId"`
	CreateTime      time.Time `json:"createTime"`
	UpdateTime      time.Time `json:"updateTime"`
	DeleteTime      time.Time `json:"deleteTime"`
}

type Operator struct {
	ManualOperator bool   `json:"manualOperator"`
	OperatorName   string `json:"operatorName"`
	OperatorId     string `json:"operatorId"`
	TenantId       string `json:"tenantId"`
}

func GetOperator(c *gin.Context) *Operator {
	v, _ := c.Get(interceptor.VisitorIdentityKey)

	visitor, _ := v.(*interceptor.VisitorIdentity)

	return &Operator{
		ManualOperator: true,
		OperatorId:     visitor.AccountId,
		OperatorName:   visitor.AccountName,
		TenantId:       visitor.TenantId,
	}
}

func (o *Operator) ConvertToDTO() *clusterpb.OperatorDTO {
	return &clusterpb.OperatorDTO{
		Id:       o.OperatorId,
		Name:     o.OperatorName,
		TenantId: o.TenantId,
	}
}

func (p *PageRequest) ConvertToDTO() *clusterpb.PageDTO {
	return &clusterpb.PageDTO{
		Page:     int32(p.Page),
		PageSize: int32(p.PageSize),
	}
}

func ParsePageFromDTO(dto *clusterpb.PageDTO) *Page {
	return &Page{
		Page:     int(dto.Page),
		PageSize: int(dto.PageSize),
		Total:    int(dto.Total),
	}
}

func ParseUsageFromDTO(dto *clusterpb.UsageDTO) (usage *Usage) {
	usage = &Usage{
		Total:     dto.Total,
		Used:      dto.Used,
		UsageRate: dto.UsageRate,
	}
	return
}

func ConvertVersionDTO(code string) (dto *clusterpb.ClusterVersionDTO) {
	version := knowledge.ClusterVersionFromCode(code)

	dto = &clusterpb.ClusterVersionDTO{
		Code: version.Code,
		Name: version.Name,
	}

	return
}

func ConvertTypeDTO(code string) (dto *clusterpb.ClusterTypeDTO) {
	t := knowledge.ClusterTypeFromCode(code)

	dto = &clusterpb.ClusterTypeDTO{
		Code: t.Code,
		Name: t.Name,
	}
	return
}

func ConvertRecoverInfoDTO(sourceClusterId string, backupRecordId int64) (dto *clusterpb.RecoverInfoDTO) {
	return &clusterpb.RecoverInfoDTO{
		SourceClusterId: sourceClusterId,
		BackupRecordId:  backupRecordId,
	}
}
