package geek

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

var server_test_db = map[string]string{
	"Tom":  "630",
	"Tom2": "631",
	"Tom3": "632",
}

func TestServer(t *testing.T) {
	a := assert.New(t)
	g := NewGroup("scores", 2<<10, false, GetterFunc(
		func(key string) ([]byte, bool, time.Time) {
			log.Println("[SlowDB] search key", key)
			if v, ok := server_test_db[key]; ok {
				return []byte(v), true, time.Time{}
			}
			return nil, false, time.Time{}
		}))
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 50000 + r.Intn(100)
	addr := fmt.Sprintf("localhost:%d", port)

	// 添加peerPicker
	picker := NewClientPicker(addr)

	g.RegisterPeers(picker)

	view, err := g.Get("Tom")
	if err != nil {
		t.Fatal(err)
	}
	s, err := g.Delete("Tom")
	a.True(s)
	a.Nil(err)
	if !reflect.DeepEqual(view.String(), "630") {
		t.Errorf("Tom %s(actual)/%s(ok)", view.String(), "630")
	}
	view, err = g.Get("Unknown")
	if err == nil || view.String() != "" {
		t.Errorf("Unknown not exists, but got %s", view.String())
	}
}
