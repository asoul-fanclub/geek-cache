# geek-cache

a distributed read-only cache, based [groupcache](https://github.com/golang/groupcache), using etcd as a registry, supports efficient concurrent reading.

## Install

The package supports 3 last Go versions and requires a Go version with modules support.

`go get github.com/Makonike/geek-cache`

## Usage

Be sure to install **etcd v3**(port 2379), grpcurl(for your tests), protobuf v3.

## Interface

- Get

Get the value with specific group and key. Node will find from the peer node which was chosen by hash function, if not found, will get locally by given method.

- HGet

Like Get, but always get locally.

- Delete

Delete the value if the value is changed.

## Config

In your application, you can configure like following:

- Server

```go
func NewServer(self string, opts ...ServerOptions) (*Server, error)

server, err := geek.NewServer(addr, geek.ServiceName("your-service-name"))
```

- Client

```go
registry.GlobalClientConfig = &clientv3.Config{
	Endpoints:   []string{"127.0.0.1:2379"}, // etcd address
	DialTimeout: 5 * time.Second, // the timeout for failing to establish a connection
}
```

- Picker and Consistent Hash

```go
picker := geek.NewClientPicker(addr, geek.PickerServiceName("geek-cache"), geek.ConsHashOptions(consistenthash.HashFunc(crc32.ChecksumIEEE), consistenthash.Replicas(150)))
```

## Test

Write the following code for testing.

main.go

```go
package main

import (
	"flag"
	"log"
	"strconv"
	"time"

	"github.com/Makonike/geek-cache/geek"
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
	g := geek.NewGroup("scores", 2<<10, false, geek.GetterFunc(
		func(key string) ([]byte, bool, time.Time) {
			log.Println("[SlowDB] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), true, time.Time{}
			}
			return nil, false, time.Time{}
		}))
	g2 := geek.NewGroup("scores", 2<<10, true, geek.GetterFunc(
		func(key string) ([]byte, bool, time.Time) {
			log.Println("[SlowDB] search hot key", key)
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
	g2.RegisterPeers(picker)

	for {
		err = server.Start()
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}
```

a.sh

```shell
#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server
./server -port=8001 &
./server -port=8002 &
./server -port=8003 &

echo ">>> start test"
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8001 pb.GroupCache/Get 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom1"}' 127.0.0.1:8001 pb.GroupCache/Get 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom2"}' 127.0.0.1:8001 pb.GroupCache/Get 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8002 pb.GroupCache/Get 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom1"}' 127.0.0.1:8002 pb.GroupCache/Get 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom2"}' 127.0.0.1:8002 pb.GroupCache/Get 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8003 pb.GroupCache/Get 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom1"}' 127.0.0.1:8003 pb.GroupCache/Get 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom2"}' 127.0.0.1:8003 pb.GroupCache/Get 

sleep 3

grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8001 pb.GroupCache/HGet 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom1"}' 127.0.0.1:8001 pb.GroupCache/HGet 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom2"}' 127.0.0.1:8001 pb.GroupCache/HGet 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8002 pb.GroupCache/HGet 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom1"}' 127.0.0.1:8002 pb.GroupCache/HGet 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom2"}' 127.0.0.1:8002 pb.GroupCache/HGet 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8003 pb.GroupCache/HGet 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom1"}' 127.0.0.1:8003 pb.GroupCache/HGet 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom2"}' 127.0.0.1:8003 pb.GroupCache/HGet 

sleep 3

kill -9 `lsof -ti:8002`;

sleep 3

grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8001 pb.GroupCache/Get 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8003 pb.GroupCache/Get 

sleep 3

grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8001 pb.GroupCache/HGet 
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8003 pb.GroupCache/HGet 
wait
```

Running the shell, then you can see the results following.

```bash
$ ./a.sh 
>>> start test
2023/03/22 22:35:29 [127.0.0.1:8003] register service success
2023/03/22 22:35:29 [127.0.0.1:8002] register service success
2023/03/22 22:35:29 [127.0.0.1:8001] register service success
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8001] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:29 [Server 127.0.0.1:8001] Pick peer 127.0.0.1:8002
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8002] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:29 [Server 127.0.0.1:8002] Pick peer 127.0.0.1:8002
2023/03/22 22:35:29 [SlowDB] search key Tom
{
  "value": "NjMw"
}
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8001] Recv RPC Request for get- (scores)/(Tom1)
2023/03/22 22:35:29 [Server 127.0.0.1:8001] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom1)
2023/03/22 22:35:29 [Server 127.0.0.1:8003] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [SlowDB] search key Tom1
{
  "value": "NjMx"
}
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8001] Recv RPC Request for get- (scores)/(Tom2)
2023/03/22 22:35:29 [Server 127.0.0.1:8001] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom2)
2023/03/22 22:35:29 [Server 127.0.0.1:8003] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [SlowDB] search key Tom2
{
  "value": "NjMy"
}
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8002] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:29 [Server 127.0.0.1:8002] Pick peer 127.0.0.1:8002
2023/03/22 22:35:29 [Geek-Cache] hit
{
  "value": "NjMw"
}
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8002] Recv RPC Request for get- (scores)/(Tom1)
2023/03/22 22:35:29 [Server 127.0.0.1:8002] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom1)
2023/03/22 22:35:29 [Server 127.0.0.1:8003] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [Geek-Cache] hit
{
  "value": "NjMx"
}
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8002] Recv RPC Request for get- (scores)/(Tom2)
2023/03/22 22:35:29 [Server 127.0.0.1:8002] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom2)
2023/03/22 22:35:29 [Server 127.0.0.1:8003] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [Geek-Cache] hit
{
  "value": "NjMy"
}
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:29 [Server 127.0.0.1:8003] Pick peer 127.0.0.1:8002
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8002] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:29 [Server 127.0.0.1:8002] Pick peer 127.0.0.1:8002
2023/03/22 22:35:29 [Geek-Cache] hit
{
  "value": "NjMw"
}
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom1)
2023/03/22 22:35:29 [Server 127.0.0.1:8003] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [Geek-Cache] hit
{
  "value": "NjMx"
}
2023/03/22 22:35:29 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom2)
2023/03/22 22:35:29 [Server 127.0.0.1:8003] Pick peer 127.0.0.1:8003
2023/03/22 22:35:29 [Geek-Cache] hit
{
  "value": "NjMy"
}
2023/03/22 22:35:32 [Geek-Cache 127.0.0.1:8001] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:32 [SlowDB] search hot key Tom
{
  "value": "NjMw"
}
2023/03/22 22:35:32 [Geek-Cache 127.0.0.1:8001] Recv RPC Request for get- (scores)/(Tom1)
2023/03/22 22:35:32 [SlowDB] search hot key Tom1
{
  "value": "NjMx"
}
2023/03/22 22:35:32 [Geek-Cache 127.0.0.1:8001] Recv RPC Request for get- (scores)/(Tom2)
2023/03/22 22:35:32 [SlowDB] search hot key Tom2
{
  "value": "NjMy"
}
2023/03/22 22:35:32 [Geek-Cache 127.0.0.1:8002] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:32 [SlowDB] search hot key Tom
{
  "value": "NjMw"
}
2023/03/22 22:35:32 [Geek-Cache 127.0.0.1:8002] Recv RPC Request for get- (scores)/(Tom1)
2023/03/22 22:35:32 [SlowDB] search hot key Tom1
{
  "value": "NjMx"
}
2023/03/22 22:35:32 [Geek-Cache 127.0.0.1:8002] Recv RPC Request for get- (scores)/(Tom2)
2023/03/22 22:35:32 [SlowDB] search hot key Tom2
{
  "value": "NjMy"
}
2023/03/22 22:35:32 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:32 [SlowDB] search hot key Tom
{
  "value": "NjMw"
}
2023/03/22 22:35:32 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom1)
2023/03/22 22:35:32 [SlowDB] search hot key Tom1
{
  "value": "NjMx"
}
2023/03/22 22:35:32 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom2)
2023/03/22 22:35:32 [SlowDB] search hot key Tom2
{
  "value": "NjMy"
}
./a.sh：行 36: 37053 已杀死               ./server -port=8002
2023/03/22 22:35:38 [Geek-Cache 127.0.0.1:8001] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:38 [Server 127.0.0.1:8001] Pick peer 127.0.0.1:8001
2023/03/22 22:35:38 [SlowDB] search key Tom
{
  "value": "NjMw"
}
2023/03/22 22:35:38 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:38 [Server 127.0.0.1:8003] Pick peer 127.0.0.1:8001
2023/03/22 22:35:38 [Geek-Cache 127.0.0.1:8001] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:38 [Server 127.0.0.1:8001] Pick peer 127.0.0.1:8001
2023/03/22 22:35:38 [Geek-Cache] hit
{
  "value": "NjMw"
}
2023/03/22 22:35:41 [Geek-Cache 127.0.0.1:8001] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:41 [Geek-Cache] hit
{
  "value": "NjMw"
}
2023/03/22 22:35:41 [Geek-Cache 127.0.0.1:8003] Recv RPC Request for get- (scores)/(Tom)
2023/03/22 22:35:41 [Geek-Cache] hit
{
  "value": "NjMw"
}
```

## Tech Stack

GoLang+GRPC+ETCD

## Feature

- 使用一致性哈希解决缓存副本冗余、Rehashing开销大、缓存雪崩的问题
- 使用SingleFlight解决缓存击穿问题
- 使用ProtoBuf进行节点间通信，编码报文，提高效率
- 构造虚拟节点使得请求映射负载均衡，提⾼可⽤性
- 实现了过期删除策略，惰性清除与定期清除，防⽌经常扫描内存⽽导致开销过⼤
- 使用LRU缓存淘汰算法解决资源限制的问题
- 使用ETCD服务发现动态更新哈希环
- 支持并发读
- 支持Hot Key缓存预热