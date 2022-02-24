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
	"errors"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"

	"github.com/pingcap-inc/tiem/models/common"

	"github.com/pingcap-inc/tiem/deployment"
	mock_deployment "github.com/pingcap-inc/tiem/test/mockdeployment"

	"github.com/pingcap-inc/tiem/test/mockutilcdc"
	"github.com/pingcap-inc/tiem/test/mockutilpd"
	"github.com/pingcap-inc/tiem/test/mockutiltidbhttp"
	"github.com/pingcap-inc/tiem/test/mockutiltikv"
	"github.com/pingcap-inc/tiem/util/api/cdc"
	"github.com/pingcap-inc/tiem/util/api/pd"
	"github.com/pingcap-inc/tiem/util/api/tidb/http"
	"github.com/pingcap-inc/tiem/util/api/tikv"

	"github.com/pingcap-inc/tiem/test/mockmodels/mockparametergroup"

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

	t.Run("success", func(t *testing.T) {
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)

		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
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
	})

	t.Run("query cluster parameter fail", func(t *testing.T) {
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)

		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
				return "1", []*parameter.ClusterParamDetail{}, 1, errors.New("query cluster parameter fail")
			})

		_, _, err := mockManager.QueryClusterParameters(context.TODO(), cluster.QueryClusterParametersReq{
			ClusterID:   "1",
			PageRequest: structs.PageRequest{Page: 1, PageSize: 10},
		})
		assert.Error(t, err)
	})
}

func TestManager_UpdateClusterParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		configRW := mockconfig.NewMockReaderWriter(ctrl)
		models.SetConfigReaderWriter(configRW)

		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
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
			DoAndReturn(func(ctx context.Context, clusterID string) (*management.Cluster, []*management.ClusterInstance, []*management.DBUser, error) {
				return mockCluster(), mockClusterInstances(), mockDBUsers(), nil
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
	})

	t.Run("query cluster parameter fail", func(t *testing.T) {
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)

		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
				return "1", []*parameter.ClusterParamDetail{}, 1, errors.New("query cluster parameter fail")
			})

		_, err := mockManager.UpdateClusterParameters(context.TODO(), cluster.UpdateClusterParametersReq{
			ClusterID: "1",
			Params: []structs.ClusterParameterSampleInfo{
				{
					ParamId:   "1",
					RealValue: structs.ParameterRealValue{ClusterValue: "10"},
				},
			},
			Reboot: false,
		}, true)
		assert.Error(t, err)
	})
}

func TestManager_ApplyParameterGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
		models.SetParameterGroupReaderWriter(parameterGroupRW)
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		configRW := mockconfig.NewMockReaderWriter(ctrl)
		models.SetConfigReaderWriter(configRW)

		parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, parameterGroupId, paramName, instanceType string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
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
							RangeType:      1,
							Unit:           "MB",
							UnitOptions:    "[\"KB\", \"MB\", \"GB\"]",
						},
						DefaultValue: "1",
						Note:         "param1",
					},
				}, nil
			})
		clusterManagementRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterID string) (*management.Cluster, []*management.ClusterInstance, []*management.DBUser, error) {
				return mockCluster(), mockClusterInstances(), mockDBUsers(), nil
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
		}, true)
		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ParamGroupID)
	})
}

func TestManager_ApplyParameterGroup_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	clusterManagementRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, clusterID string) (*management.Cluster, []*management.ClusterInstance, []*management.DBUser, error) {
			return nil, nil, nil, errors.New("cluster id is null")
		})
	_, err := mockManager.ApplyParameterGroup(context.TODO(), message.ApplyParameterGroupReq{
		ParamGroupId: "",
		ClusterID:    "",
		Reboot:       false,
	}, true)
	assert.Error(t, err)
}

func TestManager_PersistApplyParameterGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
		models.SetParameterGroupReaderWriter(parameterGroupRW)
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)

		parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, parameterGroupId, paramName, instanceType string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
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
							RangeType:      1,
							Unit:           "MB",
							UnitOptions:    "[\"KB\", \"MB\", \"GB\"]",
						},
						DefaultValue: "1",
						Note:         "param1",
					},
				}, nil
			})
		clusterParameterRW.EXPECT().ApplyClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, parameterGroupId string, clusterId string, param []*parameter.ClusterParameterMapping) error {
				return nil
			})

		resp, err := mockManager.PersistApplyParameterGroup(context.TODO(), message.ApplyParameterGroupReq{
			ParamGroupId: "1",
			ClusterID:    "1",
			Reboot:       false,
		}, false)
		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ParamGroupID)
	})

	t.Run("get parameter group fail", func(t *testing.T) {
		parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
		models.SetParameterGroupReaderWriter(parameterGroupRW)

		parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, parameterGroupId, paramName, instanceType string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
				return &parametergroup.ParameterGroup{ID: "1"}, []*parametergroup.ParamDetail{}, errors.New("get parameter group fail")
			})

		_, err := mockManager.PersistApplyParameterGroup(context.TODO(), message.ApplyParameterGroupReq{
			ParamGroupId: "1",
			ClusterID:    "1",
			Reboot:       false,
		}, false)
		assert.Error(t, err)
	})
}

func TestManager_InspectClusterParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("inspect cluster parameter", func(t *testing.T) {
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)

		mockCDCApiService := mockutilcdc.NewMockCDCApiService(ctrl)
		cdc.ApiService = mockCDCApiService
		mockPDApiService := mockutilpd.NewMockPDApiService(ctrl)
		pd.ApiService = mockPDApiService
		mockTiDBApiService := mockutiltidbhttp.NewMockTiDBApiService(ctrl)
		http.ApiService = mockTiDBApiService
		mockTiKVApiService := mockutiltikv.NewMockTiKVApiService(ctrl)
		tikv.ApiService = mockTiKVApiService
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(string(tomlConfigContent), nil)
		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(string(tomlConfigContent), nil)

		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
				return "1", []*parameter.ClusterParamDetail{
					{
						Parameter: parametergroup.Parameter{
							ID:           "1",
							Category:     "security",
							Name:         "enable-sem",
							InstanceType: "TiDB",
							HasApply:     1,
							Type:         2,
						},
						RealValue: "{\"clusterValue\":\"true\"}",
					},
					{
						Parameter: parametergroup.Parameter{
							ID:           "2",
							Category:     "basic",
							Name:         "oom-action",
							InstanceType: "TiKV",
							HasApply:     1,
							Type:         1,
						},
						RealValue: "{\"clusterValue\":\"cancel\"}",
					},
					{
						Parameter: parametergroup.Parameter{
							ID:           "3",
							Category:     "log.file",
							Name:         "max-size",
							InstanceType: "PD",
							HasApply:     1,
							Type:         0,
						},
						RealValue: "{\"clusterValue\":\"102400\"}",
					},
					{
						Parameter: parametergroup.Parameter{
							ID:           "4",
							Category:     "basic",
							Name:         "quota-backend-bytes",
							InstanceType: "CDC",
							HasApply:     1,
							Type:         0,
						},
						RealValue: "{\"clusterValue\":\"8589934592\"}",
					},
					{
						Parameter: parametergroup.Parameter{
							ID:           "5",
							Category:     "label-property",
							Name:         "value",
							InstanceType: "TiFlash",
							HasApply:     1,
							Type:         4,
						},
						RealValue: "{\"clusterValue\":\"[\\\"test\\\"]\"}",
					},
					{
						Parameter: parametergroup.Parameter{
							ID:           "test1",
							Category:     "basic",
							Name:         "param1",
							InstanceType: "TiKV",
							HasApply:     1,
							Type:         0,
						},
						DefaultValue: "10",
						RealValue:    "{\"clusterValue\":\"\"}",
					},
					{
						Parameter: parametergroup.Parameter{
							ID:           "test2",
							Category:     "basic",
							Name:         "param1",
							InstanceType: "TiKV",
							HasApply:     0,
							Type:         0,
						},
						DefaultValue: "10",
						RealValue:    "{\"clusterValue\":\"1\"}",
					},
				}, 7, nil
			})
		clusterManagementRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterID string) (*management.Cluster, []*management.ClusterInstance, []*management.DBUser, error) {
				return mockCluster(), mockClusterAllInstances(), mockDBUsers(), nil
			})

		mockPDApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, nil)
		mockTiDBApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, nil)
		mockTiKVApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, nil)

		resp, err := mockManager.InspectClusterParameters(context.TODO(), cluster.InspectParametersReq{
			ClusterID: "123",
		})
		assert.NoError(t, err)
		assert.Equal(t, len(resp.Params), 5)
	})

	t.Run("inspect tidb instance parameter", func(t *testing.T) {
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)
		mockTiDBApiService := mockutiltidbhttp.NewMockTiDBApiService(ctrl)
		http.ApiService = mockTiDBApiService

		mockTiDBApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, nil)
		clusterManagementRW.EXPECT().GetInstance(gomock.Any(), gomock.Any()).
			Return(&management.ClusterInstance{
				Entity:    common.Entity{ID: "1", Status: string(constants.ClusterInstanceRunning)},
				Type:      "TiDB",
				Version:   "v5.0.0",
				ClusterID: "123",
				HostIP:    []string{"127.0.0.1"},
				Ports:     []int32{10000, 10001},
			}, nil)
		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
				return "1", []*parameter.ClusterParamDetail{
					{
						Parameter: parametergroup.Parameter{
							ID:           "1",
							Category:     "security",
							Name:         "enable-sem",
							InstanceType: "TiDB",
							HasApply:     1,
							Type:         2,
						},
						RealValue: "{\"clusterValue\":\"true\"}",
					},
				}, 1, nil
			})

		resp, err := mockManager.InspectClusterParameters(context.TODO(), cluster.InspectParametersReq{
			InstanceID: "1",
		})
		assert.NoError(t, err)
		assert.Equal(t, len(resp.Params), 1)
	})

	t.Run("inspect tikv instance parameter", func(t *testing.T) {
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)
		mockTiKVApiService := mockutiltikv.NewMockTiKVApiService(ctrl)
		tikv.ApiService = mockTiKVApiService

		mockTiKVApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, nil)
		clusterManagementRW.EXPECT().GetInstance(gomock.Any(), gomock.Any()).
			Return(&management.ClusterInstance{
				Entity:    common.Entity{ID: "1", Status: string(constants.ClusterInstanceRunning)},
				Type:      "TiKV",
				Version:   "v5.0.0",
				ClusterID: "123",
				HostIP:    []string{"127.0.0.1"},
				Ports:     []int32{10000, 10001},
			}, nil)
		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
				return "1", []*parameter.ClusterParamDetail{
					{
						Parameter: parametergroup.Parameter{
							ID:           "2",
							Category:     "basic",
							Name:         "oom-action",
							InstanceType: "TiKV",
							HasApply:     1,
							Type:         1,
						},
						RealValue: "{\"clusterValue\":\"cancel\"}",
					},
				}, 1, nil
			})

		resp, err := mockManager.InspectClusterParameters(context.TODO(), cluster.InspectParametersReq{
			InstanceID: "1",
		})
		assert.NoError(t, err)
		assert.Equal(t, len(resp.Params), 1)
	})

	t.Run("inspect pd instance parameter", func(t *testing.T) {
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)
		mockPDApiService := mockutilpd.NewMockPDApiService(ctrl)
		pd.ApiService = mockPDApiService

		mockPDApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, nil)
		clusterManagementRW.EXPECT().GetInstance(gomock.Any(), gomock.Any()).
			Return(&management.ClusterInstance{
				Entity:    common.Entity{ID: "1", Status: string(constants.ClusterInstanceRunning)},
				Type:      "PD",
				Version:   "v5.0.0",
				ClusterID: "123",
				HostIP:    []string{"127.0.0.1"},
				Ports:     []int32{10000, 10001},
			}, nil)
		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
				return "1", []*parameter.ClusterParamDetail{
					{
						Parameter: parametergroup.Parameter{
							ID:           "3",
							Category:     "log.file",
							Name:         "max-size",
							InstanceType: "PD",
							HasApply:     1,
							Type:         0,
						},
						RealValue: "{\"clusterValue\":\"102400\"}",
					},
				}, 1, nil
			})

		resp, err := mockManager.InspectClusterParameters(context.TODO(), cluster.InspectParametersReq{
			InstanceID: "1",
		})
		assert.NoError(t, err)
		assert.Equal(t, len(resp.Params), 1)
	})

	t.Run("inspect cdc and tiflash instance parameter", func(t *testing.T) {
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(string(tomlConfigContent), nil)

		clusterManagementRW.EXPECT().GetInstance(gomock.Any(), gomock.Any()).
			Return(&management.ClusterInstance{
				Entity:    common.Entity{ID: "1", Status: string(constants.ClusterInstanceRunning)},
				Type:      "CDC",
				Version:   "v5.0.0",
				ClusterID: "123",
				HostIP:    []string{"127.0.0.1"},
				Ports:     []int32{10000, 10001},
			}, nil)
		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
				return "1", []*parameter.ClusterParamDetail{
					{
						Parameter: parametergroup.Parameter{
							ID:           "4",
							Category:     "basic",
							Name:         "quota-backend-bytes",
							InstanceType: "CDC",
							HasApply:     1,
							Type:         0,
						},
						RealValue: "{\"clusterValue\":\"8589934592\"}",
					},
				}, 1, nil
			})

		resp, err := mockManager.InspectClusterParameters(context.TODO(), cluster.InspectParametersReq{
			InstanceID: "1",
		})
		assert.NoError(t, err)
		assert.Equal(t, len(resp.Params), 1)
	})

	t.Run("query cluster parameter error", func(t *testing.T) {
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)

		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
				return "1", []*parameter.ClusterParamDetail{}, 1, errors.New("query cluster parameter error")
			})

		_, err := mockManager.InspectClusterParameters(context.TODO(), cluster.InspectParametersReq{
			ClusterID: "123",
		})
		assert.Error(t, err)
	})

	t.Run("get meta error", func(t *testing.T) {
		clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
		models.SetClusterParameterReaderWriter(clusterParameterRW)
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)

		clusterParameterRW.EXPECT().QueryClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId, name, instanceType string, offset, size int) (paramGroupId string, params []*parameter.ClusterParamDetail, total int64, err error) {
				return "1", []*parameter.ClusterParamDetail{
					{
						Parameter: parametergroup.Parameter{
							ID:             "1",
							Category:       "basic",
							Name:           "param1",
							InstanceType:   "TiKV",
							HasApply:       1,
							SystemVariable: "",
							Type:           0,
						},
						DefaultValue: "10",
						RealValue:    "{\"clusterValue\":\"1\"}",
					},
				}, 3, nil
			})
		clusterManagementRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterID string) (*management.Cluster, []*management.ClusterInstance, []*management.DBUser, error) {
				return mockCluster(), mockClusterInstances(), mockDBUsers(), errors.New("get meta fail")
			})

		_, err := mockManager.InspectClusterParameters(context.TODO(), cluster.InspectParametersReq{
			ClusterID: "123",
		})
		assert.Error(t, err)
	})

	t.Run("id is empty", func(t *testing.T) {
		_, err := mockManager.InspectClusterParameters(context.TODO(), cluster.InspectParametersReq{
			InstanceID: "",
			ClusterID:  "",
		})
		assert.Error(t, err)
	})
}

func TestManager_inspectTiDBComponentInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	instances := mockClusterInstances()
	t.Run("request tidb api error", func(t *testing.T) {
		mockTiDBApiService := mockutiltidbhttp.NewMockTiDBApiService(ctrl)
		http.ApiService = mockTiDBApiService

		mockTiDBApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, errors.New("request tidb api error"))

		_, err := inspectTiDBComponentInstance(context.TODO(), instances[0], []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
			},
		})
		assert.Error(t, err)
	})
}

func TestManager_inspectTiKVComponentInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	instances := mockClusterInstances()
	t.Run("request tikv api error", func(t *testing.T) {
		mockTiKVApiService := mockutiltikv.NewMockTiKVApiService(ctrl)
		tikv.ApiService = mockTiKVApiService

		mockTiKVApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, errors.New("request tikv api error"))

		_, err := inspectTiKVComponentInstance(context.TODO(), instances[0], []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "TiKV",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
			},
		})
		assert.Error(t, err)
	})
}

func TestManager_inspectPDComponentInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	instances := mockClusterInstances()
	t.Run("request pd api error", func(t *testing.T) {
		mockPDApiService := mockutilpd.NewMockPDApiService(ctrl)
		pd.ApiService = mockPDApiService

		mockPDApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, errors.New("request pd api error"))

		_, err := inspectPDComponentInstance(context.TODO(), instances[0], []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "PD",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
			},
		})
		assert.Error(t, err)
	})
}

func TestManager_inspectInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	instances := mockClusterInstances()
	t.Run("inspect instances success", func(t *testing.T) {
		inspectInstanceInfo, err := inspectInstancesByApiAndConfig(context.TODO(), instances[0], apiContent, []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, inspectInstanceInfo)
		assert.Equal(t, len(inspectInstanceInfo.ParameterInfos), 0)
	})
	t.Run("inspect instances diff", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService
		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(string(tomlConfigContent), nil)

		inspectInstanceInfo, err := inspectInstancesByApiAndConfig(context.TODO(), instances[0], apiContent, []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
			},
			{
				ParamId:      "2",
				Category:     "basic",
				Name:         "auto-compaction-retention",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "1h"},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, inspectInstanceInfo)
		assert.Equal(t, len(inspectInstanceInfo.ParameterInfos), 0)
	})
	t.Run("unmarshal error", func(t *testing.T) {
		apiContent := []byte(`Unmarshal error`)
		_, err := inspectInstancesByApiAndConfig(context.TODO(), instances[0], apiContent, []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
			},
		})
		assert.Error(t, err)
	})
}

func TestManager_inspectApiParameter(t *testing.T) {
	instances := mockClusterInstances()
	t.Run("unmarshal error", func(t *testing.T) {
		apiContent := []byte(`Unmarshal error`)
		_, _, err := inspectApiParameter(context.TODO(), instances[0], apiContent, []structs.ClusterParameterInfo{})
		assert.Error(t, err)
	})

	t.Run("inspect value equals", func(t *testing.T) {
		inspectApiParams, instDiffParams, err := inspectApiParameter(context.TODO(), instances[0], apiContent, []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
			},
			{
				ParamId:      "2",
				Category:     "security",
				Name:         "enable-sem",
				InstanceType: "TiDB",
				Type:         2,
				RealValue:    structs.ParameterRealValue{ClusterValue: "true"},
			},
			{
				ParamId:      "3",
				Category:     "log.file",
				Name:         "max-size",
				InstanceType: "TiDB",
				Type:         0,
				RealValue:    structs.ParameterRealValue{ClusterValue: "102400"},
			},
			{
				ParamId:      "4",
				Category:     "coprocessor",
				Name:         "region-max-size",
				InstanceType: "TiDB",
				Type:         1,
				Range:        []string{"KB", "MB", "GB"},
				RangeType:    1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "144MB"},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, len(inspectApiParams), 0)
		assert.Equal(t, len(instDiffParams), 0)
	})
	t.Run("inspect value not equals", func(t *testing.T) {
		inspectApiParams, instDiffParams, err := inspectApiParameter(context.TODO(), instances[0], apiContent, []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
			},
			{
				ParamId:      "2",
				Category:     "security",
				Name:         "enable-sem",
				InstanceType: "TiDB",
				Type:         2,
				RealValue:    structs.ParameterRealValue{ClusterValue: "false"},
			},
			{
				ParamId:      "3",
				Category:     "log.file",
				Name:         "max-size",
				InstanceType: "TiDB",
				Type:         0,
				RealValue:    structs.ParameterRealValue{ClusterValue: "102400"},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, len(inspectApiParams), 1)
		assert.Equal(t, len(instDiffParams), 0)
	})
	t.Run("inspect diff parameter", func(t *testing.T) {
		inspectApiParams, instDiffParams, err := inspectApiParameter(context.TODO(), instances[0], apiContent, []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
			},
			{
				ParamId:      "2",
				Category:     "security",
				Name:         "enable-sem",
				InstanceType: "TiDB",
				Type:         2,
				RealValue:    structs.ParameterRealValue{ClusterValue: "true"},
			},
			{
				ParamId:      "3",
				Category:     "log.file",
				Name:         "max-size",
				InstanceType: "TiDB",
				Type:         0,
				RealValue:    structs.ParameterRealValue{ClusterValue: "102400"},
			},
			{
				ParamId:      "4",
				Category:     "log.file",
				Name:         "min-size",
				InstanceType: "TiDB",
				Type:         0,
				RealValue:    structs.ParameterRealValue{ClusterValue: "1024"},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, len(inspectApiParams), 0)
		assert.Equal(t, len(instDiffParams), 1)
	})
	t.Run("convert real parameter error", func(t *testing.T) {
		_, _, err := inspectApiParameter(context.TODO(), instances[0], apiContent, []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "basic",
				Name:         "oom-action",
				InstanceType: "TiDB",
				Type:         0,
				RealValue:    structs.ParameterRealValue{ClusterValue: "log"},
			},
		})
		assert.Error(t, err)
	})
}

func TestManager_inspectConfigParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	instances := mockClusterInstances()

	t.Run("inspect value equals", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService
		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(string(tomlConfigContent), nil)

		inspectParams, err := inspectConfigParameter(context.TODO(), instances[0], []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "dashboard",
				Name:         "public-path-prefix",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "/dashboard"},
			},
			{
				ParamId:      "2",
				Category:     "label-property",
				Name:         "key",
				InstanceType: "TiDB",
				Type:         4,
				RealValue:    structs.ParameterRealValue{ClusterValue: "[\"test\"]"},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, len(inspectParams), 0)
	})

	t.Run("inspect value not equals", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService
		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(string(tomlConfigContent), nil)

		inspectParams, err := inspectConfigParameter(context.TODO(), instances[0], []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "dashboard",
				Name:         "public-path-prefix",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "/dashboards"},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, len(inspectParams), 1)
	})

	t.Run("request pull error", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService
		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(string(tomlConfigContent), errors.New("request pull error"))

		_, err := inspectConfigParameter(context.TODO(), instances[0], []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "dashboard",
				Name:         "public-path-prefix",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "/dashboards"},
			},
		})
		assert.Error(t, err)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService
		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return("unmarshal error", nil)

		_, err := inspectConfigParameter(context.TODO(), instances[0], []structs.ClusterParameterInfo{
			{
				ParamId:      "1",
				Category:     "dashboard",
				Name:         "public-path-prefix",
				InstanceType: "TiDB",
				Type:         1,
				RealValue:    structs.ParameterRealValue{ClusterValue: "/dashboard"},
			},
		})
		assert.Error(t, err)
	})
}
