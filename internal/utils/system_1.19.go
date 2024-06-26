//go:build go1.19

package utils

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
