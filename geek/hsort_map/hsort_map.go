package hsort_map

// hsort_map 根据hash值排序的map
// 最重要的是提供一个根据hash值进行删除的功能
type HSortMap interface {
	Get(key string) []byte
	Exist(key string) bool
	Put(key string, value []byte)
	Delete(key string) []byte
	DeleteByHashRange(lhash string, rhash string)
}
