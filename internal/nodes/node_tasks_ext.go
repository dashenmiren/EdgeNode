//go:build !plus

package nodes

import "github.com/dashenmiren/EdgeNode/internal/rpc"

func (this *Node) execScriptsChangedTask() error {
	// stub
	return nil
}

func (this *Node) execUAMPolicyChangedTask(rpcClient *rpc.RPCClient) error {
	// stub
	return nil
}

func (this *Node) execHTTPCCPolicyChangedTask(rpcClient *rpc.RPCClient) error {
	// stub
	return nil
}

func (this *Node) execHTTP3PolicyChangedTask(rpcClient *rpc.RPCClient) error {
	// stub
	return nil
}

func (this *Node) execHTTPPagesPolicyChangedTask(rpcClient *rpc.RPCClient) error {
	// stub
	return nil
}

func (this *Node) execNetworkSecurityPolicyChangedTask(rpcClient *rpc.RPCClient) error {
	// stub
	return nil
}

func (this *Node) execPlanChangedTask(rpcClient *rpc.RPCClient) error {
	return nil
}
