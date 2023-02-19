package cache

import (
	"github.com/Makonike/geek-cache/geek"
	"testing"
)

func TestCache_Get(t *testing.T) {
	cache := NewCache(1000, ALLKEYS_LRU)
	cache.Add("a", &geek.ByteView{make([]byte, 1000)})
	cache.Add("b", &geek.ByteView{make([]byte, 1000)})

}
