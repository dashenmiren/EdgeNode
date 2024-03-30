package compressions

import (
	"compress/gzip"
	"io"

	teaconst "github.com/TeaOSLab/EdgeNode/internal/const"
	"github.com/TeaOSLab/EdgeNode/internal/utils"
)

var sharedGzipWriterPool *WriterPool

func init() {
	if !teaconst.IsMain {
		return
	}

	var maxSize = utils.SystemMemoryGB() * 256
	if maxSize == 0 {
		maxSize = 256
	}
	sharedGzipWriterPool = NewWriterPool(maxSize, gzip.BestCompression, func(writer io.Writer, level int) (Writer, error) {
		return newGzipWriter(writer, level)
	})
}
