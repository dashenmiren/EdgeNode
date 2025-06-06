// Copyright 2023 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package fsutils_test

import (
	"github.com/dashenmiren/EdgeNode/internal/utils/fs"
	"testing"
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
