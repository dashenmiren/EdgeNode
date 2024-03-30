package compressions

import (
	"io"

	teaconst "github.com/dashenmiren/EdgeNode/internal/const"
	"github.com/dashenmiren/EdgeNode/internal/utils"
	"github.com/klauspost/compress/zstd"
)

var sharedZSTDWriterPool *WriterPool

func init() {
	if !teaconst.IsMain {
		return
	}

	var maxSize = utils.SystemMemoryGB() * 256
	if maxSize == 0 {
		maxSize = 256
	}
	sharedZSTDWriterPool = NewWriterPool(maxSize, int(zstd.SpeedBestCompression), func(writer io.Writer, level int) (Writer, error) {
		return newZSTDWriter(writer, level)
	})
}
