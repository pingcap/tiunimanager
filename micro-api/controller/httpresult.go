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

package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"net/http"
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

func HandleHttpResponse(c *gin.Context, err error,
	withStatusCode func() (common.TIEM_ERROR_CODE, string),
	withData func() interface{}, withPage func() Page) {

	if err != nil {
		framework.LogWithContext(c).Error(err.Error())
		c.JSON(http.StatusInternalServerError, Fail(500, err.Error()))
		return
	}

	if withStatusCode != nil {
		code, message := withStatusCode()
		if code > 0 {
			framework.LogWithContext(c).Error(message)
			c.JSON(mapHttpCode(int(code)), Fail(int(code), message))
			return
		}
	}

	if withPage != nil {
		c.JSON(http.StatusOK, SuccessWithPage(withData(), withPage()))
	} else {
		c.JSON(http.StatusOK, Success(withData()))
	}
}

func mapHttpCode(code int) (httpCode int) {
	// todo map to httpcode
	return 200
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

