package goman_test

import (
	"runtime"
	"testing"

	"github.com/dashenmiren/EdgeNode/internal/goman"
)

func TestNewTaskGroup(t *testing.T) {
	var group = goman.NewTaskGroup()
	var m = map[int]bool{}

	for i := 0; i < runtime.NumCPU()*2; i++ {
		var index = i
		group.Run(func() {
			t.Log("task", index)

			group.Lock()
			_, ok := m[index]
			if ok {
				t.Error("duplicated:", index)
			}
			m[index] = true
			group.Unlock()
		})
	}
	group.Wait()
}
