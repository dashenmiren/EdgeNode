// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package fsutils_test

import (
	fsutils "github.com/dashenmiren/EdgeNode/internal/utils/fs"
	"github.com/dashenmiren/EdgeNode/internal/utils/testutils"
	"github.com/iwind/TeaGo/assert"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestLimiter_SetThreads(t *testing.T) {
	var limiter = fsutils.NewLimiter(4)

	var concurrent = 1024

	var wg = sync.WaitGroup{}
	wg.Add(concurrent)

	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()

			limiter.SetThreads(rand.Int() % 128)
			limiter.TryAck()
		}()
	}

	wg.Wait()
}

func TestLimiter_Ack(t *testing.T) {
	var a = assert.NewAssertion(t)

	{
		var limiter = fsutils.NewLimiter(4)
		a.IsTrue(limiter.FreeThreads() == 4)
		limiter.Ack()
		a.IsTrue(limiter.FreeThreads() == 3)
		limiter.Ack()
		a.IsTrue(limiter.FreeThreads() == 2)
		limiter.Release()
		a.IsTrue(limiter.FreeThreads() == 3)
		limiter.Release()
		a.IsTrue(limiter.FreeThreads() == 4)
	}
}

func TestLimiter_TryAck(t *testing.T) {
	var a = assert.NewAssertion(t)

	{
		var limiter = fsutils.NewLimiter(4)
		var count = limiter.FreeThreads()
		a.IsTrue(count == 4)
		for i := 0; i < count; i++ {
			limiter.Ack()
		}
		a.IsTrue(limiter.FreeThreads() == 0)
		a.IsFalse(limiter.TryAck())
		a.IsTrue(limiter.FreeThreads() == 0)
	}

	{
		var limiter = fsutils.NewLimiter(4)
		var count = limiter.FreeThreads()
		a.IsTrue(count == 4)
		for i := 0; i < count-1; i++ {
			limiter.Ack()
		}
		a.IsTrue(limiter.FreeThreads() == 1)
		a.IsTrue(limiter.TryAck())
		a.IsTrue(limiter.FreeThreads() == 0)
	}
}

func TestLimiter_TryAck2(t *testing.T) {
	if !testutils.IsSingleTesting() {
		return
	}

	var a = assert.NewAssertion(t)

	{
		var limiter = fsutils.NewLimiter(4)
		var count = limiter.FreeThreads()
		a.IsTrue(count == 4)
		for i := 0; i < count-1; i++ {
			limiter.Ack()
		}
		a.IsTrue(limiter.FreeThreads() == 1)
		a.IsTrue(limiter.TryAck())
		a.IsFalse(limiter.TryAck())
		a.IsFalse(limiter.TryAck())

		limiter.Release()
		a.IsTrue(limiter.TryAck())
	}
}

func TestLimiter_Timout(t *testing.T) {
	var timeout = time.NewTimer(100 * time.Millisecond)

	var r = make(chan bool, 1)
	r <- true

	var before = time.Now()
	select {
	case <-r:
	case <-timeout.C:
	}
	t.Log(time.Since(before).Seconds()*1000, "ms")

	timeout.Stop()

	before = time.Now()
	timeout.Reset(100 * time.Millisecond)
	<-timeout.C
	t.Log(time.Since(before).Seconds()*1000, "ms")
}
