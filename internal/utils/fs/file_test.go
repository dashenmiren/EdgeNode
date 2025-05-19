// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package fsutils_test

import (
	fsutils "github.com/dashenmiren/EdgeNode/internal/utils/fs"
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestFileFlags(t *testing.T) {
	var a = assert.NewAssertion(t)
	a.IsTrue(fsutils.FlagRead&fsutils.FlagRead == fsutils.FlagRead)
	a.IsTrue(fsutils.FlagWrite&fsutils.FlagWrite != fsutils.FlagRead)
	a.IsTrue((fsutils.FlagWrite|fsutils.FlagRead)&fsutils.FlagRead == fsutils.FlagRead)
}
