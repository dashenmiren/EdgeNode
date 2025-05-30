package nodes

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs"
	"github.com/dashenmiren/EdgeNode/internal/utils/fasttime"
	"github.com/dashenmiren/EdgeNode/internal/utils/goman"
	"github.com/cespare/xxhash/v2"
	"github.com/pires/go-proxyproto"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SharedHTTPClientPool HTTP客户端池单例
var SharedHTTPClientPool = NewHTTPClientPool()

const httpClientProxyProtocolTag = "@ProxyProtocol@"
const maxHTTPRedirects = 8

// HTTPClientPool 客户端池
type HTTPClientPool struct {
	clientsMap map[uint64]*HTTPClient // origin key => client

	cleanTicker *time.Ticker

	locker sync.RWMutex
}

// NewHTTPClientPool 获取新对象
func NewHTTPClientPool() *HTTPClientPool {
	var pool = &HTTPClientPool{
		cleanTicker: time.NewTicker(1 * time.Hour),
		clientsMap:  map[uint64]*HTTPClient{},
	}

	goman.New(func() {
		pool.cleanClients()
	})

	return pool
}

// Client 根据地址获取客户端
func (this *HTTPClientPool) Client(req *HTTPRequest,
	origin *serverconfigs.OriginConfig,
	originAddr string,
	proxyProtocol *serverconfigs.ProxyProtocolConfig,
	followRedirects bool) (rawClient *http.Client, err error) {
	if origin.Addr == nil {
		return nil, errors.New("origin addr should not be empty (originId:" + strconv.FormatInt(origin.Id, 10) + ")")
	}

	if req == nil || req.RawReq == nil || req.RawReq.URL == nil {
		err = errors.New("invalid request url")
		return
	}
	var originHost = req.RawReq.URL.Host
	var urlPort = req.RawReq.URL.Port()
	if len(urlPort) == 0 {
		if req.RawReq.URL.Scheme == "http" {
			urlPort = "80"
		} else {
			urlPort = "443"
		}

		originHost += ":" + urlPort
	}

	var rawKey = origin.UniqueKey() + "@" + originAddr + "@" + originHost

	// if we are under available ProxyProtocol, we add client ip to key to make every client unique
	var isProxyProtocol = false
	if proxyProtocol != nil && proxyProtocol.IsOn {
		rawKey += httpClientProxyProtocolTag + req.requestRemoteAddr(true)
		isProxyProtocol = true
	}

	// follow redirects
	if followRedirects {
		rawKey += "@follow"
	}

	var key = xxhash.Sum64String(rawKey)

	var isLnRequest = origin.Id == 0

	this.locker.RLock()
	client, found := this.clientsMap[key]
	this.locker.RUnlock()
	if found {
		client.UpdateAccessTime()
		return client.RawClient(), nil
	}

	// 这里不能使用RLock，避免因为并发生成多个同样的client实例
	this.locker.Lock()
	defer this.locker.Unlock()

	// 再次查找
	client, found = this.clientsMap[key]
	if found {
		client.UpdateAccessTime()
		return client.RawClient(), nil
	}

	var maxConnections = origin.MaxConns
	var connectionTimeout = origin.ConnTimeoutDuration()
	var readTimeout = origin.ReadTimeoutDuration()
	var idleTimeout = origin.IdleTimeoutDuration()
	var idleConns = origin.MaxIdleConns

	// 超时时间
	if connectionTimeout <= 0 {
		connectionTimeout = 15 * time.Second
	}

	if idleTimeout <= 0 {
		idleTimeout = 2 * time.Minute
	}

	var numberCPU = runtime.NumCPU()
	if numberCPU < 8 {
		numberCPU = 8
	}
	if maxConnections <= 0 {
		maxConnections = numberCPU * 64
	}

	if idleConns <= 0 {
		idleConns = numberCPU * 16
	}

	if isProxyProtocol { // ProxyProtocol无需保持太多空闲连接
		idleConns = 3
	} else if isLnRequest { // 可以判断为Ln节点请求
		maxConnections *= 8
		idleConns *= 8
		idleTimeout *= 4
	} else if sharedNodeConfig != nil && sharedNodeConfig.Level > 1 {
		// Ln节点可以适当增加连接数
		maxConnections *= 2
		idleConns *= 2
	}

	// TLS通讯
	var tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	if origin.Cert != nil {
		var obj = origin.Cert.CertObject()
		if obj != nil {
			tlsConfig.InsecureSkipVerify = false
			tlsConfig.Certificates = []tls.Certificate{*obj}
			if len(origin.Cert.ServerName) > 0 {
				tlsConfig.ServerName = origin.Cert.ServerName
			}
		}
	}

	var transport = &HTTPClientTransport{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				var realAddr = originAddr

				// for redirections
				if followRedirects && originHost != addr {
					realAddr = addr
				}

				// connect
				conn, dialErr := (&net.Dialer{
					Timeout:   connectionTimeout,
					KeepAlive: 1 * time.Minute,
				}).DialContext(ctx, network, realAddr)
				if dialErr != nil {
					return nil, dialErr
				}

				// handle PROXY protocol
				proxyErr := this.handlePROXYProtocol(conn, req, proxyProtocol)
				if proxyErr != nil {
					return nil, proxyErr
				}

				return NewOriginConn(conn), nil
			},
			MaxIdleConns:          0,
			MaxIdleConnsPerHost:   idleConns,
			MaxConnsPerHost:       maxConnections,
			IdleConnTimeout:       idleTimeout,
			ExpectContinueTimeout: 1 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			TLSClientConfig:       tlsConfig,
			ReadBufferSize:        8 * 1024,
			Proxy:                 nil,
		},
	}

	// support http/2
	if origin.HTTP2Enabled && origin.Addr != nil && origin.Addr.Protocol == serverconfigs.ProtocolHTTPS {
		_ = http2.ConfigureTransport(transport.Transport)
	}

	rawClient = &http.Client{
		Timeout:   readTimeout,
		Transport: transport,
		CheckRedirect: func(targetReq *http.Request, via []*http.Request) error {
			// follow redirects
			if followRedirects && len(via) <= maxHTTPRedirects {
				return nil
			}

			return http.ErrUseLastResponse
		},
	}

	this.clientsMap[key] = NewHTTPClient(rawClient, isProxyProtocol)

	return rawClient, nil
}

// 清理不使用的Client
func (this *HTTPClientPool) cleanClients() {
	for range this.cleanTicker.C {
		var nowTime = fasttime.Now().Unix()

		var expiredKeys []uint64
		var expiredClients = []*HTTPClient{}

		// lookup expired clients
		this.locker.RLock()
		for k, client := range this.clientsMap {
			if client.AccessTime() < nowTime-86400 ||
				(client.IsProxyProtocol() && client.AccessTime() < nowTime-3600) { // 超过 N 秒没有调用就关闭
				expiredKeys = append(expiredKeys, k)
				expiredClients = append(expiredClients, client)
			}
		}
		this.locker.RUnlock()

		// remove expired keys
		if len(expiredKeys) > 0 {
			this.locker.Lock()
			for _, k := range expiredKeys {
				delete(this.clientsMap, k)
			}
			this.locker.Unlock()
		}

		// close expired clients
		if len(expiredClients) > 0 {
			for _, client := range expiredClients {
				client.Close()
			}
		}
	}
}

// 支持PROXY Protocol
func (this *HTTPClientPool) handlePROXYProtocol(conn net.Conn, req *HTTPRequest, proxyProtocol *serverconfigs.ProxyProtocolConfig) error {
	if proxyProtocol != nil &&
		proxyProtocol.IsOn &&
		(proxyProtocol.Version == serverconfigs.ProxyProtocolVersion1 || proxyProtocol.Version == serverconfigs.ProxyProtocolVersion2) {
		var remoteAddr = req.requestRemoteAddr(true)
		var transportProtocol = proxyproto.TCPv4
		if strings.Contains(remoteAddr, ":") {
			transportProtocol = proxyproto.TCPv6
		}
		var destAddr = conn.RemoteAddr()
		var reqConn = req.RawReq.Context().Value(HTTPConnContextKey)
		if reqConn != nil {
			destAddr = reqConn.(net.Conn).LocalAddr()
		}
		var header = proxyproto.Header{
			Version:           byte(proxyProtocol.Version),
			Command:           proxyproto.PROXY,
			TransportProtocol: transportProtocol,
			SourceAddr: &net.TCPAddr{
				IP:   net.ParseIP(remoteAddr),
				Port: req.requestRemotePort(),
			},
			DestinationAddr: destAddr,
		}
		_, err := header.WriteTo(conn)
		if err != nil {
			_ = conn.Close()
			return err
		}
		return nil
	}

	return nil
}
