package utils

import "testing"

func TestSystemMemoryGB(t *testing.T) {
	t.Log(SystemMemoryGB())
	t.Log(SystemMemoryGB())
	t.Log(SystemMemoryGB())
	t.Log(SystemMemoryBytes())
	t.Log(SystemMemoryBytes())
	t.Log(SystemMemoryBytes()>>30, "GB")
}
