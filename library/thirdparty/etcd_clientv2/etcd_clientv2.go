package etcd_clientv2

import (
	"context"
	"strings"
	"time"

	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/library/common"

	"go.etcd.io/etcd/client/v2"
)

type EtcdClient struct {
	cli     client.Client
	keysAPI client.KeysAPI
}

var etcdClient *EtcdClient

func InitEtcdClient(etcdAddress []string) *EtcdClient {
	// Determine whether to include 'http://'
	for i, addr := range etcdAddress {
		if !strings.HasPrefix(addr, common.HttpProtocol) {
			etcdAddress[i] = common.HttpProtocol + addr
		}
	}
	cli, err := client.New(client.Config{
		Endpoints: etcdAddress,
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	})
	if err != nil {
		panic(err)
	}
	keysAPI := client.NewKeysAPI(cli)
	etcdClient = &EtcdClient{
		cli:     cli,
		keysAPI: keysAPI,
	}
	return etcdClient
}

func (etcd *EtcdClient) SetWithTtl(key, value string, ttl int) error {
	ctx, cancel := context.WithTimeout(context.Background(), framework.EtcdTimeOut)
	options := &client.SetOptions{}
	if ttl > 0 {
		options.TTL = time.Second * time.Duration(ttl)
	}
	_, err := etcd.keysAPI.Set(ctx, key, value, options)
	defer cancel()
	return err
}

func (etcd *EtcdClient) Set(key, value string) error {
	return etcd.SetWithTtl(key, value, 0)
}

func (etcd *EtcdClient) Get(key string) (*client.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), framework.EtcdTimeOut)
	resp, err := etcd.keysAPI.Get(ctx, key, nil)
	defer cancel()
	return resp, err
}
