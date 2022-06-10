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
	"github.com/pingcap-inc/tiunimanager/common/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/pingcap-inc/tiunimanager/library/framework"
)

// HandleJsonRequestWithBuiltReq
// @Description: handle common json request with a built request
// validate request using validator.New().Struct(req)
// serialize request using json.Marshal(req)
// @Parameter c
// @Parameter req
// @Parameter appenders append request data out of http body, such as id in path
// @return bool
// @return requestBody
func HandleJsonRequestWithBuiltReq(c *gin.Context, req interface{}) (string, bool) {
	return HandleRequest(c,
		req,
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

// HandleJsonRequestFromQuery
// @Description: handle common json request from query
// build request using c.ShouldBindQuery(req, binding.JSON)
// validate request using validator.New().Struct(req)
// serialize request using json.Marshal(req)
// @Parameter c
// @Parameter req
// @Parameter appenders append request data out of http body, such as id in path
// @return ok
// @return requestBody
func HandleJsonRequestFromQuery(c *gin.Context,
	req interface{},
	appenders ...func(c *gin.Context, req interface{}) error,
) (requestBody string, ok bool) {
	return HandleRequest(c,
		req,
		func(c *gin.Context, req interface{}) error {
			err := c.ShouldBindQuery(req)
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
) (requestBody string, ok bool) {
	return HandleRequest(c,
		req,
		func(c *gin.Context, req interface{}) error {
			err := c.ShouldBindBodyWith(req, binding.JSON)
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
	serializer func(req interface{}) ([]byte, error)) (string, bool) {

	requestContent := ""
	ok := errors.OfNullable(nil).
		BreakIf(func() error {
			return errors.OfNullable(builder(c, req)).
				Map(func(err error) error {
					return errors.NewError(errors.TIUNIMANAGER_MARSHAL_ERROR, err.Error())
				}).
				Present()
		}).
		BreakIf(func() error {
			return errors.OfNullable(validator(req)).
				Map(func(err error) error {
					return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, err.Error())
				}).
				Present()
		}).
		BreakIf(func() error {
			bytes, err := serializer(req)
			return errors.OfNullable(err).
				Map(func(err error) error {
					return errors.NewError(errors.TIUNIMANAGER_MARSHAL_ERROR, err.Error())
				}).
				Else(func() {
					requestContent = string(bytes)
				}).
				Present()
		}).
		If(func(err error) {
			framework.LogWithContext(c).Errorf("handle http request failed, %s", err.Error())
			c.JSON(http.StatusBadRequest, Fail(int(err.(errors.EMError).GetCode()), err.Error()))
		}).
		IfNil()
	return requestContent, ok
}
