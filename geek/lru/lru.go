package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         // 允许使用的最大内存
	nbytes    int64                         // 当前已使用内存
	ll        *list.List                    // 双向链表
	cache     map[string]*list.Element      // 字典缓存
	OnEvicted func(key string, value Value) // 某条记录被删除时的回调函数
}

// entry 双向链表的节点类型
// 通过key可以在记录删除时，删除字典缓存中的映射
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int // 返回缓存数据占用的内存大小
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nbytes:    0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 获取缓存数据
// 如果获取成功，将节点从移到队尾。并返回找到的值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if v, ok := c.cache[key]; ok {
		c.ll.MoveToBack(v)
		kv := v.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 缓存淘汰
// 从双向链表中取出队头删除，并删除cache字典缓存中的映射
// 如果有回调函数，则调用回调函数
func (c *Cache) RemoveOldest() {
	v := c.ll.Front()
	if v != nil {
		c.ll.Remove(v)
		kv := v.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key) + kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// todo: 这里有bug，如果key已经存在，则无法更新
// Add 添加或删除缓存
// 此处是先加后删。如果加入一个较大值会OOM
func (c *Cache) Add(key string, value Value) {
	if v, ok := c.cache[key]; ok {
		c.ll.MoveToBack(v)
		kv := v.Value.(*entry)
		// 加上差值
		c.nbytes += int64(value.Len() - kv.value.Len())
	} else {
		v := c.ll.PushBack(&entry{key, value})
		c.cache[key] = v
		c.nbytes += int64(len(key) + value.Len())
	}
	// 超过缓存限制就需要删除旧节点
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len 返回缓存节点数
func (c *Cache) Len() int {
	return c.ll.Len()
}
