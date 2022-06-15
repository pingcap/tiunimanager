/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

// Package classification TiUniManager API.
//
// the purpose of this application is to provide an application
// that is using plain go code to define an API
//
// This should demonstrate all the possible comment annotations
// that are available to turn go code into a fully compliant swagger 2.0 spec

// @title TiUniManager API
// @version v1
// @description This is a sample TiUniManager-API server.
// @BasePath /api/v1

// swagger:meta
package controller

import (
	"net/http"
	"time"

	"github.com/asim/go-micro/v3/client"
	"github.com/gin-gonic/gin"
)

// DefaultTimeout
// todo adjust timeout for async flow task
var DefaultTimeout = func(o *client.CallOptions) {
	o.RequestTimeout = time.Minute * 5
	o.DialTimeout = time.Minute * 5
}

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, Success("hello"))
}

func HelloPage(c *gin.Context) {
	c.JSON(http.StatusOK, Success("hello world"))
}
