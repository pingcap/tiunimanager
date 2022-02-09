/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 ******************************************************************************/

/*******************************************************************************
 * @File: clusterconfig_test.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/8 15:25
*******************************************************************************/

package cdc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/pingcap-inc/tiem/message/cluster"
)

func TestSecondMicro_ApiEditConfig_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(nil)
		r.ParseForm()
	}))
	defer server.Close()

	ipAndPort := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Errorf(err.Error())
	}
	type args struct {
		req cluster.ApiEditConfigReq
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
				req: cluster.ApiEditConfigReq{
					InstanceHost: host,
					InstancePort: uint(port),
					Headers:      map[string]string{},
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
			"config map is null",
			args{
				req: cluster.ApiEditConfigReq{
					InstanceHost: host,
					InstancePort: uint(port),
					Headers:      nil,
					ConfigMap:    nil,
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
			hasSuc, err := ApiService.ApiEditConfig(context.TODO(), tt.args.req)
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
func TestSecondMicro_ApiEditConfig_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(nil)
		r.ParseForm()
	}))
	defer server.Close()

	ipAndPort := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Errorf(err.Error())
	}
	type args struct {
		req cluster.ApiEditConfigReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, hasSuc bool) bool
	}{
		{
			"err1",
			args{
				req: cluster.ApiEditConfigReq{
					InstanceHost: host,
					InstancePort: uint(port),
					Headers:      map[string]string{},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasSuc, err := ApiService.ApiEditConfig(context.TODO(), tt.args.req)
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