/**
    @Author:     ZonzeeLi
    @Project:    grpc-gateway-demo
    @CreateDate: 2022/8/17
    @UpdateDate: xxx
    @Note:       重试
**/

package retry

import (
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
)

func RetryOpts() []grpc_retry.CallOption {
	return []grpc_retry.CallOption{
		grpc_retry.WithMax(3),
		grpc_retry.WithPerRetryTimeout(2 * time.Second),
	}
}
