//go:build !plus

package nodes

import "crypto/tls"

func (this *BaseListener) calculateFingerprint(clientInfo *tls.ClientHelloInfo) []byte {
	return nil
}
