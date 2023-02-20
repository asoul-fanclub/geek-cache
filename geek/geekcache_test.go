package geek

import (
	"testing"
	"time"
)

var db = map[string]string{
	"Tom":   "630",
	"Jack":  "742",
	"Amy":   "601",
	"Alice": "653",
}

func TestGroup_Get(t *testing.T) {
	loads := make(map[string]int)
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, bool, *time.Time) {
			if v, ok := db[key]; ok {
				loads[key] += 1
				return []byte(v), true, nil
			}
			return nil, false, nil
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
