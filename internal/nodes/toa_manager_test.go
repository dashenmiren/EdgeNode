package nodes

import (
	"github.com/dashenmiren/EdgeCommon/pkg/nodeconfigs"
	"testing"
)

func TestTOAManager_Run(t *testing.T) {
	manager := NewTOAManager()
	err := manager.Run(&nodeconfigs.TOAConfig{
		IsOn: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
