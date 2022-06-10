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
	"github.com/pingcap-inc/tiunimanager/common/errors"
	"github.com/pingcap-inc/tiunimanager/library/framework"
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
	withStatusCode func() (errors.EM_ERROR_CODE, string),
	withData func() (interface{}, error), withPage func() Page) {
	if err != nil {
		framework.LogWithContext(c).Error(err.Error())
		c.JSON(http.StatusInternalServerError, Fail(int(errors.TIUNIMANAGER_CLUSTER_SERVER_CALL_ERROR), err.Error()))
		return
	}

	if withStatusCode != nil {
		code, message := withStatusCode()
		if code != errors.TIUNIMANAGER_SUCCESS {
			framework.LogWithContext(c).Error(message)
			c.JSON(code.GetHttpCode(), Fail(int(code), message))
			return
		}
	}

	data, err := withData()
	if err != nil {
		framework.LogWithContext(c).Error(err.Error())
		c.JSON(http.StatusInternalServerError, Fail(int(errors.TIUNIMANAGER_UNRECOGNIZED_ERROR), err.Error()))
		return
	}
	if withPage != nil {
		c.JSON(http.StatusOK, SuccessWithPage(data, withPage()))
	} else {
		c.JSON(http.StatusOK, Success(data))
	}
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
