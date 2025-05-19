//go:build !plus

package nodes

// 检查套餐
func (this *HTTPRequest) doPlanBefore() (blocked bool) {
	// stub
	return false
}
