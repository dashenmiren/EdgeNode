// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package agents

import (
	"github.com/iwind/TeaGo/logs"
	"testing"
)

func TestNewIPCacheMap(t *testing.T) {
	var cacheMap = NewIPCacheMap(3)

	t.Log("====")
	cacheMap.Add("1")
	cacheMap.Add("2")
	logs.PrintAsJSON(cacheMap.m, t)
	logs.PrintAsJSON(cacheMap.list, t)

	t.Log("====")
	cacheMap.Add("3")
	logs.PrintAsJSON(cacheMap.m, t)
	logs.PrintAsJSON(cacheMap.list, t)

	t.Log("====")
	cacheMap.Add("4")
	logs.PrintAsJSON(cacheMap.m, t)
	logs.PrintAsJSON(cacheMap.list, t)

	t.Log("====")
	cacheMap.Add("3")
	logs.PrintAsJSON(cacheMap.m, t)
	logs.PrintAsJSON(cacheMap.list, t)
}
