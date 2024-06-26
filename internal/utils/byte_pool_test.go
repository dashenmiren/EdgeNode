package utils_test

import (
	"bytes"
	"runtime"
	"sync"
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/utils"
)

func TestBytePool_Memory(t *testing.T) {
	var stat1 = &runtime.MemStats{}
	runtime.ReadMemStats(stat1)

	var pool = utils.NewBytePool(32 * 1024)
	for i := 0; i < 20480; i++ {
		pool.Put(make([]byte, 32*1024))
	}

	//pool.Purge()

	//time.Sleep(60 * time.Second)

	runtime.GC()

	var stat2 = &runtime.MemStats{}
	runtime.ReadMemStats(stat2)
	t.Log((stat2.HeapInuse-stat1.HeapInuse)/1024/1024, "MB,")
}

func BenchmarkBytePool_Get(b *testing.B) {
	runtime.GOMAXPROCS(1)

	var pool = utils.NewBytePool(1)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf = pool.Get()
		_ = buf
		pool.Put(buf)
	}
}

func BenchmarkBytePool_Get_Parallel(b *testing.B) {
	runtime.GOMAXPROCS(1)

	var pool = utils.NewBytePool(1024)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var buf = pool.Get()
			pool.Put(buf)
		}
	})
}

func BenchmarkBytePool_Get_Sync(b *testing.B) {
	runtime.GOMAXPROCS(1)

	var pool = &sync.Pool{
		New: func() any {
			return make([]byte, 1024)
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var buf = pool.Get()
			pool.Put(buf)
		}
	})
}

func BenchmarkBytePool_Copy(b *testing.B) {
	var data = bytes.Repeat([]byte{'A'}, 8<<10)

	var pool = &sync.Pool{
		New: func() any {
			return make([]byte, 8<<10)
		},
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var buf = pool.Get().([]byte)
			copy(buf, data)
			pool.Put(buf)
		}
	})
}
