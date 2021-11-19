
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

package domain

import (
	ctx "context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_defaultContextParser(t *testing.T) {
	ctx := defaultContextParser("")
	assert.NotNil(t, ctx)
}

func TestInitFlowMap(t *testing.T) {
	current := FlowWorkDefineMap
	defer func() {
		FlowWorkDefineMap = current
	}()
	InitFlowMap()
	assert.LessOrEqual(t, 10, len(FlowWorkDefineMap))
}

func TestClusterEndWithPersist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(ctx.TODO())
		agg := &ClusterAggregation{
			Cluster: &Cluster{WorkFlowId: 999},
			ConfigModified: false,
			FlowModified: false,
		}
		flowCtx.SetData(contextClusterKey, agg)
		ret := CompositeExecutor(ClusterEnd, ClusterPersist)(task, flowCtx)

		assert.Equal(t, true, ret)
		assert.Equal(t, 0, int(agg.Cluster.WorkFlowId))
		assert.Equal(t, true, agg.FlowModified)
		assert.Equal(t, TaskStatusFinished, task.Status)
	})
}

func TestClusterEnd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(ctx.TODO())
		agg := &ClusterAggregation{
			Cluster: &Cluster{WorkFlowId: 999},
			ConfigModified: false,
			FlowModified: false,
		}
		flowCtx.SetData(contextClusterKey, agg)
		ret := ClusterEnd(task, flowCtx)

		assert.Equal(t, true, ret)
		assert.Equal(t, 0, int(agg.Cluster.WorkFlowId))
		assert.Equal(t, true, agg.FlowModified)
		assert.Equal(t, TaskStatusFinished, task.Status)

	})
}


func TestClusterFailWithPersist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(ctx.TODO())
		agg := &ClusterAggregation{
			Cluster: &Cluster{WorkFlowId: 999},
			ConfigModified: false,
			FlowModified: false,
		}
		flowCtx.SetData(contextClusterKey, agg)
		ret := CompositeExecutor(ClusterFail, ClusterPersist)(task, flowCtx)

		assert.Equal(t, true, ret)
		assert.Equal(t, 0, int(agg.Cluster.WorkFlowId))
		assert.Equal(t, true, agg.FlowModified)
		assert.Equal(t, TaskStatusError, task.Status)
	})
}

func TestClusterFail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		task := &TaskEntity{
			Id: 123,
		}
		flowCtx := NewFlowContext(ctx.TODO())
		agg := &ClusterAggregation{
			Cluster: &Cluster{WorkFlowId: 999},
			ConfigModified: false,
			FlowModified: false,
		}
		flowCtx.SetData(contextClusterKey, agg)
		ret := ClusterFail(task, flowCtx)

		assert.Equal(t, true, ret)
		assert.Equal(t, 0, int(agg.Cluster.WorkFlowId))
		assert.Equal(t, true, agg.FlowModified)
		assert.Equal(t, TaskStatusError, task.Status)

	})
}