package compressions

import (
	"io"

	teaconst "github.com/TeaOSLab/EdgeNode/internal/const"
	"github.com/TeaOSLab/EdgeNode/internal/utils"
	"github.com/andybalholm/brotli"
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
