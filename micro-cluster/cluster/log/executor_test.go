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
 ******************************************************************************/

/*******************************************************************************
 * @File: executor_test.go
 * @Description: cluster log executor
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/29 17:05
*******************************************************************************/

package log

import (
	"context"
	"testing"

	"github.com/pingcap-inc/tiem/library/secondparty"
	mock_secondparty_v2 "github.com/pingcap-inc/tiem/test/mocksecondparty_v2"

	"github.com/golang/mock/gomock"

	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"

	"github.com/alecthomas/assert"
	"github.com/pingcap-inc/tiem/workflow"
)

func TestExecutor_collectorClusterLogConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock2rdService := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mock2rdService

	t.Run("success", func(t *testing.T) {
		mock2rdService.EXPECT().Transfer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		ctx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		ctx.SetData(contextClusterMeta, mockClusterMeta())
		err := collectorClusterLogConfig(mockWorkFlowAggregation().CurrentNode, ctx)
		assert.NoError(t, err)
	})
}

func TestExecutor_buildCollectorClusterLogConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		configs, err := buildCollectorClusterLogConfig(context.TODO(), []*handler.InstanceLogInfo{
			{
				ClusterID:    "123",
				InstanceType: "TiDB",
				IP:           "172.16.1.1",
				DataDir:      "/mnt/sda/123/tidb-data",
				DeployDir:    "/mnt/sda/123/tidb-deploy",
				LogDir:       "/mnt/sda/123/tidb-deploy/123/tidb-log",
			},
			{
				ClusterID:    "123",
				InstanceType: "PD",
				IP:           "172.16.1.2",
				DataDir:      "/mnt/sda/123/pd-data",
				DeployDir:    "/mnt/sda/123/pd-deploy",
				LogDir:       "/mnt/sda/123/pd-deploy/123/tidb-log",
			},
			{
				ClusterID:    "123",
				InstanceType: "TiKV",
				IP:           "172.16.1.3",
				DataDir:      "/mnt/sda/123/tikv-data",
				DeployDir:    "/mnt/sda/123/tikv-deploy",
				LogDir:       "/mnt/sda/123/tikv-deploy/123/tidb-log",
			},
			{
				ClusterID:    "123",
				InstanceType: "TiFlash",
				IP:           "172.16.1.4",
				DataDir:      "/mnt/sda/123/tiflash-data",
				DeployDir:    "/mnt/sda/123/tiflash-deploy",
				LogDir:       "/mnt/sda/123/tiflash-deploy/123/tidb-log",
			},
			{
				ClusterID:    "123",
				InstanceType: "CDC",
				IP:           "172.16.1.5",
				DataDir:      "/mnt/sda/123/cdc-data",
				DeployDir:    "/mnt/sda/123/cdc-deploy",
				LogDir:       "/mnt/sda/123/cdc-deploy/123/tidb-log",
			},
			{
				ClusterID:    "124",
				InstanceType: "CDC",
				IP:           "172.16.1.5",
				DataDir:      "/mnt/sda/123/cdc-data",
				DeployDir:    "/mnt/sda/123/cdc-deploy",
				LogDir:       "/mnt/sda/123/cdc-deploy/123/tidb-log",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(configs))

		assert.Equal(t, "/mnt/sda/123/tidb-deploy/123/tidb-log/tidb.log", configs[0].TiDB.Var.Paths[0])
		assert.Equal(t, "/mnt/sda/123/pd-deploy/123/tidb-log/pd.log", configs[0].PD.Var.Paths[0])
		assert.Equal(t, "/mnt/sda/123/tikv-deploy/123/tidb-log/tikv.log", configs[0].TiKV.Var.Paths[0])
	})
}

func TestExecutor_buildCollectorModuleDetail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		detail := buildCollectorModuleDetail("123", "127.0.0.1", "/mnt/sda/123/deploy-pd/123/pd.log")
		assert.NotNil(t, detail)
		assert.Equal(t, "123", detail.Input.Fields.ClusterId)
		assert.Equal(t, "127.0.0.1", detail.Input.Fields.Ip)
		assert.Equal(t, "/mnt/sda/123/deploy-pd/123/pd.log", detail.Var.Paths[0])
	})
}

func TestExecutor_listClusterHosts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hosts := listClusterHosts(mockClusterMeta())
		assert.NotNil(t, hosts)
	})
}

func TestExecutor_defaultEnd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		err := defaultEnd(mockWorkFlowAggregation().CurrentNode, ctx)
		assert.NoError(t, err)
	})
}
