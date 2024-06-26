package metrics

import (
	"testing"

	"github.com/dashenmiren/EdgeCommon/pkg/serverconfigs"
)

func TestNewManager(t *testing.T) {
	var manager = NewManager()
	{
		manager.Update([]*serverconfigs.MetricItemConfig{})
		for _, task := range manager.taskMap {
			t.Log(task.item.Id)
		}
	}
	{
		t.Log("====")
		manager.Update([]*serverconfigs.MetricItemConfig{
			{
				Id: 1,
			},
			{
				Id: 2,
			},
			{
				Id: 3,
			},
		})
		for _, task := range manager.taskMap {
			t.Log("task:", task.item.Id)
		}
	}

	{
		t.Log("====")
		manager.Update([]*serverconfigs.MetricItemConfig{
			{
				Id: 1,
			},
			{
				Id: 2,
			},
		})
		for _, task := range manager.taskMap {
			t.Log("task:", task.item.Id)
		}
	}

	{
		t.Log("====")
		manager.Update([]*serverconfigs.MetricItemConfig{
			{
				Id:      1,
				Version: 1,
			},
		})
		for _, task := range manager.taskMap {
			t.Log("task:", task.item.Id)
		}
	}
}
