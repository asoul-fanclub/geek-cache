package hsort_map

// hsort_map 根据hash值排序的map
// 最重要的是提供一个根据hash值进行删除的功能
type HSortMap interface {
	Get(key []byte) interface{}
	Put(key []byte, value interface{})
	Delete(key []byte) interface{}
	DeleteByHashRange(lhash []byte, rhash []byte)
}
