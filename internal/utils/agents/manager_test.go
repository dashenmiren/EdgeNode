package agents_test

import (
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/utils/agents"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
)

func TestNewManager(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var db = agents.NewDB(Tea.Root + "/data/agents.db")
	err := db.Init()
	if err != nil {
		t.Fatal(err)
	}

	var manager = agents.NewManager()
	manager.SetDB(db)
	err = manager.Load()
	if err != nil {
		t.Fatal(err)
	}

	_, err = manager.Loop()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(manager.LookupIP("192.168.3.100"))
}
