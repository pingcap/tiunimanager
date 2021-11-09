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

	"github.com/asim/go-micro/v3/client"
	"github.com/gin-gonic/gin"
)

type ResultMark struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CommonFileResult struct {
	ResultMark
	Data interface{} `json:"data"`
}

var DefaultTimeout = func(o *client.CallOptions) {
	o.RequestTimeout = time.Second * 30
	o.DialTimeout = time.Second * 30
}

func Success(data interface{}) *CommonFileResult {
	return &CommonFileResult{ResultMark: ResultMark{0, "OK"}, Data: data}
}

func Fail(code int, message string) *CommonFileResult {
	return &CommonFileResult{ResultMark{code, message}, struct{}{}}
}

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, Success("hello"))
}
