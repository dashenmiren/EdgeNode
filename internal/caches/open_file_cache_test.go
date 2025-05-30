// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package caches_test

import (
	"github.com/dashenmiren/EdgeNode/internal/caches"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"github.com/iwind/TeaGo/types"
	"testing"
	"time"
)

func TestNewOpenFileCache_Close(t *testing.T) {
	cache, err := caches.NewOpenFileCache(1024)
	if err != nil {
		t.Fatal(err)
	}
	cache.Debug()
	cache.Put("a.txt", caches.NewOpenFile(nil, nil, nil, 0, 1<<20))
	cache.Put("b.txt", caches.NewOpenFile(nil, nil, nil, 0, 1<<20))
	cache.Put("b.txt", caches.NewOpenFile(nil, nil, nil, 0, 1<<20))
	cache.Put("b.txt", caches.NewOpenFile(nil, nil, nil, 0, 1<<20))
	cache.Put("c.txt", caches.NewOpenFile(nil, nil, nil, 0, 1<<20))

	cache.Get("b.txt")
	cache.Get("d.txt") // not exist
	cache.Close("a.txt")

	if testutils.IsSingleTesting() {
		time.Sleep(100 * time.Second)
	}
}

func TestNewOpenFileCache_OverSize(t *testing.T) {
	cache, err := caches.NewOpenFileCache(1024)
	if err != nil {
		t.Fatal(err)
	}

	cache.SetCapacity(1 << 30)

	cache.Debug()

	for i := 0; i < 100; i++ {
		cache.Put("a"+types.String(i)+".txt", caches.NewOpenFile(nil, nil, nil, 0, 128<<20))
	}

	if testutils.IsSingleTesting() {
		time.Sleep(100 * time.Second)
	}
}

func TestNewOpenFileCache_CloseAll(t *testing.T) {
	cache, err := caches.NewOpenFileCache(1024)
	if err != nil {
		t.Fatal(err)
	}
	cache.Debug()
	cache.Put("a.txt", caches.NewOpenFile(nil, nil, nil, 0, 1))
	cache.Put("b.txt", caches.NewOpenFile(nil, nil, nil, 0, 1))
	cache.Put("c.txt", caches.NewOpenFile(nil, nil, nil, 0, 1))
	cache.Get("b.txt")
	cache.Get("d.txt")
	cache.CloseAll()

	if testutils.IsSingleTesting() {
		time.Sleep(6 * time.Second)
	}
}
