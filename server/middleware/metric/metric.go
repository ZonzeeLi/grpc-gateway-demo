/**
    @Author:     ZonzeeLi
    @Project:    grpc-gateway-demo
    @CreateDate: 2022/8/19
    @UpdateDate: xxx
    @Note:       监控指标
**/

package metric

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

var (
	Reg           = prometheus.NewRegistry()
	CounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "greet_demo_server",
		Help: "Total number of RPCs handled on the server.",
	}, []string{"greet"})
	GrpcMetrics = grpc_prometheus.NewServerMetrics()
)

func InitMetric() {
	Reg.MustRegister(GrpcMetrics, CounterMetric)
	CounterMetric.WithLabelValues("lzz")
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return GrpcMetrics.UnaryServerInterceptor()
}
