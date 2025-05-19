// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package percpu_test

import (
	"github.com/dashenmiren/EdgeNode/internal/utils/percpu"
	"github.com/dashenmiren/EdgeNode/internal/utils/zero"
	"testing"
)

func TestChan_C(t *testing.T) {
	var c = percpu.NewChan[zero.Zero](10)
	c.C() <- zero.Zero{}

	t.Log(<-c.C())

	select {
	case <-c.C():
		t.Fatal("should not return from here")
	default:
		t.Log("ok")
	}
}
