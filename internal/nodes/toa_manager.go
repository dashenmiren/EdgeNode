// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .
//go:build !plus

package nodes

import "github.com/dashenmiren/EdgeCommon/pkg/nodeconfigs"

var sharedTOAManager = NewTOAManager()

type TOAManager struct {
}

func NewTOAManager() *TOAManager {
	return &TOAManager{}
}

func (this *TOAManager) Apply(config *nodeconfigs.TOAConfig) error {
	return nil
}

func (this *TOAManager) Config() *nodeconfigs.TOAConfig {
	return nil
}

func (this *TOAManager) Quit() error {
	return nil
}

func (this *TOAManager) SendMsg(msg string) error {
	return nil
}
