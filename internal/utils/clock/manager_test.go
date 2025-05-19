package clock_test

import (
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/utils/clock"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
)

func TestReadServer(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	t.Log(clock.NewClockManager().ReadServer("pool.ntp.org"))
}
