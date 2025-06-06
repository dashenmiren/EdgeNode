// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package fsutils

import (
	"runtime"
	"time"
)

var maxThreads = runtime.NumCPU()
var WriterLimiter = NewLimiter(max(maxThreads*16, 32))
var ReaderLimiter = NewLimiter(max(maxThreads*16, 32))

type Limiter struct {
	threads      chan struct{}
	count        int
	countDefault int
	timers       chan *time.Timer
}

func NewLimiter(threads int) *Limiter {
	if threads < 32 {
		threads = 32
	}
	if threads > 1024 {
		threads = 1024
	}

	var threadsChan = make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		threadsChan <- struct{}{}
	}

	return &Limiter{
		countDefault: threads,
		count:        threads,
		threads:      threadsChan,
		timers:       make(chan *time.Timer, 4096),
	}
}

func (this *Limiter) SetThreads(newThreads int) {
	if newThreads <= 0 {
		newThreads = this.countDefault
	}

	if newThreads != this.count {
		var threadsChan = make(chan struct{}, newThreads)
		for i := 0; i < newThreads; i++ {
			threadsChan <- struct{}{}
		}

		this.threads = threadsChan
		this.count = newThreads
	}
}

func (this *Limiter) Ack() {
	<-this.threads
}

func (this *Limiter) TryAck() bool {
	const timeoutDuration = 500 * time.Millisecond

	var timeout *time.Timer
	select {
	case timeout = <-this.timers:
		timeout.Reset(timeoutDuration)
	default:
		timeout = time.NewTimer(timeoutDuration)
	}

	defer func() {
		timeout.Stop()

		select {
		case this.timers <- timeout:
		default:
		}
	}()

	select {
	case <-this.threads:
		return true
	case <-timeout.C:
		return false
	}
}

func (this *Limiter) Release() {
	select {
	case this.threads <- struct{}{}:
	default:
		// 由于容量可能有变化，这里忽略多余的thread
	}
}

func (this *Limiter) FreeThreads() int {
	return len(this.threads)
}
