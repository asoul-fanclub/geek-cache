package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (Value, error)
	AddWithExpiration(key string, value Value, expirationTime time.Time)
	Add(key string, value Value)
}

// cache struct
type cache struct {
	lock      sync.Mutex
	cache     map[string]*list.Element      // map cache
	expires   map[string]time.Time          // The expiration time of key
	ll        *list.List                    // 双向链表
	OnEvicted func(key string, value Value) // The callback function when a record is deleted
	maxBytes  int64                         // The maximum memory allowed
	nbytes    int64                         // The memory is currently in use
}

// 通过key可以在记录删除时，删除字典缓存中的映射
type entry struct {
	key   string
	value Value
}

func NewCache(maxSize int64) Cache {
	answer := cache{
		cache:    make(map[string]*list.Element),
		expires:  make(map[string]time.Time),
		nbytes:   0,
		ll:       list.New(),
		maxBytes: maxSize,
	}
	go func() {
		ticker := time.Tick(1 * time.Hour)
		for {
			select {
			case <-ticker:
				answer.periodicMemoryClean()
			}
		}
	}()
	return &answer
}

func (c *cache) Get(key string) (Value, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// check for expiration
	expirationTime, ok := c.expires[key]
	if ok && expirationTime.Before(time.Now()) {
		c.nbytes -= int64(c.cache[key].Value.(*entry).value.Len())
		value := c.cache[key].Value.(*entry).value
		delete(c.cache, key)
		delete(c.expires, key)
		// rollback
		if c.OnEvicted != nil {
			c.OnEvicted(key, value)
		}
		return nil, fmt.Errorf("cache miss error")
	}
	// get value
	if v, ok := c.cache[key]; ok {
		return v.Value.(*entry).value, nil
	}
	return nil, fmt.Errorf("cache miss error")
}

func (c *cache) Add(key string, value Value) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// Check whether the key already exists
	if _, ok := c.cache[key]; ok {
		c.nbytes += int64(value.Len() - c.cache[key].Value.(*entry).value.Len())
		c.cache[key].Value = &entry{key, value}
		delete(c.expires, key)
	} else {
		c.nbytes += int64(len(key) + value.Len())
		c.cache[key] = c.ll.PushBack(&entry{key, value})
	}
	c.freeMemoryIfNeeded()
}

func (c *cache) AddWithExpiration(key string, value Value, expirationTime time.Time) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// Check whether the key already exists
	if _, ok := c.cache[key]; ok {
		c.nbytes += int64(value.Len() - c.cache[key].Value.(*entry).value.Len())
	} else {
		c.nbytes += int64(len(key) + value.Len())
	}
	c.cache[key] = c.ll.PushBack(&entry{key, value})
	c.expires[key] = expirationTime

	c.freeMemoryIfNeeded()
}

// lockless !!!
func (c *cache) freeMemoryIfNeeded() {
	// 只有一种淘汰策略，lru
	for c.nbytes > c.maxBytes {
		v := c.ll.Front()
		if v != nil {
			c.ll.Remove(v)
			kv := v.Value.(*entry)
			delete(c.cache, kv.key)
			delete(c.expires, kv.key)
			c.nbytes -= int64(len(kv.key) + kv.value.Len())
			if c.OnEvicted != nil {
				c.OnEvicted(kv.key, kv.value)
			}
		}
	}
}

func (c *cache) periodicMemoryClean() {
	c.lock.Lock()
	defer c.lock.Unlock()
	n := len(c.expires) / 10
	for key := range c.expires {
		// check for expiration
		if c.expires[key].Before(time.Now()) {
			c.nbytes -= int64(len(key) + c.cache[key].Value.(*entry).value.Len())
			delete(c.expires, key)
			delete(c.cache, key)
		}
		n--
		if n == 0 {
			break
		}
	}
}

type Value interface {
	Len() int // return data size
}
