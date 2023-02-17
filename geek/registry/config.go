package registry

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	GlobalClientConfig *clientv3.Config = defaultEtcdConfig
)
