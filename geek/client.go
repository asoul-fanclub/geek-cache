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
	name        string // name of remote server, e.g. ip:port
	serviceName string // name of service, e.g. geek-cache
}

// NewClient creates a new client
func NewClient(name, serviceName string) (*Client, error) {
	return &Client{
		name:        name,
		serviceName: serviceName,
	}, nil
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

	conn, err := registry.EtcdDial(cli, c.serviceName, c.name)
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
		return nil, fmt.Errorf("could not get %s-%s from peer %s", group, key, c.name)
	}
	return resp.GetValue(), nil
}

// resure implemented
var _ PeerGetter = (*Client)(nil)
