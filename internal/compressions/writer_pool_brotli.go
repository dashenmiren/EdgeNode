package compressions

import (
	"io"

	"github.com/andybalholm/brotli"
	teaconst "github.com/dashenmiren/EdgeNode/internal/const"
	"github.com/dashenmiren/EdgeNode/internal/utils"
)

var sharedBrotliWriterPool *WriterPool

func init() {
	if !teaconst.IsMain {
		return
	}

	var maxSize = utils.SystemMemoryGB() * 256
	if maxSize == 0 {
		maxSize = 256
	}
	sharedBrotliWriterPool = NewWriterPool(maxSize, brotli.BestCompression, func(writer io.Writer, level int) (Writer, error) {
		return newBrotliWriter(writer, level)
	})
}
