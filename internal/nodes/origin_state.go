package nodes

import "github.com/dashenmiren/EdgeCommon/pkg/serverconfigs"

type OriginState struct {
	CountFails   int64
	UpdatedAt    int64
	Config       *serverconfigs.OriginConfig
	Addr         string
	TLSHost      string
	ReverseProxy *serverconfigs.ReverseProxyConfig
}
