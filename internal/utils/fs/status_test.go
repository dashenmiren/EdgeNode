package fsutils_test

import (
	"testing"

	fsutils "github.com/TeaOSLab/EdgeNode/internal/utils/fs"
	"github.com/iwind/TeaGo/assert"
)

func TestWrites(t *testing.T) {
	var a = assert.NewAssertion(t)

	for i := 0; i < int(fsutils.DiskMaxWrites); i++ {
		fsutils.WriteBegin()
	}
	a.IsFalse(fsutils.WriteReady())

	fsutils.WriteEnd()
	a.IsTrue(fsutils.WriteReady())
}

func BenchmarkWrites(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fsutils.WriteReady()
			fsutils.WriteBegin()
			fsutils.WriteEnd()
		}
	})
}
