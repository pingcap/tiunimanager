/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package workflow

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/models"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockworkflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

var doNodeName1 = func(node *wfModel.WorkFlowNode, context *FlowContext) bool {
	node.Success("success")
	return true
}
var doNodeName2 = func(node *wfModel.WorkFlowNode, context *FlowContext) bool {
	node.Success("success")
	return true
}
var doSuccess = func(node *wfModel.WorkFlowNode, context *FlowContext) bool {
	node.Success("success")
	return true
}
var doFail = func(node *wfModel.WorkFlowNode, context *FlowContext) bool {
	node.Success("success")
	return true
}

func init() {
	models.MockDB()
}

func TestFlowManager_RegisterWorkFlow(t *testing.T) {
	manager := GetWorkFlowService()
	manager.RegisterWorkFlow(context.TODO(), "flowName",
		&WorkFlowDefine{
			FlowName: "flowName",
			TaskNodes: map[string]*NodeDefine{
				"start":         {"nodeName1", "nodeName1Done", "fail", SyncFuncNode, doNodeName1},
				"nodeName1Done": {"nodeName2", "nodeName2Done", "fail", SyncFuncNode, doNodeName2},
				"nodeName2Done": {"end", "", "", SyncFuncNode, doSuccess},
				"fail":          {"fail", "", "", SyncFuncNode, doFail},
			},
		})

	define, err := manager.GetWorkFlowDefine(context.TODO(), "flowName")
	assert.NoError(t, err)
	assert.Equal(t, "flowName", define.FlowName)
	assert.Equal(t, "nodeName1", define.TaskNodes["start"].Name)
	assert.Equal(t, "nodeName2", define.TaskNodes["nodeName1Done"].Name)
	assert.Equal(t, "end", define.TaskNodes["nodeName2Done"].Name)
	assert.Equal(t, "fail", define.TaskNodes["fail"].Name)
}

func TestFlowManager_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowRW := mockworkflow.NewMockReaderWriter(ctrl)
	mockFlowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetWorkFlowReaderWriter(mockFlowRW)

	manager := GetWorkFlowService()
	manager.RegisterWorkFlow(context.TODO(), "flowName",
		&WorkFlowDefine{
			FlowName: "flowName",
			TaskNodes: map[string]*NodeDefine{
				"start":         {"nodeName1", "nodeName1Done", "fail", SyncFuncNode, doNodeName1},
				"nodeName1Done": {"nodeName2", "nodeName2Done", "fail", SyncFuncNode, doNodeName2},
				"nodeName2Done": {"end", "", "", SyncFuncNode, doSuccess},
				"fail":          {"fail", "", "", SyncFuncNode, doFail},
			},
		})

	_, errRegister := manager.GetWorkFlowDefine(context.TODO(), "flowName")
	assert.NoError(t, errRegister)

	flow, errCreate := manager.CreateWorkFlow(context.TODO(), "clusterId", "flowName")
	assert.NoError(t, errCreate)
	errStart := manager.Start(context.TODO(), flow)
	assert.NoError(t, errStart)
}

func TestFlowManager_AddContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowRW := mockworkflow.NewMockReaderWriter(ctrl)
	mockFlowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetWorkFlowReaderWriter(mockFlowRW)

	manager := GetWorkFlowService()
	manager.RegisterWorkFlow(context.TODO(), "flowName",
		&WorkFlowDefine{
			FlowName: "flowName",
			TaskNodes: map[string]*NodeDefine{
				"start":         {"nodeName1", "nodeName1Done", "fail", SyncFuncNode, doNodeName1},
				"nodeName1Done": {"nodeName2", "nodeName2Done", "fail", SyncFuncNode, doNodeName2},
				"nodeName2Done": {"end", "", "", SyncFuncNode, doSuccess},
				"fail":          {"fail", "", "", SyncFuncNode, doFail},
			},
		})

	_, errRegister := manager.GetWorkFlowDefine(context.TODO(), "flowName")
	assert.NoError(t, errRegister)

	flow, errCreate := manager.CreateWorkFlow(context.TODO(), "clusterId", "flowName")
	assert.NoError(t, errCreate)

	manager.AddContext(flow, "key", "value")
	assert.Equal(t, "value", flow.Context.GetData("key"))
}

func TestFlowManager_Destroy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowRW := mockworkflow.NewMockReaderWriter(ctrl)
	mockFlowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetWorkFlowReaderWriter(mockFlowRW)

	manager := GetWorkFlowService()
	manager.RegisterWorkFlow(context.TODO(), "flowName",
		&WorkFlowDefine{
			FlowName: "flowName",
			TaskNodes: map[string]*NodeDefine{
				"start":         {"nodeName1", "nodeName1Done", "fail", SyncFuncNode, doNodeName1},
				"nodeName1Done": {"nodeName2", "nodeName2Done", "fail", SyncFuncNode, doNodeName2},
				"nodeName2Done": {"end", "", "", SyncFuncNode, doSuccess},
				"fail":          {"fail", "", "", SyncFuncNode, doFail},
			},
		})

	_, errRegister := manager.GetWorkFlowDefine(context.TODO(), "flowName")
	assert.NoError(t, errRegister)

	flow, errCreate := manager.CreateWorkFlow(context.TODO(), "clusterId", "flowName")
	assert.NoError(t, errCreate)

	errDestroy := manager.Destroy(context.TODO(), flow, "reason")
	assert.NoError(t, errDestroy)
}

func TestFlowManager_Complete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowRW := mockworkflow.NewMockReaderWriter(ctrl)
	mockFlowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetWorkFlowReaderWriter(mockFlowRW)

	manager := GetWorkFlowService()
	manager.RegisterWorkFlow(context.TODO(), "flowName",
		&WorkFlowDefine{
			FlowName: "flowName",
			TaskNodes: map[string]*NodeDefine{
				"start":         {"nodeName1", "nodeName1Done", "fail", SyncFuncNode, doNodeName1},
				"nodeName1Done": {"nodeName2", "nodeName2Done", "fail", SyncFuncNode, doNodeName2},
				"nodeName2Done": {"end", "", "", SyncFuncNode, doSuccess},
				"fail":          {"fail", "", "", SyncFuncNode, doFail},
			},
		})

	_, errRegister := manager.GetWorkFlowDefine(context.TODO(), "flowName")
	assert.NoError(t, errRegister)

	flow, errCreate := manager.CreateWorkFlow(context.TODO(), "clusterId", "flowName")
	assert.NoError(t, errCreate)

	manager.Complete(context.TODO(), flow, true)
	assert.Equal(t, string(constants.WorkFlowStatusFinished), flow.Flow.Status)
}

func TestFlowManager_ListWorkFlows(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowRW := mockworkflow.NewMockReaderWriter(ctrl)
	mockFlowRW.EXPECT().QueryWorkFlows(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), nil).AnyTimes()
	models.SetWorkFlowReaderWriter(mockFlowRW)

	manager := GetWorkFlowService()
	_, _, err := manager.ListWorkFlows(context.TODO(), "", "", "", 1, 10)
	assert.NoError(t, err)
}

func TestFlowManager_DetailWorkFlow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowRW := mockworkflow.NewMockReaderWriter(ctrl)
	mockFlowRW.EXPECT().QueryDetailWorkFlow(gomock.Any(), gomock.Any()).Return(&wfModel.WorkFlow{
		Name: "flowName",
	}, nil, nil).AnyTimes()
	models.SetWorkFlowReaderWriter(mockFlowRW)

	manager := GetWorkFlowService()
	manager.RegisterWorkFlow(context.TODO(), "flowName",
		&WorkFlowDefine{
			FlowName: "flowName",
			TaskNodes: map[string]*NodeDefine{
				"start":         {"nodeName1", "nodeName1Done", "fail", SyncFuncNode, doNodeName1},
				"nodeName1Done": {"nodeName2", "nodeName2Done", "fail", SyncFuncNode, doNodeName2},
				"nodeName2Done": {"end", "", "", SyncFuncNode, doSuccess},
				"fail":          {"fail", "", "", SyncFuncNode, doFail},
			},
		})
	_, err := manager.DetailWorkFlow(context.TODO(), "")
	assert.NoError(t, err)
}

func TestFlowManager_AsyncStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowRW := mockworkflow.NewMockReaderWriter(ctrl)
	mockFlowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	models.SetWorkFlowReaderWriter(mockFlowRW)

	manager := GetWorkFlowService()
	manager.RegisterWorkFlow(context.TODO(), "flowName",
		&WorkFlowDefine{
			FlowName: "flowName",
			TaskNodes: map[string]*NodeDefine{
				"start":         {"nodeName1", "nodeName1Done", "fail", SyncFuncNode, doNodeName1},
				"nodeName1Done": {"nodeName2", "nodeName2Done", "fail", SyncFuncNode, doNodeName2},
				"nodeName2Done": {"end", "", "", SyncFuncNode, doSuccess},
				"fail":          {"fail", "", "", SyncFuncNode, doFail},
			},
		})

	_, errRegister := manager.GetWorkFlowDefine(context.TODO(), "flowName")
	assert.NoError(t, errRegister)

	flow, errCreate := manager.CreateWorkFlow(context.TODO(), "clusterId", "flowName")
	assert.NoError(t, errCreate)
	errStart := manager.AsyncStart(context.TODO(), flow)
	assert.NoError(t, errStart)
}
