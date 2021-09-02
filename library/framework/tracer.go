package framework

import (
	"context"
	"io"
	"time"

	"github.com/asim/go-micro/v3/metadata"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
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

func GinOpenTracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// todo tracer
		//var parentSpan opentracing.Span
		//
		//tracer := GlobalTracer
		//
		//spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		//if err != nil {
		//	parentSpan = tracer.StartSpan(c.Request.URL.Path)
		//	defer parentSpan.Finish()
		//} else {
		//	parentSpan = opentracing.StartSpan(
		//		c.Request.URL.Path,
		//		opentracing.ChildOf(spCtx),
		//		opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
		//		ext.SpanKindRPCServer,
		//	)
		//	defer parentSpan.Finish()
		//}
		//c.Set("Tracer", tracer)
		//c.Set("ParentSpan", parentSpan)
		//c.Next()
	}
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
	if ctx == nil {
		return ""
	}
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
	if ctx == nil {
		return nil
	}
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
