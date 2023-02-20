package geek

import (
	"time"

	c "github.com/Makonike/geek-cache/geek/cache"
)

// cache 实例化lru，封装get和add。
type cache struct {
	lruCache   c.Cache
	cacheBytes int64
}

func (cache *cache) add(key string, value ByteView) {
	// lazy load
	cache.lruCacheLazyLoadIfNeed()
	cache.lruCache.Add(key, value)
}

func (cache *cache) lruCacheLazyLoadIfNeed() {
	if cache.lruCache == nil {
		lock.Lock()
		defer lock.Unlock()
		if cache.lruCache == nil {
			cache.lruCache = c.NewLRUCache(cache.cacheBytes)
		}
	}
}

func (cache *cache) get(key string) (value ByteView, ok bool) {
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
