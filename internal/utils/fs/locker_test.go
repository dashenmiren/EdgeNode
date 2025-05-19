package fsutils_test

import (
	"testing"

	fsutils "github.com/dashenmiren/EdgeNode/internal/utils/fs"
)

func TestLocker_Lock(t *testing.T) {
	var path = "/tmp/file-test"
	var locker = fsutils.NewLocker(path)
	err := locker.Lock()
	if err != nil {
		t.Fatal(err)
	}
	_ = locker.Release()

	var locker2 = fsutils.NewLocker(path)
	err = locker2.Lock()
	if err != nil {
		t.Fatal(err)
	}
}
