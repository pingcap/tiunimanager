// Package classification TiCP API.
//
// the purpose of this application is to provide an application
// that is using plain go code to define an API
//
// This should demonstrate all the possible comment annotations
// that are available to turn go code into a fully compliant swagger 2.0 spec

// @title TiCP API
// @version v1
// @description This is a sample TiCP-API server.
// @BasePath /api/v1

// swagger:meta
package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
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

func Success(data interface{}) CommonResult {
	return CommonResult{ResultMark: ResultMark{0, "OK"}, Data: data}
}

func SuccessWithPage(data interface{}, page Page) ResultWithPage {
	return ResultWithPage{ResultMark: ResultMark{0, "OK"}, Data: data, Page: page}
}

func Fail(code int, message string) CommonResult {
	return CommonResult{ResultMark{code, message}, struct{}{}}
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
