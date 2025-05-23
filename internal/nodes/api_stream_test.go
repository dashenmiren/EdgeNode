package nodes

import (
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"testing"
)

func TestAPIStream_Start(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	apiStream := NewAPIStream()
	apiStream.Start()
}
