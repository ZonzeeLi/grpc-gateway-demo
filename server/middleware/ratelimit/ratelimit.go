/**
    @Author:     ZonzeeLi
    @Project:    grpc-gateway-demo
    @CreateDate: 2022/8/22
    @UpdateDate: xxx
    @Note:       限流
**/

package ratelimit

type alwaysPassLimiter struct{}

func (*alwaysPassLimiter) Limit() bool {
	return false
}

func InitRatelimit() *alwaysPassLimiter {
	return new(alwaysPassLimiter)
}
