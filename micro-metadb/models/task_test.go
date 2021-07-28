package models

import (
	"gorm.io/gorm"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestBatchSaveTasks(t *testing.T) {
	type args struct {
		tasks []*TaskDO
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wants    []func (a args, r []*TaskDO) bool
	}{
		{"normal", args{[]*TaskDO{
			{TaskName: "task1", Data: Data{BizId: "111"}},
			{TaskName: "task2", Data: Data{BizId: "222"}},
			{TaskName: "task3", Data: Data{BizId: "333"}},
		}}, false, []func (a args, r []*TaskDO) bool{
			func (a args, r []*TaskDO) bool {return len(a.tasks) == len(r)},
			func (a args, r []*TaskDO) bool {return r[2].ID > 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewTasks, err := BatchSaveTasks(tt.args.tasks)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchSaveTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, gotNewTasks) {
					t.Errorf("BatchSaveTasks() test error, testname = %v, assert = %v, args = %v, gotNewTasks = %v", tt.name, i, tt.args, gotNewTasks)
				}
			}
		})
	}
}

func TestCreateFlow(t *testing.T) {
	type args struct {
		flowName    string
		statusAlias string
		bizId       string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wants    []func (args args, flow *FlowDO) bool
	}{
		{"normal", args{flowName: "testFlow", statusAlias: "创建中", bizId: "TestCreateFlow1"}, false, []func (args args, flow *FlowDO) bool{
			func (args args, flow *FlowDO) bool {return flow.FlowName == args.flowName},
			func (args args, flow *FlowDO) bool {return flow.StatusAlias == args.statusAlias},
			func (args args, flow *FlowDO) bool {return flow.BizId == args.bizId},
			func (args args, flow *FlowDO) bool {return flow.ID > 0},
			func (args args, flow *FlowDO) bool {return flow.Status == 0},
		}},
		{"empty biz id", args{flowName: "testFlow", statusAlias: "创建中", bizId: ""}, true, []func (args args, flow *FlowDO) bool{
			func (args args, flow *FlowDO) bool {return flow.ID == 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlow, err := CreateFlow(tt.args.flowName, tt.args.statusAlias, tt.args.bizId)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, gotFlow) {
					t.Errorf("CreateFlow() test error, testname = %v, assert = %v, args = %v, gotFlow = %v", tt.name, i, tt.args, gotFlow)
				}
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	type args struct {
		parentType     int8
		parentId       string
		taskName       string
		bizId          string
		taskReturnType string
		parameters     string
		result         string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wants    []func (a args, r *TaskDO) bool
	}{
		{"normal", args{parentType: 1, parentId: "pi1",taskName: "tn1", bizId: "TestCreateTask1", taskReturnType: "trt1", parameters: "p1", result: "r1"}, false, []func (a args, r *TaskDO) bool{
			func (args args, r *TaskDO) bool {return r.ID > 0},
			func (args args, r *TaskDO) bool {return r.ParentType == args.parentType},
			func (args args, r *TaskDO) bool {return r.ParentId == args.parentId},
			func (args args, r *TaskDO) bool {return r.TaskName == args.taskName},
			func (args args, r *TaskDO) bool {return r.BizId == args.bizId},
			func (args args, r *TaskDO) bool {return r.TaskReturnType == args.taskReturnType},
			func (args args, r *TaskDO) bool {return r.Parameters == args.parameters},
			func (args args, r *TaskDO) bool {return r.Result == args.result},
		}},
		{"empty biz id", args{parentType: 1, parentId: "pi1",taskName: "tn1", bizId: "", taskReturnType: "trt1", parameters: "p1", result: "r1"}, true, []func (a args, r *TaskDO) bool{
			func (args args, r *TaskDO) bool {return r.ID == 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTask, err := CreateTask(tt.args.parentType, tt.args.parentId, tt.args.taskName, tt.args.bizId, tt.args.taskReturnType, tt.args.parameters, tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, gotTask) {
					t.Errorf("CreateFlow() test error, testname = %v, assert = %v, args = %v, gotFlow = %v", tt.name, i, tt.args, gotTask)
				}
			}
		})
	}
}

func TestFetchFlow(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		flow, _ := CreateFlow("name", "创建中", "TestFetchFlow1")
		got, err := FetchFlow(flow.ID)
		if err != nil {
			t.Errorf("TestFetchFlow() error = %v", err)
		}

		if got.ID != flow.ID ||
			got.BizId != flow.BizId ||
			got.Status != flow.Status ||
			got.CreatedAt.Unix() != flow.CreatedAt.Unix() ||
			got.FlowName != flow.FlowName {
			t.Errorf("FetchFlow() got = %v, want %v", got, flow)
		}
	})
}

func TestFetchFlowDetail(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		flow, _ := CreateFlow("name", "创建中", "TestFetchFlowDetail1")
		CreateTask(0, strconv.Itoa(int(flow.ID)), "task1", "TestFetchFlowDetail1", "trt1", "p1", "r1")
		CreateTask(0, strconv.Itoa(int(flow.ID)), "task2", "TestFetchFlowDetail2", "trt2", "p2", "r2")
		CreateTask(0, "22", "task3", "TestFetchFlowDetail3", "trt3", "p3", "r3")
		CreateTask(1, strconv.Itoa(int(flow.ID)), "task4", "TestFetchFlowDetail4", "trt4", "p4", "r4")

		gotFlow, gotTasks, err := FetchFlowDetail(flow.ID)
		if err != nil {
			t.Errorf("TestFetchFlow() error = %v", err)
		}

		if gotFlow.ID != flow.ID ||
			gotFlow.BizId != flow.BizId ||
			gotFlow.Status != flow.Status ||
			gotFlow.CreatedAt.Unix() != flow.CreatedAt.Unix() ||
			gotFlow.FlowName != flow.FlowName {
			t.Errorf("FetchFlowDetail() gotFlow = %v, want %v", gotFlow, flow)
		}

		if len(gotTasks) != 2 {
			t.Errorf("FetchFlowDetail() gotTasks = %v", gotTasks)
		}

		for _, task := range gotTasks {
			if task.ParentType != 0 || task.ParentId != strconv.Itoa(int(flow.ID)) {
				t.Errorf("FetchFlowDetail() gotTasks = %v", gotTasks)
			}
		}
	})
}

func TestFetchTask(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		task, _ := CreateTask(1, "p1", "tn", "bi", "trt", "p", "r")

		got, err := FetchTask(task.ID)
		if err != nil {
			t.Errorf("TestFetchFlow() error = %v", err)
		}

		if got.ID != task.ID ||
			got.TaskName != task.TaskName ||
			got.Status != task.Status ||
			got.CreatedAt.Unix() != task.CreatedAt.Unix() ||
			got.Result != task.Result {
			t.Errorf("FetchFlowDetail() got = %v, want %v", got, task)
		}
	})
}

func TestFlowDO_BeforeCreate(t *testing.T) {
	type fields struct {
		Data        Data
		FlowName    string
		StatusAlias string
	}
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"normal", fields{Data: Data{Status: 9}}, args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do := &FlowDO{
				Data:        tt.fields.Data,
				FlowName:    tt.fields.FlowName,
				StatusAlias: tt.fields.StatusAlias,
			}
			if err := do.BeforeCreate(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("BeforeCreate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if do.Status != 0 {
				t.Errorf("BeforeCreate() error, status %v", do.Status)
			}
		})
	}
}

func TestFlowDO_TableName(t *testing.T) {
	type fields struct {
		Data        Data
		FlowName    string
		StatusAlias string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"normal", fields{}, "flows"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do := FlowDO{
				Data:        tt.fields.Data,
				FlowName:    tt.fields.FlowName,
				StatusAlias: tt.fields.StatusAlias,
			}
			if got := do.TableName(); got != tt.want {
				t.Errorf("TableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryTask(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		CreateTask(0, "1", "task1", "TestQueryTask1", "trt1", "p1", "r1")
		CreateTask(0, "1", "task2", "TestQueryTask2", "trt2", "p2", "r2")
		CreateTask(0, "1", "task3", "TestQueryTask3", "trt3", "p3", "r3")
		CreateTask(1, "1", "task4", "TestQueryTask4", "trt4", "p4", "r4")

		gotTasks, err := QueryTask("TestQueryTask1", "whatever")

		if err != nil {
			t.Errorf("QueryTask() error = %v", err)
			return
		}

		if len(gotTasks) != 1 {
			t.Errorf("QueryTask() gotTasks got %v want 1", len(gotTasks))

		}

		if gotTasks[0].BizId != "TestQueryTask1" ||
			gotTasks[0].Parameters != "p1"{
			t.Errorf("QueryTask() gotTasks %v", gotTasks)
		}
	})
}

func TestTaskDO_BeforeCreate(t *testing.T) {
	type fields struct {
		Data           Data
		ParentType     int8
		ParentId       string
		TaskName       string
		TaskReturnType string
		Parameters     string
		Result         string
	}
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"normal", fields{Data: Data{Status: 9}}, args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do := &TaskDO{
				Data:           tt.fields.Data,
				ParentType:     tt.fields.ParentType,
				ParentId:       tt.fields.ParentId,
				TaskName:       tt.fields.TaskName,
				TaskReturnType: tt.fields.TaskReturnType,
				Parameters:     tt.fields.Parameters,
				Result:         tt.fields.Result,
			}
			if err := do.BeforeCreate(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("BeforeCreate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if do.Status != 0 {
				t.Errorf("BeforeCreate() error, status %v", do.Status)
			}
		})
	}
}

func TestTaskDO_TableName(t *testing.T) {
	type fields struct {
		Data           Data
		ParentType     int8
		ParentId       string
		TaskName       string
		TaskReturnType string
		Parameters     string
		Result         string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"normal", fields{}, "tasks"},
	}
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do := TaskDO{
				Data:           tt.fields.Data,
				ParentType:     tt.fields.ParentType,
				ParentId:       tt.fields.ParentId,
				TaskName:       tt.fields.TaskName,
				TaskReturnType: tt.fields.TaskReturnType,
				Parameters:     tt.fields.Parameters,
				Result:         tt.fields.Result,
			}
			if got := do.TableName(); got != tt.want {
				t.Errorf("TableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateFlow(t *testing.T) {
	now := time.Now()
	ff, _ := CreateFlow("test", "", "TestUpdateFlow1")
	type args struct {
		flow FlowDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func (a args, r FlowDO) bool
	}{
		{"normal", args{flow: *ff}, false, []func (a args, r FlowDO) bool{
			func (a args, r FlowDO) bool{return a.flow.ID == r.ID},
			func (a args, r FlowDO) bool{return r.UpdatedAt.After(now)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlow, err := UpdateFlow(tt.args.flow)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, gotFlow) {
					t.Errorf("UpdateFlow() test error, testname = %v, assert = %v, args = %v, gotFlow = %v", tt.name, i, tt.args, gotFlow)
				}
			}
		})
	}
}

func TestUpdateFlowAndTasks(t *testing.T) {
	type args struct {
		flow  FlowDO
		tasks []TaskDO
	}
	tests := []struct {
		name    string
		args    args
		want    FlowDO
		want1   []TaskDO
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := UpdateFlowAndTasks(tt.args.flow, tt.args.tasks)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateFlowAndTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateFlowAndTasks() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("UpdateFlowAndTasks() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	now := time.Now()
	type args struct {
		task TaskDO
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wants   	[]func (a args, r TaskDO) bool

	}{
		{"normal", args{task: TaskDO{Data: Data{BizId: "TestUpdateTask1"}}}, false, []func (a args, r TaskDO) bool{
			func (a args, r TaskDO) bool{return r.UpdatedAt.After(now)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewTask, err := UpdateTask(tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, gotNewTask) {
					t.Errorf("UpdateFlow() test error, testname = %v, assert = %v, args = %v, gotFlow = %v", tt.name, i, tt.args, gotNewTask)
				}
			}
		})
	}
}

func TestBatchFetchFlows(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		flow1, _ := CreateFlow("flow1", "创建中", "121212")
		flow2, _ := CreateFlow("flow2", "创建中", "121212")

		got, err := BatchFetchFlows([]uint{0, flow1.ID, 0, flow2.ID, 0})
		if err != nil {
			t.Errorf("TestBatchFetchFlows() error = %v", err)
		}

		if len(got) != 2 {
			t.Errorf("FetchFlowDetail() got = %v, want %v", len(got), 2)
		}

		if got[1].FlowName != "flow2" {
			t.Errorf("FetchFlowDetail() got = %v, want %v", got[1].FlowName, "flow2")
		}
	})
}