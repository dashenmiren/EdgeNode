// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package bfs

import (
	fsutils "github.com/dashenmiren/EdgeNode/internal/utils/fs"
	memutils "github.com/dashenmiren/EdgeNode/internal/utils/mem"
	"time"
)

type FSOptions struct {
	MaxOpenFiles int
	BytesPerSync int64
	SyncTimeout  time.Duration
	MaxSyncFiles int
}

func (this *FSOptions) EnsureDefaults() {
	if this.MaxOpenFiles <= 0 {
		// 根据内存计算最大打开文件数
		var maxOpenFiles = memutils.SystemMemoryGB() * 128
		if maxOpenFiles > (8 << 10) {
			maxOpenFiles = 8 << 10
		}
		this.MaxOpenFiles = maxOpenFiles
	}
	if this.BytesPerSync <= 0 {
		if fsutils.DiskIsFast() {
			this.BytesPerSync = 1 << 20 // TODO 根据硬盘实际写入速度进行调整
		} else {
			this.BytesPerSync = 512 << 10
		}
	}
	if this.SyncTimeout <= 0 {
		this.SyncTimeout = 1 * time.Second
	}
	if this.MaxSyncFiles <= 0 {
		this.MaxSyncFiles = 32
	}
}

var DefaultFSOptions = &FSOptions{
	MaxOpenFiles: 1 << 10,
	BytesPerSync: 512 << 10,
	SyncTimeout:  1 * time.Second,
	MaxSyncFiles: 32,
}
