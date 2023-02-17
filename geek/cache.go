package geek

import (
	"sync"

	"github.com/Makonike/geek-cache/geek/lru"
)

// cache 解决并发控制，实例化lru，封装get和add。
type cache struct {
	lock       sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// 懒加载
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.lru == nil {
		return
	}
	if v, find := c.lru.Get(key); find {
		return v.(ByteView), true
	}
	return
}
