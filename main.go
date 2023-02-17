package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Makonike/geek-cache/geek"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.Parse()
	var mysql = map[string]string{
		"Tom":  "630",
		"Tom1": "631",
		"Tom2": "632",
	}
	g := geek.NewGroup("scores", 2<<10, geek.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not found", key)
		}))

	addrMap := map[int]string{
		8001: "8001",
		8002: "8002",
		8003: "8003",
	}
	var addr string = "127.0.0.1:" + addrMap[port]

	server, err := geek.NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}

	addrs := make([]string, 0)
	for _, addr := range addrMap {
		addrs = append(addrs, "127.0.0.1:"+addr)
	}

	// set client address
	// TODO: will be substituted with etcd service discovery
	server.Set(addrs...)
	g.RegisterPeers(server)
	for {
		err = server.Start()
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}
