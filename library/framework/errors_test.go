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
	"errors"
	"fmt"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleError(t *testing.T) {
	cause := errors.New("cause")
	te := WrapError(common.TIEM_BACKUP_RECORD_CREATE_FAILED, "backup", cause)

	assert.Equal(t, cause, te.cause)
}

func TestCustomizeMessageError(t *testing.T) {
	te := NewTiEMError(common.TIEM_BACKUP_RECORD_CREATE_FAILED, "sss")

	assert.Equal(t, "sss", te.msg)
}

func TestWrapError(t *testing.T) {
	cause := errors.New("cause")
	te := WrapError(common.TIEM_BACKUP_RECORD_CREATE_FAILED, "backup", cause)

	assert.Equal(t, cause, te.cause)
}

func TestIsError(t *testing.T) {
	cause := errors.New("cause")

	simple := SimpleError(common.TIEM_PARAMETER_INVALID)
	assert.True(t, simple.Is(simple))
	assert.False(t, simple.Is(cause))

	te := WrapError(common.TIEM_BACKUP_RECORD_CREATE_FAILED, "backup", cause)

	assert.True(t, te.Is(cause))

	te2 := WrapError(common.TIEM_ACCOUNT_NOT_FOUND, "account", te)
	assert.True(t, te2.Is(te))
	assert.True(t, te2.Is(cause))

}

func TestError(t *testing.T) {
	cause := errors.New("cause")
	te := WrapError(common.TIEM_BACKUP_RECORD_CREATE_FAILED, "backup", cause)
	assert.Equal(t, fmt.Sprintf("[%d]backup, cause:cause", common.TIEM_BACKUP_RECORD_CREATE_FAILED), te.Error())

	t2 := SimpleError(common.TIEM_BACKUP_RECORD_QUERY_FAILED)
	assert.Equal(t, fmt.Sprintf("[%d]query backup record failed", common.TIEM_BACKUP_RECORD_QUERY_FAILED), t2.Error())
}

func TestErrorBuilder(t *testing.T) {
	cause := errors.New("cause")

	err := ErrorBuilder().Code(common.TIEM_ADD_TOKEN_FAILED).
		Format("my error %s", "sss").
		Cause(cause).
		Trace("123123").
		Build()

	assert.Equal(t, int(common.TIEM_ADD_TOKEN_FAILED), int(err.GetCode()))
	assert.Equal(t, "[400]my error sss, cause:cause", err.Error())
	assert.Equal(t, "my error sss", err.GetMsg())
	assert.Equal(t, "add token failed", err.GetCodeText())

	assert.Equal(t, "123123", err.GetTraceId())
	assert.Equal(t, cause, err.GetCause())

}

func TestUnwrapError(t *testing.T) {
	cause := errors.New("cause")
	te := WrapError(common.TIEM_BACKUP_RECORD_CREATE_FAILED, "backup", cause)

	assert.Equal(t, cause, te.Unwrap())
}

func TestNewTiEMErrorf(t *testing.T) {
	got := NewTiEMErrorf(common.TIEM_METADB_SERVER_CALL_ERROR, "1 %s 2", "insert")
	assert.Equal(t, "[9998]1 insert 2", got.Error())
}
