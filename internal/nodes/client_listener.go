// Copyright 2021 GoEdge goedge.cdn@gmail.com. All rights reserved.

package nodes

import (
	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/dashenmiren/EdgeNode/internal/firewalls"
	"github.com/dashenmiren/EdgeNode/internal/iplibrary"
	"github.com/dashenmiren/EdgeNode/internal/waf"
	"net"
)

// ClientListener 客户端网络监听
type ClientListener struct {
	rawListener net.Listener
	isHTTP      bool
	isTLS       bool
}

func NewClientListener(listener net.Listener, isHTTP bool) *ClientListener {
	return &ClientListener{
		rawListener: listener,
		isHTTP:      isHTTP,
	}
}

func (this *ClientListener) SetIsTLS(isTLS bool) {
	this.isTLS = isTLS
}

func (this *ClientListener) IsTLS() bool {
	return this.isTLS
}

func (this *ClientListener) Accept() (net.Conn, error) {
	conn, err := this.rawListener.Accept()
	if err != nil {
		return nil, err
	}

	// 是否在WAF名单中
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	var isInAllowList = false
	if err == nil {
		canGoNext, inAllowList, expiresAt := iplibrary.AllowIP(ip, 0)
		isInAllowList = inAllowList
		if !canGoNext {
			firewalls.DropTemporaryTo(ip, expiresAt)
		} else {
			if !waf.SharedIPWhiteList.Contains(waf.IPTypeAll, firewallconfigs.FirewallScopeGlobal, 0, ip) {
				var ok bool
				expiresAt, ok = waf.SharedIPBlackList.ContainsExpires(waf.IPTypeAll, firewallconfigs.FirewallScopeGlobal, 0, ip)
				if ok {
					canGoNext = false
					firewalls.DropTemporaryTo(ip, expiresAt)
				}
			}
		}

		if !canGoNext {
			tcpConn, ok := conn.(*net.TCPConn)
			if ok {
				_ = tcpConn.SetLinger(0)
			}

			_ = conn.Close()

			return this.Accept()
		}
	}

	return NewClientConn(conn, this.isHTTP, this.isTLS, isInAllowList), nil
}

func (this *ClientListener) Close() error {
	return this.rawListener.Close()
}

func (this *ClientListener) Addr() net.Addr {
	return this.rawListener.Addr()
}
