package cache

import (
	"container/list"
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (Value, bool)
	Add(key string, value Value)
	AddWithExpiration(key string, value Value, expirationTime time.Time)
	Delete(key string) bool
}

type Value interface {
	Len() int // return data size
}

// cache struct
type lruCache struct {
	lock      sync.Mutex
	cacheMap  map[string]*list.Element      // map cache
	expires   map[string]time.Time          // The expiration time of key
	ll        *list.List                    // linked list
	OnEvicted func(key string, value Value) // The callback function when a record is deleted
	maxBytes  int64                         // The maximum memory allowed
	nbytes    int64                         // The memory is currently in use
}

// 通过key可以在记录删除时，删除字典缓存中的映射
type entry struct {
	key   string
	value Value
}

func NewLRUCache(maxSize int64) *lruCache {
	answer := lruCache{
		cacheMap: make(map[string]*list.Element),
		expires:  make(map[string]time.Time),
		nbytes:   0,
		ll:       list.New(),
		maxBytes: maxSize,
	}
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
		v := c.cacheMap[key].Value.(*entry)
		c.nbytes -= int64(v.value.Len() + len(v.key))
		c.ll.Remove(c.cacheMap[key])
		delete(c.cacheMap, key)
		delete(c.expires, key)
		// rollback
		if c.OnEvicted != nil {
			c.OnEvicted(key, v.value)
		}
		return nil, false
	}
	// get value
	if v, ok2 := c.cacheMap[key]; ok2 {
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
	delete(c.cacheMap, key)
	delete(c.expires, key)
	return true
}

func (c *lruCache) baseAdd(key string, value Value) {
	// Check whether the key already exists
	if _, ok := c.cacheMap[key]; ok {
		c.nbytes += int64(value.Len() - c.getValueSizeByKey(key))
		// update value
		c.cacheMap[key].Value = &entry{key, value}
		// popular
		c.ll.MoveToBack(c.cacheMap[key])
	} else {
		c.nbytes += int64(len(key) + value.Len())
		c.cacheMap[key] = c.ll.PushBack(&entry{key, value})
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
			delete(c.cacheMap, kv.key)
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
			delete(c.cacheMap, key)
		}
		n--
		if n == 0 {
			break
		}
	}
}

func (c *lruCache) getValueSizeByKey(key string) int {
	return c.cacheMap[key].Value.(*entry).value.Len()
}
