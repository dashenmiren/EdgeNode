package compressions

import (
	"io"

	teaconst "github.com/TeaOSLab/EdgeNode/internal/const"
	"github.com/TeaOSLab/EdgeNode/internal/utils"
)

var sharedZSTDReaderPool *ReaderPool

func init() {
	if !teaconst.IsMain {
		return
	}

	var maxSize = utils.SystemMemoryGB() * 256
	if maxSize == 0 {
		maxSize = 256
	}
	sharedZSTDReaderPool = NewReaderPool(maxSize, func(reader io.Reader) (Reader, error) {
		return newZSTDReader(reader)
	})
}
