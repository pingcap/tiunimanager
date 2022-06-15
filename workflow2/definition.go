/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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

package workflow2

import (
	"github.com/pingcap/tiunimanager/models/workflow"
)

type WorkFlowDefine struct {
	FlowName      string
	TaskNodes     map[string]*NodeDefine
	ContextParser func(string) *FlowContext
}

type NodeExecutor func(task *workflow.WorkFlowNode, context *FlowContext) error

type NodeDefine struct {
	Name         string
	SuccessEvent string
	FailEvent    string
	ReturnType   NodeReturnType
	Executor     NodeExecutor
}

func (define *WorkFlowDefine) getNodeNameList() []string {
	var nodeNames []string
	node := define.TaskNodes["start"]

	for node != nil {
		nodeNames = append(nodeNames, node.Name)
		node = define.TaskNodes[node.SuccessEvent]
	}
	return nodeNames
}

func (define *WorkFlowDefine) getNodeDefineKeyByName(nodeName string) string {
	for key, value := range define.TaskNodes {
		if nodeName == value.Name {
			return key
		}
	}
	return ""
}

func (define *WorkFlowDefine) isFailNode(nodeName string) bool {
	for _, value := range define.TaskNodes {
		if value.FailEvent == nodeName {
			return true
		}
	}
	return false
}

func CompositeExecutor(executors ...NodeExecutor) NodeExecutor {
	return func(node *workflow.WorkFlowNode, context *FlowContext) error {
		for _, executor := range executors {
			if err := executor(node, context); err == nil {
				continue
			} else {
				return err
			}
		}
		return nil
	}
}
