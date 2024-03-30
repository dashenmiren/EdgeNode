package ttlcache

import (
	"sync"
	"time"

	"github.com/dashenmiren/EdgeNode/internal/goman"
	"github.com/dashenmiren/EdgeNode/internal/zero"
)

var SharedManager = NewManager()

type GCAble interface {
	GC()
}

type Manager struct {
	ticker *time.Ticker
	locker sync.Mutex

	cacheMap map[GCAble]zero.Zero
}

func NewManager() *Manager {
	var manager = &Manager{
		ticker:   time.NewTicker(2 * time.Second),
		cacheMap: map[GCAble]zero.Zero{},
	}

	goman.New(func() {
		manager.init()
	})

	return manager
}

func (this *Manager) init() {
	for range this.ticker.C {
		this.locker.Lock()
		for cache := range this.cacheMap {
			cache.GC()
		}
		this.locker.Unlock()
	}
}

func (this *Manager) Add(cache GCAble) {
	this.locker.Lock()
	this.cacheMap[cache] = zero.New()
	this.locker.Unlock()
}

func (this *Manager) Remove(cache GCAble) {
	this.locker.Lock()
	delete(this.cacheMap, cache)
	this.locker.Unlock()
}

func (this *Manager) Count() int {
	this.locker.Lock()
	defer this.locker.Unlock()
	return len(this.cacheMap)
}
