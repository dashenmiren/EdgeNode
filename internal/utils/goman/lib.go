// Copyright 2021 GoEdge goedge.cdn@gmail.com. All rights reserved.

package goman

import (
	teaconst "github.com/dashenmiren/EdgeNode/internal/const"
	"runtime"
	"sync"
	"time"
)

var locker = &sync.Mutex{}
var instanceMap = map[uint64]*Instance{} // id => *Instance
var instanceId = uint64(0)

// New 新创建goroutine
func New(f func()) {
	if !teaconst.IsMain {
		return
	}

	_, file, line, _ := runtime.Caller(1)

	go func() {
		locker.Lock()
		instanceId++

		var instance = &Instance{
			Id:          instanceId,
			CreatedTime: time.Now(),
		}

		instance.File = file
		instance.Line = line

		instanceMap[instanceId] = instance
		locker.Unlock()

		// run function
		f()

		locker.Lock()
		delete(instanceMap, instanceId)
		locker.Unlock()
	}()
}

// NewWithArgs 创建带有参数的goroutine
func NewWithArgs(f func(args ...interface{}), args ...interface{}) {
	if !teaconst.IsMain {
		return
	}

	_, file, line, _ := runtime.Caller(1)

	go func() {
		locker.Lock()
		instanceId++

		var instance = &Instance{
			Id:          instanceId,
			CreatedTime: time.Now(),
		}

		instance.File = file
		instance.Line = line

		instanceMap[instanceId] = instance
		locker.Unlock()

		// run function
		f(args...)

		locker.Lock()
		delete(instanceMap, instanceId)
		locker.Unlock()
	}()
}

// List 列出所有正在运行goroutine
func List() []*Instance {
	locker.Lock()
	defer locker.Unlock()

	var result = []*Instance{}
	for _, instance := range instanceMap {
		result = append(result, instance)
	}
	return result
}
