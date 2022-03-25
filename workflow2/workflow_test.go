/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package workflow2

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/common"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockworkflow"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var doNodeName1 = func(node *wfModel.WorkFlowNode, context *FlowContext) error {
	fmt.Println("doNodeName1")
	return nil
}
var doNodeName2 = func(node *wfModel.WorkFlowNode, context *FlowContext) error {
	fmt.Println("doNodeName2")
	return nil
}
var doSuccess = func(node *wfModel.WorkFlowNode, context *FlowContext) error {
	fmt.Println("doSuccess")
	return nil
}
var doFail = func(node *wfModel.WorkFlowNode, context *FlowContext) error {
	fmt.Println("doFail")
	return nil
}
var defaultSuccess = func(node *wfModel.WorkFlowNode, context *FlowContext) error {
	return nil
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
				"fail":          {"end", "", "", SyncFuncNode, doFail},
			},
		})

	define, err := manager.GetWorkFlowDefine(context.TODO(), "flowName")
	assert.NoError(t, err)
	assert.Equal(t, "flowName", define.FlowName)
	assert.Equal(t, "nodeName1", define.TaskNodes["start"].Name)
	assert.Equal(t, "nodeName2", define.TaskNodes["nodeName1Done"].Name)
	assert.Equal(t, "end", define.TaskNodes["nodeName2Done"].Name)
	assert.Equal(t, "end", define.TaskNodes["fail"].Name)
}

func TestFlowManager_Start_case1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowRW := mockworkflow.NewMockReaderWriter(ctrl)
	mockFlowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(&wfModel.WorkFlow{
		Entity: common.Entity{
			Status:   constants.WorkFlowStatusInitializing,
			TenantId: framework.GetTenantIDFromContext(context.TODO()),
			ID:       "testflowId",
		},
	}, nil).AnyTimes()
	mockFlowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockFlowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFlowRW.EXPECT().UpdateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFlowRW.EXPECT().GetWorkFlow(gomock.Any(), gomock.Any()).Return(&wfModel.WorkFlow{
		Entity: common.Entity{
			Status:   constants.WorkFlowStatusInitializing,
			TenantId: framework.GetTenantIDFromContext(context.TODO()),
			ID:       "testflowId",
		},
	}, nil).AnyTimes()
	models.SetWorkFlowReaderWriter(mockFlowRW)

	manager := GetWorkFlowService()
	manager.RegisterWorkFlow(context.TODO(), "flowName",
		&WorkFlowDefine{
			FlowName: "flowName",
			TaskNodes: map[string]*NodeDefine{
				"start":         {"nodeName1", "nodeName1Done", "fail", SyncFuncNode, doNodeName1},
				"nodeName1Done": {"nodeName2", "nodeName2Done", "fail", SyncFuncNode, doNodeName2},
				"nodeName2Done": {"end", "", "", SyncFuncNode, CompositeExecutor(doFail, defaultSuccess)},
				"fail":          {"end", "", "", SyncFuncNode, doFail},
			},
		})

	_, errRegister := manager.GetWorkFlowDefine(context.TODO(), "flowName")
	assert.NoError(t, errRegister)

	flow, errCreate := manager.CreateWorkFlow(context.TODO(), "clusterId", BizTypeCluster, "flowName")
	assert.NoError(t, errCreate)
	errStart := manager.Start(context.TODO(), flow)
	assert.NoError(t, errStart)

	time.Sleep(10 * time.Second)
}
