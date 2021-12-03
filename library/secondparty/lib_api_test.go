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

/*******************************************************************************
 * @File: lib_api_test.go
 * @Description: test tidb component uses api to update parameters
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/2 16:50
*******************************************************************************/

package secondparty

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var secondMicro3 *SecondMicro

func init() {
	secondMicro3 = &SecondMicro{}
}

func TestSecondMicro_ApiEditConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(nil)
		r.ParseForm()
	}))
	defer server.Close()

	ipAndPort := strings.TrimLeft(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Errorf(err.Error())
	}
	type args struct {
		req ApiEditConfigReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, hasSuc bool) bool
	}{
		{
			"normal",
			args{
				req: ApiEditConfigReq{
					TiDBClusterComponent: "tidb",
					InstanceHost:         host,
					InstancePort:         uint(port),
					Headers:              map[string]string{},
					ConfigMap: map[string]interface{}{
						"binlog_cache": 1024,
					},
				},
			},
			false,
			[]func(a args, hasSuc bool) bool{
				func(a args, hasSuc bool) bool { return hasSuc },
			},
		},
		{
			"normal2",
			args{
				req: ApiEditConfigReq{
					TiDBClusterComponent: "tikv",
					InstanceHost:         host,
					InstancePort:         uint(port),
					Headers:              map[string]string{},
					ConfigMap: map[string]interface{}{
						"binlog_cache": 1024,
					},
				},
			},
			false,
			[]func(a args, hasSuc bool) bool{
				func(a args, hasSuc bool) bool { return hasSuc },
			},
		},
		{
			"normal3",
			args{
				req: ApiEditConfigReq{
					TiDBClusterComponent: "pd",
					InstanceHost:         host,
					InstancePort:         uint(port),
					Headers:              map[string]string{},
					ConfigMap: map[string]interface{}{
						"binlog_cache": 1024,
					},
				},
			},
			false,
			[]func(a args, hasSuc bool) bool{
				func(a args, hasSuc bool) bool { return hasSuc },
			},
		},
		{
			"cluster component is null",
			args{
				req: ApiEditConfigReq{
					TiDBClusterComponent: "",
					InstanceHost:         host,
					InstancePort:         uint(port),
					Headers:              map[string]string{},
					ConfigMap: map[string]interface{}{
						"binlog_cache": 1024,
					},
				},
			},
			true,
			[]func(a args, hasSuc bool) bool{
				func(a args, hasSuc bool) bool { return hasSuc },
			},
		},
		{
			"config map is null",
			args{
				req: ApiEditConfigReq{
					TiDBClusterComponent: "pd",
					InstanceHost:         host,
					InstancePort:         uint(port),
					Headers:              nil,
					ConfigMap:            nil,
				},
			},
			true,
			[]func(a args, hasSuc bool) bool{
				func(a args, hasSuc bool) bool { return hasSuc },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasSuc, err := secondMicro3.ApiEditConfig(context.TODO(), tt.args.req)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("ApiEditConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, hasSuc) {
					t.Errorf("ApiEditConfig() test error, testname = %v, assert %v, args = %v, got param id = %v", tt.name, i, tt.args, hasSuc)
				}
			}
		})
	}
}
