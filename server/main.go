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
	"grpc_gateway_demo/server/middleware/auth"
	"grpc_gateway_demo/server/middleware/logging"
	"grpc_gateway_demo/server/middleware/ratelimit"
	"grpc_gateway_demo/server/middleware/recovery"
	"grpc_gateway_demo/server/middleware/retry"
	"grpc_gateway_demo/server/middleware/tracing"
	"grpc_gateway_demo/server/proto"

	grpc_metric "grpc_gateway_demo/server/middleware/metric"

	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ratelimit "github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	tracer    opentracing.Tracer
	zapLogger *zap.Logger
)

type server struct {
	proto.UnimplementedGreeterServer
}

func (s *server) Greet(ctx context.Context, in *proto.HelloRequest) (*proto.HelloResponse, error) {
	grpc_metric.CounterMetric.WithLabelValues(in.GetGreet()).Inc()
	var span opentracing.Span
	var traceId jaeger.TraceID
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentCtx := parent.Context()
		span = tracer.StartSpan("Greet-server", opentracing.ChildOf(parentCtx))
	} else {
		span = tracer.StartSpan("Greet-server")
	}
	if sc, ok := span.Context().(jaeger.SpanContext); ok {
		traceId = sc.TraceID()
		zapLogger.Info("request TraceID", zap.Any("TraceID", traceId))
	}
	defer func() {
		span.SetTag("server", in.GetGreet())
		span.Finish()
	}()
	return &proto.HelloResponse{Resp: "你好，你的年纪是" + strconv.Itoa(int(in.GetAge().GetAge()))}, nil
}

// AuthFuncOverride 如果开启则覆盖掉鉴权中间件
//func (s *server) AuthFuncOverride(ctx context.Context, method string) (context.Context, error) {
//	zapLogger.Info("client is calling method:", zap.Any("method", method))
//	return ctx, nil
//}

func httpRun() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	retryOpts := retry.RetryOpts()
	zapOpts := logging.ZapOpts()
	opentracingOpts := tracing.OpentracingOpts()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpc_zap.UnaryClientInterceptor(zapLogger, zapOpts...),
			grpc_validator.UnaryClientInterceptor(),
			grpc_retry.UnaryClientInterceptor(retryOpts...),
			grpc_opentracing.UnaryClientInterceptor(opentracingOpts...)),
	}
	gwMux := runtime.NewServeMux()
	err := proto.RegisterGreeterHandlerFromEndpoint(ctx, gwMux, ":9090", opts)
	if err != nil {
		zapLogger.Fatal("failed RegisterGreeterHandler", zap.Error(err))
	}

	return http.ListenAndServe(":8080", tracing.Wrapper(gwMux))
}

func main() {
	var closer io.Closer
	tracer, closer = tracing.InitTrace("Greet", "127.0.0.1:6831")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	zapLogger = logging.InitZap()
	authFunc := auth.InitAuth()
	recoverOpts := recovery.RecoverOpts()
	opentracingOpts := tracing.OpentracingOpts()
	limit := ratelimit.InitRatelimit()
	grpc_metric.InitMetric()

	// grpc
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		grpc_opentracing.UnaryServerInterceptor(opentracingOpts...),
		grpc_validator.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(zapLogger),
		grpc_recovery.UnaryServerInterceptor(recoverOpts...),
		grpc_auth.UnaryServerInterceptor(authFunc),
		grpc_metric.UnaryServerInterceptor(),
		grpc_ratelimit.UnaryServerInterceptor(limit),
	),
	)
	proto.RegisterGreeterServer(s, &server{})
	grpc_metric.GrpcMetrics.InitializeMetrics(s)

	go func() {
		httpServer := &http.Server{Handler: promhttp.HandlerFor(grpc_metric.Reg, promhttp.HandlerOpts{}), Addr: fmt.Sprintf("localhost:%d", 9092)}
		if err := httpServer.ListenAndServe(); err != nil {
			zapLogger.Fatal("failed promhttp", zap.Error(err))
		}
	}()

	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		zapLogger.Fatal("failed listen", zap.Error(err))
	}
	go func() {
		if err := s.Serve(l); err != nil {
			zapLogger.Fatal("failed serve", zap.Error(err))
		}
	}()
	// http
	if err = httpRun(); err != nil {
		zapLogger.Fatal("failed http", zap.Error(err))
	}
}
