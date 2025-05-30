// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved.

package fnv_test

import (
	"github.com/dashenmiren/EdgeNode/internal/utils/fnv"
	"github.com/iwind/TeaGo/types"
	"testing"
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
			_ = fnv.HashString("abcdefh")
		}
	})
}

func BenchmarkHashString_Long(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = fnv.HashString("HELLO,WORLDHELLO,WORLDHELLO,WORLDHELLO,WORLDHELLO,WORLDHELLO,WORLD")
		}
	})
}
