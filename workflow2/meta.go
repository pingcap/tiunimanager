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

package workflow2

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/deployment"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/metrics"
	"github.com/pingcap-inc/tiem/models"
	dbModel "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/workflow"
	"runtime/debug"
	"sync"
	"time"
)

// WorkFlowMeta workflow meta info with workflow definition and nodes
type WorkFlowMeta struct {
	Flow              *workflow.WorkFlow
	Define            *WorkFlowDefine
	CurrentNode       *workflow.WorkFlowNode
	CurrentNodeDefine *NodeDefine
	Nodes             []*workflow.WorkFlowNode
	Context           *FlowContext
	IsFailNode        bool
}

type FlowContext struct {
	context.Context
	FlowData map[string]string
	mutex    sync.Mutex
}

func NewFlowContext(ctx context.Context, data map[string]string) *FlowContext {
	flowCtx := &FlowContext{
		ctx,
		data,
		sync.Mutex{},
	}
	return flowCtx.InitFlowContext()
}

func (c *FlowContext) GetData(key string, data interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.FlowData[key]; !ok {
		return nil
	} else {
		return json.Unmarshal([]byte(c.FlowData[key]), data)
	}
}

func (c *FlowContext) SetData(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.FlowData[key] = string(data)
	return nil
}

func handleWorkFlowMetrics(flow *workflow.WorkFlow) {
	metrics.HandleWorkFlowMetrics(metrics.WorkFlowLabel{
		BizType: flow.BizType,
		Name:    flow.Name,
		Status:  flow.Status,
	})
}

func handleWorkFlowNodeMetrics(flow *WorkFlowMeta, node *workflow.WorkFlowNode) {
	metrics.HandleWorkFlowNodeMetrics(metrics.WorkFlowNodeLabel{
		BizType:  flow.Flow.BizType,
		FlowName: flow.Flow.Name,
		Node:     node.Name,
		Status:   node.Status,
	})
}

func (c *FlowContext) InitFlowContext() *FlowContext {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Context = framework.NewMicroContextWithKeyValuePairs(c.Context, map[string]string{
		framework.TiEM_X_TRACE_ID_KEY:  c.FlowData[framework.TiEM_X_TRACE_ID_KEY],
		framework.TiEM_X_USER_ID_KEY:   c.FlowData[framework.TiEM_X_USER_ID_KEY],
		framework.TiEM_X_TENANT_ID_KEY: c.FlowData[framework.TiEM_X_TENANT_ID_KEY],
	})
	return c
}

func (c *FlowContext) GetContextString() string {
	data, err := json.Marshal(c.FlowData)
	if err != nil {
		framework.LogWithContext(c).Warnf("json marshal flow context data failed %s", err.Error())
		return ""
	}
	return string(data)
}

func NewWorkFlowMeta(ctx context.Context, flowId string) (*WorkFlowMeta, error) {
	flow, nodes, err := models.GetWorkFlowReaderWriter().QueryDetailWorkFlow(ctx, flowId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get workflow meta by flowId %s failed, %s", flowId, err.Error())
		return nil, err
	}

	define, err := GetWorkFlowService().GetWorkFlowDefine(ctx, flow.Name)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get workflow define by flowName %s failed, %s", flow.Name, err.Error())
		return nil, err
	}

	flowData := make(map[string]string)
	if flow.Context != "" {
		err = json.Unmarshal([]byte(flow.Context), &flowData)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("json unmarshal flow context %s failed, %s", flow.Context, err.Error())
			return nil, err
		}
	}

	var latest *workflow.WorkFlowNode
	var latestNodeDefineKey string
	for _, node := range nodes {
		if latest == nil || latest.CreatedAt.Before(node.CreatedAt) {
			latest = node
		}
	}

	isNodeFail := false
	var currentNodeDefine *NodeDefine
	if latest != nil {
		latestNodeDefineKey = define.getNodeDefineKeyByName(latest.Name)
		if latestNodeDefineKey == "" {
			framework.LogWithContext(ctx).Errorf("get node define key by node name %s failed", latest.Name)
			return nil, fmt.Errorf("get node define key by node name %s failed", latest.Name)
		}
		currentNodeDefine = define.TaskNodes[latestNodeDefineKey]
		switch latest.Status {
		case constants.WorkFlowStatusInitializing:
		case constants.WorkFlowStatusProcessing:
			currentNodeDefine = define.TaskNodes[latestNodeDefineKey]
		case constants.WorkFlowStatusFinished:
			currentNodeDefine = define.TaskNodes[currentNodeDefine.SuccessEvent]
		case constants.WorkFlowStatusError:
			currentNodeDefine = define.TaskNodes[currentNodeDefine.FailEvent]
		}
	} else {
		currentNodeDefine = define.TaskNodes["start"]
	}

	var current *workflow.WorkFlowNode
	if currentNodeDefine != nil {
		isNodeFail = define.isFailNode(currentNodeDefine.Name)
		current = &workflow.WorkFlowNode{
			Entity: dbModel.Entity{
				TenantId: flow.TenantId,
				Status:   constants.WorkFlowStatusInitializing,
			},
			Name:       currentNodeDefine.Name,
			BizID:      flow.BizID,
			ParentID:   flow.ID,
			ReturnType: string(currentNodeDefine.ReturnType),
			StartTime:  time.Now(),
		}
	} else {
		isNodeFail = define.isFailNode(latest.Name)
	}

	meta := &WorkFlowMeta{
		Flow:              flow,
		Nodes:             nodes,
		Define:            define,
		CurrentNode:       current,
		CurrentNodeDefine: currentNodeDefine,
		Context:           NewFlowContext(ctx, flowData),
		IsFailNode:        isNodeFail,
	}

	return meta, nil
}

func (flow *WorkFlowMeta) Restore() {
	data, err := json.Marshal(flow.Context.FlowData)
	if err != nil {
		framework.LogWithContext(flow.Context).Warnf("json marshal flow context data failed %s", err.Error())
		return
	}
	flow.Flow.Context = string(data)
	current, err := models.GetWorkFlowReaderWriter().GetWorkFlow(flow.Context, flow.Flow.ID)
	if err != nil {
		framework.LogWithContext(flow.Context).Warnf("get workflow by id %s failed %s", flow.Flow.ID, err.Error())
		return
	}
	if (constants.WorkFlowStatusStopped == current.Status ||
		constants.WorkFlowStatusCanceling == current.Status ||
		constants.WorkFlowStatusCanceled == current.Status) && constants.WorkFlowStatusCanceled != flow.Flow.Status {
		//interrupt status update by api, do not overwrite it
		flow.Flow.Status = current.Status
	}
	err = models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.Flow, flow.Nodes)
	if err != nil {
		framework.LogWithContext(flow.Context).Warnf("update workflow detail %+v failed %s", flow, err.Error())
	}
	//framework.LogWithContext(flow.Context).Infof("restore workflow %+v success", flow.Flow)
}

func (flow *WorkFlowMeta) CheckNeedPause() {
	if flow.CurrentNodeDefine.FailEvent == "pause" {
		//pause wait for manual handle
		flow.Flow.Status = constants.WorkFlowStatusStopped
	}
}

func (flow *WorkFlowMeta) Execute() {
	defer func() {
		if r := recover(); r != nil {
			framework.LogWithContext(flow.Context).Errorf("recover from workflow %s, node %s, stacktrace %s", flow.Flow.Name, flow.CurrentNode.Name, string(debug.Stack()))
			execErr := errors.NewErrorf(errors.TIEM_PANIC, "%v", r)
			if flow.CurrentNode != nil {
				flow.CurrentNode.Fail(execErr)
			}
		}
	}()

	if flow.CurrentNode != nil {
		framework.LogWithContext(flow.Context).Infof("begin execute workflow %s, id %s, node name %s", flow.Flow.Name, flow.Flow.ID, flow.CurrentNode.Name)
		defer framework.LogWithContext(flow.Context).Infof("end execute workflow %s, id %s, node name %s", flow.Flow.Name, flow.Flow.ID, flow.CurrentNode.Name)
	}

	if flow.Flow.Status == constants.WorkFlowStatusInitializing {
		flow.Flow.Status = constants.WorkFlowStatusProcessing
	}
	if flow.CurrentNodeDefine == nil {
		if !flow.IsFailNode {
			flow.Flow.Status = constants.WorkFlowStatusFinished
		} else {
			flow.Flow.Status = constants.WorkFlowStatusError
		}
		handleWorkFlowMetrics(flow.Flow)
		flow.Restore()
		return
	}

	node := flow.CurrentNode
	nodeDefine := flow.CurrentNodeDefine

	handleWorkFlowNodeMetrics(flow, node)
	_, err := models.GetWorkFlowReaderWriter().CreateWorkFlowNode(flow.Context, node)
	if err != nil {
		framework.LogWithContext(flow.Context).Warnf("create workflow node, node %s failed %s", node.Name, err.Error())
		return
	}

	flow.Nodes = append(flow.Nodes, node)
	node.Processing()
	handleWorkFlowNodeMetrics(flow, node)
	flow.Restore()

	err = nodeDefine.Executor(node, flow.Context)
	if err != nil {
		framework.LogWithContext(flow.Context).Infof("workflow %s of bizId %s do node %s failed, %s", flow.Flow.ID, flow.Flow.BizID, node.Name, err.Error())
		node.Fail(err)
		handleWorkFlowNodeMetrics(flow, node)
		flow.CheckNeedPause()
		flow.Restore()
		return
	}

	switch nodeDefine.ReturnType {
	case SyncFuncNode:
		node.Success()
		handleWorkFlowNodeMetrics(flow, node)
		flow.Restore()
		return
	case PollingNode:
		if node.Status == constants.WorkFlowStatusFinished {
			flow.Restore()
			handleWorkFlowNodeMetrics(flow, node)
			return
		}
		ticker := time.NewTicker(3 * time.Second)
		sequence := int32(0)
		for range ticker.C {
			sequence++
			if sequence > maxPollingSequence {
				node.Fail(errors.Error(errors.TIEM_WORKFLOW_NODE_POLLING_TIME_OUT))
				handleWorkFlowNodeMetrics(flow, node)
				flow.CheckNeedPause()
				flow.Restore()
				return
			}
			framework.LogWithContext(flow.Context).Debugf("polling node waiting, sequence %d, nodeId %s, nodeName %s", sequence, node.ID, node.Name)

			op, err := deployment.M.GetStatus(flow.Context, node.OperationID)
			if err != nil {
				framework.LogWithContext(flow.Context).Errorf("call deployment GetStatus %s, failed %s", node.OperationID, err.Error())
				node.Fail(errors.NewError(errors.TIEM_TASK_FAILED, err.Error()))
				handleWorkFlowNodeMetrics(flow, node)
				flow.CheckNeedPause()
				flow.Restore()
				return
			}
			if op.Status == deployment.Error {
				framework.LogWithContext(flow.Context).Errorf("call deployment GetStatus %s, response error %s", node.OperationID, op.ErrorStr)
				node.Fail(errors.NewError(errors.TIEM_TASK_FAILED, op.ErrorStr))
				handleWorkFlowNodeMetrics(flow, node)
				flow.CheckNeedPause()
				flow.Restore()
				return
			}
			if op.Status == deployment.Finished {
				if op.Result != "" {
					node.Success(op.Result)
				} else {
					node.Success(nil)
				}
				handleWorkFlowNodeMetrics(flow, node)
				flow.Restore()
				return
			}
		}

	}
}
