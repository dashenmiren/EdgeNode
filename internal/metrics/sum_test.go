// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package metrics_test

import (
	"github.com/dashenmiren/EdgeNode/internal/metrics"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"runtime"
	"testing"
)

func BenchmarkSumStat(b *testing.B) {
	runtime.GOMAXPROCS(2)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			metrics.UniqueKey(1, []string{"1.2.3.4"}, timeutil.Format("Ymd"), 1, 1)
		}
	})
}
