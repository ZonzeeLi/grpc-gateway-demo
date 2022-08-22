/**
    @Author:     ZonzeeLi
    @Project:    grpc-gateway-demo
    @CreateDate: 2022/8/17
    @UpdateDate: xxx
    @Note:       grpc-gateway网关服务框架
**/

package main

import (
	"context"
	"fmt"
	"grpc_gateway_demo/client/proto"
	"net/http"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"io"
	"log"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	reg := prometheus.NewRegistry()
	grpcMetrics := grpc_prometheus.NewClientMetrics()
	reg.MustRegister(grpcMetrics)

	tracer, closer := InitTrace("GreetClient", "172.28.5.39:6831")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	span := tracer.StartSpan("greet client")
	span.SetTag("client", "greetStart")
	conn, err := grpc.Dial("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpc_opentracing.UnaryClientInterceptor(),
			grpc_retry.UnaryClientInterceptor(),
			grpcMetrics.UnaryClientInterceptor(),
		))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// promHttp
	httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: fmt.Sprintf("localhost:%d", 9094)}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal("failed promhttp:%v", err)
		}
	}()
	// grpc
	c := proto.NewGreeterClient(conn)
	r, err := c.Greet(context.Background(), &proto.HelloRequest{
		Greet: "lzz",
		Age: &proto.HelloRequest_Message{
			Age: 10,
		},
	}, grpc_retry.WithMax(3), grpc_retry.WithPerRetryTimeout(2*time.Second))
	if err != nil {
		log.Fatalf("failed err:%v", err)
	}
	fmt.Println(r.GetResp())
}

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
