package rpc_test

import (
	"testing"

	"github.com/TeaOSLab/EdgeNode/internal/rpc"
)

func TestNewCallStat(t *testing.T) {
	var stat = rpc.NewCallStat(10)
	stat.Add(true, 1)
	stat.Add(true, 2)
	stat.Add(true, 3)
	stat.Add(false, 4)
	stat.Add(true, 0)
	stat.Add(true, 1)
	t.Log(stat.Sum())
}
