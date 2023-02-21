package registry

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

var (
	GlobalClientConfig *clientv3.Config = defaultEtcdConfig
	defaultEtcdConfig                   = &clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// 在租赁模式添加一对kv至etcd
func etcdAdd(c *clientv3.Client, lid clientv3.LeaseID, service, addr string) error {
	em, err := endpoints.NewManager(c, service)
	if err != nil {
		return err
	}
	return em.AddEndpoint(c.Ctx(), service+"/"+addr, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(lid))
}

// Register register a service to etcd
// no return if not error
func Register(service, addr string, stop chan error) error {
	cli, err := clientv3.New(*GlobalClientConfig)
	if err != nil {
		return fmt.Errorf("create etcd client failed: %v", err)
	}
	defer cli.Close()
	// create a lease for 5 seconds
	resp, err := cli.Grant(context.Background(), 2)
	if err != nil {
		return fmt.Errorf("create lease failed: %v", err)
	}
	leaseId := resp.ID
	// register service
	err = etcdAdd(cli, leaseId, service, addr)
	if err != nil {
		return fmt.Errorf("add etcd record failed: %v", err)
	}
	// set heartbeat
	ch, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("set keepalive failed: %v", err)
	}
	log.Printf("[%s] register service success", addr)
	for {
		select {
		case err := <-stop:
			if err != nil {
				log.Println(err)
			}
		case <-cli.Ctx().Done():
			log.Println("service closed")
			return nil
		case _, ok := <-ch:
			// 监听租约
			if !ok {
				log.Println("keepalive channel closed")
				_, err := cli.Revoke(context.Background(), leaseId)
				return err
			}
		}
	}
}
