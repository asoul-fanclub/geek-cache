package geek

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/Makonike/geek-cache/geek/pb"
	registry "github.com/Makonike/geek-cache/geek/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	addr        string // name of remote server, e.g. ip:port
	serviceName string // name of service, e.g. geek-cache
}

// NewClient creates a new client
func NewClient(addr, serviceName string) *Client {
	return &Client{
		addr:        addr,
		serviceName: serviceName,
	}
}

// Get send the url for getting specific group and key,
// and return the result
func (c *Client) Get(group, key string) ([]byte, error) {
	cli, err := clientv3.New(*registry.GlobalClientConfig)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer cli.Close()

	conn, err := registry.EtcdDial(cli, c.serviceName, c.addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	grpcCLient := pb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := grpcCLient.Get(ctx, &pb.Request{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get %s-%s from peer %s", group, key, c.addr)
	}
	return resp.GetValue(), nil
}

// Delete send the url for getting specific group and key,
// and return the result
func (c *Client) Delete(group string, key string) (bool, error) {
	cli, err := clientv3.New(*registry.GlobalClientConfig)

	if err != nil {
		log.Fatal(err)
		return false, err
	}
	defer cli.Close()

	conn, err := registry.EtcdDial(cli, c.serviceName, c.addr)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	grpcCLient := pb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := grpcCLient.Delete(ctx, &pb.Request{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return false, fmt.Errorf("could not get %s-%s from peer %s", group, key, c.addr)
	}
	return resp.GetValue(), nil
}

// resure implemented
var _ PeerGetter = (*Client)(nil)
