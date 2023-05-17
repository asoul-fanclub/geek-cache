package hsort_map

import (
	"container/list"
	"strings"
	"sync"
)

// NewGeekMap 底层通过sync.map来实现
// 推荐理由：1. 并发安全 2. map实现读写都是o(1) 3. 实现简单
// 不推荐理由：1. 删除一个区间的时候会遍历一遍map
type GeekMap struct {
	m sync.Map
}

func NewGeekMap(hash Hash) *GeekMap {
	return &GeekMap{
		m: sync.Map{},
	}
}

func (t *GeekMap) Get(key string) (*list.Element, bool) {
	v, b := t.m.Load(key)
	return v.(*list.Element), b
}

func (t *GeekMap) Exist(key string) bool {
	_, v := t.m.Load(key)
	return v
}

func (t *GeekMap) Put(key string, value *list.Element) {
	t.m.Store(key, value)
}

func (t *GeekMap) Delete(key string) bool {
	_, b := t.m.LoadAndDelete(key)
	return b
}

func (t *GeekMap) DeleteByHashRange(lhash string, rhash string) int {
	answer := 0
	t.m.Range(func(key any, value any) bool {
		if strings.Compare(key.(string), lhash) >= 0 && strings.Compare(key.(string), rhash) < 0 {
			t.m.Delete(key)
			answer++
		}
		return true
	})
	return answer
}
