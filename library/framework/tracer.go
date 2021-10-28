
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
	"io"
	"time"

	"github.com/asim/go-micro/v3/metadata"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

type Tracer opentracing.Tracer

// trace id:
//    *gin.Context
//			key $TiEM_X_TRACE_ID_NAME
//    micro-ctx
//			metadata key $TiEM_X_TRACE_ID_NAME
//    normal-ctx
//			key traceIDCtxKey
var TiEM_X_TRACE_ID_NAME = "Tiem-X-Trace-Id"

type traceIDCtxKeyType struct{}

var traceIDCtxKey traceIDCtxKeyType

func NewTracerFromArgs(args *ClientArgs) *Tracer {
	jaegerTracer, _, err := NewJaegerTracer("tiem", args.TracerAddress)
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

	return tracer, closer, err
}

type BackgroundTask struct {
	fromCtx        context.Context
	fromTraceID    string
	currentCtx     context.Context
	currentTraceID string
	fn             func(context.Context) error
	comments       string
	// close to notify error return is ready
	retNotifyCh chan struct{}
	retErr      error
}

// NewBackgroundTask return a new background task but do not start it in this function call.
func NewBackgroundTask(fromCtx context.Context, comments string, fn func(context.Context) error) *BackgroundTask {
	currentTraceID := uuidutil.GenerateID()
	var fromTraceID string
	if fromCtx == nil {
		fromTraceID = ""
	} else {
		fromTraceID = GetTraceIDFromContext(fromCtx)
	}
	t := &BackgroundTask{
		fromCtx:        fromCtx,
		fromTraceID:    fromTraceID,
		currentCtx:     newMicroContextWithTraceID(context.Background(), currentTraceID),
		currentTraceID: currentTraceID,
		fn:             fn,
		comments:       comments,
		retNotifyCh:    make(chan struct{}),
	}
	return t
}

// Exec exec this task in current goroutine
func (p *BackgroundTask) Exec() error {
	if len(p.fromTraceID) > 0 {
		LogWithContext(p.currentCtx).Infof("start new background task from traceID %s with comments: %s",
			p.fromTraceID, p.comments,
		)
	} else {
		LogWithContext(p.currentCtx).Infof("start new background task with comments: %s", p.comments)
	}
	err := p.fn(p.currentCtx)
	p.retErr = err
	close(p.retNotifyCh)
	if err != nil {
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

// make a new micro ctx base on the gin ctx
func NewMicroCtxFromGinCtx(c *gin.Context) context.Context {
	var ctx context.Context
	ctx = c
	id := getTraceIDFromGinContext(c)
	parentSpan := getParentSpanFromGinContext(c)
	if parentSpan != nil {
		ctx = opentracing.ContextWithSpan(ctx, parentSpan)
	}
	return newMicroContextWithTraceID(ctx, id)
}

func newMicroContextWithTraceID(ctx context.Context, traceID string) context.Context {
	md, ok := metadata.FromContext(ctx)
	if ok {
	} else {
		md = make(map[string]string)
	}
	md[TiEM_X_TRACE_ID_NAME] = traceID
	return metadata.NewContext(ctx, md)
}

func getTraceIDFromMicroContext(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if ok {
		return md[TiEM_X_TRACE_ID_NAME]
	} else {
		return ""
	}
}

func getTraceIDFromGinContext(ctx *gin.Context) string {
	id := ctx.GetString(TiEM_X_TRACE_ID_NAME)
	if len(id) <= 0 {
		return ""
	} else {
		return id
	}
}

func getTraceIDFromNormalContext(ctx context.Context) string {
	v := ctx.Value(traceIDCtxKey)
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
