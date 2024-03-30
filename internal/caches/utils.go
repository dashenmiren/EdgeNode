package caches

import (
	"net"
	"strings"

	"github.com/dashenmiren/EdgeCommon/pkg/configutils"
)

func ParseHost(key string) string {
	var schemeIndex = strings.Index(key, "://")
	if schemeIndex <= 0 {
		return ""
	}

	var firstSlashIndex = strings.Index(key[schemeIndex+3:], "/")
	if firstSlashIndex <= 0 {
		return ""
	}

	var host = key[schemeIndex+3 : schemeIndex+3+firstSlashIndex]

	hostPart, _, err := net.SplitHostPort(host)
	if err == nil && len(hostPart) > 0 {
		host = configutils.QuoteIP(hostPart)
	}

	return host
}
