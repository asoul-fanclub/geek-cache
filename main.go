package main

import (
	"fmt"
	"geek-cache/geek"
	"log"
	"sync"
)

func main() {

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
	var addr string = "localhost:9999"
	server, err := geek.NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}
	server.Set(addr)
	g.RegisterPeers(server)
	log.Println("geek-cache is running at", addr)

	go func() {
		err = server.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(4)
	go GetTomScore(g, &wg)
	go GetTomScore(g, &wg)
	go GetTomScore(g, &wg)
	go GetTomScore(g, &wg)

}

func GetTomScore(g *geek.Group, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("get Tom...")
	view, err := g.Get("Tom")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(view.String())
}
