package hsort_map

import "container/list"

// HSortMap 根据hash值排序的map
// 最重要的是提供一个根据hash值进行删除的功能
type HSortMap interface {
	Get(key string) (*list.Element, bool)
	Exist(key string) bool
	Put(key string, value *list.Element)
	Delete(key string) bool
	DeleteByHashRange(lhash string, rhash string) int
}
