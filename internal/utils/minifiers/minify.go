//go:build !plus

package minifiers

import (
	"net/http"

	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs"
)

// MinifyResponse minify response body
func MinifyResponse(config *serverconfigs.HTTPPageOptimizationConfig, url string, resp *http.Response) error {
	// stub
	return nil
}
