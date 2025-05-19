// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package clock_test

import (
	"github.com/dashenmiren/EdgeNode/internal/utils/clock"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"testing"
)

func TestReadServer(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	t.Log(clock.NewClockManager().ReadServer("pool.ntp.org"))
}
