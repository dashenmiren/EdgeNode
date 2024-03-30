package fnv_test

import (
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/utils/fnv"
	"github.com/iwind/TeaGo/types"
)

func TestHash(t *testing.T) {
	for _, key := range []string{"costarring", "liquid", "hello"} {
		var h = fnv.HashString(key)
		t.Log(key + " => " + types.String(h))
	}
}

func BenchmarkHashString(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fnv.HashString("abcdefh")
		}
	})
}
