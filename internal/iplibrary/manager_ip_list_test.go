package iplibrary

import (
	"testing"
	"time"

	"github.com/dashenmiren/EdgeNode/internal/utils"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"github.com/iwind/TeaGo/logs"
)

func TestIPListManager_init(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var manager = NewIPListManager()
	manager.init()
	t.Log(manager.listMap)
	t.Log(SharedServerListManager.blackMap)
	logs.PrintAsJSON(GlobalBlackIPList.sortedItems, t)
}

func TestIPListManager_check(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var manager = NewIPListManager()
	manager.init()

	var before = time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()
	t.Log(SharedServerListManager.FindBlackList(23, true).Contains(utils.IP2Long("127.0.0.2")))
	t.Log(GlobalBlackIPList.Contains(utils.IP2Long("127.0.0.6")))
}

func TestIPListManager_loop(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var manager = NewIPListManager()
	manager.Start()
	err := manager.loop()
	if err != nil {
		t.Fatal(err)
	}
}
