package iplibrary_test

import (
	"github.com/dashenmiren/EdgeCommon/pkg/iputils"
	"github.com/dashenmiren/EdgeNode/internal/iplibrary"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"github.com/iwind/TeaGo/logs"
	"testing"
	"time"
)

func TestIPListManager_init(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var manager = iplibrary.NewIPListManager()
	manager.Init()
	t.Log(manager.ListMap())
	t.Log(iplibrary.SharedServerListManager.BlackMap())
	logs.PrintAsJSON(iplibrary.GlobalBlackIPList.SortedRangeItems(), t)
}

func TestIPListManager_check(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var manager = iplibrary.NewIPListManager()
	manager.Init()

	var before = time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()
	t.Log(iplibrary.SharedServerListManager.FindBlackList(23, true).Contains(iputils.ToBytes("127.0.0.2")))
	t.Log(iplibrary.GlobalBlackIPList.Contains(iputils.ToBytes("127.0.0.6")))
}

func TestIPListManager_loop(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var manager = iplibrary.NewIPListManager()
	manager.Start()
	err := manager.Loop()
	if err != nil {
		t.Fatal(err)
	}
}
