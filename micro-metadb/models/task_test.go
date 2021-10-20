
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

package models

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"strconv"
	"testing"
	"time"
)

func TestBatchSaveTasks(t *testing.T) {
	type args struct {
		tasks []*TaskDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, r []*TaskDO) bool
	}{
		{"normal", args{[]*TaskDO{
			{Name: "task1", Data: Data{BizId: "111", Status: 1}},
			{Name: "task2", Data: Data{BizId: "222", Status: 1}},
			{Name: "task3", Data: Data{BizId: "333", Status: 1}},
		}}, false, []func(a args, r []*TaskDO) bool{
			func(a args, r []*TaskDO) bool { return len(a.tasks) == len(r) },
			func(a args, r []*TaskDO) bool { return r[2].ID > 0 },
			func(a args, r []*TaskDO) bool { return r[0].Status == a.tasks[0].Status },
			func(a args, r []*TaskDO) bool { return r[0].Status == 1 },
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewTasks, err := BatchSaveTasks(MetaDB, tt.args.tasks)
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
		name    string
		args    args
		wantErr bool
		wants   []func(args args, flow *FlowDO) bool
	}{
		{"normal", args{flowName: "testFlow", statusAlias: "创建中", bizId: "TestCreateFlow1"}, false, []func(args args, flow *FlowDO) bool{
			func(args args, flow *FlowDO) bool { return flow.Name == args.flowName },
			func(args args, flow *FlowDO) bool { return flow.StatusAlias == args.statusAlias },
			func(args args, flow *FlowDO) bool { return flow.BizId == args.bizId },
			func(args args, flow *FlowDO) bool { return flow.ID > 0 },
			func(args args, flow *FlowDO) bool { return flow.Status == 0 },
		}},
		{"empty biz id", args{flowName: "testFlow", statusAlias: "创建中", bizId: ""}, false, []func(args args, flow *FlowDO) bool{
			func(args args, flow *FlowDO) bool { return flow.ID > 0 },
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlow, err := CreateFlow(MetaDB, tt.args.flowName, tt.args.statusAlias, tt.args.bizId, "111")
			defer MetaDB.Delete(gotFlow)
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
		name    string
		args    args
		wantErr bool
		wants   []func(a args, r *TaskDO) bool
	}{
		{"normal", args{parentType: 1, parentId: "pi1", taskName: "tn1", bizId: "TestCreateTask1", taskReturnType: "trt1", parameters: "p1", result: "r1"}, false, []func(a args, r *TaskDO) bool{
			func(args args, r *TaskDO) bool { return r.ID > 0 },
			func(args args, r *TaskDO) bool { return r.ParentType == args.parentType },
			func(args args, r *TaskDO) bool { return r.ParentId == args.parentId },
			func(args args, r *TaskDO) bool { return r.Name == args.taskName },
			func(args args, r *TaskDO) bool { return r.BizId == args.bizId },
			func(args args, r *TaskDO) bool { return r.ReturnType == args.taskReturnType },
			func(args args, r *TaskDO) bool { return r.Parameters == args.parameters },
			func(args args, r *TaskDO) bool { return r.Result == args.result },
		}},
		{"empty biz id", args{parentType: 1, parentId: "pi1", taskName: "tn1", bizId: "", taskReturnType: "trt1", parameters: "p1", result: "r1"}, false, []func(a args, r *TaskDO) bool{
			func(args args, r *TaskDO) bool { return r.ID > 0 },
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTask, err := CreateTask(MetaDB, tt.args.parentType, tt.args.parentId, tt.args.taskName, tt.args.bizId, tt.args.taskReturnType, tt.args.parameters, tt.args.result)
			defer MetaDB.Delete(gotTask)
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
		flow, _ := CreateFlow(MetaDB, "name", "创建中", "TestFetchFlow1", "111")
		defer MetaDB.Delete(flow)
		got, err := FetchFlow(MetaDB, flow.ID)
		if err != nil {
			t.Errorf("TestFetchFlow() error = %v", err)
		}

		if got.ID != flow.ID ||
			got.BizId != flow.BizId ||
			got.Status != flow.Status ||
			got.CreatedAt.Unix() != flow.CreatedAt.Unix() ||
			got.Name != flow.Name {
			t.Errorf("FetchFlow() got = %v, want %v", got, flow)
		}
	})
}

func TestFetchFlowDetail(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		flow, _ := CreateFlow(MetaDB, "name", "创建中", "TestFetchFlowDetail1", "111")
		defer MetaDB.Delete(flow)

		CreateTask(MetaDB, 0, strconv.Itoa(int(flow.ID)), "task1", "TestFetchFlowDetail1", "trt1", "p1", "r1")
		CreateTask(MetaDB, 0, strconv.Itoa(int(flow.ID)), "task2", "TestFetchFlowDetail2", "trt2", "p2", "r2")
		CreateTask(MetaDB, 0, "22", "task3", "TestFetchFlowDetail3", "trt3", "p3", "r3")
		CreateTask(MetaDB, 1, strconv.Itoa(int(flow.ID)), "task4", "TestFetchFlowDetail4", "trt4", "p4", "r4")

		gotFlow, gotTasks, err := FetchFlowDetail(MetaDB, flow.ID)
		if err != nil {
			t.Errorf("TestFetchFlow() error = %v", err)
		}

		if gotFlow.ID != flow.ID ||
			gotFlow.BizId != flow.BizId ||
			gotFlow.Status != flow.Status ||
			gotFlow.CreatedAt.Unix() != flow.CreatedAt.Unix() ||
			gotFlow.Name != flow.Name {
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
		task, _ := CreateTask(MetaDB, 1, "p1", "tn", "bi", "trt", "p", "r")

		got, err := FetchTask(MetaDB, task.ID)
		if err != nil {
			t.Errorf("TestFetchFlow() error = %v", err)
		}

		if got.ID != task.ID ||
			got.Name != task.Name ||
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
				Name:        tt.fields.FlowName,
				StatusAlias: tt.fields.StatusAlias,
			}
			if err := do.BeforeCreate(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("BeforeCreate() error = %v, wantErr %v", err, tt.wantErr)
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
				Name:        tt.fields.FlowName,
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
		CreateTask(MetaDB, 0, "1", "task1", "TestQueryTask1", "trt1", "p1", "r1")
		CreateTask(MetaDB, 0, "1", "task2", "TestQueryTask2", "trt2", "p2", "r2")
		CreateTask(MetaDB, 0, "1", "task3", "TestQueryTask3", "trt3", "p3", "r3")
		CreateTask(MetaDB, 1, "1", "task4", "TestQueryTask4", "trt4", "p4", "r4")

		gotTasks, err := QueryTask(MetaDB, "TestQueryTask1", "whatever")

		if err != nil {
			t.Errorf("QueryTask() error = %v", err)
			return
		}

		if len(gotTasks) != 1 {
			t.Errorf("QueryTask() gotTasks got %v want 1", len(gotTasks))

		}

		if gotTasks[0].BizId != "TestQueryTask1" ||
			gotTasks[0].Parameters != "p1" {
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
				Data:       tt.fields.Data,
				ParentType: tt.fields.ParentType,
				ParentId:   tt.fields.ParentId,
				Name:       tt.fields.TaskName,
				ReturnType: tt.fields.TaskReturnType,
				Parameters: tt.fields.Parameters,
				Result:     tt.fields.Result,
			}
			if err := do.BeforeCreate(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("BeforeCreate() error = %v, wantErr %v", err, tt.wantErr)
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
				Data:       tt.fields.Data,
				ParentType: tt.fields.ParentType,
				ParentId:   tt.fields.ParentId,
				Name:       tt.fields.TaskName,
				ReturnType: tt.fields.TaskReturnType,
				Parameters: tt.fields.Parameters,
				Result:     tt.fields.Result,
			}
			if got := do.TableName(); got != tt.want {
				t.Errorf("TableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateFlow(t *testing.T) {
	now := time.Now()
	ff, _ := CreateFlow(MetaDB, "test", "", "TestUpdateFlow1", "111")
	defer MetaDB.Delete(ff)

	type args struct {
		flow FlowDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, r FlowDO) bool
	}{
		{"normal", args{flow: *ff}, false, []func(a args, r FlowDO) bool{
			func(a args, r FlowDO) bool { return a.flow.ID == r.ID },
			func(a args, r FlowDO) bool { return r.UpdatedAt.After(now) },
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlow, err := UpdateFlowStatus(MetaDB, tt.args.flow)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateFlowStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, gotFlow) {
					t.Errorf("UpdateFlowStatus() test error, testname = %v, assert = %v, args = %v, gotFlow = %v", tt.name, i, tt.args, gotFlow)
				}
			}
		})
	}
}

//func TestUpdateFlowAndTasks(t *testing.T) {
//	flow, _ := CreateFlow("TestUpdateFlowAndTasks", "TestUpdateFlowAndTasks", "TestUpdateFlowAndTasks")
//	task, _ := CreateTask(0, strconv.Itoa(int(flow.ID)), "taskName", "bizId" , "taskReturnType" , "parameters", "result" )
//
//	type args struct {
//		flow  FlowDO
//		tasks []TaskDO
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//		wants   []func (a args, r FlowDO, tasks []TaskDO, index int) bool
//	}{
//		{"normal", args{flow: *flow, tasks: []TaskDO{*task, {Data: Data{BizId: "bizId2222"}, ParentType: 0, ParentId: strconv.Itoa(int(flow.ID)), Name: "newTaskName"}}}, false, []func (a args, r FlowDO, tasks []TaskDO, index int) bool{
//			func (a args, r FlowDO, tasks []TaskDO, index int) bool {return r.ID == flow.ID},
//			func (a args, r FlowDO, tasks []TaskDO, index int) bool {return index == int(r.Status)},
//			func (a args, r FlowDO, tasks []TaskDO, index int) bool {return 2 == len(tasks)},
//			func (a args, r FlowDO, tasks []TaskDO, index int) bool {return tasks[0].ID == task.ID},
//			func (a args, r FlowDO, tasks []TaskDO, index int) bool {
//				return index == int(tasks[0].Status)
//			},
//			func (a args, r FlowDO, tasks []TaskDO, index int) bool {
//				return tasks[1].ID > task.ID
//			},
//		}},
//		{"no flow", args{flow: FlowDO{Data: Data{ID: 87878787}}, tasks: []TaskDO{}}, true, []func (a args, r FlowDO, tasks []TaskDO, index int) bool{}},
//		{"no flow id", args{flow: FlowDO{Data: Data{}}, tasks: []TaskDO{}}, true, []func (a args, r FlowDO, tasks []TaskDO, index int) bool{}},
//		{"no tasks", args{flow: *flow, tasks: []TaskDO{}}, false, []func (a args, r FlowDO, tasks []TaskDO, index int) bool{
//			func (a args, r FlowDO, tasks []TaskDO, index int) bool {return r.ID == flow.ID},
//			func (a args, r FlowDO, tasks []TaskDO, index int) bool {
//				return index == int(r.Status)
//			},
//			func (a args, r FlowDO, tasks []TaskDO, index int) bool {return 0 == len(tasks)},
//		}},
//	}
//	for index, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			tt.args.flow.Status = int8(index + 1)
//			if len(tt.args.tasks) > 0 {
//				tt.args.tasks[0].Status = int8(index + 1)
//			}
//			gotFlow, gotTasks, err := UpdateFlowAndTasks(tt.args.flow, tt.args.tasks)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("UpdateFlowAndTasks() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			for i, assert := range tt.wants {
//				if !assert(tt.args, gotFlow, gotTasks, index + 1) {
//					t.Errorf("UpdateFlowStatus() test error, assert = %v, args = %v, gotFlow = %v, gotTasks = %v", i, tt.args, gotFlow, gotTasks)
//				}
//			}
//		})
//	}
//}

func TestUpdateTask(t *testing.T) {
	task, _ := CreateTask(MetaDB, 0, "28282828", "taskName", "bizId", "taskReturnType", "parameters", "result")

	type args struct {
		task TaskDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, r TaskDO, index int) bool
	}{
		{"normal", args{task: *task}, false, []func(a args, r TaskDO, index int) bool{
			func(a args, r TaskDO, index int) bool { return int(r.Status) == index },
		}},
	}
	for index, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.task.Status = int8(index + 1)
			gotNewTask, err := UpdateTask(MetaDB, tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, gotNewTask, index+1) {
					t.Errorf("UpdateFlowStatus() test error, testname = %v, assert = %v, args = %v, gotFlow = %v", tt.name, i, tt.args, gotNewTask)
				}
			}
		})
	}
}

func TestBatchFetchFlows(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		flow1, _ := CreateFlow(MetaDB, "flow1", "创建中", "121212", "111")
		flow2, _ := CreateFlow(MetaDB, "flow2", "创建中", "121212", "111")
		defer MetaDB.Delete(flow1)
		defer MetaDB.Delete(flow2)

		got, err := BatchFetchFlows(MetaDB, []uint{0, flow1.ID, 0, flow2.ID, 0})
		if err != nil {
			t.Errorf("TestBatchFetchFlows() error = %v", err)
		}

		if len(got) != 2 {
			t.Errorf("FetchFlowDetail() got = %v, want %v", len(got), 2)
		}

		if got[1].Name != "flow2" {
			t.Errorf("FetchFlowDetail() got = %v, want %v", got[1].Name, "flow2")
		}
	})
}

func TestListFlows(t *testing.T) {
	MetaDB.Model(FlowDO{}).Where("id > 0").Delete(&FlowDO{})
	flow1, _ := CreateFlow(MetaDB, "flow1", "TestListFlows", "121212", "111")
	flow2, _ := CreateFlow(MetaDB, "TestListFlows", "TestListFlows", "9999999", "111")
	defer MetaDB.Delete(flow1)
	defer MetaDB.Delete(flow2)

	t.Run("normal", func(t *testing.T) {
		got, total, err := ListFlows(MetaDB, "", "", -1, 0, 10)
		assert.NoError(t, err)
		assert.True(t, total >= 2)
		fmt.Println(got[0].ID)
		assert.Equal(t, flow1.StatusAlias, got[0].StatusAlias)
		assert.Equal(t, flow2.Name, got[1].StatusAlias)
	})
	t.Run("status", func(t *testing.T) {
		got, _, err := ListFlows(MetaDB, "", "", 2, 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(got))
	})
	t.Run("bizId", func(t *testing.T) {
		got, _, err := ListFlows(MetaDB, "9999999", "", -1, 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, flow2.StatusAlias, got[0].StatusAlias)
		assert.Equal(t, "9999999", got[0].BizId)
	})
	t.Run("keyword", func(t *testing.T) {
		got, _, err := ListFlows(MetaDB, "", "TestList", -1, 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, flow2.StatusAlias, got[0].StatusAlias)
		assert.Equal(t, "9999999", got[0].BizId)
	})
	t.Run("length", func(t *testing.T) {
		_, total, err := ListFlows(MetaDB, "", "", 0, 0, 1)
		assert.NoError(t, err)
		assert.True(t, total > 2)
	})
	t.Run("offset", func(t *testing.T) {
		got, total, err := ListFlows(MetaDB, "", "", 0, 1, 1)
		assert.NoError(t, err)
		assert.True(t, total > 2)
		assert.Equal(t, flow2.StatusAlias, got[0].StatusAlias)
		assert.Equal(t, "9999999", got[0].BizId)
	})
}