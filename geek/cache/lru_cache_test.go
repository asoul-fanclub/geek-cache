package cache

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
	"time"
)

// 检测并发情况下是否会出现问题
func TestCache_GetAndAdd(t *testing.T) {
	var wg sync.WaitGroup
	cache := NewLRUCache(LRUCacheSize(1000000000))
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000000; i++ {
			cache.Add(strconv.Itoa(i), &testValue{"东神牛逼"})
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000000; i++ {
			cache.Add(strconv.Itoa(i+1000000), &testValue{"欧神牛逼"})
		}
	}()
	wg.Wait()
	a := assert.New(t)
	for i := 1000000; i < 2000000; i++ {
		t, _ := cache.Get(strconv.Itoa(i))
		a.Equal("欧神牛逼", t.(*testValue).b)
		// todo: 效率有点低
	}
}

// 检测lru算法
func TestCache_FreeMemory(t *testing.T) {
	a := assert.New(t)
	// 测试LRU
	cache := NewLRUCache(LRUCacheSize(90))
	for i := 0; i < 10; i++ {
		cache.Add(strconv.Itoa(i), &testValue{"123456789"})
	}
	// key为0的被淘汰
	_, f0 := cache.Get("0")
	a.False(f0)
	// key为1的未被淘汰
	v1, f1 := cache.Get("1")
	a.True(f1)
	a.Equal(v1.(*testValue).b, "123456789")
	// 添加一个缓存
	cache.Add("a", &testValue{"123456789"})
	// key为2的缓存被淘汰
	_, f2 := cache.Get("2")
	a.False(f2)
	// key为3未被淘汰
	v3, f3 := cache.Get("3")
	a.True(f3)
	a.Equal(v3.(*testValue).b, "123456789")
}

// 检测AddWithExpiration的算法中lru逻辑
func TestCache_FreeMemory2(t *testing.T) {
	timeout := time.Now().Add(3 * time.Second)
	a := assert.New(t)
	// 测试LRU
	cache := NewLRUCache(LRUCacheSize(90))
	for i := 0; i < 10; i++ {
		cache.AddWithExpiration(strconv.Itoa(i), &testValue{"123456789"}, timeout)
	}
	// key为0的被淘汰
	_, f0 := cache.Get("0")
	a.False(f0)
	// key为1的未被淘汰
	v1, f1 := cache.Get("1")
	a.True(f1)
	a.Equal(v1.(*testValue).b, "123456789")
	// 添加一个缓存
	cache.AddWithExpiration("a", &testValue{"123456789"}, timeout)
	// key为2的缓存被淘汰
	_, f2 := cache.Get("2")
	a.False(f2)
	// key为3未被淘汰
	v3, f3 := cache.Get("3")
	a.True(f3)
	a.Equal(v3.(*testValue).b, "123456789")
}

// 测试超时
func TestCache_AddWithExpiration(t *testing.T) {
	a := assert.New(t)
	cache := NewLRUCache(LRUCacheSize(100))
	cache.AddWithExpiration("1", &testValue{"123456789"}, time.Now().Add(3*time.Second))
	time.Sleep(2 * time.Second)
	v1, _ := cache.Get("1")
	a.Equal("123456789", v1.(*testValue).b)
	time.Sleep(2 * time.Second)
	_, f := cache.Get("1")
	a.False(f)
}

// 测试删除
func TestCache_Delete(t *testing.T) {
	a := assert.New(t)
	cache := NewLRUCache(LRUCacheSize(100))
	cache.Add("1", &testValue{"123456789"})
	cache.Add("2", &testValue{"123456789"})
	_, f1 := cache.Get("1")
	_, f2 := cache.Get("2")
	a.True(f1)
	a.True(f2)
	cache.Delete("1")
	_, f3 := cache.Get("1")
	_, f4 := cache.Get("2")
	a.False(f3)
	a.True(f4)
}

// ByteView 只读的字节视图，用于缓存数据
type testValue struct {
	b string
}

func (b *testValue) Len() int {
	return len(b.b)
}
