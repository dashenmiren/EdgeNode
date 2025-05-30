// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved.

package linkedlist_test

import (
	"github.com/dashenmiren/EdgeNode/internal/utils/linkedlist"
	"github.com/iwind/TeaGo/types"
	"runtime"
	"strconv"
	"testing"
)

func TestNewList_Memory(t *testing.T) {
	var stat1 = &runtime.MemStats{}
	runtime.ReadMemStats(stat1)

	var list = linkedlist.NewList[int]()
	for i := 0; i < 1_000_000; i++ {
		var item = &linkedlist.Item[int]{}
		list.Push(item)
	}

	runtime.GC()

	var stat2 = &runtime.MemStats{}
	runtime.ReadMemStats(stat2)
	t.Log((stat2.HeapInuse-stat1.HeapInuse)>>20, "MB")
	t.Log(list.Len())

	var count = 0
	list.Range(func(item *linkedlist.Item[int]) (goNext bool) {
		count++
		return true
	})
	t.Log(count)
}

func TestNewList_Memory_String(t *testing.T) {
	var stat1 = &runtime.MemStats{}
	runtime.ReadMemStats(stat1)

	var list = linkedlist.NewList[string]()
	for i := 0; i < 1_000_000; i++ {
		var item = &linkedlist.Item[string]{}
		item.Value = strconv.Itoa(i)
		list.Push(item)
	}

	runtime.GC()

	var stat2 = &runtime.MemStats{}
	runtime.ReadMemStats(stat2)
	t.Log((stat2.HeapInuse-stat1.HeapInuse)>>20, "MB")
	t.Log(list.Len())
}

func TestList_Push(t *testing.T) {
	var list = linkedlist.NewList[int]()
	list.Push(linkedlist.NewItem(1))
	list.Push(linkedlist.NewItem(2))

	var item3 = linkedlist.NewItem(3)
	list.Push(item3)

	var item4 = linkedlist.NewItem(4)
	list.Push(item4)
	list.Range(func(item *linkedlist.Item[int]) (goNext bool) {
		t.Log(item.Value)
		return true
	})

	t.Log("=== after push 3 ===")
	list.Push(item3)
	list.Range(func(item *linkedlist.Item[int]) (goNext bool) {
		t.Log(item.Value)
		return true
	})

	t.Log("=== after push 4 ===")
	list.Push(item4)
	list.Push(item3)
	list.Push(item3)
	list.Push(item3)
	list.Push(item4)
	list.Push(item4)
	list.Range(func(item *linkedlist.Item[int]) (goNext bool) {
		t.Log(item.Value)
		return true
	})

	t.Log("=== after remove 3 ===")
	list.Remove(item3)
	list.Range(func(item *linkedlist.Item[int]) (goNext bool) {
		t.Log(item.Value)
		return true
	})
}

func TestList_Shift(t *testing.T) {
	var list = linkedlist.NewList[int]()
	list.Push(linkedlist.NewItem(1))
	list.Push(linkedlist.NewItem(2))
	list.Push(linkedlist.NewItem(3))
	list.Push(linkedlist.NewItem(4))

	for i := 0; i < 10; i++ {
		t.Log("=== before shift " + types.String(i) + " ===")
		list.Range(func(item *linkedlist.Item[int]) (goNext bool) {
			t.Log(item.Value)
			return true
		})

		t.Logf("shift: %+v", list.Shift())

		t.Log("=== after shift  " + types.String(i) + " ===")
		list.Range(func(item *linkedlist.Item[int]) (goNext bool) {
			t.Log(item.Value)
			return true
		})
	}
}

func TestList_RangeReverse(t *testing.T) {
	var list = linkedlist.NewList[int]()
	list.Push(linkedlist.NewItem(1))
	list.Push(linkedlist.NewItem(2))

	var item3 = linkedlist.NewItem(3)
	list.Push(item3)

	list.Push(linkedlist.NewItem(4))

	//list.Push(item3)
	//list.Remove(item3)
	list.RangeReverse(func(item *linkedlist.Item[int]) (goNext bool) {
		t.Log(item.Value)
		return true
	})
}

func BenchmarkList_Add(b *testing.B) {
	var list = linkedlist.NewList[int]()
	for i := 0; i < b.N; i++ {
		var item = &linkedlist.Item[int]{}
		list.Push(item)
	}
}
