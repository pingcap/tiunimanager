
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

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
