package nodes

import (
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
)

func TestAPIStream_Start(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	apiStream := NewAPIStream()
	apiStream.Start()
}
