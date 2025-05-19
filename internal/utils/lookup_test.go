package utils

import "testing"

func TestLookupCNAME(t *testing.T) {
	t.Log(LookupCNAME("www.yun4s.cn"))
}
