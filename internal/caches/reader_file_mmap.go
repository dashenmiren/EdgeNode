// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .
//go:build !plus

package caches

import (
	"errors"
	"io"
)

type MMAPFileReader struct {
	FileReader
}

func (this *MMAPFileReader) CopyBodyTo(writer io.Writer) (int, error) {
	// stub
	return 0, errors.New("not implemented")
}
