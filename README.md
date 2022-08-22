## gRPC_gateway_demo

一个使用grpc-gateway实现http+rpc双重入口的简单网关服务demo，主要内容为服务端的拦截器使用（客户端只简单添加了部分）。拦截器的实现方法均为第三方库[github.com/grpc-ecosystem/go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)。

## 主要内容

以下方法，http和grpc的服务端都已实现，（客户端只实现部分，按实际需求来说客户端本身也不需要）。

- 鉴权：简单的认证，已添加JWT的方法实现
- 校验：Proto文件参数校验
- 日志：Zap自定义日志记录
- 监控指标：Prometheus监控客户端和服务端埋点统计
- 限流：拒绝访问
- 重试：可根据参数自定重试次数、重试间隔、最大重试时间...
- 重启恢复：异常恢复
- 链路追踪：Jaeger+OpenTracing的链路追踪

## 后续

后续会对拦截器（中间件）的逐一方法进行自己的实现，并进行压测和在一定配置的情况下能顶住多少并发请求。并且会尝试加入客户端负载均衡、跨域、代理、重定向、缓存、熔断等....