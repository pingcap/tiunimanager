/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
	"go.etcd.io/etcd/client/pkg/v3/transport"
	"strings"
	"time"

	"github.com/pingcap/tiunimanager/common/constants"

	clientv2 "go.etcd.io/etcd/client/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdTimeOut etcd time out
const (
	etcdTimeOut      = time.Second * 3
	etcdV2TlsTimeOut = 10 * time.Second
	httpProtocol     = "https://"
)

type EtcdClientV3 struct {
	cli *clientv3.Client
}

type EtcdClientV2 struct {
	cli     clientv2.Client
	keysAPI clientv2.KeysAPI
}

type EtcdCertTransport struct {
	PeerTLSInfo   transport.TLSInfo
	ServerTLSInfo transport.TLSInfo
	ClientTLSInfo transport.TLSInfo
}

var etcdClientV3 *EtcdClientV3
var etcdClientV2 *EtcdClientV2
var EtcdCert EtcdCertTransport

func InitEtcdClient(etcdAddress []string) *EtcdClientV3 {
	tlsInfo := EtcdCert.ClientTLSInfo
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		panic(err)
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddress,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})

	if err != nil {
		panic(err)
	}
	// Wait for the ETCD server to be ready in a loop.
	for {
		ctx, cancel := context.WithTimeout(context.Background(), etcdTimeOut)
		_, err = cli.MemberList(ctx)
		cancel()
		if err != nil {
			LogForkFile(constants.LogFileSystem).Warnf("connect etcd server [%v] failed, err: %v\n", etcdAddress, err)
			continue
		}
		break
	}
	etcdClientV3 = &EtcdClientV3{cli: cli}
	return etcdClientV3
}

func (etcd *EtcdClientV3) Put(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeOut)
	_, err := etcd.cli.Put(ctx, key, value)
	defer cancel()
	return err
}

func (etcd *EtcdClientV3) Get(key string) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeOut)
	resp, err := etcd.cli.Get(ctx, key)
	defer cancel()
	return resp, err
}

func (etcd *EtcdClientV3) Watch(key string, ops ...clientv3.OpOption) (clientv3.WatchChan, error) {
	rch := etcd.cli.Watch(context.Background(), key, ops...)
	return rch, nil
}

func (etcd *EtcdClientV3) Lease() clientv3.Lease {
	return clientv3.NewLease(etcd.cli)
}

func InitEtcdClientV2(etcdAddress []string) *EtcdClientV2 {
	// Determine whether to include 'http://'
	for i, addr := range etcdAddress {
		if !strings.HasPrefix(addr, httpProtocol) {
			etcdAddress[i] = httpProtocol + addr
		}
	}
	tls := EtcdCert.ClientTLSInfo
	ts, err := transport.NewTransport(tls, etcdV2TlsTimeOut)
	if err != nil {
		panic(err)
	}
	cli, err := clientv2.New(clientv2.Config{
		Endpoints: etcdAddress,
		Transport: ts,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	})

	if err != nil {
		panic(err)
	}
	keysAPI := clientv2.NewKeysAPI(cli)
	etcdClientV2 = &EtcdClientV2{
		cli:     cli,
		keysAPI: keysAPI,
	}
	return etcdClientV2
}

func (etcd *EtcdClientV2) SetWithTtl(key, value string, ttl int) error {
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeOut)
	options := &clientv2.SetOptions{}
	if ttl > 0 {
		options.TTL = time.Second * time.Duration(ttl)
	}
	_, err := etcd.keysAPI.Set(ctx, key, value, options)
	defer cancel()
	return err
}

func (etcd *EtcdClientV2) Set(key, value string) error {
	return etcd.SetWithTtl(key, value, 0)
}

func (etcd *EtcdClientV2) Get(key string) (*clientv2.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), etcdTimeOut)
	resp, err := etcd.keysAPI.Get(ctx, key, nil)
	defer cancel()
	return resp, err
}