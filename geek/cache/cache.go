package cache

import (
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
	cache     map[string]*entry             // map cache
	expires   map[string]time.Time          // The expiration time of key
	OnEvicted func(key string, value Value) // The callback function when a record is deleted
	maxBytes  int64                         // The maximum memory allowed
	nbytes    int64                         // The memory is currently in use
}

// 通过key可以在记录删除时，删除字典缓存中的映射
type entry struct {
	key   string
	value Value
}

func (c *cache) Get(key string) (Value, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// check for expiration
	expirationTime, ok := c.expires[key]
	if ok && expirationTime.Before(time.Now()) {
		c.nbytes -= int64(c.cache[key].value.Len())
		delete(c.cache, key)
		delete(c.expires, key)
		return nil, fmt.Errorf("cache miss error")
	}
	// get value
	if v, ok := c.cache[key]; ok {
		return v.value, nil
	}
	return nil, fmt.Errorf("cache miss error")
}

func (c *cache) Add(key string, value Value) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// Check whether the key already exists
	if _, ok := c.cache[key]; !ok {
		c.nbytes += int64(value.Len() - c.cache[key].value.Len())
		c.cache[key] = &entry{key, value}
		delete(c.expires, key)
	} else {
		c.nbytes += int64(len(key) + value.Len())
		c.cache[key] = &entry{key, value}
	}
	c.freeMemoryIfNeeded()
	c.lock.Lock()
}

func (c *cache) AddWithExpiration(key string, value Value, expirationTime time.Time) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// Check whether the key already exists
	if _, ok := c.cache[key]; ok {
		c.nbytes += int64(value.Len() - c.cache[key].value.Len())
	} else {
		c.nbytes += int64(len(key) + value.Len())
	}
	c.cache[key] = &entry{key, value}
	c.expires[key] = expirationTime

	c.freeMemoryIfNeeded()
	c.lock.Lock()
}

// free memory
func (c *cache) freeMemoryIfNeeded() {

}

type Value interface {
	Len() int // return data size
}
