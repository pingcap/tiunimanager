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

package framework

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/common"
)

//
// TiEMError
// @Description: TiEM business error
// Always get TiEMError from TiEMErrorBuilder.build(), limited to SimpleError, NewTiEMError, NewTiEMErrorf, WrapError, ErrorBuilder().build()
type TiEMError struct {
	code  common.TIEM_ERROR_CODE
	msg   string
	cause error
	traceId string
}

func SimpleError(code common.TIEM_ERROR_CODE) TiEMError {
	return ErrorBuilder().Code(code).Build()
}

func NewTiEMError(code common.TIEM_ERROR_CODE, msg string) TiEMError {
	return ErrorBuilder().Code(code).Message(msg).Build()
}

func NewTiEMErrorf(code common.TIEM_ERROR_CODE, format string, a ...interface{}) TiEMError {
	return ErrorBuilder().Code(code).Message(fmt.Sprintf(format, a...)).Build()
}

func WrapError(code common.TIEM_ERROR_CODE, msg string, err error) TiEMError {
	return ErrorBuilder().Code(code).Message(msg).Cause(err).Build()
}

func (e TiEMError) Is(err error) bool {
	if e == err {
		return true
	}

	if e.cause == nil {
		return false
	}
	
	if cause, ok := e.cause.(TiEMError); ok {
		return cause.Is(err)
	} else {
		return e.cause == err
	}

}

func (e TiEMError) Error() string {
	if e.cause == nil {
		return fmt.Sprintf("[%d]%s", e.GetCode(), e.GetMsg())
	} else {
		return fmt.Sprintf("[%d]%s, cause:%s", e.GetCode(), e.GetMsg(), e.GetCause().Error())
	}
}

type TiEMErrorBuilder struct {
	template TiEMError
}

// ErrorBuilder
// @return TiEMErrorBuilder
func ErrorBuilder() TiEMErrorBuilder {
	return TiEMErrorBuilder{
		template: TiEMError{},
	}
}

func (t TiEMErrorBuilder) Code(code common.TIEM_ERROR_CODE) TiEMErrorBuilder {
	t.template.code = code
	return t
}

func (t TiEMErrorBuilder) Message(msg string) TiEMErrorBuilder {
	t.template.msg = msg
	return t
}

func (t TiEMErrorBuilder) Format(format string, args... interface{}) TiEMErrorBuilder {
	t.template.msg = fmt.Sprintf(format, args...)
	return t
}

func (t TiEMErrorBuilder) Cause(cause error) TiEMErrorBuilder {
	t.template.cause = cause
	return t
}

func (t TiEMErrorBuilder) Trace(traceId string) TiEMErrorBuilder {
	t.template.traceId = traceId
	return t
}

func (t TiEMErrorBuilder) Context(ctx context.Context) TiEMErrorBuilder {
	traceId := GetTraceIDFromContext(ctx)
	t.template.traceId = traceId
	return t
}

// Build
// @Description: the only way to get a efficient TiEMError
// @Receiver t
// @return TiEMError
func (t TiEMErrorBuilder) Build() TiEMError {

	return t.template
}

func (e TiEMError) GetCode() common.TIEM_ERROR_CODE {
	return e.code
}

func (e TiEMError) GetMsg() string {
	if len(e.msg) == 0 {
		return e.GetCodeText()
	}
	return e.msg
}

func (e TiEMError) GetCodeText() string {
	return e.code.Explain()
}

func (e TiEMError) GetCause() error {
	return e.cause
}

func (e TiEMError) GetTraceId() string {
	return e.traceId
}

func (e TiEMError) Unwrap() error {
	return e.cause
}