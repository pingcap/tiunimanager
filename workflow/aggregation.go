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
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/models"
	dbModel "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/workflow"
	secondpartyModel "github.com/pingcap-inc/tiem/models/workflow/secondparty"
	"time"
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
		return nil, errors.NewEMErrorf(errors.TIEM_FLOW_NOT_FOUND, "empty workflow definition")
	}
	flowData := make(map[string]interface{})

	flow := define.getInstance(ctx, bizId, flowData)
	_, err := models.GetWorkFlowReaderWriter().CreateWorkFlow(ctx, flow.Flow)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create workflow %+v failed %s", flow.Flow, err.Error())
		return nil, err
	}
	return flow, nil
}

func (flow *WorkFlowAggregation) start(ctx context.Context) {
	flow.Flow.Status = constants.WorkFlowStatusProcessing
	start := flow.Define.TaskNodes["start"]
	result := flow.handle(start)
	flow.complete(result)
	err := models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("update workflow detail %+v failed %s", flow, err.Error())
	}
}

func (flow *WorkFlowAggregation) asyncStart(ctx context.Context) {
	go flow.start(ctx)
}

func (flow *WorkFlowAggregation) destroy(ctx context.Context, reason string) {
	flow.Flow.Status = constants.WorkFlowStatusCanceled

	if flow.CurrentNode != nil {
		flow.CurrentNode.Fail(errors.NewError(errors.TIEM_TASK_CANCELED, reason))
	}
	err := models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("update workflow detail %+v failed %s", flow, err.Error())
	}
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

func (flow *WorkFlowAggregation) executeTask(node *workflow.WorkFlowNode, nodeDefine *NodeDefine) (execErr error) {
	defer func() {
		if r := recover(); r != nil {
			framework.LogWithContext(flow.Context).Errorf("recover from workflow %s, node %s", flow.Flow.Name, node.Name)
			execErr = errors.NewEMErrorf(errors.TIEM_PANIC, "%v", r)
			node.Fail(execErr)
		}
	}()

	flow.CurrentNode = node
	flow.Nodes = append(flow.Nodes, node)
	node.Processing()
	data, err := json.Marshal(flow.Context.FlowData)
	if err != nil {
		framework.Log().Warnf("json marshal flow context data failed %s", err.Error())
	}
	flow.Flow.Context = string(data)
	err = models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)
	if err != nil {
		framework.Log().Warnf("update workflow %s detail of bizId %s failed %s", flow.Flow.ID, flow.Flow.BizID, err.Error())
	}

	err = nodeDefine.Executor(node, &flow.Context)
	if err != nil {
		framework.LogWithContext(flow.Context).Infof("workflow %s of bizId %s do node %s failed, %s", flow.Flow.ID, flow.Flow.BizID, node.Name, err.Error())
		node.Fail(err)
	}

	return err
}

func (flow *WorkFlowAggregation) handleTaskError(node *workflow.WorkFlowNode, nodeDefine *NodeDefine) {
	flow.FlowError = fmt.Errorf(node.Result)
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
		Entity: dbModel.Entity{
			TenantId: flow.Flow.TenantId,
			Status:   constants.WorkFlowStatusInitializing,
		},
		Name:       nodeDefine.Name,
		BizID:      flow.Flow.BizID,
		ParentID:   flow.Flow.ID,
		ReturnType: string(nodeDefine.ReturnType),
		StartTime:  time.Now(),
	}

	_, err := models.GetWorkFlowReaderWriter().CreateWorkFlowNode(flow.Context, node)
	if err != nil {
		framework.Log().Warnf("create workflow node, node %s failed %s", node.Name, err.Error())
	}
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
		sequence := int32(0)
		for range ticker.C {
			sequence++
			if sequence > maxPollingSequence {
				node.Fail(errors.Error(errors.TIEM_WORKFLOW_NODE_POLLING_TIME_OUT))
				flow.handleTaskError(node, nodeDefine)
				return false
			}
			framework.LogWithContext(flow.Context).Infof("polling node waiting, sequence %d, nodeId %s, nodeName %s", sequence, node.ID, node.Name)

			resp, err := secondparty.Manager.GetOperationStatusByWorkFlowNodeID(flow.Context, node.ID)
			if err != nil {
				framework.LogWithContext(flow.Context).Error(err)
				node.Fail(fmt.Errorf("call secondparty GetOperationStatusByWorkFlowNodeID %s, failed %s", node.ID, err.Error()))
				flow.handleTaskError(node, nodeDefine)
				return false
			}
			if resp.Status == secondpartyModel.OperationStatus_Error {
				node.Fail(fmt.Errorf("call secondparty GetOperationStatusByWorkFlowNodeID %s, response error %s", node.ID, resp.ErrorStr))
				flow.handleTaskError(node, nodeDefine)
				return false
			}
			if resp.Status == secondpartyModel.OperationStatus_Finished {
				if resp.Result != "" {
					node.Success(resp.Result)
				} else {
					node.Success(nil)
				}

				return flow.handle(flow.Define.TaskNodes[nodeDefine.SuccessEvent])
			}
		}
	}
	return true
}
