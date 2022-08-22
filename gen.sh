# 生成pb.go文件
protoc -I=proto --go_out=proto --go-grpc_out=proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative hello.proto

# 生成pg.gw.go 用于启动http服务
protoc -I=proto --grpc-gateway_out=proto --grpc-gateway_opt=paths=source_relative hello.proto

# 生成pb.validate.go 验证器
protoc -I=proto --govalidators_out=proto --govalidators_opt=paths=source_relative hello.proto
