// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .
//go:build !linux

package executils

import "os/exec"

func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
