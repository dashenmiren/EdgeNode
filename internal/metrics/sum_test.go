package metrics_test

import (
	"runtime"
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/metrics"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

func BenchmarkSumStat(b *testing.B) {
	runtime.GOMAXPROCS(2)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			metrics.SumStat(1, []string{"1.2.3.4"}, timeutil.Format("Ymd"), 1, 1)
		}
	})
}
