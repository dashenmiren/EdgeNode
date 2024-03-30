package caches

import "github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"

type FileDir struct {
	Path     string
	Capacity *shared.SizeCapacity
	IsFull   bool
}
