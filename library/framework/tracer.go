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
	"io"
	"time"

	"github.com/pingcap-inc/tiem/util/uuidutil"

	"github.com/asim/go-micro/v3/metadata"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"

	"github.com/asim/go-micro/v3/server"
)

type Tracer opentracing.Tracer

// trace id:
//    *gin.Context
//			key $TiEM_X_TRACE_ID_KEY
//    micro-ctx
//			metadata key $TiEM_X_TRACE_ID_KEY
//    normal-ctx
//			key traceIDCtxKey
const TiEM_X_TRACE_ID_KEY = "Em-X-Trace-Id"

type traceCtxKeyType string

const traceIDCtxKey = traceCtxKeyType(TiEM_X_TRACE_ID_KEY)
const traceMicroServiceNameCtxKey = traceCtxKeyType("Em-X-Micro-Service-Name")
const traceMicroEndpointNameCtxKey = traceCtxKeyType("Em-X-Micro-Endpoint-Name")

const TiEM_X_USER_ID_KEY = "Em-X-User-Id"

const TiEM_X_USER_NAME_KEY = "Em-X-User-Name"

const TiEM_X_TENANT_ID_KEY = "Em-X-Tenant-Id"

func NewTracerFromArgs(args *ClientArgs) *Tracer {
	jaegerTracer, _, err := NewJaegerTracer("em", args.TracerAddress)
	if err != nil {
		panic("init tracer failed")
	}
	opentracing.SetGlobalTracer(jaegerTracer)
	return (*Tracer)(&jaegerTracer)
}

func NewJaegerTracer(serviceName string, addr string) (opentracing.Tracer, io.Closer, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}

	sender, err := jaeger.NewUDPTransport(addr, 0)
	if err != nil {
		return nil, nil, err
	}

	reporter := jaeger.NewRemoteReporter(sender)
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Reporter(reporter),
	)
	context.Background()
	return tracer, closer, err
}

type BackgroundTask struct {
	fromCtx        context.Context
	fromTraceID    string
	currentCtx     context.Context
	currentTraceID string
	fn             func(context.Context) error
	span           opentracing.Span
	operationName  string
	// close to notify error return is ready
	retNotifyCh chan struct{}
	retErr      error
}

// NewBackgroundMicroCtx is deprecated and now is only a noop function which the arg `fromMicroCtx` is returned.
// Please use functions like `StartBackgroundTask` instead.
func NewBackgroundMicroCtx(fromMicroCtx context.Context, newTraceIDFlag bool) context.Context {
	return fromMicroCtx
}

// NewBackgroundMicroCtx return a new background micro ctx.
// A new traceID would be generated if newTraceIDFlag is true. Otherwise, the old traceID would be used again.
// The new ctx returned is never canceled, and has no deadline.
func newBackgroundMicroCtx(fromMicroCtx context.Context, newTraceIDFlag bool) context.Context {
	if fromMicroCtx == nil {
		retCtx := NewMicroContextWithKeyValuePairs(
			context.Background(),
			map[string]string{
				TiEM_X_TRACE_ID_KEY: uuidutil.GenerateID(),
			},
		)
		LogWithContext(retCtx).Warn("current ctx is created from a nil micro ctx")
		return retCtx
	}
	traceID := getStringValueFromMicroContext(fromMicroCtx, TiEM_X_TRACE_ID_KEY)
	if newTraceIDFlag {
		traceID = uuidutil.GenerateID()
	}
	retCtx := NewMicroContextWithKeyValuePairs(
		context.Background(),
		map[string]string{
			TiEM_X_TRACE_ID_KEY:  traceID,
			TiEM_X_USER_ID_KEY:   getStringValueFromMicroContext(fromMicroCtx, TiEM_X_USER_ID_KEY),
			TiEM_X_USER_NAME_KEY: getStringValueFromMicroContext(fromMicroCtx, TiEM_X_USER_NAME_KEY),
			TiEM_X_TENANT_ID_KEY: getStringValueFromMicroContext(fromMicroCtx, TiEM_X_TENANT_ID_KEY),
		},
	)
	if newTraceIDFlag {
		LogWithContext(retCtx).Infof("current ctx is created from the ctx with traceID %s",
			getStringValueFromMicroContext(fromMicroCtx, TiEM_X_TRACE_ID_KEY))
	} else {
		LogWithContext(retCtx).Infof("new ctx is created")
	}
	return retCtx
}

func getCurrentServiceName() string {
	if Current == nil {
		return ""
	}
	m := Current.GetServiceMeta()
	if m == nil {
		return ""
	}
	return string(m.ServiceName)
}

func getCurrentTracer() opentracing.Tracer {
	if Current == nil {
		return opentracing.GlobalTracer()
	}
	return *Current.GetTracer()
}

// NewBackgroundTask return a new background task but do not start it in this function call.
func NewBackgroundTask(fromCtx context.Context, operationName string, fn func(context.Context) error) *BackgroundTask {
	currentCtxInfo := getCurrentMicroCtxInfo(fromCtx)
	if len(operationName) <= 0 {
		operationName = fmt.Sprintf("%s BackgroundTask", currentCtxInfo)
	}
	var fromTraceID string
	if fromCtx == nil {
		fromTraceID = ""
	} else {
		fromTraceID = GetTraceIDFromContext(fromCtx)
	}
	newCtx := newBackgroundMicroCtx(fromCtx, false)
	span := getCurrentTracer().StartSpan(operationName)
	span.LogKV("spawn-from", currentCtxInfo)
	newCtx = opentracing.ContextWithSpan(newCtx, span)
	span.SetTag(TiEM_X_TRACE_ID_KEY, GetTraceIDFromContext(newCtx))
	currentTraceID := getStringValueFromMicroContext(newCtx, TiEM_X_TRACE_ID_KEY)
	t := &BackgroundTask{
		fromCtx:        fromCtx,
		fromTraceID:    fromTraceID,
		currentCtx:     newCtx,
		currentTraceID: currentTraceID,
		fn:             fn,
		span:           span,
		operationName:  operationName,
		retNotifyCh:    make(chan struct{}),
	}
	return t
}

func getCurrentMicroCtxInfo(ctx context.Context) string {
	serviceName := GetMicroServiceNameFromContext(ctx)
	endpointName := GetMicroEndpointNameFromContext(ctx)
	if len(serviceName) <= 0 || len(endpointName) <= 0 {
		return getCurrentServiceName()
	} else {
		return fmt.Sprintf("%s.%s", serviceName, endpointName)
	}
}

// GetContext get context from task
func (p *BackgroundTask) GetContext() context.Context {
	return p.currentCtx
}

// Exec exec this task in current goroutine
func (p *BackgroundTask) Exec() error {
	defer p.span.Finish()
	if len(p.fromTraceID) > 0 {
		LogWithContext(p.currentCtx).Infof("start new background task from traceID %s with comments: %s",
			p.fromTraceID, p.operationName,
		)
	} else {
		LogWithContext(p.currentCtx).Infof("start new background task with comments: %s", p.operationName)
	}
	err := p.fn(p.currentCtx)
	p.retErr = err
	close(p.retNotifyCh)
	if err != nil {
		p.span.LogFields(opentracinglog.String("error", err.Error()))
		p.span.SetTag("error", true)
		LogWithContext(p.currentCtx).Errorf("background task returned an error: %s", err)
	} else {
		LogWithContext(p.currentCtx).Info("background task finished successfully")
	}
	return err
}

// Sync get the task's final return error value syncronously.
func (p *BackgroundTask) Sync() error {
	<-p.retNotifyCh
	return p.retErr
}

// StartBackgroundTask new and start a background task in a newly created background goroutine.
// Here is example:
/* func example1(ctx context.Context) {
	// current ctx
	// could be a gin ctx, micro ctx, gorm ctx, background task ctx or nil ctx
	currentCtx := ctx
	// start a background task in a newly created background goroutine
	runningTask := StartBackgroundTask(currentCtx, "tidb backup routine", func(bgTaskCtx context.Context) error {
		// do the backup task
		return nil
	})
	// do something else
	// get return err of the task
	err := runningTask.Sync()
	_ = err
} */
func StartBackgroundTask(fromCtx context.Context, comments string, fn func(context.Context) error) *BackgroundTask {
	t := NewBackgroundTask(fromCtx, comments, fn)
	go func() {
		t.Exec()
	}()
	return t
}

// NewMicroCtxFromGinCtx make a new micro ctx base on the gin ctx
func NewMicroCtxFromGinCtx(c *gin.Context) context.Context {
	var ctx context.Context
	ctx = c
	traceID := getTraceIDFromGinContext(c)
	userID := getStringValueFromGinContext(c, TiEM_X_USER_ID_KEY)
	userName := getStringValueFromGinContext(c, TiEM_X_USER_NAME_KEY)
	tenantID := getStringValueFromGinContext(c, TiEM_X_TENANT_ID_KEY)
	parentSpan := getParentSpanFromGinContext(c)
	if parentSpan != nil {
		ctx = opentracing.ContextWithSpan(ctx, parentSpan)
	}
	return NewMicroContextWithKeyValuePairs(ctx, map[string]string{
		TiEM_X_TRACE_ID_KEY:  traceID,
		TiEM_X_USER_ID_KEY:   userID,
		TiEM_X_USER_NAME_KEY: userName,
		TiEM_X_TENANT_ID_KEY: tenantID,
	})
}

func ForkMicroCtx(ctx context.Context) context.Context {
	return newMicroContextWithTraceID(context.Background(), getTraceIDFromMicroContext(ctx))
}

func newMicroContextWithTraceID(ctx context.Context, traceID string) context.Context {
	md, ok := metadata.FromContext(ctx)
	if ok {
	} else {
		md = make(map[string]string)
	}
	md[TiEM_X_TRACE_ID_KEY] = traceID
	return metadata.NewContext(ctx, md)
}

func getTraceIDFromMicroContext(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if ok {
		return md[TiEM_X_TRACE_ID_KEY]
	} else {
		return ""
	}
}

func getTraceIDFromGinContext(ctx *gin.Context) string {
	id := ctx.GetString(TiEM_X_TRACE_ID_KEY)
	if len(id) <= 0 {
		return ""
	} else {
		return id
	}
}

func getStringValueFromNormalContext(ctx context.Context, key traceCtxKeyType) string {
	v := ctx.Value(key)
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	} else {
		return ""
	}
}

func getTraceIDFromNormalContext(ctx context.Context) string {
	return getStringValueFromNormalContext(ctx, traceIDCtxKey)
}

// GetTraceIDFromContext Get TraceID from ctx
func GetTraceIDFromContext(ctx context.Context) string {
	AssertWithInfo(ctx != nil, "ctx should not be nil")
	switch v := ctx.(type) {
	case *gin.Context:
		return getTraceIDFromGinContext(v)
	default:
	}
	id := getTraceIDFromMicroContext(ctx)
	if len(id) > 0 {
		return id
	}
	return getTraceIDFromNormalContext(ctx)
}

func getStringValueFromGinContext(ctx *gin.Context, key string) string {
	v := ctx.GetString(key)
	if len(v) <= 0 {
		return ""
	} else {
		return v
	}
}

func getStringValueFromMicroContext(ctx context.Context, key string) string {
	md, ok := metadata.FromContext(ctx)
	if ok {
		return md[key]
	} else {
		return ""
	}
}

func getStringValueFromContext(ctx context.Context, key string) string {
	AssertWithInfo(ctx != nil, "ctx should not be nil")
	switch v := ctx.(type) {
	case *gin.Context:
		return getStringValueFromGinContext(v, key)
	default:
	}
	return getStringValueFromMicroContext(ctx, key)
}

func NewMicroContextWithKeyValuePairs(ctx context.Context, pairs map[string]string) context.Context {
	md, ok := metadata.FromContext(ctx)
	if ok {
	} else {
		md = make(map[string]string)
	}
	for k, v := range pairs {
		md[k] = v
	}
	return metadata.NewContext(ctx, md)
}

// GetUserIDFromContext Get UserID from ctx
func GetUserIDFromContext(ctx context.Context) string {
	return getStringValueFromContext(ctx, TiEM_X_USER_ID_KEY)
}

// GetTenantIDFromContext Get TenantID from ctx
func GetTenantIDFromContext(ctx context.Context) string {
	return getStringValueFromContext(ctx, TiEM_X_TENANT_ID_KEY)
}

// GetMicroServiceNameFromContext Get MicroServiceName from ctx, something like "em.cluster"
func GetMicroServiceNameFromContext(ctx context.Context) string {
	return getStringValueFromNormalContext(ctx, traceMicroServiceNameCtxKey)
}

// GetMicroEndpointNameFromContext Get MicroEndpointName from ctx, something like "ClusterService.Login"
func GetMicroEndpointNameFromContext(ctx context.Context) string {
	return getStringValueFromNormalContext(ctx, traceMicroEndpointNameCtxKey)
}

func getParentSpanFromGinContext(ctx context.Context) opentracing.Span {
	AssertWithInfo(ctx != nil, "ctx should not be nil")
	switch v := ctx.(type) {
	case *gin.Context:
		span, existFlag := v.Get("ParentSpan")
		if existFlag {
			return span.(opentracing.Span)
		} else {
			return nil
		}
	default:
		return nil
	}
}

// NewMicroHandlerWrapper inject serviceName and endpointName into current ctx
func NewMicroHandlerWrapper() server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			serviceName := req.Service()
			endpointName := req.Endpoint()
			ctx = context.WithValue(ctx, traceMicroServiceNameCtxKey, serviceName)
			ctx = context.WithValue(ctx, traceMicroEndpointNameCtxKey, endpointName)
			return h(ctx, req, rsp)
		}
	}
}
