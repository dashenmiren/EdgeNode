package nodes

import (
	"net"

	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs"
)

type UnixListener struct {
	BaseListener

	Listener net.Listener
}

func (this *UnixListener) Serve() error {
	// TODO
	// TODO 注意管理 CountActiveConnections
	return nil
}

func (this *UnixListener) Close() error {
	// TODO
	return nil
}

func (this *UnixListener) Reload(group *serverconfigs.ServerAddressGroup) {
	this.Group = group
	this.Reset()
}
