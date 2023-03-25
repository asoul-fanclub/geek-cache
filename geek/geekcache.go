package geek

import (
	"fmt"
	"log"
	"time"

	"github.com/Makonike/geek-cache/geek/singleflight"
)

var (
	groups  = make(map[string]*Group)
	hgroups = make(map[string]*Group)
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
func NewGroup(name string, cacheBytes int64, isHot bool, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
		loader: &singleflight.Group{},
	}
	if !isHot {
		groups[name] = g
	} else {
		hgroups[name] = g
	}
	return g
}

func GetGroup(name string) *Group {
	g := groups[name]
	return g
}

func GetHGroup(name string) *Group {
	g := hgroups[name]
	return g
}

func (g *Group) HGet(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// make sure requests for the key only execute once in concurrent condition
	v, err := g.loader.Do(key, func() (interface{}, error) {
		return g.getLocally(key)
	})

	if err == nil {
		return v.(ByteView), nil
	}
	return ByteView{}, err
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	return g.load(key)
}

// get from peer first, then get locally
func (g *Group) load(key string) (ByteView, error) {
	// make sure requests for the key only execute once in concurrent condition
	v, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok, isSelf := g.peers.PickPeer(key); ok {
				if isSelf {
					if v, ok := g.mainCache.get(key); ok {
						log.Println("[Geek-Cache] hit")
						return v, nil
					}
				} else {
					if value, err := g.getFromPeer(peer, key); err == nil {
						return value, nil
					} else {
						log.Println("[Geek-Cache] Failed to get from peer", err)
					}
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

func (g *Group) Delete(key string) (bool, error) {
	if key == "" {
		return true, fmt.Errorf("key is required")
	}
	// Peer is not set, delete from local
	if g.peers == nil {
		return g.mainCache.delete(key), nil
	}
	// The peer is set,
	peer, ok, isSelf := g.peers.PickPeer(key)
	if !ok {
		return false, nil
	}
	if isSelf {
		return g.mainCache.delete(key), nil
	} else {
		//use other server to delete the key-value
		success, err := g.deleteFromPeer(peer, key)
		return success, err
	}
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

func (g *Group) deleteFromPeer(peer PeerGetter, key string) (bool, error) {
	success, err := peer.Delete(g.name, key)
	if err != nil {
		return false, err
	}
	return success, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// have a try again
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Geek-Cache] hit")
		return v, nil
	}
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
