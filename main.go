package main

import (
	"flag"
	"fmt"
	"geek-cache/geek"
	"log"
	"net/http"
)

var test_http_peers_db = map[string]string{
	"Tom":  "630",
	"Tom1": "631",
	"Tom2": "632",
}

func createGroup() *geek.Group {
	return geek.NewGroup("scores", 2<<10, geek.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := test_http_peers_db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%v not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, gee *geek.Group) {
	peers, _ := geek.NewServer(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("geek-cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *geek.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := gee.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-TYpe", "application/octet-stream")
		w.Write(view.ByteSLice())
	}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geek-Cache Server Port")
	flag.BoolVar(&api, "api", false, "Start an api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], []string(addrs), gee)
}
