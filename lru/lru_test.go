package lru

import (
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestCache_Get(t *testing.T) {
	key := "学rust有种每天重置记忆的感觉，这种把握不住记忆的感觉真是太好了。"
	key2 := "woca"
	a := String(key)
	lru := New(int64(a.Len()*2), nil)
	lru.Add(key, a)
	v, hit := lru.Get(key)
	if !hit {
		t.Fatalf("cache miss for key %v, expected hit", key)
	}
	if v != a {
		t.Errorf("hit value expected %v, but %v got", a, v)
	}
	lru.Add(key2, a)
	v, hit = lru.Get(key2)
	if !hit {
		t.Fatalf("cache miss for key %v, expected hit", key2)
	}
	if v != a {
		t.Errorf("hit value expected %v, but %v got", a, v)
	}
	v, hit = lru.Get(key)
	if hit {
		t.Fatalf("cache miss for key %v, expected hit", key2)
	}
	if v == a {
		t.Errorf("hit value expected %v, but %v got", a, v)
	}
}

func TestCache_Add(t *testing.T) {
	lru := New(int64(5000), nil)
	lru.Add("key", String("1"))
	lru.Add("key", String("122"))
	if lru.nbytes != int64(len("key")+len("111")) {
		t.Fatalf("excepted %v but got %v", int64(len("key")+len("111")), lru.nbytes)
	}
}
