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
