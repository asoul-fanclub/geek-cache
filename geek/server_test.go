package geek

import (
	"fmt"
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
	g := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, bool, *time.Time) {
			log.Println("[SlowDB] search key", key)
			if v, ok := server_test_db[key]; ok {
				return []byte(v), true, nil
			}
			return nil, true, nil
		}))
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 50000 + r.Intn(100)
	addr := fmt.Sprintf("localhost:%d", port)

	// 启动服务
	server, err := NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("geek-cache is running at", addr)

	go func() {
		err := server.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()
	// 添加peerPicker
	picker := NewClientPicker()
	picker.Set(addr)
	g.RegisterPeers(picker)

	defer func() {
		server.Stop()
		DestroyGroup(g.name)
	}()

	view, err := g.Get("Tom")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(view.String(), "630") {
		t.Errorf("Tom %s(actual)/%s(ok)", view.String(), "630")
	}
	view, err = g.Get("Unknown")
	if err == nil || view.String() != "" {
		t.Errorf("Unknown not exists, but got %s", view.String())
	}
}
