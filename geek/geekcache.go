package geek

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Makonike/geek-cache/geek/singleflight"
)

var (
	lock   sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	name      string              // group name
	getter    Getter              // 缓存未名中时的callback
	mainCache cache               // main cache
	peers     PeerPicker          // pick function
	loader    *singleflight.Group // make sure that each key is only fetched once
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called multiple times")
	}
	g.peers = peers
}

// NewGroup 新创建一个Group
// 如果存在同名的group会进行覆盖
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	lock.Lock()
	defer lock.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
		loader: &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	lock.RLock()
	g := groups[name]
	lock.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Geek-Cache] hit")
		return v, nil
	}
	return g.load(key)
}

// get from peer first, then get locally
func (g *Group) load(key string) (ByteView, error) {
	// make sure requests for the key only execute once in concurrent condition
	v, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				} else {
					log.Println("[Geek-Cache] Failed to get from peer", err)
				}
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return v.(ByteView), nil
	}
	return ByteView{}, err
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{
		b: cloneBytes(bytes),
	}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, f, expirationTime := g.getter.Get(key)
	if !f {
		return ByteView{}, fmt.Errorf("data not found")
	}
	bw := ByteView{cloneBytes(bytes)}
	if !expirationTime.IsZero() {
		g.mainCache.addWithExpiration(key, bw, expirationTime)
	} else {
		g.mainCache.add(key, bw)
	}
	return bw, nil
}

// Getter loads data for a key locally
// call back when a key cache missed
// impl by user
type Getter interface {
	Get(key string) ([]byte, bool, time.Time)
}

type GetterFunc func(key string) ([]byte, bool, time.Time)

func (f GetterFunc) Get(key string) ([]byte, bool, time.Time) {
	return f(key)
}

func DestroyGroup(name string) {
	g := GetGroup(name)
	if g != nil {
		delete(groups, name)
		log.Printf("Destroy cache [%s]", name)
	}
}
