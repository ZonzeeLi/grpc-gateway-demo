/**
    @Author:     ZonzeeLi
    @Project:    grpc-gateway-demo
    @CreateDate: 2022/8/19
    @UpdateDate: xxx
    @Note:       链路追踪
**/

package tracing

import (
	"fmt"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
	"io"
	"net/http"
)

func InitTrace(serviceName string, host string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: host,
			LogSpans:           true,
		},
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger), config.Metrics(metrics.NullFactory))
	if err != nil {
		panic(fmt.Sprintf("Init failed: %v\n", err))
	}

	return tracer, closer
}

func OpentracingOpts() []grpc_opentracing.Option {
	return []grpc_opentracing.Option{
		grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
	}
}

var grpcGatewayTag = opentracing.Tag{Key: string(ext.Component), Value: "grpc-gateway"}

func Wrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentSpanContext, err := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
		if err == nil || err == opentracing.ErrSpanContextNotFound {
			serverSpan := opentracing.GlobalTracer().StartSpan(
				"Greet-HTTP-Server",
				// this is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
				ext.RPCServerOption(parentSpanContext),
				grpcGatewayTag,
			)
			r = r.WithContext(opentracing.ContextWithSpan(r.Context(), serverSpan))

			trace, ok := serverSpan.Context().(jaeger.SpanContext)
			if ok {
				w.Header().Set(jaeger.TraceContextHeaderName, fmt.Sprint(trace.TraceID()))
			}
			fmt.Printf("spanID:%v, parentID:%v\n", trace.SpanID(), trace.ParentID())
			fmt.Println(trace.TraceID())
			defer serverSpan.Finish()
		}
		h.ServeHTTP(w, r)
	})
}
