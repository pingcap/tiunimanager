package framework

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

type Tracer opentracing.Tracer

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
