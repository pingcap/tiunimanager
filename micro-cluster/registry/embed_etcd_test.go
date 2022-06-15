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

package registry

import (
	"net/url"
	"strings"
	"testing"

	"github.com/pingcap/tiunimanager/library/framework"
)

func TestParseEtcdConfig(t *testing.T) {
	f := framework.InitBaseFrameworkForUt(framework.MetaDBService)
	clientArgs = f.GetClientArgs()
	log = f.Log()

	tests := []struct {
		name    string
		args    *EmbedEtcdConfig
		asserts []func(args *EmbedEtcdConfig, resp *EmbedEtcdConfig) bool
	}{
		{"normal", &EmbedEtcdConfig{
			Name:           "etcd0",
			Index:          0,
			Dir:            "./testdata/etcd/data_etcd0",
			ClientUrl:      "127.0.0.1:4101",
			PeerUrl:        "127.0.0.1:4102",
			EtcdClientUrls: []string{"127.0.0.1:4101"},
			EtcdPeerUrls:   []string{"127.0.0.1:4102"},
		},
			[]func(args *EmbedEtcdConfig, resp *EmbedEtcdConfig) bool{
				func(args *EmbedEtcdConfig, resp *EmbedEtcdConfig) bool { return args.Name == resp.Name },
				func(args *EmbedEtcdConfig, resp *EmbedEtcdConfig) bool { return args.Dir == resp.Dir },
				func(args *EmbedEtcdConfig, resp *EmbedEtcdConfig) bool { return args.ClientUrl == resp.ClientUrl },
				func(args *EmbedEtcdConfig, resp *EmbedEtcdConfig) bool { return args.PeerUrl == resp.PeerUrl },
				func(args *EmbedEtcdConfig, resp *EmbedEtcdConfig) bool { return args.Index == resp.Index },
			},
		},
	}

	config, err := parseEtcdConfig()
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, assert := range tt.asserts {
				if !assert(tt.args, config) {
					t.Errorf("parseEtcdConfig assert false, assert index = %v, expire = %v, got = %v", i, tt.args, config)
				}
			}
		})
	}
}

func TestParsePeers(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		asserts []func(args []string, resp []url.URL) bool
	}{
		{
			"normal",
			[]string{
				"127.0.0.1:4102",
				"127.0.0.1:5102",
				"127.0.0.1:6102",
			},
			[]func(args []string, resp []url.URL) bool{
				func(args []string, resp []url.URL) bool {
					return len(resp) == 3
				},
				func(args []string, resp []url.URL) bool {
					if len(args) > 0 && len(resp) > 0 {
						u, err := url.Parse(httpProtocol + args[0])
						if err != nil {
							return false
						}
						return u.Path == resp[0].Path
					}
					return false
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, assert := range tt.asserts {
				ret := parsePeers(tt.args)
				if !assert(tt.args, ret) {
					t.Errorf("parsePeers assert false, assert index = %v, expire = %v, got = %v", i, tt.args, ret)
				}
			}
		})
	}
}

func TestParseClients(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		asserts []func(args []string, resp []url.URL) bool
	}{
		{
			"normal",
			[]string{
				"127.0.0.1:4101",
				"127.0.0.1:5101",
				"127.0.0.1:6101",
			},
			[]func(args []string, resp []url.URL) bool{
				func(args []string, resp []url.URL) bool {
					return len(resp) == 3
				},
				func(args []string, resp []url.URL) bool {
					if len(args) > 0 && len(resp) > 0 {
						u, err := url.Parse(httpProtocol + args[0])
						if err != nil {
							return false
						}
						return u.Path == resp[0].Path
					}
					return false
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, assert := range tt.asserts {
				ret := parseClients(tt.args)
				if !assert(tt.args, ret) {
					t.Errorf("parseClients assert false, assert index = %v, expire = %v, got = %v", i, tt.args, ret)
				}
			}
		})
	}
}

func TestParseInitialCluster(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		asserts []func(args []string, resp string) bool
	}{
		{
			"normal",
			[]string{
				"127.0.0.1:4102",
				"127.0.0.1:5102",
				"127.0.0.1:6102",
			},
			[]func(args []string, resp string) bool{
				func(args []string, resp string) bool {
					return len(resp) > 0
				},
				func(args []string, resp string) bool {
					return len(args) == 3
				},
				func(args []string, resp string) bool {
					return strings.Contains(resp, args[0])
				},
				func(args []string, resp string) bool {
					return strings.Contains(resp, args[1])
				},
				func(args []string, resp string) bool {
					return strings.Contains(resp, args[2])
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, assert := range tt.asserts {
				ret := parseInitialCluster(tt.args)
				if !assert(tt.args, ret) {
					t.Errorf("parseInitialCluster assert false, assert index = %v, expire = %v, got = %v", i, tt.args, ret)
				}
			}
		})
	}
}
