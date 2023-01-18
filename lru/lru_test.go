package lru

import (
	"reflect"
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

func TestCache_RemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "v3"
	lru := New(int64(len(k1+k2+v1+v2)), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))
	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("removeOLdest Failed, expected remove oldest key %v, but not removed", k1)
	}
}

func TestCache_OnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(5000), callback)
	lru.Add("key1", String("wwwwww"))
	lru.Add("k2", String("wwwwww"))
	lru.Add("k--3", String("wwwwww"))
	lru.RemoveOldest()
	lru.RemoveOldest()
	lru.RemoveOldest()
	lru.RemoveOldest()
	lru.RemoveOldest()
	expected := []string{"key1", "k2", "k--3"}
	if !reflect.DeepEqual(expected, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expected)
	}
}
