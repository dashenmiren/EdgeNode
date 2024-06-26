package iplibrary

import (
	"runtime"
	"testing"
	"time"

	"github.com/dashenmiren/EdgeNode/internal/utils"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"github.com/iwind/TeaGo/assert"
)

func TestIPItem_Contains(t *testing.T) {
	a := assert.NewAssertion(t)

	{
		item := &IPItem{
			IPFrom:    utils.IP2Long("192.168.1.100"),
			IPTo:      0,
			ExpiredAt: 0,
		}
		a.IsTrue(item.Contains(utils.IP2Long("192.168.1.100")))
	}

	{
		item := &IPItem{
			IPFrom:    utils.IP2Long("192.168.1.100"),
			IPTo:      0,
			ExpiredAt: time.Now().Unix() + 1,
		}
		a.IsTrue(item.Contains(utils.IP2Long("192.168.1.100")))
	}

	{
		item := &IPItem{
			IPFrom:    utils.IP2Long("192.168.1.100"),
			IPTo:      0,
			ExpiredAt: time.Now().Unix() - 1,
		}
		a.IsFalse(item.Contains(utils.IP2Long("192.168.1.100")))
	}
	{
		item := &IPItem{
			IPFrom:    utils.IP2Long("192.168.1.100"),
			IPTo:      0,
			ExpiredAt: 0,
		}
		a.IsFalse(item.Contains(utils.IP2Long("192.168.1.101")))
	}

	{
		item := &IPItem{
			IPFrom:    utils.IP2Long("192.168.1.1"),
			IPTo:      utils.IP2Long("192.168.1.101"),
			ExpiredAt: 0,
		}
		a.IsTrue(item.Contains(utils.IP2Long("192.168.1.100")))
	}

	{
		item := &IPItem{
			IPFrom:    utils.IP2Long("192.168.1.1"),
			IPTo:      utils.IP2Long("192.168.1.100"),
			ExpiredAt: 0,
		}
		a.IsTrue(item.Contains(utils.IP2Long("192.168.1.100")))
	}

	{
		item := &IPItem{
			IPFrom:    utils.IP2Long("192.168.1.1"),
			IPTo:      utils.IP2Long("192.168.1.101"),
			ExpiredAt: 0,
		}
		a.IsTrue(item.Contains(utils.IP2Long("192.168.1.1")))
	}
}

func TestIPItem_Memory(t *testing.T) {
	var isSingleTest = testutils.IsSingleTesting()

	var list = NewIPList()
	var count = 100
	if isSingleTest {
		count = 2_000_000
	}
	for i := 0; i < count; i++ {
		list.Add(&IPItem{
			Type:       "ip",
			Id:         uint64(i),
			IPFrom:     utils.IP2Long("192.168.1.1"),
			IPTo:       0,
			ExpiredAt:  time.Now().Unix(),
			EventLevel: "",
		})
	}
	t.Log("waiting")
	if isSingleTest {
		time.Sleep(10 * time.Second)
	}
}

func BenchmarkIPItem_Contains(b *testing.B) {
	runtime.GOMAXPROCS(1)

	item := &IPItem{
		IPFrom:    utils.IP2Long("192.168.1.1"),
		IPTo:      utils.IP2Long("192.168.1.101"),
		ExpiredAt: 0,
	}
	ip := utils.IP2Long("192.168.1.1")
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10_000; j++ {
			item.Contains(ip)
		}
	}
}
