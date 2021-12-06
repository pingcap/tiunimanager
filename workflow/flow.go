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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap/errors"
	//"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	//"github.com/pingcap-inc/tiem/library/secondparty"
	common2 "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/workflow"
	"time"
)

//
// FlowWorkAggregation
// @Description: flowwork aggregation with flowwork definition and nodes
//
type FlowWorkAggregation struct {
	FlowWork    *workflow.WorkFlow
	Define      *FlowWorkDefine
	CurrentNode *workflow.WorkFlowNode
	Nodes       []*workflow.WorkFlowNode
	Context     FlowContext
	FlowError   error
}

type FlowContext struct {
	context.Context
	FlowData map[string]interface{}
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

func createFlowWork(ctx context.Context, bizId string, define *FlowWorkDefine) (*FlowWorkAggregation, error) {
	framework.LogWithContext(ctx).Infof("create flowwork %v for bizId %s", define, bizId)
	if define == nil {
		return nil, framework.SimpleError(common.TIEM_FLOW_NOT_FOUND)
	}
	flowData := make(map[string]interface{})

	flow := define.getInstance(framework.ForkMicroCtx(ctx), bizId, flowData)
	_, err := models.GetWorkFlowReaderWriter().CreateWorkFlow(ctx, flow.FlowWork)
	if err != nil {
		return nil, err
	}
	//TaskRepo.AddFlowWork(ctx, flow.FlowWork)
	return flow, nil
}

func (flow *FlowWorkAggregation) GetFlowNodeNameList() []string {
	var nodeNames []string
	node := flow.Define.TaskNodes["start"]

	for node != nil && node.Name != "end" && node.Name != "fail" {
		nodeNames = append(nodeNames, node.Name)
		node = flow.Define.TaskNodes[node.SuccessEvent]
	}
	return nodeNames
}

func (flow *FlowWorkAggregation) start() {
	flow.FlowWork.Status = constants.WorkFlowStatusProcessing
	start := flow.Define.TaskNodes["start"]
	result := flow.handle(start)
	flow.complete(result)
	_ = models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.FlowWork, flow.Nodes)
	//TaskRepo.Persist(flow.Context, flow)
}

func (flow *FlowWorkAggregation) asyncStart() {
	go flow.start()
}

func (flow *FlowWorkAggregation) destroy(reason string) {
	flow.FlowWork.Status = constants.WorkFlowStatusCanceled

	if flow.CurrentNode != nil {
		flow.CurrentNode.Fail(framework.NewTiEMError(common.TIEM_TASK_CANCELED, reason))
	}
	_ = models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.FlowWork, flow.Nodes)
	//TaskRepo.Persist(flow.Context, flow)
}

func (flow FlowWorkAggregation) complete(success bool) {
	if success {
		flow.FlowWork.Status = constants.WorkFlowStatusFinished
	} else {
		flow.FlowWork.Status = constants.WorkFlowStatusError
	}
}

func (flow *FlowWorkAggregation) addContext(key string, value interface{}) {
	flow.Context.SetData(key, value)
}

func (flow *FlowWorkAggregation) executeTask(node *workflow.WorkFlowNode, nodeDefine *NodeDefine) bool {
	flow.CurrentNode = node
	flow.Nodes = append(flow.Nodes, node)
	node.Processing()
	_ = models.GetWorkFlowReaderWriter().UpdateWorkFlowDetail(flow.Context, flow.FlowWork, flow.Nodes)
	//TaskRepo.Persist(flow.Context, flow)

	return nodeDefine.Executor(node, &flow.Context)
}

func (flow *FlowWorkAggregation) handleTaskError(node *workflow.WorkFlowNode, nodeDefine *NodeDefine) {
	flow.FlowError = errors.New(node.Result)
	if "" != nodeDefine.FailEvent {
		flow.handle(flow.Define.TaskNodes[nodeDefine.FailEvent])
	} else {
		framework.Log().Warnf("no fail event in flow definition, flowname %s", nodeDefine.Name)
	}
}

func (flow *FlowWorkAggregation) handle(nodeDefine *NodeDefine) bool {
	if nodeDefine == nil {
		flow.FlowWork.Status = constants.WorkFlowStatusFinished
		return true
	}
	node := &workflow.WorkFlowNode{
		Entities: common2.Entities{
			Status: constants.WorkFlowStatusInitializing,
		},
		Name:       nodeDefine.Name,
		BizID:      flow.FlowWork.BizID,
		ParentID:   flow.FlowWork.ID,
		ReturnType: string(nodeDefine.ReturnType),
		StartTime:  time.Now(),
	}

	_, _ = models.GetWorkFlowReaderWriter().CreateWorkFlowNode(flow.Context, node)
	//TaskRepo.AddFlowTask(flow.Context, task, flow.FlowWork.ID)
	handleSuccess := flow.executeTask(node, nodeDefine)

	if !handleSuccess {
		flow.handleTaskError(node, nodeDefine)
		return false
	}

	switch nodeDefine.ReturnType {
	case workflow.SyncFuncNode:
		return flow.handle(flow.Define.TaskNodes[nodeDefine.SuccessEvent])
	case workflow.PollingNode:
		//todo: wait tiup bizid become string
		/*
			ticker := time.NewTicker(3 * time.Second)
			sequence := 0
			for range ticker.C {
				if sequence += 1; sequence > 200 {
					task.Fail(framework.SimpleError(common.TIEM_TASK_POLLING_TIME_OUT))
					flow.handleTaskError(task, taskDefine)
					return false
				}
				framework.LogWithContext(flow.Context).Infof("polling task waiting, sequence %d, taskId %d, taskName %s", sequence, task.ID, task.Name)

				stat, statString, err := secondparty.SecondParty.MicroSrvGetTaskStatusByBizID(flow.Context, uint64(task.ID))
				if err != nil {
					framework.LogWithContext(flow.Context).Error(err)
					task.Fail(framework.WrapError(common.TIEM_TASK_FAILED, common.TIEM_TASK_FAILED.Explain(), err))
					flow.handleTaskError(task, taskDefine)
					return false
				}
				if stat == dbpb.TiupTaskStatus_Error {
					task.Fail(framework.NewTiEMError(common.TIEM_TASK_FAILED, statString))
					flow.handleTaskError(task, taskDefine)
					return false
				}
				if stat == dbpb.TiupTaskStatus_Finished {
					task.Success(statString)
					return flow.handle(flow.Define.TaskNodes[taskDefine.SuccessEvent])
				}
			}
		*/
	}
	return true
}
