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
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/micro-api/security"
	cluster "github.com/pingcap/tiem/micro-cluster/proto"
	"net/http"
	"time"
)

type ResultMark struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
}

type CommonResult struct {
	ResultMark
	Data    interface{}  `json:"data"`
}

type ResultWithPage struct {
	ResultMark
	Data interface{} `json:"data"`
	Page Page        `json:"page"`
}

func BuildCommonResult(code int, message string, data interface{}) (result *CommonResult) {
	result = &CommonResult{}
	result.Code = code
	result.Message = message
	result.Data = data

	return
}

func BuildResultWithPage(code int, message string, page *Page, data interface{}) (result *ResultWithPage) {
	result = &ResultWithPage{}
	result.Code = code
	result.Message = message
	result.Data = data
	result.Page = *page

	return
}


func Success(data interface{}) *CommonResult {
	return &CommonResult{ResultMark: ResultMark{0, "OK"}, Data: data}
}

func SuccessWithPage(data interface{}, page Page) *ResultWithPage {
	return &ResultWithPage{ResultMark: ResultMark{0, "OK"}, Data: data, Page: page}
}

func Fail(code int, message string) *CommonResult {
	return &CommonResult{ResultMark{code, message}, struct{}{}}
}

type Usage struct {
	Total     float32
	Used      float32
	UsageRate float32
}

type Page struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	Total     int    `json:"total"`
}

type PageRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
}

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, Success("hello"))
}

func HelloPage(c *gin.Context) {
	c.JSON(http.StatusOK, Success("hello world"))
}

type StatusInfo struct {
	StatusCode			string
	StatusName			string
	InProcessFlowId 	int
	CreateTime			time.Time
	UpdateTime			time.Time
	DeleteTime			time.Time
}

type Operator struct {
	ManualOperator 		bool
	OperatorName 		string
	OperatorId   		string
	TenantId			string
}

func GetOperator(c *gin.Context) *Operator {
	v, _ := c.Get(security.VisitorIdentityKey)
	
	visitor, _ := v.(*security.VisitorIdentity)

	return &Operator{
		ManualOperator: true,
		OperatorId: visitor.AccountId,
		OperatorName: visitor.AccountName,
		TenantId: visitor.TenantId,
	}
}

func (o *Operator) ConvertToDTO() *cluster.OperatorDTO{
	return &cluster.OperatorDTO{
		Id: o.OperatorId,
		Name: o.OperatorName,
		TenantId: o.TenantId,
	}
}

func (p *PageRequest) ConvertToDTO() *cluster.PageDTO {
	return &cluster.PageDTO{
		Page:     int32(p.Page),
		PageSize: int32(p.PageSize),
	}
}

func ParsePageFromDTO(dto *cluster.PageDTO) *Page {
	return &Page{
		Page:     int(dto.Page),
		PageSize: int(dto.PageSize),
		Total: int(dto.Total),
	}
}

func ParseUsageFromDTO(dto *cluster.UsageDTO) (usage *Usage) {
	usage = &Usage{
		Total: dto.Total,
		Used: dto.Used,
		UsageRate: dto.UsageRate,
	}
	return
}

func ConvertVersionDTO(code string) (dto *cluster.ClusterVersionDTO) {
	version := knowledge.ClusterVersionFromCode(code)

	dto = &cluster.ClusterVersionDTO{
		Code: version.Code,
		Name: version.Name,
	}

	return
}

func ConvertTypeDTO(code string) (dto *cluster.ClusterTypeDTO) {
	t := knowledge.ClusterTypeFromCode(code)

	dto = &cluster.ClusterTypeDTO{
		Code: t.Code,
		Name: t.Name,
	}
	return
}
