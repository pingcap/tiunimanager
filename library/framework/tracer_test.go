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

package framework

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestNewTracerFromArgs(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		args := &ClientArgs{}
		got := NewTracerFromArgs(args)
		fmt.Println(got)
	})
}

type MyContext struct{}

func (m MyContext) Deadline() (deadline time.Time, ok bool) {
	panic("implement me")
}

func (m MyContext) Done() <-chan struct{} {
	panic("implement me")
}

func (m MyContext) Err() error {
	panic("implement me")
}

func (m MyContext) Value(key interface{}) interface{} {
	if key == traceIDCtxKey {
		return "traceIDCtxValue"
	} else {
		return nil
	}
}

func TestGetTraceIDFromContext(t *testing.T) {
	t.Run("gin", func(t *testing.T) {
		ginContext := &gin.Context{}
		ginContext.Set(TiEM_X_TRACE_ID_KEY, "111")
		got := GetTraceIDFromContext(ginContext)
		assert.Equal(t, "111", got)
	})
	t.Run("micro", func(t *testing.T) {
		microContext := newMicroContextWithTraceID(&gin.Context{}, "222")
		got := GetTraceIDFromContext(microContext)
		assert.Equal(t, "222", got)
	})

	t.Run("normal", func(t *testing.T) {
		got := GetTraceIDFromContext(MyContext{})
		assert.Equal(t, "traceIDCtxValue", got)
	})
}

func TestNewJaegerTracer(t *testing.T) {
	_, _, err := NewJaegerTracer("em", "127.0.0.1:999")
	assert.NoError(t, err)
}

func TestNewMicroCtxFromGinCtx(t *testing.T) {
	ctx := &gin.Context{}
	ctx.Set(TiEM_X_TRACE_ID_KEY, "111")
	got := NewMicroCtxFromGinCtx(ctx)
	assert.True(t, got.Value(TiEM_X_TRACE_ID_KEY) != "")
	assert.True(t, got.Value(TiEM_X_TRACE_ID_KEY) == ctx.Value(TiEM_X_TRACE_ID_KEY))
}

func Test_getParentSpanFromGinContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want opentracing.Span
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getParentSpanFromGinContext(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getParentSpanFromGinContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_MicroContextWithTraceID(t *testing.T) {
	got := newMicroContextWithTraceID(&gin.Context{}, "111")
	assert.Equal(t, "111", getTraceIDFromMicroContext(got))
}

func Test_getParentSpanFromGinContext1(t *testing.T) {
	t.Run("micro", func(t *testing.T) {
		micro := newMicroContextWithTraceID(&gin.Context{}, "111")
		got := getParentSpanFromGinContext(micro)
		assert.Equal(t, nil, got)
	})
}

func Test_GetStringValuesFromContext(t *testing.T) {
	c := &gin.Context{}
	traceID := "traceID"
	userID := "userID"
	userName := "userName"
	tenantID := "tenantID"
	c.Set(TiEM_X_TRACE_ID_KEY, traceID)
	c.Set(TiEM_X_USER_ID_KEY, userID)
	c.Set(TiEM_X_USER_NAME_KEY, userName)
	c.Set(TiEM_X_TENANT_ID_KEY, tenantID)
	assert.Equal(t, traceID, GetTraceIDFromContext(c))
	assert.Equal(t, userID, GetUserIDFromContext(c))
	assert.Equal(t, userName, GetUserNameFromContext(c))
	assert.Equal(t, tenantID, GetTenantIDFromContext(c))
	ctx := NewMicroCtxFromGinCtx(c)
	assert.Equal(t, traceID, GetTraceIDFromContext(ctx))
	assert.Equal(t, userID, GetUserIDFromContext(ctx))
	assert.Equal(t, userName, GetUserNameFromContext(ctx))
	assert.Equal(t, tenantID, GetTenantIDFromContext(ctx))
}

func Test_NewBackgroundMicroCtx(t *testing.T) {
	c := &gin.Context{}
	traceID := "traceID"
	userID := "userID"
	userName := "userName"
	tenantID := "tenantID"
	c.Set(TiEM_X_TRACE_ID_KEY, traceID)
	c.Set(TiEM_X_USER_ID_KEY, userID)
	c.Set(TiEM_X_USER_NAME_KEY, userName)
	c.Set(TiEM_X_TENANT_ID_KEY, tenantID)
	assert.Equal(t, traceID, GetTraceIDFromContext(c))
	assert.Equal(t, userID, GetUserIDFromContext(c))
	assert.Equal(t, userName, GetUserNameFromContext(c))
	assert.Equal(t, tenantID, GetTenantIDFromContext(c))
	ctx := NewMicroCtxFromGinCtx(c)
	assert.Equal(t, traceID, GetTraceIDFromContext(ctx))
	assert.Equal(t, userID, GetUserIDFromContext(ctx))
	assert.Equal(t, userName, GetUserNameFromContext(ctx))
	assert.Equal(t, tenantID, GetTenantIDFromContext(ctx))
	newCtxWithSameTraceID := NewBackgroundMicroCtx(ctx, false)
	assert.Equal(t, traceID, GetTraceIDFromContext(newCtxWithSameTraceID))
	newCtxWithDifferentTraceID := NewBackgroundMicroCtx(ctx, true)
	assert.NotEqual(t, traceID, GetTraceIDFromContext(newCtxWithDifferentTraceID))
	assert.NotEqual(t, traceID, NewBackgroundMicroCtx(nil, false))
	assert.NotEqual(t, traceID, NewBackgroundMicroCtx(nil, true))
	assert.NotEqual(t, "", NewBackgroundMicroCtx(nil, false))
}
