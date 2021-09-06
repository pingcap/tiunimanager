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

	"github.com/asim/go-micro/v3/client"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
)

type ResultMark struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CommonResult struct {
	ResultMark
	Data interface{} `json:"data"`
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

var DefaultTimeout = func(o *client.CallOptions) {
	o.RequestTimeout = time.Second * 30
	o.DialTimeout = time.Second * 30
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
	Total     float32 `json:"total"`
	Used      float32 `json:"used"`
	UsageRate float32 `json:"usageRate"`
}

type Page struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

var DefaultPageRequest PageRequest = PageRequest{
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

func (o *Operator) ConvertToDTO() *cluster.OperatorDTO {
	return &cluster.OperatorDTO{
		Id:       o.OperatorId,
		Name:     o.OperatorName,
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
		Total:    int(dto.Total),
	}
}

func ParseUsageFromDTO(dto *cluster.UsageDTO) (usage *Usage) {
	usage = &Usage{
		Total:     dto.Total,
		Used:      dto.Used,
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
