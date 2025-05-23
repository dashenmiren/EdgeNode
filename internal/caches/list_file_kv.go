// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .

package caches

import (
	"fmt"
	"github.com/dashenmiren/EdgeNode/internal/remotelogs"
	"github.com/dashenmiren/EdgeNode/internal/utils/fasttime"
	"github.com/dashenmiren/EdgeNode/internal/utils/fnv"
	"github.com/dashenmiren/EdgeNode/internal/utils/goman"
	memutils "github.com/dashenmiren/EdgeNode/internal/utils/mem"
	"github.com/dashenmiren/EdgeNode/internal/utils/ttlcache"
	"github.com/iwind/TeaGo/types"
	"strings"
	"testing"
)

const countKVStores = 10

type KVFileList struct {
	dir    string
	stores [countKVStores]*KVListFileStore

	onAdd    func(item *Item)
	onRemove func(item *Item)

	memCache *ttlcache.Cache[int64]
}

func NewKVFileList(dir string) *KVFileList {
	var memGB = memutils.SystemMemoryGB()
	if memGB <= 0 {
		memGB = 1
	}
	var maxCachePieces = 32
	var maxCacheItems = memGB << 15

	var memCache = ttlcache.NewCache[int64](ttlcache.NewPiecesOption(maxCachePieces), ttlcache.NewMaxItemsOption(maxCacheItems))

	dir = strings.TrimSuffix(dir, "/")

	var stores = [countKVStores]*KVListFileStore{}
	for i := 0; i < countKVStores; i++ {
		stores[i] = NewKVListFileStore(dir+"/db-"+types.String(i)+".store", memCache)
	}

	return &KVFileList{
		dir:      dir,
		stores:   stores,
		memCache: memCache,
	}
}

// Init 初始化
func (this *KVFileList) Init() error {
	remotelogs.Println("CACHE", "loading database from '"+this.dir+"' ...")

	var group = goman.NewTaskGroup()
	var lastErr error

	for _, store := range this.stores {
		var storeCopy = store
		group.Run(func() {
			err := storeCopy.Open()
			if err != nil {
				lastErr = fmt.Errorf("open store '"+storeCopy.Path()+"' failed: %w", err)
			}
		})
	}
	group.Wait()

	return lastErr
}

// Reset 重置数据
func (this *KVFileList) Reset() error {
	this.memCache.Clean()
	return nil
}

// Add 添加内容
func (this *KVFileList) Add(hash string, item *Item) error {
	err := this.getStore(hash).AddItem(hash, item)
	if err != nil {
		return err
	}

	if this.onAdd != nil {
		this.onAdd(item)
	}

	if item.ExpiresAt > 0 {
		this.memCache.Write(hash, item.HeaderSize+item.BodySize, min(item.ExpiresAt, fasttime.Now().Unix()+3600))
	}

	return nil
}

// Exist 检查内容是否存在
func (this *KVFileList) Exist(hash string) (bool, int64, error) {
	// read from cache
	var cacheItem = this.memCache.Read(hash)
	if cacheItem != nil {
		return true, cacheItem.Value, nil
	}

	return this.getStore(hash).ExistItem(hash)
}

// ExistQuick 快速检查内容是否存在
func (this *KVFileList) ExistQuick(hash string) (bool, error) {
	// read from cache
	if this.memCache.Read(hash) != nil {
		return true, nil
	}

	return this.getStore(hash).ExistQuickItem(hash)
}

// CleanPrefix 清除某个前缀的缓存
func (this *KVFileList) CleanPrefix(prefix string) error {
	var group = goman.NewTaskGroup()
	var lastErr error
	for _, store := range this.stores {
		var storeCopy = store
		group.Run(func() {
			err := storeCopy.CleanItemsWithPrefix(prefix)
			if err != nil {
				lastErr = err
			}
		})
	}
	group.Wait()
	return lastErr
}

// CleanMatchKey 清除通配符匹配的Key
func (this *KVFileList) CleanMatchKey(key string) error {
	var group = goman.NewTaskGroup()
	var lastErr error
	for _, store := range this.stores {
		var storeCopy = store
		group.Run(func() {
			err := storeCopy.CleanItemsWithWildcardKey(key)
			if err != nil {
				lastErr = err
			}
		})
	}
	group.Wait()
	return lastErr
}

// CleanMatchPrefix 清除通配符匹配的前缀
func (this *KVFileList) CleanMatchPrefix(prefix string) error {
	var group = goman.NewTaskGroup()
	var lastErr error
	for _, store := range this.stores {
		var storeCopy = store
		group.Run(func() {
			err := storeCopy.CleanItemsWithWildcardPrefix(prefix)
			if err != nil {
				lastErr = err
			}
		})
	}
	group.Wait()
	return lastErr
}

// Remove 删除内容
func (this *KVFileList) Remove(hash string) error {
	err := this.getStore(hash).RemoveItem(hash)
	if err != nil {
		return err
	}

	if this.onRemove != nil {
		// when remove file item, no any extra information needed
		this.onRemove(nil)
	}

	// remove from cache
	this.memCache.Delete(hash)

	return nil
}

// Purge 清理过期数据
func (this *KVFileList) Purge(count int, callback func(hash string) error) (int, error) {
	count /= countKVStores
	if count <= 0 {
		count = 100
	}

	var countFound = 0
	var lastErr error
	for _, store := range this.stores {
		purgeCount, err := store.PurgeItems(count, callback)
		countFound += purgeCount
		if err != nil {
			lastErr = err
		}
	}

	return countFound, lastErr
}

// PurgeLFU 清理LFU数据
func (this *KVFileList) PurgeLFU(count int, callback func(hash string) error) error {
	count /= countKVStores
	if count <= 0 {
		count = 100
	}

	var lastErr error
	for _, store := range this.stores {
		err := store.PurgeLFUItems(count, callback)
		if err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// CleanAll 清除所有缓存
func (this *KVFileList) CleanAll() error {
	var group = goman.NewTaskGroup()
	var lastErr error
	for _, store := range this.stores {
		var storeCopy = store
		group.Run(func() {
			err := storeCopy.RemoveAllItems()
			if err != nil {
				lastErr = err
			}
		})
	}
	group.Wait()

	this.memCache.Clean()

	return lastErr
}

// Stat 统计
func (this *KVFileList) Stat(check func(hash string) bool) (*Stat, error) {
	var stat = &Stat{}

	var group = goman.NewTaskGroup()

	var lastErr error
	for _, store := range this.stores {
		var storeCopy = store
		group.Run(func() {
			storeStat, err := storeCopy.StatItems()
			if err != nil {
				lastErr = err
				return
			}

			group.Lock()
			stat.Size += storeStat.Size
			stat.ValueSize += storeStat.ValueSize
			stat.Count += storeStat.Count
			group.Unlock()
		})
	}

	group.Wait()

	return stat, lastErr
}

// Count 总数量
func (this *KVFileList) Count() (int64, error) {
	var count int64

	var group = goman.NewTaskGroup()

	var lastErr error
	for _, store := range this.stores {
		var storeCopy = store
		group.Run(func() {
			countStoreItems, err := storeCopy.CountItems()
			if err != nil {
				lastErr = err
				return
			}

			group.Lock()
			count += countStoreItems
			group.Unlock()
		})
	}

	group.Wait()

	return count, lastErr
}

// OnAdd 添加事件
func (this *KVFileList) OnAdd(fn func(item *Item)) {
	this.onAdd = fn
}

// OnRemove 删除事件
func (this *KVFileList) OnRemove(fn func(item *Item)) {
	this.onRemove = fn
}

// Close 关闭
func (this *KVFileList) Close() error {
	var lastErr error
	var group = goman.NewTaskGroup()
	for _, store := range this.stores {
		var storeCopy = store
		group.Run(func() {
			err := storeCopy.Close()
			if err != nil {
				lastErr = err
			}
		})
	}
	group.Wait()

	this.memCache.Destroy()

	return lastErr
}

// IncreaseHit 增加点击量
func (this *KVFileList) IncreaseHit(hash string) error {
	// do nothing
	return nil
}

func (this *KVFileList) TestInspect(t *testing.T) error {
	for _, store := range this.stores {
		err := store.TestInspect(t)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *KVFileList) getStore(hash string) *KVListFileStore {
	return this.stores[fnv.HashString(hash)%countKVStores]
}
