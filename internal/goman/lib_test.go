package goman_test

import (
	"testing"
	"time"

	"github.com/dashenmiren/EdgeNode/internal/goman"
)

func TestNew(t *testing.T) {
	goman.New(func() {
		t.Log("Hello")

		t.Log(goman.List())
	})

	time.Sleep(1 * time.Second)
	t.Log(goman.List())

	time.Sleep(1 * time.Second)
}

func TestNewWithArgs(t *testing.T) {
	goman.NewWithArgs(func(args ...interface{}) {
		t.Log(args[0], args[1])
	}, 1, 2)
	time.Sleep(1 * time.Second)
}
