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

package domain

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/test/mocksecondparty"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"

	copywriting2 "github.com/pingcap-inc/tiem/library/copywriting"
)

func testStart(task *TaskEntity, context *FlowContext) bool {
	context.SetData("testKey", 999)
	task.Success("testStart")
	return true
}

func testDoing(task *TaskEntity, context *FlowContext) bool {
	task.Success("doing")
	return true
}

func testError(task *TaskEntity, context *FlowContext) bool {
	task.Fail(errors.New("some error"))
	return false
}

func initFlow() {
	FlowWorkDefineMap = map[string]*FlowWorkDefine{
		"testFlow": {
			FlowName:    "testFlow",
			StatusAlias: "test",
			TaskNodes: map[string]*TaskDefine{
				"start":     {"doStart", "startDone", "fail", SyncFuncTask, testStart},
				"startDone": {"doSomething", "done", "fail", SyncFuncTask, testDoing},
				"done":      {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail":      {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		"testFlow2": {
			FlowName:    "testFlow2",
			StatusAlias: "test2",
			TaskNodes: map[string]*TaskDefine{
				"start":     {"doStart", "startDone", "fail", CallbackTask, testStart},
				"startDone": {"doSomething", "done", "fail", SyncFuncTask, testDoing},
				"done":      {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail":      {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		"testFlow3": {
			FlowName:    "testFlow3",
			StatusAlias: "test3",
			TaskNodes: map[string]*TaskDefine{
				"start":     {"doStart", "startDone", "fail", SyncFuncTask, testError},
				"startDone": {"doSomething", "done", "fail", SyncFuncTask, testDoing},
				"fail":      {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		"testFlow4": {
			FlowName:    "testFlow4",
			StatusAlias: "test4",
			TaskNodes: map[string]*TaskDefine{
				"start":     {"doStart", "startDone", "fail", PollingTasK, testStart},
				"startDone": {"doSomething", "done", "fail", SyncFuncTask, testDoing},
				"done":      {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail":      {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		FlowCreateCluster: {
			FlowName:    FlowCreateCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowCreateCluster),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		FlowDeleteCluster: {
			FlowName:    FlowDeleteCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowDeleteCluster),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		FlowBackupCluster: {
			FlowName:    FlowBackupCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowBackupCluster),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		FlowRecoverCluster: {
			FlowName:    FlowRecoverCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowRecoverCluster),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		FlowModifyParameters: {
			FlowName:    FlowModifyParameters,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowModifyParameters),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		FlowExportData: {
			FlowName:    FlowExportData,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowExportData),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		FlowImportData: {
			FlowName:    FlowImportData,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowImportData),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		FlowRestartCluster: {
			FlowName:    FlowRestartCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowRestartCluster),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				return nil
			},
		},
		FlowStopCluster: {
			FlowName:    FlowStopCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowStopCluster),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				return nil
			},
		},
		FlowTakeoverCluster: {
			FlowName:    FlowTakeoverCluster,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowTakeoverCluster),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
		FlowBuildLogConfig: {
			FlowName:    FlowBuildLogConfig,
			StatusAlias: copywriting2.DisplayByDefault(copywriting2.CWFlowBuildLogConfig),
			TaskNodes: map[string]*TaskDefine{
				"start": {"doing", "done", "fail", SyncFuncTask, func(task *TaskEntity, context *FlowContext) bool {
					return true
				}},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail": {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				return nil
			},
		},
	}
}

func TestCreateFlowWork(t *testing.T) {
	type args struct {
		bizId      string
		defineName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(args args, a *FlowWorkAggregation) bool
	}{
		{"normal", args{"1111", "testFlow"}, false, []func(args args, a *FlowWorkAggregation) bool{
			func(args args, a *FlowWorkAggregation) bool { return a.FlowWork.Id > 0 },
			func(args args, a *FlowWorkAggregation) bool { return a.Define.FlowName == "testFlow" },
		}},
		{"define not existed", args{"1111", "not existed"}, true, []func(args args, a *FlowWorkAggregation) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateFlowWork(context.TODO(), tt.args.bizId, tt.args.defineName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFlowWork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(tt.args, got) {
					t.Errorf("CreateAccount() assert got false, index = %v, got = %v", i, got)
				}
			}
		})
	}
}

func TestFlowWorkAggregation_Destroy(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		flow, _ := CreateFlowWork(context.TODO(), "111", "testFlow2", nil)
		flow.Start()
		flow.Destroy()
		if !flow.FlowWork.Finished() {
			t.Errorf("Start() finished")
		}

		if flow.FlowWork.Status != TaskStatusError {
			t.Errorf("Start() FlowWork status wrong, want = %v, got = %v", TaskStatusError, flow.FlowWork.Status)
		}

		if flow.CurrentTask.Status != TaskStatusError {
			t.Errorf("Start() CurrentTask status wrong, want = %v, got = %v", TaskStatusError, flow.CurrentTask.Status)
		}

	})
}

func TestFlowWorkAggregation_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTiup := mocksecondparty.NewMockMicroSrv(ctrl)
	mockTiup.EXPECT().MicroSrvGetTaskStatusByBizID(gomock.Any(), gomock.Any()).Return(dbpb.TiupTaskStatus_Finished, "success", nil)
	secondparty.SecondParty = mockTiup

	t.Run("normal", func(t *testing.T) {
		flow, _ := CreateFlowWork(context.TODO(), "1111", "testFlow", nil)
		flow.Start()
		if !flow.FlowWork.Finished() {
			t.Errorf("Start() not finished")
		}

		value, ok := flow.Context.GetData("testKey").(int)
		if !ok || 999 != value {
			t.Errorf("Start() get wrong GetData of testKey, want = 999, got = %v", 999)
		}
	})
	t.Run("callback", func(t *testing.T) {
		flow, _ := CreateFlowWork(context.TODO(), "2222", "testFlow2", nil)
		flow.Start()
		if !flow.FlowWork.Finished() {
			t.Errorf("Start() finished")
		}
		if flow.CurrentTask.TaskName != "doStart" {
			t.Errorf("Start() CurrentTask wrong, want = %v, got = %v", "doStart", flow.CurrentTask.TaskName)
		}
	})
	t.Run("polling", func(t *testing.T) {
		flow, _ := CreateFlowWork(context.TODO(), "4444", "testFlow4", nil)
		flow.Start()

		value, ok := flow.Context.GetData("testKey").(int)
		if !ok || 999 != value {
			t.Errorf("Start() get wrong GetData of testKey, want = 999, got = %v", 999)
		}
	})
	t.Run("error", func(t *testing.T) {
		flow, _ := CreateFlowWork(context.TODO(), "3333", "testFlow3", nil)
		flow.Start()
		if !flow.FlowWork.Finished() {
			t.Errorf("Start() finished")
		}
		if flow.FlowWork.Status != TaskStatusError {
			t.Errorf("Start() FlowWork status wrong, want = %v, got = %v", TaskStatusError, flow.FlowWork.Status)
		}
	})
}

func TestFlowWorkEntity_Finished(t *testing.T) {
	type fields struct {
		Id             uint
		FlowName       string
		StatusAlias    string
		BizId          string
		Status         TaskStatus
		ContextContent string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"init", fields{Status: TaskStatusInit}, false},
		{"processing", fields{Status: TaskStatusProcessing}, false},
		{"finished", fields{Status: TaskStatusFinished}, true},
		{"error", fields{Status: TaskStatusError}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := FlowWorkEntity{
				Id:             tt.fields.Id,
				FlowName:       tt.fields.FlowName,
				StatusAlias:    tt.fields.StatusAlias,
				BizId:          tt.fields.BizId,
				Status:         tt.fields.Status,
				ContextContent: tt.fields.ContextContent,
			}
			if got := c.Finished(); got != tt.want {
				t.Errorf("Finished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDefaultMaintainTask(t *testing.T) {
	tests := []struct {
		name string
		want *CronTaskEntity
	}{
		{"mock", &CronTaskEntity{
			Name:   "maintain",
			Cron:   "0 0 21 ? ? ? ",
			Status: CronStatusValid,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDefaultMaintainTask(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDefaultMaintainTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskEntity_Fail(t1 *testing.T) {
	type fields struct {
		Id             uint
		Status         TaskStatus
		TaskName       string
		TaskReturnType TaskReturnType
		BizId          string
		Parameters     string
		Result         string
	}
	type args struct {
		//e error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TaskEntity{
				Id:             tt.fields.Id,
				Status:         tt.fields.Status,
				TaskName:       tt.fields.TaskName,
				TaskReturnType: tt.fields.TaskReturnType,
				BizId:          tt.fields.BizId,
				Parameters:     tt.fields.Parameters,
				Result:         tt.fields.Result,
			}
			fmt.Println(t)
		})
	}
}

func TestTaskEntity_Processing(t1 *testing.T) {
	type fields struct {
		Id             uint
		Status         TaskStatus
		TaskName       string
		TaskReturnType TaskReturnType
		BizId          string
		Parameters     string
		Result         string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TaskEntity{
				Id:             tt.fields.Id,
				Status:         tt.fields.Status,
				TaskName:       tt.fields.TaskName,
				TaskReturnType: tt.fields.TaskReturnType,
				BizId:          tt.fields.BizId,
				Parameters:     tt.fields.Parameters,
				Result:         tt.fields.Result,
			}
			fmt.Println(t)

		})
	}
}

func TestTaskEntity_Success(t1 *testing.T) {
	type fields struct {
		Id             uint
		Status         TaskStatus
		TaskName       string
		TaskReturnType TaskReturnType
		BizId          string
		Parameters     string
		Result         string
	}
	type args struct {
		//result interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TaskEntity{
				Id:             tt.fields.Id,
				Status:         tt.fields.Status,
				TaskName:       tt.fields.TaskName,
				TaskReturnType: tt.fields.TaskReturnType,
				BizId:          tt.fields.BizId,
				Parameters:     tt.fields.Parameters,
				Result:         tt.fields.Result,
			}
			fmt.Println(t)

		})
	}
}

func TestFlowWorkAggregation_AddContext(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got, _ := CreateFlowWork(context.TODO(), "1111", "testFlow", nil)
		got.AddContext("TestFlowWorkAggregation_AddContext", 123)
		v, ok := got.Context.GetData("TestFlowWorkAggregation_AddContext").(int)
		if !ok || v != 123 {
			t.Errorf("AddContext() get wrong GetData, got = %v, want = %v", v, 123)
		}
	})
}

func TestFlowWorkAggregation_GetAllTaskDef(t *testing.T) {
	flow , err := CreateFlowWork(context.TODO(), "testFlow", "testFlow", &Operator{})
	if err == nil {
		def := flow.GetAllTaskDef()
		assert.True(t, len(def) > 0)
	}
}

func TestFlowWorkAggregation_ExtractTaskDTO(t *testing.T) {
	flow := FlowWorkAggregation {
		Tasks: []*TaskEntity{
			{Id: 1, Status: 1, TaskName: "name1"},
			{Id: 2, Status: 0, TaskName: "name1"},
		},
	}
	tasks := flow.ExtractTaskDTO()
	assert.Equal(t, 2, len(tasks))
	assert.Equal(t, int64(1), tasks[0].Id)

}