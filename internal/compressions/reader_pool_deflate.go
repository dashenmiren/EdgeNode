package compressions

import (
	"io"

	teaconst "github.com/dashenmiren/EdgeNode/internal/const"
	"github.com/dashenmiren/EdgeNode/internal/utils"
)

var sharedDeflateReaderPool *ReaderPool

func init() {
	if !teaconst.IsMain {
		return
	}

	var maxSize = utils.SystemMemoryGB() * 256
	if maxSize == 0 {
		maxSize = 256
	}
	sharedDeflateReaderPool = NewReaderPool(maxSize, func(reader io.Reader) (Reader, error) {
		return newDeflateReader(reader)
	})
}
