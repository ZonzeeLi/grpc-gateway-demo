/**
    @Author:     ZonzeeLi
    @Project:    grpc-gateway-demo
    @CreateDate: 2022/8/17
    @UpdateDate: xxx
    @Note:       目录判断
**/

package utils

import "os"

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
