// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved.

package compressions

import (
	teaconst "github.com/dashenmiren/EdgeNode/internal/const"
	"io"
)

var sharedBrotliWriterPool *WriterPool

func init() {
	if !teaconst.IsMain {
		return
	}

	sharedBrotliWriterPool = NewWriterPool(CalculatePoolSize(), func(writer io.Writer, level int) (Writer, error) {
		return newBrotliWriter(writer)
	})
}
