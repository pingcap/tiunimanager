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
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/pingcap-inc/tiem/file-server/controller"
	"github.com/pingcap-inc/tiem/library/framework"
	"net/http"
)

// HandleJsonRequestWithBuiltReq
// @Description: handle common json request with a built request
// validate request using validator.New().Struct(req)
// serialize request using json.Marshal(req)
// @Parameter c
// @Parameter req
// @Parameter appenders append request data out of http body, such as id in path
// @return err
// @return requestBody
func HandleJsonRequestWithBuiltReq(c *gin.Context, req interface{}) (requestBody string, err error) {
	return HandleRequest(c,
		&req,
		func(c *gin.Context, req interface{}) error {
			return nil
		},
		func(req interface{}) error {
			return validator.New().Struct(req)
		},
		func(req interface{}) ([]byte, error) {
			return json.Marshal(req)
		},
	)
}

// HandleJsonRequestFromBody
// @Description: handle common json request from body
// build request using c.ShouldBindBodyWith(req, binding.JSON)
// validate request using validator.New().Struct(req)
// serialize request using json.Marshal(req)
// @Parameter c
// @Parameter req
// @Parameter appenders append request data out of http body, such as id in path
// @return err
// @return requestBody
func HandleJsonRequestFromBody(c *gin.Context,
	req interface{},
	appenders ...func(c *gin.Context, req interface{}) error,
) (requestBody string, err error) {
	return HandleRequest(c,
		&req,
		func(c *gin.Context, req interface{}) error {
			err = c.ShouldBindBodyWith(req, binding.JSON)
			if err != nil {
				return err
			}
			for _, appender := range appenders {
				err = appender(c, req)
				if err != nil {
					return err
				}
			}
			return nil
		},
		func(req interface{}) error {
			return validator.New().Struct(req)
		},
		func(req interface{}) ([]byte, error) {
			return json.Marshal(req)
		},
	)
}

// HandleRequest
// @Description: handle request
// @Parameter c
// @Parameter req
// @Parameter builder build request from gin.Context
// @Parameter validator validate request
// @Parameter serializer serialize request to string
// @return err
// @return requestBody
func HandleRequest(c *gin.Context,
	req interface{},
	builder func(c *gin.Context, req interface{}) error,
	validator func(req interface{}) error,
	serializer func(req interface{}) ([]byte, error)) (requestBody string, err error) {

	err = builder(c, req)
	if err != nil {
		framework.LogWithContext(c).Errorf("get request failed, %s", err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	err = validator(req)
	if err != nil {
		framework.LogWithContext(c).Errorf("validate request failed, %s", err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	requestBodyBytes, err := serializer(req)
	if err != nil {
		framework.LogWithContext(c).Errorf("validate request failed, %s", err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	} else {
		requestBody = string(requestBodyBytes)
		return
	}
}
