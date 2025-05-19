// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package fsutils_test

import (
	"testing"

	fsutils "github.com/dashenmiren/EdgeNode/internal/utils/fs"
)

func TestCheckDiskWritingSpeed(t *testing.T) {
	t.Log(fsutils.CheckDiskWritingSpeed())
}

func TestCheckDiskIsFast(t *testing.T) {
	t.Log(fsutils.CheckDiskIsFast())
}
