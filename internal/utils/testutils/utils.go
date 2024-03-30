package testutils

import "os"

// IsSingleTesting 判断当前测试环境是否为单个函数测试
func IsSingleTesting() bool {
	for _, arg := range os.Args {
		if arg == "-test.run" {
			return true
		}
	}
	return false
}
