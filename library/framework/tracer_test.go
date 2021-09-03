package framework

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestNewTracerFromArgs(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		args := &ClientArgs{}
		got := NewTracerFromArgs(args)
		fmt.Println(got)
	})
}

type MyContext struct {}

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
		ginContext.Set(TiEM_X_TRACE_ID_NAME, "111")
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
	t.Run("nil", func(t *testing.T) {
		got := GetTraceIDFromContext(nil)
		assert.Equal(t, "", got)
	})
}

func TestNewJaegerTracer(t *testing.T) {
	_, _, err := NewJaegerTracer("tiem", "127.0.0.1:999")
	assert.NoError(t, err)
}

func TestNewMicroCtxFromGinCtx(t *testing.T) {
	ctx := &gin.Context{}
	ctx.Set(TiEM_X_TRACE_ID_NAME, "111")
	got := NewMicroCtxFromGinCtx(ctx)
	assert.True(t, got.Value(TiEM_X_TRACE_ID_NAME) != "")
	assert.True(t, got.Value(TiEM_X_TRACE_ID_NAME) == ctx.Value(TiEM_X_TRACE_ID_NAME))
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
	t.Run("nil", func(t *testing.T) {
		got := getParentSpanFromGinContext(nil)
		assert.Equal(t, nil, got)
	})
	t.Run("micro", func(t *testing.T) {
		micro := newMicroContextWithTraceID(&gin.Context{}, "111")
		got := getParentSpanFromGinContext(micro)
		assert.Equal(t, nil, got)
	})
}