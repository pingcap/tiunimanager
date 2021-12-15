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
	"encoding/json"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/models"
	secondparty2 "github.com/pingcap-inc/tiem/models/workflow/secondparty"
	"github.com/pingcap/errors"

	//"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"

	//"github.com/pingcap-inc/tiem/library/secondparty"
	"time"

	common2 "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/workflow"
)

// WorkFlowAggregation workflow aggregation with workflow definition and nodes
type WorkFlowAggregation struct {
	Flow        *workflow.WorkFlow
	Define      *WorkFlowDefine
	CurrentNode *workflow.WorkFlowNode
	Nodes       []*workflow.WorkFlowNode
	Context     FlowContext
	FlowError   error
}

type FlowContext struct {
	context.Context
	FlowData map[string]interface{}
}

type WorkFlowDetail struct {
	Flow      *structs.WorkFlowInfo
	Nodes     []*structs.WorkFlowNodeInfo
	NodeNames []string
}

func NewFlowContext(ctx context.Context) *FlowContext {
	return &FlowContext{
		ctx,
		map[string]interface{}{},
	}
}

func (c FlowContext) GetData(key string) interface{} {
	return c.FlowData[key]
}

func (c FlowContext) SetData(key string, value interface{}) {
	c.FlowData[key] = value
}

func createFlowWork(ctx context.Context, bizId string, define *WorkFlowDefine) (*WorkFlowAggregation, error) {
	framework.LogWithContext(ctx).Infof("create flowwork %v for bizId %s", define, bizId)
	if define == nil {
		return nil, framework.SimpleError(common.TIEM_FLOW_NOT_FOUND)
	}
	flowData := make(map[string]interface{})

	flow := define.getInstance(ctx, bizId, flowData)
	_, err := models.GetWorkFlowReaderWriter().CreateWorkFlow(ctx, flow.Flow)
	if err != nil {
		return nil, err
	}
	//TaskRepo.AddFlowWork(ctx, flow.FlowWork)
	return flow, nil
}

func (flow *WorkFlowAggregation) start() {
	flow.Flow.Status = constants.WorkFlowStatusProcessing
	start := flow.Define.TaskNodes["start"]
	result := flow.handle(start)
	flow.complete(result)
	_ = models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)
	//TaskRepo.Persist(flow.Context, flow)
}

func (flow *WorkFlowAggregation) asyncStart() {
	go flow.start()
}

func (flow *WorkFlowAggregation) destroy(reason string) {
	flow.Flow.Status = constants.WorkFlowStatusCanceled

	if flow.CurrentNode != nil {
		flow.CurrentNode.Fail(framework.NewTiEMError(common.TIEM_TASK_CANCELED, reason))
	}
	_ = models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)
	//TaskRepo.Persist(flow.Context, flow)
}

func (flow WorkFlowAggregation) complete(success bool) {
	if success {
		flow.Flow.Status = constants.WorkFlowStatusFinished
	} else {
		flow.Flow.Status = constants.WorkFlowStatusError
	}
}

func (flow *WorkFlowAggregation) addContext(key string, value interface{}) {
	flow.Context.SetData(key, value)
	data, err := json.Marshal(flow.Context.FlowData)
	if err != nil {
		framework.Log().Warnf("json marshal flow context data failed %s", err.Error())
		return
	}
	flow.Flow.Context = string(data)
}

func (flow *WorkFlowAggregation) executeTask(node *workflow.WorkFlowNode, nodeDefine *NodeDefine) error {
	flow.CurrentNode = node
	flow.Nodes = append(flow.Nodes, node)
	node.Processing()
	_ = models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)
	//TaskRepo.Persist(flow.Context, flow)
	err := nodeDefine.Executor(node, &flow.Context)
	if err != nil {
		node.Fail(err)
	}

	return err
}

func (flow *WorkFlowAggregation) handleTaskError(node *workflow.WorkFlowNode, nodeDefine *NodeDefine) {
	flow.FlowError = errors.New(node.Result)
	if "" != nodeDefine.FailEvent {
		flow.handle(flow.Define.TaskNodes[nodeDefine.FailEvent])
	} else {
		framework.Log().Warnf("no fail event in flow definition, flowname %s", nodeDefine.Name)
	}
}

func (flow *WorkFlowAggregation) handle(nodeDefine *NodeDefine) bool {
	if nodeDefine == nil {
		flow.Flow.Status = constants.WorkFlowStatusFinished
		return true
	}
	node := &workflow.WorkFlowNode{
		Entity: common2.Entity{
			Status: constants.WorkFlowStatusInitializing,
		},
		Name:       nodeDefine.Name,
		BizID:      flow.Flow.BizID,
		ParentID:   flow.Flow.ID,
		ReturnType: string(nodeDefine.ReturnType),
		StartTime:  time.Now(),
	}

	_, _ = models.GetWorkFlowReaderWriter().CreateWorkFlowNode(flow.Context, node)
	//TaskRepo.AddFlowTask(flow.Context, task, flow.FlowWork.ID)
	handleError := flow.executeTask(node, nodeDefine)
	if handleError != nil {
		flow.handleTaskError(node, nodeDefine)
		return false
	}

	switch nodeDefine.ReturnType {
	case SyncFuncNode:
		if node.Result == "" {
			node.Success()
		}
		return flow.handle(flow.Define.TaskNodes[nodeDefine.SuccessEvent])
	case PollingNode:
		ticker := time.NewTicker(3 * time.Second)
		for range ticker.C {
			framework.LogWithContext(flow.Context).Infof("polling node waiting, nodeId %s, nodeName %s", node.ID, node.Name)

			resp, err := secondparty.Manager.GetOperationStatusByWorkFlowNodeID(flow.Context, node.ID)
			if err != nil {
				framework.LogWithContext(flow.Context).Error(err)
				node.Fail(framework.WrapError(common.TIEM_TASK_FAILED, common.TIEM_TASK_FAILED.Explain(), err))
				flow.handleTaskError(node, nodeDefine)
				return false
			}
			if resp.Status == secondparty2.OperationStatus_Error {
				node.Fail(framework.NewTiEMError(common.TIEM_TASK_FAILED, resp.ErrorStr))
				flow.handleTaskError(node, nodeDefine)
				return false
			}
			if resp.Status == secondparty2.OperationStatus_Finished {
				node.Success(resp.Result)
				return flow.handle(flow.Define.TaskNodes[nodeDefine.SuccessEvent])
			}
		}
	}
	return true
}
