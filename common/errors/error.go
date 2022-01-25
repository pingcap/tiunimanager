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

package errors

import (
	"fmt"
)

//
// EMError
// @Description: EM business error
// Always get EMError from EMErrorBuilder.build(), limited to Error, NewEMError, NewErrorf, WrapError, ErrorBuilder().build()
type EMError struct {
	code  EM_ERROR_CODE
	msg   string
	cause error
}

func Error(code EM_ERROR_CODE) EMError {
	return ErrorBuilder().Code(code).Build()
}

func NewError(code EM_ERROR_CODE, msg string) EMError {
	return ErrorBuilder().Code(code).Message(msg).Build()
}

func NewErrorf(code EM_ERROR_CODE, format string, a ...interface{}) EMError {
	return ErrorBuilder().Code(code).Message(fmt.Sprintf(format, a...)).Build()
}

func WrapError(code EM_ERROR_CODE, msg string, err error) EMError {
	return ErrorBuilder().Code(code).Message(msg).Cause(err).Build()
}

func (e EMError) Is(err error) bool {
	if e == err {
		return true
	}

	if e.cause == nil {
		return false
	}

	if cause, ok := e.cause.(EMError); ok {
		return cause.Is(err)
	} else {
		return e.cause == err
	}

}

func (e EMError) Error() string {
	errInfo := fmt.Sprintf("[%d]%s", e.GetCode(), e.GetCodeText())
	if len(e.msg) > 0  {
		errInfo = fmt.Sprintf("%s	%s", errInfo, e.msg)
	}

	if e.cause != nil {
		errInfo = fmt.Sprintf("%s	cause: %s", errInfo, e.cause.Error())
	}
	return errInfo
}

type EMErrorBuilder struct {
	template EMError
}

// ErrorBuilder
// @return EMErrorBuilder
func ErrorBuilder() EMErrorBuilder {
	return EMErrorBuilder{
		template: EMError{},
	}
}

func (t EMErrorBuilder) Code(code EM_ERROR_CODE) EMErrorBuilder {
	t.template.code = code
	return t
}

func (t EMErrorBuilder) Message(msg string) EMErrorBuilder {
	t.template.msg = msg
	return t
}

func (t EMErrorBuilder) Format(format string, args... interface{}) EMErrorBuilder {
	t.template.msg = fmt.Sprintf(format, args...)
	return t
}

func (t EMErrorBuilder) Cause(cause error) EMErrorBuilder {
	t.template.cause = cause
	return t
}

// Build
// @Description: the only way to get a efficient EMError
// @Receiver t
// @return EMError
func (t EMErrorBuilder) Build() EMError {
	return t.template
}

func (e EMError) GetCode() EM_ERROR_CODE {
	return e.code
}

func (e EMError) GetMsg() string {
	if len(e.msg) == 0 {
		return e.GetCodeText()
	}
	return e.msg
}

func (e EMError) GetCodeText() string {
	return e.code.Explain()
}

func (e EMError) GetCause() error {
	return e.cause
}

func (e EMError) Unwrap() error {
	return e.cause
}
