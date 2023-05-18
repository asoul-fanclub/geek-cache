package cache

import "time"

/**
 * 需要提供的能力：
 * 事实上就是在hsort_map的基础上给缓存加上跳淘汰策略，超时时间等
 * 1. 可以设定缓存大小
 * 2. 并发安全
 * 3. 可以删除一个hash区间的数据
 * 4. 指定key的超时时间
 */
type Cache interface {
	Get(key string) (Value, bool)
	Add(key string, value Value)
	AddWithExpiration(key string, value Value, expirationTime time.Time)
	Delete(key string) bool
	DeleteByHashRange(lhash string, rhash string) int
}

type Value interface {
	Len() int // return data size
}
