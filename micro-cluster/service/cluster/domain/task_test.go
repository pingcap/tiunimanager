package domain

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func testStart(task *TaskEntity, context *FlowContext) bool {
	context.put("testKey", 999)
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
				"start":        {"doStart", "startDone", "fail", SyncFuncTask, testStart},
				"startDone": {"doSomething", "done", "fail", SyncFuncTask, testDoing},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail":        {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				c := make(FlowContext)
				return &c
			},
		},
		"testFlow2": {
			FlowName:    "testFlow2",
			StatusAlias: "test2",
			TaskNodes: map[string]*TaskDefine{
				"start":        {"doStart", "startDone", "fail", CallbackTask, testStart},
				"startDone": {"doSomething", "done", "fail", SyncFuncTask, testDoing},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail":        {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				c := make(FlowContext)
				return &c
			},
		},
		"testFlow3": {
			FlowName:    "testFlow3",
			StatusAlias: "test3",
			TaskNodes: map[string]*TaskDefine{
				"start":        {"doStart", "startDone", "fail", SyncFuncTask, testError},
				"startDone": {"doSomething", "done", "fail", SyncFuncTask, testDoing},
				"fail":        {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				c := make(FlowContext)
				return &c
			},
		},
		"testFlow4": {
			FlowName:    "testFlow4",
			StatusAlias: "test4",
			TaskNodes: map[string]*TaskDefine{
				"start":        {"doStart", "startDone", "fail", PollingTasK, testStart},
				"startDone": {"doSomething", "done", "fail", SyncFuncTask, testDoing},
				"done": {"end", "", "", SyncFuncTask, DefaultEnd},
				"fail":        {"fail", "", "", SyncFuncTask, DefaultFail},
			},
			ContextParser: func(s string) *FlowContext {
				// todo parse context
				c := make(FlowContext)
				return &c
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
			func(args args, a *FlowWorkAggregation) bool{return a.FlowWork.Id > 0},
			func(args args, a *FlowWorkAggregation) bool{return a.Define.FlowName == "testFlow"},
		}},
		{"define not existed", args{"1111", "not existed"}, true, []func(args args, a *FlowWorkAggregation) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateFlowWork(tt.args.bizId, tt.args.defineName)
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
		flow, _ := CreateFlowWork("111", "testFlow2")
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
	t.Run("normal", func(t *testing.T) {
		flow, _ := CreateFlowWork("1111", "testFlow")
		flow.Start()
		if !flow.FlowWork.Finished() {
			t.Errorf("Start() not finished")
		}

		value, ok := flow.Context.value("testKey").(int)
		if !ok || 999 != value {
			t.Errorf("Start() get wrong value of testKey, want = 999, got = %v", 999)
		}
	})
	t.Run("callback", func(t *testing.T) {
		flow, _ := CreateFlowWork("2222", "testFlow2")
		flow.Start()
		if flow.FlowWork.Finished() {
			t.Errorf("Start() finished")
		}
		if flow.CurrentTask.TaskName != "doStart" {
			t.Errorf("Start() CurrentTask wrong, want = %v, got = %v", "doStart", flow.CurrentTask.TaskName)
		}
	})
	t.Run("polling", func(t *testing.T) {
		flow, _ := CreateFlowWork("4444", "testFlow4")
		flow.Start()
		if !flow.FlowWork.Finished() {
			t.Errorf("Start() not finished")
		}

		value, ok := flow.Context.value("testKey").(int)
		if !ok || 999 != value {
			t.Errorf("Start() get wrong value of testKey, want = 999, got = %v", 999)
		}
	})
	t.Run("error", func(t *testing.T) {
		flow, _ := CreateFlowWork("3333", "testFlow3")
		flow.Start()
		if !flow.FlowWork.Finished() {
			t.Errorf("Start() finished")
		}
		if flow.FlowWork.Status != TaskStatusFinished {
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
			Name: "maintain",
			Cron: "0 0 21 ? ? ? ",
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
		e error
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
		result interface{}
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
		got, _ := CreateFlowWork("1111", "testFlow")
		got.AddContext("TestFlowWorkAggregation_AddContext", 123)
		v, ok := got.Context.value("TestFlowWorkAggregation_AddContext").(int)
		if !ok || v != 123 {
			t.Errorf("AddContext() get wrong value, got = %v, want = %v", v, 123)
		}
	})
}