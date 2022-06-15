/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
 * @File: http_test.go
 * @Description: test util http method
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/2 17:44
*******************************************************************************/

package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(nil)
		r.ParseForm()
	}))
	defer server.Close()

	type args struct {
		reqURL    string
		reqParams map[string]string
		headers   map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, resp *http.Response) bool
	}{
		{
			"normal",
			args{
				reqURL:    server.URL + "/status",
				reqParams: nil,
				headers:   nil,
			},
			false,
			[]func(a args, resp *http.Response) bool{
				func(a args, resp *http.Response) bool { return resp.StatusCode == http.StatusOK },
			},
		},
		{
			"normal2",
			args{
				reqURL: server.URL + "/status",
				reqParams: map[string]string{
					"param1": "123",
				},
				headers: map[string]string{
					"token": "123",
				},
			},
			false,
			[]func(a args, resp *http.Response) bool{
				func(a args, resp *http.Response) bool { return resp.StatusCode == http.StatusOK },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Get(tt.args.reqURL, tt.args.reqParams, tt.args.headers)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, resp) {
					t.Errorf("Get() test error, testname = %v, assert %v, args = %v, got param id = %v", tt.name, i, tt.args, resp)
				}
			}
		})
	}
}

func TestPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(nil)
		r.ParseForm()
	}))
	defer server.Close()

	type args struct {
		reqURL    string
		reqParams map[string]interface{}
		headers   map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, resp *http.Response) bool
	}{
		{
			"normal",
			args{
				reqURL: server.URL + "/person",
				reqParams: map[string]interface{}{
					"param1": "123",
				},
				headers: map[string]string{
					"token": "123",
				},
			},
			false,
			[]func(a args, resp *http.Response) bool{
				func(a args, resp *http.Response) bool { return resp.StatusCode == http.StatusOK },
			},
		},
		{
			"req param is null",
			args{
				reqURL:    server.URL + "/person",
				reqParams: nil,
				headers:   nil,
			},
			true,
			[]func(a args, resp *http.Response) bool{
				func(a args, resp *http.Response) bool { return resp.StatusCode == http.StatusOK },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := PostJSON(tt.args.reqURL, tt.args.reqParams, tt.args.headers)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("PostJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, resp) {
					t.Errorf("PostJSON() test error, testname = %v, assert %v, args = %v, got param id = %v", tt.name, i, tt.args, resp)
				}
			}
		})
	}
}

func TestPostForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(nil)
		r.ParseForm()
	}))
	defer server.Close()

	type args struct {
		reqURL    string
		reqParams map[string]interface{}
		headers   map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, resp *http.Response) bool
	}{
		{
			"normal",
			args{
				reqURL: server.URL + "/person",
				reqParams: map[string]interface{}{
					"param1": "123",
				},
				headers: map[string]string{
					"token": "123",
				},
			},
			false,
			[]func(a args, resp *http.Response) bool{
				func(a args, resp *http.Response) bool { return resp.StatusCode == http.StatusOK },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := PostForm(tt.args.reqURL, tt.args.reqParams, tt.args.headers)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("PostForm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, resp) {
					t.Errorf("PostForm() test error, testname = %v, assert %v, args = %v, got param id = %v", tt.name, i, tt.args, resp)
				}
			}
		})
	}
}

func TestPostFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(nil)
		r.ParseForm()
	}))
	defer server.Close()

	type args struct {
		reqURL    string
		reqParams map[string]interface{}
		headers   map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, resp *http.Response) bool
	}{
		{
			"normal",
			args{
				reqURL: server.URL + "/images",
				reqParams: map[string]interface{}{
					"param1": "123",
				},
				headers: map[string]string{},
			},
			false,
			[]func(a args, resp *http.Response) bool{
				func(a args, resp *http.Response) bool { return resp.StatusCode == http.StatusOK },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := make([]UploadFile, 1)
			files[0] = UploadFile{
				Name:     "http_test.go",
				Filepath: "./http_test.go",
			}
			resp, err := PostFile(tt.args.reqURL, tt.args.reqParams, files, tt.args.headers)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("PostJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, resp) {
					t.Errorf("PostJSON() test error, testname = %v, assert %v, args = %v, got param id = %v", tt.name, i, tt.args, resp)
				}
			}
		})
	}
}
