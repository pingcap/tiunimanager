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

package workflow

import (
	"context"
)

// TaskRepo todo: replace db interface
var TaskRepo TaskRepository

type TaskRepository interface {
	AddFlowWork(ctx context.Context, flowWork *FlowWorkEntity) error
	AddFlowTask(ctx context.Context, task *TaskEntity, flowId string) error
	Persist(ctx context.Context, flowWork *FlowWorkAggregation) error
	LoadFlowWork(ctx context.Context, id uint) (*FlowWorkEntity, error)
	Load(ctx context.Context, id uint) (flowWork *FlowWorkAggregation, err error)
	ListFlows(ctx context.Context, bizId, keyword string, status int, page int, pageSize int) ([]*FlowWorkEntity, int, error)
}
