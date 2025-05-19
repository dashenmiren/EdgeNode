//go:build !plus

package nodes

import "net/http"

func (this *HTTPRequest) processHTTP3Headers(respHeader http.Header) {
	// stub
}
