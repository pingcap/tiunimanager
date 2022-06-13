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
 * @File: manager_test.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/17 14:47
*******************************************************************************/

package log

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/pingcap/tiunimanager/test/mockmodels/mockconfig"
	mock_workflow_service "github.com/pingcap/tiunimanager/test/mockworkflow"
	workflow "github.com/pingcap/tiunimanager/workflow2"

	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/pingcap/tiunimanager/models/cluster/management"
	"github.com/pingcap/tiunimanager/models/common"

	"github.com/golang/mock/gomock"
	"github.com/pingcap/tiunimanager/test/mockmodels/mockclustermanagement"

	"github.com/alecthomas/assert"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/message/cluster"

	"github.com/pingcap/tiunimanager/models"
)

func TestManager_BuildClusterLogConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	configRW := mockconfig.NewMockReaderWriter(ctrl)
	models.SetConfigReaderWriter(configRW)

	clusterManagementRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterID string) (*management.Cluster, []*management.ClusterInstance, []*management.DBUser, error) {
			return mockCluster(), mockClusterInstances(), mockDBUsers(), nil
		})
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, bizId string, bizType string, flowName string) (string, error) {
			return "flowId", nil
		})
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes()
	configRW.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := mockManager.BuildClusterLogConfig(context.TODO(), "123")
	assert.NoError(t, err)
}

func TestManager_prepareSearchParams_Success1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&management.Cluster{
			Entity: common.Entity{ID: "123"},
		}, nil)

		buf, err := prepareSearchParams(context.TODO(), cluster.QueryClusterLogReq{
			ClusterID: "123",
			Module:    "tidb",
			Level:     "info",
			Ip:        "127.0.0.1",
			Message:   "hello",
			StartTime: 1630468800,
			EndTime:   1638331200,
			PageRequest: structs.PageRequest{
				Page:     1,
				PageSize: 10,
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, buf)
	})
}

func TestManager_prepareSearchParams_Success2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&management.Cluster{
			Entity: common.Entity{ID: "123"},
		}, nil)

		_, err := prepareSearchParams(context.TODO(), cluster.QueryClusterLogReq{
			StartTime: 0,
			EndTime:   1638331200,
		})
		assert.NoError(t, err)
	})
}

func TestManager_prepareSearchParams_Success3(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&management.Cluster{
			Entity: common.Entity{ID: "123"},
		}, nil)

		_, err := prepareSearchParams(context.TODO(), cluster.QueryClusterLogReq{
			StartTime: 1630468800,
			EndTime:   0,
		})
		assert.NoError(t, err)
	})
}

func TestManager_prepareSearchParams_Error1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&management.Cluster{
			Entity: common.Entity{ID: "123"},
		}, nil)

		_, err := prepareSearchParams(context.TODO(), cluster.QueryClusterLogReq{
			StartTime: 1638331200,
			EndTime:   1630468800,
		})
		assert.Error(t, err)
	})
}

var esResult = `
{
  "took" : 3,
  "timed_out" : false,
  "_shards" : {
    "total" : 3,
    "successful" : 3,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 1,
      "relation" : "eq"
    },
    "max_score" : null,
    "hits" : [
      {
        "_index" : "tiunimanager-tidb-cluster-2021.12.16",
        "_type" : "_doc",
        "_id" : "Nu30wX0BomM1rMZIIW8w",
        "_score" : null,
        "_source" : {
          "agent" : {
            "name" : "CentOS76_VM",
            "id" : "5667eca4-1022-478d-86a8-866830c61dd1",
            "type" : "filebeat",
            "ephemeral_id" : "2d0992b6-c591-4f76-b2c6-1159c1c0bfe3",
            "version" : "8.0.0"
          },
          "tidb" : {
            "tikv" : {
              "worker" : "lock-collector"
            }
          },
          "log" : {
            "file" : {
              "path" : "/mnt/sda/4BbN6j5FRbewGZ9iVNxAJQ/tikv-deploy/4BbN6j5FRbewGZ9iVNxAJQ/tidb-log/tikv.log"
            },
            "offset" : 3218,
            "level" : "INFO",
            "logger" : "mod.rs:375"
          },
          "ip" : "172.16.5.148",
          "clusterId" : "4BbN6j5FRbewGZ9iVNxAJQ",
          "message" : "stoping worker",
          "fileset" : {
            "name" : "tikv"
          },
          "type" : "tidb",
          "input" : {
            "type" : "log"
          },
          "@timestamp" : "2021-12-16T06:35:36.288Z",
          "ecs" : {
            "version" : "1.11.0"
          },
          "service" : {
            "type" : "tidb"
          },
          "host" : {
            "name" : "CentOS76_VM"
          },
          "event" : {
            "ingested" : "2021-12-16T06:35:41.743965614Z",
            "module" : "tidb",
            "dataset" : "tidb.tikv"
          }
        },
        "sort" : [
          1639636536288
        ]
      }
    ]
  }
}`

func TestManager_handleResult_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		resp, page, err := handleResult(context.TODO(), cluster.QueryClusterLogReq{
			ClusterID: "4BbN6j5FRbewGZ9iVNxAJQ",
		}, &esapi.Response{
			StatusCode: 200,
			Header:     nil,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(esResult))),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.EqualValues(t, 1, page.Total)
	})
}

func TestManager_handleResult_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		_, _, err := handleResult(context.TODO(), cluster.QueryClusterLogReq{
			ClusterID: "4BbN6j5FRbewGZ9iVNxAJQ",
		}, &esapi.Response{
			StatusCode: 400,
			Header:     nil,
			Body:       nil,
		})
		assert.Error(t, err)
	})
}
