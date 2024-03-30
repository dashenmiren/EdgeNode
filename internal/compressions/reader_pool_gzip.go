package compressions

import (
	"io"

	teaconst "github.com/TeaOSLab/EdgeNode/internal/const"
	"github.com/TeaOSLab/EdgeNode/internal/utils"
)

var sharedGzipReaderPool *ReaderPool

func init() {
	if !teaconst.IsMain {
		return
	}

	var maxSize = utils.SystemMemoryGB() * 256
	if maxSize == 0 {
		maxSize = 256
	}
	sharedGzipReaderPool = NewReaderPool(maxSize, func(reader io.Reader) (Reader, error) {
		return newGzipReader(reader)
	})
}
