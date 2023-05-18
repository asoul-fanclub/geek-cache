package cache

import (
	"container/list"
	"github.com/Makonike/geek-cache/geek/hsort_map"
	"sync"
	"time"
)

// lruCache lru缓存
type lruCache struct {
	lock      sync.Mutex
	cacheMap  hsort_map.HSortMap            // map cache
	expires   map[string]time.Time          // 存储每个key的超时时间
	ll        *list.List                    // 用于组织lru的list
	OnEvicted func(key string, value Value) // 记录被删除时的回调函数
	maxBytes  int64                         // 内存大小
	nbytes    int64                         // 当前内存大小
}

// 通过key可以在记录删除时，删除字典缓存中的映射
type entry struct {
	key   string
	value Value
}

type LRUCacheOptions func(cache *lruCache)

func LRUCacheSize(size int64) LRUCacheOptions {
	return func(lruCache *lruCache) {
		lruCache.maxBytes = size
	}
}

func LRUCacheHash(hash hsort_map.Hash) LRUCacheOptions {
	return func(lruCache *lruCache) {
		lruCache.cacheMap = hsort_map.NewHSkipList(hash)
	}
}

func NewLRUCache(opts ...LRUCacheOptions) *lruCache {
	answer := lruCache{
		expires:  make(map[string]time.Time),
		nbytes:   0,
		ll:       list.New(),
		maxBytes: 1000,
	}
	for _, opt := range opts {
		opt(&answer)
	}
	if answer.cacheMap == nil {
		answer.cacheMap = hsort_map.NewHSkipList(func(key string) string {
			return key
		})
	}
	// 启动定期清除
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			answer.periodicMemoryClean()
		}
	}()
	return &answer
}

func (c *lruCache) Get(key string) (Value, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// check for expiration
	expirationTime, ok := c.expires[key]
	if ok && expirationTime.Before(time.Now()) {
		t, _ := c.cacheMap.Get(key)
		v := t.Value.(*entry)
		c.nbytes -= int64(v.value.Len() + len(v.key))
		c.ll.Remove(t)
		c.cacheMap.Delete(key)
		delete(c.expires, key)
		// rollback
		if c.OnEvicted != nil {
			c.OnEvicted(key, v.value)
		}
		return nil, false
	}
	// get value
	if v, ok2 := c.cacheMap.Get(key); ok2 {
		c.ll.MoveToBack(v)
		return v.Value.(*entry).value, true
	}

	return nil, false
}

// add a key-value
func (c *lruCache) Add(key string, value Value) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.baseAdd(key, value)
	delete(c.expires, key)
	c.freeMemoryIfNeeded()
}

// add a key-value whth expiration
func (c *lruCache) AddWithExpiration(key string, value Value, expirationTime time.Time) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.baseAdd(key, value)
	c.expires[key] = expirationTime
	c.freeMemoryIfNeeded()
}

// delete a key-value by key
func (c *lruCache) Delete(key string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.nbytes -= int64(len(key) + c.getValueSizeByKey(key))
	c.cacheMap.Delete(key)
	delete(c.expires, key)
	return true
}

// 根据一个hash的范围来删除数据
func (c *lruCache) DeleteByHashRange(lhash string, rhash string) int {
	return c.cacheMap.DeleteByHashRange(lhash, rhash)
}

func (c *lruCache) baseAdd(key string, value Value) {
	// Check whether the key already exists
	if _, ok := c.cacheMap.Get(key); ok {
		c.nbytes += int64(value.Len() - c.getValueSizeByKey(key))
		// update value
		v, _ := c.cacheMap.Get(key)
		v.Value = &entry{key, value}
		// popular
		c.ll.MoveToBack(v)
	} else {
		c.nbytes += int64(len(key) + value.Len())
		c.cacheMap.Put(key, c.ll.PushBack(&entry{key, value}))
	}
}

// lockless !!! free Memory when the memory is insufficient
func (c *lruCache) freeMemoryIfNeeded() {
	// 只有一种淘汰策略，lru
	for c.nbytes > c.maxBytes {
		v := c.ll.Front()
		if v != nil {
			c.ll.Remove(v)
			kv := v.Value.(*entry)
			c.cacheMap.Delete(kv.key)
			delete(c.expires, kv.key)
			c.nbytes -= int64(len(kv.key) + kv.value.Len())
			if c.OnEvicted != nil {
				c.OnEvicted(kv.key, kv.value)
			}
		}
	}
}

// Scan and remove expired kv
func (c *lruCache) periodicMemoryClean() {
	c.lock.Lock()
	defer c.lock.Unlock()
	n := len(c.expires) / 10
	for key := range c.expires {
		// check for expiration
		if c.expires[key].Before(time.Now()) {
			c.nbytes -= int64(len(key) + c.getValueSizeByKey(key))
			delete(c.expires, key)
			c.cacheMap.Delete(key)
		}
		n--
		if n == 0 {
			break
		}
	}
}

func (c *lruCache) getValueSizeByKey(key string) int {
	v, _ := c.cacheMap.Get(key)
	return v.Value.(*entry).value.Len()
}
