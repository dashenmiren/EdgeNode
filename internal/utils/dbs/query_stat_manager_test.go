package dbs_test

import (
	"testing"
	"time"

	"github.com/dashenmiren/EdgeNode/internal/utils/dbs"
	"github.com/iwind/TeaGo/logs"
)

func TestQueryStatManager(t *testing.T) {
	var manager = dbs.NewQueryStatManager()
	{
		var label = manager.AddQuery("sql 1")
		time.Sleep(1 * time.Second)
		label.End()
	}
	manager.AddQuery("sql 1").End()
	manager.AddQuery("sql 2").End()
	for _, stat := range manager.TopN(10) {
		logs.PrintAsJSON(stat, t)
	}
}
