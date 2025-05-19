// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package dbs

func IsClosedErr(err error) bool {
	return err == errDBIsClosed
}
