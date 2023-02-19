package cache

import (
	"github.com/Makonike/geek-cache/geek"
	"github.com/Makonike/geek-cache/geek/utils"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
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
