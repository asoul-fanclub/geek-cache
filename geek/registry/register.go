package registy

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

var (
	DefaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:12379"},
		DialTimeout: 5 * time.Second,
	}
)

// EtcdDial request a server from grpc
// Connection can be obtained by providing an etcd client and service name
func EtcdDial(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(
		"etcd:///"+service,
		grpc.WithResolvers(etcdResolver),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
}
