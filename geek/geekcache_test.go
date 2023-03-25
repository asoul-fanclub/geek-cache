package geek

import (
	"math/rand"
	"testing"
	time "time"

	"github.com/stretchr/testify/assert"
)

var db = map[string]string{
	"Tom":   "630",
	"Jack":  "742",
	"Amy":   "601",
	"Alice": "653",
}

func TestGroup_Get(t *testing.T) {
	loads := make(map[string]int)
	gee := NewGroup("scores", 2<<10, false, GetterFunc(
		func(key string) ([]byte, bool, time.Time) {
			if v, ok := db[key]; ok {
				loads[key] += 1
				return []byte(v), true, time.Time{}
			}
			return nil, false, time.Time{}
		}),
	)
	gee1 := GetGroup("scores")
	_, err := gee1.Get("")
	if err == nil {
		t.Fatalf("Get params is empty, excepted not nil error, but nil")
	}

	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatalf("expected err is nil and value is %v, but err is %v, value is %v", v, err, view.String())
		}
		// load from callback
		if _, err := gee.Get(k); err != nil || loads[k] > 1 {
			t.Fatalf("cache %v miss", k)
		}
	}
	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the key unknown but get %v", view)
	}
}

func TestGroup_HGet(t *testing.T) {
	loads := make(map[string]int)
	gee := NewGroup("scores", 2<<10, true, GetterFunc(
		func(key string) ([]byte, bool, time.Time) {
			if v, ok := db[key]; ok {
				loads[key] += 1
				return []byte(v), true, time.Time{}
			}
			return nil, false, time.Time{}
		}),
	)
	gee1 := GetHGroup("scores")
	_, err := gee1.HGet("")
	if err == nil {
		t.Fatalf("Get params is empty, excepted not nil error, but nil")
	}

	for k, v := range db {
		if view, err := gee.HGet(k); err != nil || view.String() != v {
			t.Fatalf("expected err is nil and value is %v, but err is %v, value is %v", v, err, view.String())
		}
		// load from callback
		if _, err := gee.HGet(k); err != nil || loads[k] > 1 {
			t.Fatalf("cache %v miss", k)
		}
	}
	if view, err := gee.HGet("unknown"); err == nil {
		t.Fatalf("the key unknown but get %v", view)
	}
}

func TestGroup_Delete(t *testing.T) {
	a := assert.New(t)
	database := map[string]string{
		"Tom":   "630",
		"Jack":  "742",
		"Amy":   "601",
		"Alice": "653",
	}
	loads := make(map[string]int)
	NewGroup("scores", 2<<10, false, GetterFunc(
		func(key string) ([]byte, bool, time.Time) {
			if v, ok := database[key]; ok {
				loads[key] += 1
				return []byte(v), true, time.Time{}
			}
			return nil, false, time.Time{}
		}),
	)
	gee := GetGroup("scores")
	for k, v := range database {
		view, err := gee.Get(k)
		a.Equal(view.String(), v)
		a.Nil(err)
		// delete
		s, err2 := gee.Delete(k)
		a.True(s)
		a.Nil(err2)
		// Get it again
		_, _ = gee.Get(k)
		// check the number of loads
		a.Equal(2, loads[k])
	}
}

func TestGroup_SetTimeout(t *testing.T) {
	a := assert.New(t)
	var db2 = map[string]string{
		"Tom":   "630",
		"Jack":  "742",
		"Amy":   "601",
		"Alice": "653",
	}
	loads := make(map[string]int)
	_ = NewGroup("scores", 2<<10, false, GetterFunc(
		func(key string) ([]byte, bool, time.Time) {
			rand.Seed(time.Now().UnixNano())
			if v, ok := db2[key]; ok {
				loads[key] += 1
				// 用户设置超时时间
				timeout := time.Now().Add(time.Duration(rand.Intn(10)) * time.Second)
				return []byte(v), true, timeout
			}
			return nil, false, time.Time{}
		}),
	)
	// 读取key并校验
	gee1 := GetGroup("scores")
	v, _ := gee1.Get("Alice")
	a.Equal(v.String(), "653")

	// 过期
	db2["Alice"] = "123"
	time.Sleep(10 * time.Second)
	v2, _ := gee1.Get("Alice")
	a.Equal(v2.String(), "123")
}
