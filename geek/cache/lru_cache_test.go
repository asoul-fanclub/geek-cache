package cache

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCache_GetAndAdd(t *testing.T) {
	var wg sync.WaitGroup
	cache := NewLRUCache(1000000000)
	wg.Add(2)
	go func() {
		for i := 0; i < 1000000; i++ {
			cache.Add(strconv.Itoa(i), &testValue{"东神牛逼"})
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 1000000; i++ {
			cache.Add(strconv.Itoa(i+1000000), &testValue{"欧神牛逼"})
		}
		wg.Done()
	}()
	wg.Wait()
	a := assert.New(t)
	for i := 1000000; i < 2000000; i++ {
		t, _ := cache.Get(strconv.Itoa(i))
		a.Equal("欧神牛逼", t.(*testValue).b)
		// todo: 效率有点低
	}
}

func TestCache_FreeMemory(t *testing.T) {
	a := assert.New(t)
	// 测试LRU
	cache := NewLRUCache(90)
	for i := 0; i < 10; i++ {
		cache.Add(strconv.Itoa(i), &testValue{"123456789"})
	}
	// 第一个被淘汰
	_, f := cache.Get("0")
	a.False(f)
	// 读取第二个，并再添加一个缓存
	v1, _ := cache.Get("1")
	a.Equal("123456789", v1.(*testValue).b)
	cache.Add("a", &testValue{"123456789"})
	// 第key为2的缓存被淘汰
	_, f2 := cache.Get("2")
	a.False(f2)
}

func TestCache_AddWithExpiration(t *testing.T) {
	a := assert.New(t)
	cache := NewLRUCache(100)
	cache.AddWithExpiration("1", &testValue{"123456789"}, time.Now().Add(3*time.Second))
	time.Sleep(2 * time.Second)
	v1, _ := cache.Get("1")
	a.Equal("123456789", v1.(*testValue).b)
	time.Sleep(2 * time.Second)
	_, f := cache.Get("1")
	a.False(f)
}

// ByteView 只读的字节视图，用于缓存数据
type testValue struct {
	b string
}

func (b *testValue) Len() int {
	return len(b.b)
}
