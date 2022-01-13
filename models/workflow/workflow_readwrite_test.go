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
	"testing"
	"time"

	"github.com/pingcap-inc/tiem/models/common"
	"github.com/stretchr/testify/assert"
)

var rw *WorkFlowReadWrite

func TestFlowReadWrite_CreateWorkFlow(t *testing.T) {
	flow := &WorkFlow{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "FlowInitStatus",
		},
		Name:  "flowName",
		BizID: "clusterId",
	}
	_, err := rw.CreateWorkFlow(context.TODO(), flow)
	assert.NoError(t, err)
}

func TestFlowReadWrite_GetWorkFlow(t *testing.T) {
	flow := &WorkFlow{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "FlowInitStatus",
		},
		Name:  "flowName",
		BizID: "clusterId",
	}
	flowCreate, errCreate := rw.CreateWorkFlow(context.TODO(), flow)
	assert.NoError(t, errCreate)

	flowGet, errGet := rw.GetWorkFlow(context.TODO(), flowCreate.ID)
	assert.Equal(t, flowCreate.ID, flowGet.ID)
	assert.NoError(t, errGet)

	_, errGet = rw.GetWorkFlow(context.TODO(), "")
	assert.Equal(t, flowCreate.ID, flowGet.ID)
	assert.Error(t, errGet)

	_, errGet = rw.GetWorkFlow(context.TODO(), "sssss")
	assert.Equal(t, flowCreate.ID, flowGet.ID)
	assert.Error(t, errGet)
}

func TestFlowReadWrite_QueryWorkFlows(t *testing.T) {
	flow := &WorkFlow{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "FlowInitStatus",
		},
		Name:  "flowNameQuery",
		BizID: "clusterId",
	}
	flowCreate, errCreate := rw.CreateWorkFlow(context.TODO(), flow)
	assert.NoError(t, errCreate)

	flowQuery, total, errQuery := rw.QueryWorkFlows(context.TODO(), "", "", "flowNameQuery", "", 0, 10)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, flowCreate.ID, flowQuery[0].ID)
	assert.NoError(t, errQuery)

	_, _, _ = rw.QueryWorkFlows(context.TODO(), "fdasfdsa", "fakeBizType", "flowNameQuery", "fdsafdsaf", 0, 10)
	assert.Equal(t, 0, 0)
}

func TestFlowReadWrite_UpdateWorkFlow(t *testing.T) {
	flow := &WorkFlow{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "FlowInitStatus",
		},
		Name:  "flowName",
		BizID: "clusterId",
	}
	flowCreate, errCreate := rw.CreateWorkFlow(context.TODO(), flow)
	assert.NoError(t, errCreate)

	flowCreate.Status = "FlowEndStatus"
	flowCreate.Context = "FlowContext"
	errUpdate := rw.UpdateWorkFlow(context.TODO(), flowCreate.ID, flowCreate.Status, flowCreate.Context)
	assert.NoError(t, errUpdate)

	flowGet, errGet := rw.GetWorkFlow(context.TODO(), flowCreate.ID)
	assert.Equal(t, flowCreate.ID, flowGet.ID)
	assert.Equal(t, flowCreate.Status, flowGet.Status)
	assert.Equal(t, flowCreate.Context, flowGet.Context)
	assert.NoError(t, errGet)

	errUpdate = rw.UpdateWorkFlow(context.TODO(), "", flowCreate.Status, flowCreate.Context)
	assert.Error(t, errUpdate)

	errUpdate = rw.UpdateWorkFlow(context.TODO(), "aaaa", flowCreate.Status, flowCreate.Context)
	assert.Error(t, errUpdate)
}

func TestFlowReadWrite_DetailWorkFlow(t *testing.T) {
	flow := &WorkFlow{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "FlowInitStatus",
		},
		Name:  "flowName",
		BizID: "clusterId",
	}
	flowCreate, errFlowCreate := rw.CreateWorkFlow(context.TODO(), flow)
	assert.NoError(t, errFlowCreate)

	node := &WorkFlowNode{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "NodeInitStatus",
		},
		Parameters: "nodeParam",
		ParentID:   flowCreate.ID,
		Name:       "nodeName",
		ReturnType: "SyncFuncNode",
		Result:     "success",
		StartTime:  time.Now(),
		EndTime:    time.Now(),
	}
	nodeCreate, errNodeCreate := rw.CreateWorkFlowNode(context.TODO(), node)
	assert.NoError(t, errNodeCreate)

	flowDetail, nodeDetail, errDetail := rw.QueryDetailWorkFlow(context.TODO(), flowCreate.ID)
	assert.NoError(t, errDetail)
	assert.Equal(t, flowCreate.ID, flowDetail.ID)
	assert.Equal(t, nodeCreate.ID, nodeDetail[0].ID)

	_, _, errDetail = rw.QueryDetailWorkFlow(context.TODO(), "")
	assert.Error(t, errDetail)

	_, _, errDetail = rw.QueryDetailWorkFlow(context.TODO(), "ssssss")
	assert.Error(t, errDetail)
}

func TestFlowReadWrite_CreateWorkFlowNode(t *testing.T) {
	node := &WorkFlowNode{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "NodeInitStatus",
		},
		Parameters: "nodeParam",
		ParentID:   "flowId",
		Name:       "nodeName",
		ReturnType: "SyncFuncNode",
		Result:     "success",
		StartTime:  time.Now(),
		EndTime:    time.Now(),
	}
	_, err := rw.CreateWorkFlowNode(context.TODO(), node)
	assert.NoError(t, err)
}

func TestFlowReadWrite_UpdateWorkFlowNode(t *testing.T) {
	node := &WorkFlowNode{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "NodeInitStatus",
		},
		Parameters: "nodeParam",
		ParentID:   "flowId",
		Name:       "nodeName",
		ReturnType: "SyncFuncNode",
		Result:     "init",
		StartTime:  time.Now(),
		EndTime:    time.Now(),
	}
	nodeCreate, errCreate := rw.CreateWorkFlowNode(context.TODO(), node)
	assert.NoError(t, errCreate)

	nodeCreate.Status = "NodeEndStatus"
	nodeCreate.Result = "success"
	errUpdate := rw.UpdateWorkFlowNode(context.TODO(), nodeCreate)
	assert.NoError(t, errUpdate)

	nodeQuery, errQuery := rw.GetWorkFlowNode(context.TODO(), nodeCreate.ID)
	assert.NoError(t, errQuery)
	assert.Equal(t, nodeCreate.ID, nodeQuery.ID)
	assert.Equal(t, nodeCreate.Status, nodeQuery.Status)
	assert.Equal(t, nodeCreate.Result, nodeQuery.Result)
}

func TestFlowReadWrite_UpdateWorkFlowDetail(t *testing.T) {
	flow := &WorkFlow{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "FlowInitStatus",
		},
		Name:    "flowName",
		BizID:   "clusterId",
		Context: "flowContext",
	}
	flowCreate, errFlowCreate := rw.CreateWorkFlow(context.TODO(), flow)
	assert.NoError(t, errFlowCreate)

	node := &WorkFlowNode{
		Entity: common.Entity{
			TenantId: "tenantId",
			Status:   "NodeInitStatus",
		},
		Parameters: "nodeParam",
		ParentID:   flowCreate.ID,
		Name:       "nodeName",
		ReturnType: "SyncFuncNode",
		Result:     "init",
		StartTime:  time.Now(),
		EndTime:    time.Now(),
	}
	nodeCreate, errNodeCreate := rw.CreateWorkFlowNode(context.TODO(), node)
	assert.NoError(t, errNodeCreate)

	nodes := make([]*WorkFlowNode, 1)
	nodeCreate.Result = "success"
	nodeCreate.Status = "NodeEndStatus"
	nodes[0] = nodeCreate
	flowCreate.Status = "FlowEndStatus"
	flowCreate.Context = "FlowContextNew"
	errUpdate := rw.UpdateWorkFlowDetail(context.TODO(), flowCreate, nodes)
	assert.NoError(t, errUpdate)

	flowQuery, nodeQuery, errQuery := rw.QueryDetailWorkFlow(context.TODO(), flowCreate.ID)
	assert.NoError(t, errQuery)
	assert.Equal(t, flowCreate.Status, flowQuery.Status)
	assert.Equal(t, flowCreate.Context, flowQuery.Context)
	assert.Equal(t, nodeCreate.Status, nodeQuery[0].Status)
	assert.Equal(t, nodeCreate.Result, nodeQuery[0].Result)

	err := rw.UpdateWorkFlowDetail(context.TODO(), &WorkFlow{Entity: common.Entity{ID: "999"}}, nodes)
	assert.Error(t, err)
}
