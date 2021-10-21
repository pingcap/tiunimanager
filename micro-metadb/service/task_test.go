
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

package service

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"os"
	"testing"
)

func TestDBServiceHandler_CreateFlow(t *testing.T) {
	type args struct {
		ctx context.Context
		req *dbpb.DBCreateFlowRequest
		rsp *dbpb.DBCreateFlowResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(args *args) bool
	}{
		{"normal", args{nil, &dbpb.DBCreateFlowRequest{
			Flow: &dbpb.DBFlowDTO{FlowName: "aaa", Operator: "111111"},
		}, &dbpb.DBCreateFlowResponse{}}, false, []func(args *args) bool{
			func(args *args) bool {
				return args.req.Flow.FlowName == args.rsp.Flow.FlowName && args.rsp.Flow.Id > 0
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := "tmp/" + uuidutil.GenerateID()
			os.MkdirAll(testFile, 0755)
			d := NewDBServiceHandler(testFile, nil)
			if err := d.CreateFlow(tt.args.ctx, tt.args.req, tt.args.rsp); (err != nil) != tt.wantErr {
				t.Errorf("CreateFlow() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				for _, f := range tt.asserts {
					if !f(&tt.args) {
						t.Errorf("CreateTask() req = %v, resp %v", tt.args.req, tt.args.rsp)
					}
				}
			}
		})
	}

}

func TestDBServiceHandler_CreateTask(t *testing.T) {
	type args struct {
		ctx context.Context
		req *dbpb.DBCreateTaskRequest
		rsp *dbpb.DBCreateTaskResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DBServiceHandler{}
			if err := d.CreateTask(tt.args.ctx, tt.args.req, tt.args.rsp); (err != nil) != tt.wantErr {
				t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBServiceHandler_LoadFlow(t *testing.T) {
	type args struct {
		ctx context.Context
		req *dbpb.DBLoadFlowRequest
		rsp *dbpb.DBLoadFlowResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DBServiceHandler{}
			if err := d.LoadFlow(tt.args.ctx, tt.args.req, tt.args.rsp); (err != nil) != tt.wantErr {
				t.Errorf("LoadFlow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBServiceHandler_LoadTask(t *testing.T) {
	type args struct {
		ctx context.Context
		req *dbpb.DBLoadTaskRequest
		rsp *dbpb.DBLoadTaskResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DBServiceHandler{}
			if err := d.LoadTask(tt.args.ctx, tt.args.req, tt.args.rsp); (err != nil) != tt.wantErr {
				t.Errorf("LoadTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBServiceHandler_UpdateFlow(t *testing.T) {
	type args struct {
		ctx context.Context
		req *dbpb.DBUpdateFlowRequest
		rsp *dbpb.DBUpdateFlowResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DBServiceHandler{}
			if err := d.UpdateFlow(tt.args.ctx, tt.args.req, tt.args.rsp); (err != nil) != tt.wantErr {
				t.Errorf("UpdateFlow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBServiceHandler_UpdateTask(t *testing.T) {
	type args struct {
		ctx context.Context
		req *dbpb.DBUpdateTaskRequest
		rsp *dbpb.DBUpdateTaskResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DBServiceHandler{}
			if err := d.UpdateTask(tt.args.ctx, tt.args.req, tt.args.rsp); (err != nil) != tt.wantErr {
				t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
