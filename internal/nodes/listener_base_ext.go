// Copyright 2023 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .
//go:build !plus

package nodes

import "crypto/tls"

func (this *BaseListener) calculateFingerprint(clientInfo *tls.ClientHelloInfo) []byte {
	return nil
}
