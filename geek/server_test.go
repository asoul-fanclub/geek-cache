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
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := server_test_db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not found", key)
		}))
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 50000 + r.Intn(100)
	addr := fmt.Sprintf("localhost:%d", port)

	server, err := NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("geek-cache is running at", addr)

	server.Set(addr)
	g.RegisterPeers(server)

	go func() {
		err := server.Start()
		if err != nil {
			log.Fatal(err)
		}
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

	DestroyGroup(g.name)
}
