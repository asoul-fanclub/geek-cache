package geek

import (
	"sync"
	"time"

	c "github.com/Makonike/geek-cache/geek/cache"
)

// cache 实例化lru，封装get和add。
type cache struct {
	lock       sync.RWMutex
	lruCache   c.Cache
	cacheBytes int64
}

func (cache *cache) lruCacheLazyLoadIfNeed() {
	if cache.lruCache == nil {
		cache.lock.Lock()
		defer cache.lock.Unlock()
		if cache.lruCache == nil {
			cache.lruCache = c.NewLRUCache(c.LRUCacheSize(cache.cacheBytes))
		}
	}
}

func (cache *cache) add(key string, value ByteView) {
	// lazy load
	cache.lruCacheLazyLoadIfNeed()
	cache.lruCache.Add(key, value)
}

func (cache *cache) get(key string) (value ByteView, ok bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	if cache.lruCache == nil {
		return
	}
	if v, find := cache.lruCache.Get(key); find {
		return v.(ByteView), true
	}
	return
}

func (cache *cache) addWithExpiration(key string, value ByteView, expirationTime time.Time) {
	// lazy load
	cache.lruCacheLazyLoadIfNeed()
	cache.lruCache.AddWithExpiration(key, value, expirationTime)
}

func (cache *cache) delete(key string) bool {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	if cache.lruCache == nil {
		return true
	}
	return cache.lruCache.Delete(key)
}
