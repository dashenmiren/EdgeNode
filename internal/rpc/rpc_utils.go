package rpc

import (
	"github.com/dashenmiren/EdgeNode/internal/configs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"sync"
)

var sharedRPC *RPCClient = nil
var locker = &sync.Mutex{}

// SharedRPC RPC对象
func SharedRPC() (*RPCClient, error) {
	locker.Lock()
	defer locker.Unlock()

	if sharedRPC != nil {
		return sharedRPC, nil
	}

	config, err := configs.LoadAPIConfig()
	if err != nil {
		return nil, err
	}
	client, err := NewRPCClient(config)
	if err != nil {
		return nil, err
	}

	sharedRPC = client
	return sharedRPC, nil
}

// IsConnError 是否为连接错误
func IsConnError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为连接错误
	statusErr, ok := status.FromError(err)
	if ok {
		var errorCode = statusErr.Code()
		return errorCode == codes.Unavailable || errorCode == codes.Canceled
	}

	if strings.Contains(err.Error(), "code = Canceled") {
		return true
	}

	return false
}
