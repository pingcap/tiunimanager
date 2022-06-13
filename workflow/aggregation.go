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

/*
import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/deployment"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/metrics"
	"github.com/pingcap/tiunimanager/models"
	dbModel "github.com/pingcap/tiunimanager/models/common"
	"github.com/pingcap/tiunimanager/models/workflow"
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

func createFlowWork(ctx context.Context, bizId string, bizType string, define *WorkFlowDefine) (*WorkFlowAggregation, error) {
	framework.LogWithContext(ctx).Infof("create flowwork %v for bizId %s", define, bizId)
	if define == nil {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_FLOW_NOT_FOUND, "empty workflow definition")
	}
	flowData := make(map[string]interface{})

	flow := define.getInstance(ctx, bizId, bizType, flowData)
	handleWorkFlowMetrics(flow)

	_, err := models.GetWorkFlowReaderWriter().CreateWorkFlow(ctx, flow.Flow)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create workflow %+v failed %s", flow.Flow, err.Error())
		return nil, err
	}
	return flow, nil
}

func handleWorkFlowMetrics(flow *WorkFlowAggregation) {
	metrics.HandleWorkFlowMetrics(metrics.WorkFlowLabel{
		BizType: flow.Flow.BizType,
		Name:    flow.Flow.Name,
		Status:  flow.Flow.Status,
	})
}

func handleWorkFlowNodeMetrics(flow *WorkFlowAggregation, node *workflow.WorkFlowNode) {
	metrics.HandleWorkFlowNodeMetrics(metrics.WorkFlowNodeLabel{
		BizType:  flow.Flow.BizType,
		FlowName: flow.Flow.Name,
		Node:     node.Name,
		Status:   node.Status,
	})
}

func (flow *WorkFlowAggregation) start(ctx context.Context) {
	flow.Flow.Status = constants.WorkFlowStatusProcessing
	handleWorkFlowMetrics(flow)

	start := flow.Define.TaskNodes["start"]
	result := flow.handle(start)
	flow.complete(result)
	err := models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)

	if err != nil {
		framework.LogWithContext(ctx).Warnf("update workflow detail %+v failed %s", flow, err.Error())
	}
}

func (flow *WorkFlowAggregation) asyncStart(ctx context.Context) {
	//operationName: em.cluster.ClusterService.Login workflow.11121231
	operationName := fmt.Sprintf(
		"%s.%s workflow.%s",
		framework.GetMicroServiceNameFromContext(ctx),
		framework.GetMicroEndpointNameFromContext(ctx),
		flow.Flow.ID,
	)
	t := framework.NewBackgroundTask(
		ctx, operationName,
		func(ctx context.Context) error {
			flow.start(ctx)
			return nil
		},
	)
	flow.Context.Context = t.GetContext()
	go t.Exec()
}

func (flow *WorkFlowAggregation) destroy(ctx context.Context, reason string) {
	flow.Flow.Status = constants.WorkFlowStatusCanceled
	handleWorkFlowMetrics(flow)

	if flow.CurrentNode != nil {
		nodeDefine := flow.Define.TaskNodes[flow.CurrentNode.Name]
		err := errors.NewError(errors.TIUNIMANAGER_TASK_CANCELED, reason)
		flow.handleTaskError(flow.CurrentNode, nodeDefine, err)
	}
	err := models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("update workflow detail %+v failed %s", flow, err.Error())
	}
}

func (flow *WorkFlowAggregation) complete(success bool) {
	if success {
		flow.Flow.Status = constants.WorkFlowStatusFinished
	} else {
		flow.Flow.Status = constants.WorkFlowStatusError
	}

	handleWorkFlowMetrics(flow)
}

func (flow *WorkFlowAggregation) addContext(key string, value interface{}) {
	flow.Context.SetData(key, value)
	data, err := json.Marshal(flow.Context.FlowData)
	if err != nil {
		framework.LogWithContext(flow.Context).Warnf("json marshal flow context data failed %s", err.Error())
		return
	}
	flow.Flow.Context = string(data)
}

func (flow *WorkFlowAggregation) createTask(nodeDefine *NodeDefine) *workflow.WorkFlowNode {
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

	handleWorkFlowNodeMetrics(flow, node)

	return node
}

func (flow *WorkFlowAggregation) executeTask(node *workflow.WorkFlowNode, nodeDefine *NodeDefine) (execErr errors.EMError) {
	defer func() {
		if r := recover(); r != nil {
			framework.LogWithContext(flow.Context).Errorf(
				"recover from workflow %s, node %s, stacktrace %s",
				flow.Flow.Name, node.Name, string(debug.Stack()))
			execErr = errors.NewErrorf(errors.TIUNIMANAGER_PANIC, "%v", r)
		}
	}()

	node.Processing()
	handleWorkFlowNodeMetrics(flow, node)

	flow.CurrentNode = node
	flow.Nodes = append(flow.Nodes, node)
	data, err := json.Marshal(flow.Context.FlowData)
	if err != nil {
		framework.LogWithContext(flow.Context).Errorf("json marshal flow context data failed %s", err.Error())
		return errors.NewErrorf(errors.TIUNIMANAGER_MARSHAL_ERROR, "%v", err)
	}
	flow.Flow.Context = string(data)
	err = models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)
	if err != nil {
		framework.LogWithContext(flow.Context).Warnf("update workflow %s detail of bizId %s failed %s", flow.Flow.ID, flow.Flow.BizID, err.Error())
	}

	err = nodeDefine.Executor(node, &flow.Context)
	if err != nil {
		framework.LogWithContext(flow.Context).Errorf("workflow %s of bizId %s do node %s failed, %s", flow.Flow.ID, flow.Flow.BizID, node.Name, err.Error())
		if emErr, ok := err.(errors.EMError); ok {
			return emErr
		}
		execErr = errors.NewErrorf(errors.TIUNIMANAGER_UNRECOGNIZED_ERROR, "%v", err)
		return
	}

	return
}

func (flow *WorkFlowAggregation) handleTaskError(node *workflow.WorkFlowNode, nodeDefine *NodeDefine, err errors.EMError) {
	node.Fail(err)
	handleWorkFlowNodeMetrics(flow, node)
	flow.FlowError = fmt.Errorf(node.Result)

	if "" != nodeDefine.FailEvent {
		flow.handle(flow.Define.TaskNodes[nodeDefine.FailEvent])
	} else {
		framework.LogWithContext(flow.Context).Warnf("no fail event in flow definition, flowname %s", nodeDefine.Name)
	}
}

func (flow *WorkFlowAggregation) handleTaskSuccess(node *workflow.WorkFlowNode, result ...interface{}) {
	node.Success(result...)
	handleWorkFlowNodeMetrics(flow, node)
}

func (flow *WorkFlowAggregation) handle(nodeDefine *NodeDefine) bool {
	if nodeDefine == nil {
		flow.Flow.Status = constants.WorkFlowStatusFinished
		return true
	}

	node := flow.createTask(nodeDefine)
	_, err := models.GetWorkFlowReaderWriter().CreateWorkFlowNode(flow.Context, node)
	if err != nil {
		framework.LogWithContext(flow.Context).Warnf("create workflow node, node %s failed %s", node.Name, err.Error())
	}

	handleError := flow.executeTask(node, nodeDefine)
	if handleError.GetCode() != errors.TIUNIMANAGER_SUCCESS {
		flow.handleTaskError(node, nodeDefine, handleError)
		return false
	}

	switch nodeDefine.ReturnType {
	case SyncFuncNode:
		flow.handleTaskSuccess(node, nil)
		return flow.handle(flow.Define.TaskNodes[nodeDefine.SuccessEvent])
	case PollingNode:
		if node.Status == constants.WorkFlowStatusFinished {
			return flow.handle(flow.Define.TaskNodes[nodeDefine.SuccessEvent])
		}
		ticker := time.NewTicker(3 * time.Second)
		sequence := int32(0)
		for range ticker.C {
			sequence++
			if sequence > maxPollingSequence {
				err := errors.Error(errors.TIUNIMANAGER_WORKFLOW_NODE_POLLING_TIME_OUT)
				flow.handleTaskError(node, nodeDefine, err)
				return false
			}
			framework.LogWithContext(flow.Context).Debugf("polling node waiting, sequence %d, nodeId %s, nodeName %s", sequence, node.ID, node.Name)

			op, err := deployment.M.GetStatus(flow.Context, node.OperationID)
			if err != nil {
				framework.LogWithContext(flow.Context).Errorf("call deployment GetStatus %s, failed %s", node.OperationID, err.Error())
				err := errors.NewError(errors.TIUNIMANAGER_TASK_FAILED, err.Error())
				flow.handleTaskError(node, nodeDefine, err)
				return false
			}
			if op.Status == deployment.Error {
				framework.LogWithContext(flow.Context).Errorf("call deployment GetStatus %s, response error %s", node.OperationID, op.ErrorStr)
				err := errors.NewError(errors.TIUNIMANAGER_TASK_FAILED, op.ErrorStr)
				flow.handleTaskError(node, nodeDefine, err)
				return false
			}
			if op.Status == deployment.Finished {
				if op.Result != "" {
					flow.handleTaskSuccess(node, op.Result)
				} else {
					flow.handleTaskSuccess(node, nil)
				}

				return flow.handle(flow.Define.TaskNodes[nodeDefine.SuccessEvent])
			}
		}
	}
	return true
}
*/
