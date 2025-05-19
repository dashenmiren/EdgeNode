package compressions

import (
	"compress/flate"
	"io"

	teaconst "github.com/dashenmiren/EdgeNode/internal/const"
	"github.com/dashenmiren/EdgeNode/internal/utils"
)

var sharedDeflateWriterPool *WriterPool

func init() {
	if !teaconst.IsMain {
		return
	}

	var maxSize = utils.SystemMemoryGB() * 256
	if maxSize == 0 {
		maxSize = 256
	}
	sharedDeflateWriterPool = NewWriterPool(maxSize, flate.BestCompression, func(writer io.Writer, level int) (Writer, error) {
		return newDeflateWriter(writer, level)
	})
}
