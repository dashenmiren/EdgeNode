// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package caches_test

import (
	"fmt"
	"github.com/dashenmiren/EdgeNode/internal/caches"
	"github.com/cespare/xxhash/v2"
	"github.com/iwind/TeaGo/types"
	"strconv"
	"testing"
)

func TestParseHost(t *testing.T) {
	for _, u := range []string{
		"https://cdn.foyeseo.com/hello/world",
		"https://cdn.foyeseo.com:8080/hello/world",
		"https://cdn.foyeseo.com/hello/world?v=1&t=123",
		"https://[::1]:1234/hello/world?v=1&t=123",
		"https://[::1]/hello/world?v=1&t=123",
		"https://127.0.0.1/hello/world?v=1&t=123",
		"https:/hello/world?v=1&t=123",
		"123456",
	} {
		t.Log(u, "=>", caches.ParseHost(u))
	}
}

func TestUintString(t *testing.T) {
	t.Log(strconv.FormatUint(xxhash.Sum64String("https://cdn.foyeseo.com/"), 10))
	t.Log(strconv.FormatUint(123456789, 10))
	t.Logf("%d", 1234567890123)
}

func BenchmarkUint_String(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = strconv.FormatUint(1234567890123, 10)
	}
}

func BenchmarkUint_String2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = types.String(1234567890123)
	}
}

func BenchmarkUint_String3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%d", 1234567890123)
	}
}
