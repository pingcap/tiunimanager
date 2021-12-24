
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package framework

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdTimeOut etcd time out
const (
	EtcdTimeOut = time.Second * 3
)

type EtcdClient struct {
	cli *clientv3.Client
}

var etcdClient *EtcdClient

func InitEtcdClient(etcdAddress []string) *EtcdClient {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddress,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	// Wait for the ETCD server to be ready in a loop.
	for {
		ctx, cancel := context.WithTimeout(context.Background(), EtcdTimeOut)
		_, err = cli.MemberList(ctx)
		cancel()
		if err != nil {
			LogForkFile(constants.LogFileSystem).Warnf("connect etcd server [%v] failed, err: %v\n", etcdAddress, err)
			continue
		}
		break
	}
	etcdClient = &EtcdClient{cli: cli}
	return etcdClient
}

func (etcd *EtcdClient) Put(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), EtcdTimeOut)
	_, err := etcd.cli.Put(ctx, key, value)
	defer cancel()
	return err
}

func (etcd *EtcdClient) Get(key string) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), EtcdTimeOut)
	resp, err := etcd.cli.Get(ctx, key)
	defer cancel()
	return resp, err
}

func (etcd *EtcdClient) Watch(key string, ops ...clientv3.OpOption) (clientv3.WatchChan, error) {
	rch := etcd.cli.Watch(context.Background(), key, ops...)
	return rch, nil
}

func (etcd *EtcdClient) Lease() clientv3.Lease {
	return clientv3.NewLease(etcd.cli)
}
