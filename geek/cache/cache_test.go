package cache

import (
	"github.com/Makonike/geek-cache/geek"
	"github.com/Makonike/geek-cache/geek/utils"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCache_GetAndAdd(t *testing.T) {
	var wg sync.WaitGroup
	cache := NewCache(1000000000)
	wg.Add(2)
	go func() {
		for i := 0; i < 1000000; i++ {
			cache.Add(strconv.Itoa(i), &geek.ByteView{utils.VarStrToRaw("东神牛逼")})
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 1000000; i++ {
			cache.Add(strconv.Itoa(i+1000000), &geek.ByteView{utils.VarStrToRaw("欧神牛逼")})
		}
		wg.Done()
	}()
	wg.Wait()
	a := assert.New(t)
	for i := 1000000; i < 2000000; i++ {
		t, _ := cache.Get(strconv.Itoa(i))
		a.Equal("欧神牛逼", string(t.(*geek.ByteView).B))
		// todo: 效率有点低
	}
}

func TestCache_FreeMemory(t *testing.T) {
	a := assert.New(t)
	// 测试LRU
	cache := NewCache(90)
	for i := 0; i < 10; i++ {
		cache.Add(strconv.Itoa(i), &geek.ByteView{utils.VarStrToRaw("123456789")})
	}
	// 第一个被淘汰
	_, err := cache.Get("0")
	a.True(err != nil)
	// 读取第二个，并再添加一个缓存
	m, _ := cache.Get("1")
	a.Equal("123456789", string(m.(*geek.ByteView).B))
	cache.Add("a", &geek.ByteView{utils.VarStrToRaw("123456789")})
	// 第key为2的缓存被淘汰
	_, err2 := cache.Get("2")
	a.True(err2 != nil)
}

func TestCache_AddWithExpiration(t *testing.T) {
	a := assert.New(t)
	cache := NewCache(100)
	cache.AddWithExpiration("1", &geek.ByteView{utils.VarStrToRaw("123456789")}, time.Now().Add(3*time.Second))
	time.Sleep(2 * time.Second)
	b, _ := cache.Get("1")
	a.Equal("123456789", string(b.(*geek.ByteView).B))
	time.Sleep(2 * time.Second)
	_, err := cache.Get("1")
	a.True(err != nil)
}
