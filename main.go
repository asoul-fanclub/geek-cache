package main

import (
	"flag"
	"hash/crc32"
	"log"
	"strconv"
	"time"

	"github.com/Makonike/geek-cache/geek"
	"github.com/Makonike/geek-cache/geek/consistenthash"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.Parse()
	// mock database or other dataSource
	var mysql = map[string]string{
		"Tom":  "630",
		"Tom1": "631",
		"Tom2": "632",
	}
	// NewGroup create a Group which means a kind of sources
	// contain a func that used when misses cache
	g := geek.NewGroup("scores", 2<<10, geek.GetterFunc(
		func(key string) ([]byte, bool, time.Time) {
			log.Println("[SlowDB] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), true, time.Time{}
			}
			return nil, false, time.Time{}
		}))

	var addr string = "127.0.0.1:" + strconv.Itoa(port)

	server, err := geek.NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}

	picker := geek.NewClientPicker(addr)
	g.RegisterPeers(picker)

	for {
		err = server.Start()
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}
