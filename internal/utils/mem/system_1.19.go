// Copyright 2023 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .
//go:build go1.19

package memutils

import (
	"runtime/debug"
)

// 设置软内存最大值
func setMaxMemory(memoryGB int) {
	if memoryGB <= 0 {
		memoryGB = 1
	}

	var maxMemoryBytes = (int64(memoryGB) << 30) * 80 / 100 // 默认 80%
	debug.SetMemoryLimit(maxMemoryBytes)
}
