// Copyright 2021 GoEdge goedge.cdn@gmail.com. All rights reserved.

package iplibrary

import (
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"testing"
	"time"
)

func TestIPIsAllowed(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var manager = NewIPListManager()
	manager.Init()

	var before = time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()
	t.Log(AllowIP("127.0.0.1", 0))
	t.Log(AllowIP("127.0.0.1", 23))
}
