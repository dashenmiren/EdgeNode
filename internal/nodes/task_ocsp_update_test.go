package nodes_test

import (
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/nodes"
)

func TestOCSPUpdateTask_Loop(t *testing.T) {
	var task = &nodes.OCSPUpdateTask{}
	err := task.Loop()
	if err != nil {
		t.Fatal(err)
	}
}
