// Copyright 2021 GoEdge goedge.cdn@gmail.com. All rights reserved.

package caches_test

import (
	"encoding/json"
	"github.com/dashenmiren/EdgeNode/internal/caches"
	"github.com/dashenmiren/EdgeNode/internal/utils/fasttime"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"github.com/dashenmiren/EdgeNode/internal/utils/zero"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"runtime"
	"testing"
	"time"
)

func TestItem_Marshal(t *testing.T) {
	{
		var item = &caches.Item{}
		data, err := json.Marshal(item)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
	}

	{
		var item = &caches.Item{
			Type:       caches.ItemTypeFile,
			Key:        "https://example.com/index.html",
			ExpiresAt:  fasttime.Now().Unix(),
			HeaderSize: 1 << 10,
			BodySize:   1 << 20,
			MetaSize:   256,
		}
		data, err := json.Marshal(item)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
	}
}

func TestItems_Memory(t *testing.T) {
	var stat = &runtime.MemStats{}
	runtime.ReadMemStats(stat)
	var memory1 = stat.HeapInuse

	var items = []*caches.Item{}
	var count = 100
	if testutils.IsSingleTesting() {
		count = 10_000_000
	}
	for i := 0; i < count; i++ {
		items = append(items, &caches.Item{
			Key: types.String(i),
		})
	}

	runtime.ReadMemStats(stat)
	var memory2 = stat.HeapInuse

	t.Log(memory1, memory2, (memory2-memory1)/1024/1024, "M")

	runtime.ReadMemStats(stat)
	var memory3 = stat.HeapInuse
	t.Log(memory2, memory3, (memory3-memory2)/1024/1024, "M")

	if testutils.IsSingleTesting() {
		time.Sleep(1 * time.Second)
	}
}

func TestItems_Memory2(t *testing.T) {
	var stat = &runtime.MemStats{}
	runtime.ReadMemStats(stat)
	var memory1 = stat.HeapInuse

	var items = map[int32]map[string]zero.Zero{}
	var count = 100
	if testutils.IsSingleTesting() {
		count = 10_000_000
	}

	for i := 0; i < count; i++ {
		var week = int32((time.Now().Unix() - int64(86400*rands.Int(0, 300))) / (86400 * 7))
		m, ok := items[week]
		if !ok {
			m = map[string]zero.Zero{}
			items[week] = m
		}
		m[types.String(int64(i)*1_000_000)] = zero.New()
	}

	runtime.ReadMemStats(stat)
	var memory2 = stat.HeapInuse

	t.Log(memory1, memory2, (memory2-memory1)/1024/1024, "M")

	if testutils.IsSingleTesting() {
		time.Sleep(1 * time.Second)
	}
	for w, i := range items {
		t.Log(w, len(i))
	}
}

func TestItem_RequestURI(t *testing.T) {
	for _, u := range []string{
		"https://cdn.foyeseo.com/hello/world",
		"https://cdn.foyeseo.com:8080/hello/world",
		"https://cdn.foyeseo.com/hello/world?v=1&t=123",
	} {
		var item = &caches.Item{Key: u}
		t.Log(u, "=>", item.RequestURI())
	}
}
