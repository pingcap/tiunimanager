package service

import (
	"context"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"os"
	"testing"
)

func TestDBServiceHandler_CreateFlow(t *testing.T) {
	type args struct {
		ctx context.Context
		req *dbPb.DBCreateFlowRequest
		rsp *dbPb.DBCreateFlowResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(args *args) bool
	}{
		{"normal", args{nil, &dbPb.DBCreateFlowRequest{
			Flow: &dbPb.DBFlowDTO{FlowName: "aaa"},
		}, &dbPb.DBCreateFlowResponse{}}, false, []func(args *args) bool{
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
		req *dbPb.DBCreateTaskRequest
		rsp *dbPb.DBCreateTaskResponse
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
		req *dbPb.DBLoadFlowRequest
		rsp *dbPb.DBLoadFlowResponse
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
		req *dbPb.DBLoadTaskRequest
		rsp *dbPb.DBLoadTaskResponse
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
		req *dbPb.DBUpdateFlowRequest
		rsp *dbPb.DBUpdateFlowResponse
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
		req *dbPb.DBUpdateTaskRequest
		rsp *dbPb.DBUpdateTaskResponse
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
