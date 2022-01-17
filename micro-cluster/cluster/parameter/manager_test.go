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
 * @Description: cluster parameter cluster service unit test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/14 09:52
*******************************************************************************/

package parameter

import (
	"context"
	"testing"

	"github.com/pingcap-inc/tiem/test/mockmodels/mockparametergroup"

	"github.com/asim/go-micro/v3/errors"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/cluster/parameter"
	"github.com/pingcap-inc/tiem/models/parametergroup"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclusterparameter"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockconfig"
	"github.com/pingcap-inc/tiem/workflow"
)

func TestManager_QueryClusterParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterId, name string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
			return "1", []*parameter.ClusterParamDetail{
				&parameter.ClusterParamDetail{
					Parameter: parametergroup.Parameter{
						ID:             "1",
						Category:       "basic",
						Name:           "param1",
						InstanceType:   "TiKV",
						SystemVariable: "",
						Type:           0,
					},
					DefaultValue: "10",
					RealValue:    "{\"clusterValue\":\"1\"}",
					Note:         "test parameter",
				},
			}, 1, nil
		})

	resp, page, err := mockManager.QueryClusterParameters(context.TODO(), cluster.QueryClusterParametersReq{
		ClusterID:   "1",
		PageRequest: structs.PageRequest{Page: 1, PageSize: 10},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.EqualValues(t, 1, page.Total)
}

func TestManager_UpdateClusterParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)
	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	configRW := mockconfig.NewMockReaderWriter(ctrl)
	models.SetConfigReaderWriter(configRW)

	clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterId, name string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
			return "1", []*parameter.ClusterParamDetail{
				{
					Parameter: parametergroup.Parameter{
						ID:             "1",
						Category:       "basic",
						Name:           "param1",
						InstanceType:   "TiKV",
						SystemVariable: "",
						Type:           0,
					},
					DefaultValue: "10",
					RealValue:    "{\"clusterValue\":\"1\"}",
					Note:         "test parameter",
				},
			}, 1, nil
		})
	clusterManagementRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterID string) (*management.Cluster, []*management.ClusterInstance, error) {
			return mockCluster(), mockClusterInstances(), nil
		})
	clusterManagementRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, bizId string, bizType string, flowName string) (*workflow.WorkFlowAggregation, error) {
			return mockWorkFlowAggregation(), nil
		})
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).AnyTimes()
	configRW.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).AnyTimes()

	resp, err := mockManager.UpdateClusterParameters(context.TODO(), cluster.UpdateClusterParametersReq{
		ClusterID: "1",
		Params: []structs.ClusterParameterSampleInfo{
			{
				ParamId:   "1",
				RealValue: structs.ParameterRealValue{ClusterValue: "10"},
			},
		},
		Reboot: false,
	}, true)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

func TestManager_ApplyParameterGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	configRW := mockconfig.NewMockReaderWriter(ctrl)
	models.SetConfigReaderWriter(configRW)

	parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, parameterGroupId string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
			return &parametergroup.ParameterGroup{ID: "1"}, []*parametergroup.ParamDetail{
				{
					Parameter: parametergroup.Parameter{
						ID:             "1",
						Name:           "param1",
						InstanceType:   "TiDB",
						SystemVariable: "",
						Type:           0,
						HasApply:       1,
						UpdateSource:   0,
						Range:          "[\"0\", \"1024\"]",
					},
					DefaultValue: "1",
					Note:         "param1",
				},
			}, nil
		})
	clusterManagementRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterID string) (*management.Cluster, []*management.ClusterInstance, error) {
			return mockCluster(), mockClusterInstances(), nil
		})
	clusterManagementRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, bizId string, bizType string, flowName string) (*workflow.WorkFlowAggregation, error) {
			return mockWorkFlowAggregation(), nil
		})
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).AnyTimes()
	configRW.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).AnyTimes()

	resp, err := mockManager.ApplyParameterGroup(context.TODO(), message.ApplyParameterGroupReq{
		ParamGroupId: "1",
		ClusterID:    "1",
		Reboot:       false,
	})
	assert.NotEmpty(t, resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ParamGroupID)
}

func TestManager_ApplyParameterGroup_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	clusterManagementRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterID string) (*management.Cluster, []*management.ClusterInstance, error) {
			return nil, nil, errors.Parse("cluster id is null")
		})
	_, err := mockManager.ApplyParameterGroup(context.TODO(), message.ApplyParameterGroupReq{
		ParamGroupId: "",
		ClusterID:    "",
		Reboot:       false,
	})
	assert.Error(t, err)
}

func TestManager_InspectClusterParameters(t *testing.T) {
	parameters, err := mockManager.InspectClusterParameters(context.TODO(), cluster.InspectClusterParametersReq{ClusterID: "1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, parameters)
}
