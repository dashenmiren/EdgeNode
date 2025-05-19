// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package conns

type LingerConn interface {
	SetLinger(sec int) error
}
