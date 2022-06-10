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
 * @Description: workflow executor unit test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/16 11:19
*******************************************************************************/

package parameter

import (
	"context"
	"errors"
	workflowModel "github.com/pingcap-inc/tiunimanager/models/workflow"
	"math"
	"testing"

	"github.com/pingcap-inc/tiunimanager/micro-cluster/cluster/management/meta"

	"github.com/pingcap-inc/tiunimanager/common/constants"

	"github.com/pingcap-inc/tiunimanager/deployment"

	"github.com/pingcap-inc/tiunimanager/util/api/cdc"
	"github.com/pingcap-inc/tiunimanager/util/api/pd"
	"github.com/pingcap-inc/tiunimanager/util/api/tidb/http"
	"github.com/pingcap-inc/tiunimanager/util/api/tidb/sql"
	"github.com/pingcap-inc/tiunimanager/util/api/tikv"

	"github.com/pingcap-inc/tiunimanager/test/mockutilcdc"
	"github.com/pingcap-inc/tiunimanager/test/mockutilpd"
	"github.com/pingcap-inc/tiunimanager/test/mockutiltidbhttp"
	mockutiltidbsqlconfig "github.com/pingcap-inc/tiunimanager/test/mockutiltidbsql_config"
	"github.com/pingcap-inc/tiunimanager/test/mockutiltikv"

	"github.com/pingcap-inc/tiunimanager/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiunimanager/test/mockmodels/mockconfig"
	mock_workflow_service "github.com/pingcap-inc/tiunimanager/test/mockworkflow"

	"github.com/pingcap-inc/tiunimanager/common/structs"

	"github.com/pingcap-inc/tiunimanager/models/cluster/parameter"

	"github.com/pingcap-inc/tiunimanager/test/mockmodels/mockclusterparameter"

	"github.com/pingcap-inc/tiunimanager/models"
	"github.com/pingcap-inc/tiunimanager/test/mockmodels/mockparametergroup"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	mock_deployment "github.com/pingcap-inc/tiunimanager/test/mockdeployment"
	workflow "github.com/pingcap-inc/tiunimanager/workflow2"
)

func TestExecutor_asyncMaintenance_Success(t *testing.T) {
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

		clusterManagementRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, bizId string, bizType string, flowName string) (string, error) {
				return "flowId", nil
			})
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes()
		configRW.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).AnyTimes()

		data := map[string]interface{}{}
		data[contextModifyParameters] = mockModifyParameter()
		data[contextMaintenanceStatusChange] = true
		resp, err := asyncMaintenance(context.TODO(), mockClusterMeta(), data, constants.ClusterMaintenanceModifyParameterAndRestarting, modifyParametersDefine.FlowName)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	})

	t.Run("start maintenance fail", func(t *testing.T) {
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)

		clusterManagementRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
			Return(errors.New("set maintenance fail"))
		data := map[string]interface{}{}
		data[contextModifyParameters] = mockModifyParameter()
		data[contextMaintenanceStatusChange] = true
		_, err := asyncMaintenance(context.TODO(), mockClusterMeta(), data, constants.ClusterMaintenanceModifyParameterAndRestarting, modifyParametersDefine.FlowName)
		assert.Error(t, err)
	})

	t.Run("create workflow fail", func(t *testing.T) {
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)

		clusterManagementRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, bizId string, bizType string, flowName string) (string, error) {
				return "flowId", errors.New("create workflow fail")
			})
		data := map[string]interface{}{}
		data[contextModifyParameters] = mockModifyParameter()
		data[contextMaintenanceStatusChange] = true
		_, err := asyncMaintenance(context.TODO(), mockClusterMeta(), data, constants.ClusterMaintenanceModifyParameterAndRestarting, modifyParametersDefine.FlowName)
		assert.Error(t, err)
	})

	t.Run("async start fail", func(t *testing.T) {
		clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterManagementRW)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)

		clusterManagementRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, bizId string, bizType string, flowName string) (string, error) {
				return "flowId", nil
			})
		workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("async start fail"))
		data := map[string]interface{}{}
		data[contextModifyParameters] = mockModifyParameter()
		data[contextMaintenanceStatusChange] = true
		_, err := asyncMaintenance(context.TODO(), mockClusterMeta(), data, constants.ClusterMaintenanceModifyParameterAndRestarting, modifyParametersDefine.FlowName)
		assert.Error(t, err)
	})
}

func TestExecutor_endMaintenance_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().ClearMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		ctx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		ctx.SetData(contextClusterMeta, mockClusterMeta())
		ctx.SetData(contextModifyParameters, mockModifyParameter())
		ctx.SetData(contextMaintenanceStatusChange, true)
		err := defaultEnd(&workflowModel.WorkFlowNode{}, ctx)
		assert.NoError(t, err)
	})
}

func TestExecutor_convertRealParameterType_Success(t *testing.T) {
	applyCtx := &workflow.FlowContext{
		Context:  context.TODO(),
		FlowData: make(map[string]string),
	}
	v, err := convertRealParameterType(applyCtx, 0, "1")
	assert.NoError(t, err)
	assert.EqualValues(t, 1, v)

	v, err = convertRealParameterType(applyCtx, 0, "1.44e+06")
	assert.NoError(t, err)
	assert.EqualValues(t, 1440000, v)

	v, err = convertRealParameterType(applyCtx, 0, "123abc")
	assert.Error(t, err)

	v, err = convertRealParameterType(applyCtx, 1, "debug")
	assert.NoError(t, err)
	assert.EqualValues(t, "debug", v)

	v, err = convertRealParameterType(applyCtx, 2, "true")
	assert.NoError(t, err)
	assert.EqualValues(t, true, v)

	v, err = convertRealParameterType(applyCtx, 3, "3.00")
	assert.NoError(t, err)
	assert.EqualValues(t, 3.00, math.Trunc(v.(float64)))

	v, err = convertRealParameterType(applyCtx, 4, "[\"debug\",\"info\"]")
	assert.NoError(t, err)
	expect := []interface{}{"debug", "info"}
	assert.EqualValues(t, expect, v)
}

func TestExecutor_convertRealParameterType_Error(t *testing.T) {
	applyCtx := &workflow.FlowContext{
		Context:  context.TODO(),
		FlowData: make(map[string]string),
	}

	_, err := convertRealParameterType(applyCtx, 2, "debug")
	assert.Error(t, err)

	_, err = convertRealParameterType(applyCtx, 3, "true")
	assert.Error(t, err)

	_, err = convertRealParameterType(applyCtx, 0, "[\"debug\",\"info\"]")
	assert.Error(t, err)
}

func TestExecutor_validationParameters(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, mockModifyParameter())

		err := validationParameter(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "TiDB",
					UpdateSource:   0,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"s", "e"},
					RangeType:      1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
		})

		err := validationParameter(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.Error(t, err)
	})
}

func TestExecutor_validateRange(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		validated := ValidateRange(&ModifyClusterParameterInfo{
			ParamId:        "1",
			Name:           "test_param_1",
			InstanceType:   "TiDB",
			UpdateSource:   0,
			HasApply:       1,
			SystemVariable: "",
			Type:           0,
			Range:          []string{"1", "10"},
			RangeType:      1,
			RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
		}, true)
		assert.EqualValues(t, true, validated)

		validated = ValidateRange(&ModifyClusterParameterInfo{
			ParamId:        "2",
			Name:           "test_param_2",
			InstanceType:   "TiDB",
			UpdateSource:   0,
			HasApply:       1,
			SystemVariable: "",
			Type:           1,
			Range:          []string{"debug", "info", "warn", "error"},
			RangeType:      2,
			RealValue:      structs.ParameterRealValue{ClusterValue: "info"},
		}, true)
		assert.EqualValues(t, true, validated)

		validated = ValidateRange(&ModifyClusterParameterInfo{
			ParamId:        "3",
			Name:           "test_param_3",
			InstanceType:   "TiDB",
			UpdateSource:   0,
			HasApply:       1,
			SystemVariable: "",
			Type:           2,
			Range:          []string{"true", "false"},
			RangeType:      2,
			RealValue:      structs.ParameterRealValue{ClusterValue: "true"},
		}, true)
		assert.EqualValues(t, true, validated)

		validated = ValidateRange(&ModifyClusterParameterInfo{
			ParamId:        "4",
			Name:           "test_param_4",
			InstanceType:   "TiDB",
			UpdateSource:   0,
			HasApply:       1,
			SystemVariable: "",
			Type:           3,
			Range:          []string{"0.1", "5.2"},
			RangeType:      1,
			RealValue:      structs.ParameterRealValue{ClusterValue: "3.14"},
		}, true)

		validated = ValidateRange(&ModifyClusterParameterInfo{
			ParamId:        "5",
			Name:           "test_param_5",
			InstanceType:   "TiDB",
			UpdateSource:   0,
			HasApply:       1,
			SystemVariable: "",
			Type:           4,
			Range:          []string{"[]", "[]"},
			RangeType:      0,
			RealValue:      structs.ParameterRealValue{ClusterValue: "[]"},
		}, true)
		assert.EqualValues(t, true, validated)
	})
}

func TestExecutor_modifyParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mockCDCApiService := mockutilcdc.NewMockCDCApiService(ctrl)
		cdc.ApiService = mockCDCApiService
		mockPDApiService := mockutilpd.NewMockPDApiService(ctrl)
		pd.ApiService = mockPDApiService
		mockTiDBApiService := mockutiltidbhttp.NewMockTiDBApiService(ctrl)
		http.ApiService = mockTiDBApiService
		mockTiDBSqlConfigService := mockutiltidbsqlconfig.NewMockClusterConfigService(ctrl)
		sql.SqlService = mockTiDBSqlConfigService
		mockTiKVApiService := mockutiltikv.NewMockTiKVApiService(ctrl)
		tikv.ApiService = mockTiKVApiService

		mock2rdService.EXPECT().EditConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("1", nil)
		mock2rdService.EXPECT().ShowConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)

		mockPDApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, nil)
		mockTiDBApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, nil)
		mockTiKVApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, nil)
		mockCDCApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, nil)
		mockTiDBSqlConfigService.EXPECT().EditClusterConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, mockModifyParameter())
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.NoError(t, err)
	})

	t.Run("no cdc instance apply parameter", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mockCDCApiService := mockutilcdc.NewMockCDCApiService(ctrl)
		cdc.ApiService = mockCDCApiService

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, &meta.ClusterMeta{Cluster: mockCluster()})
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "7",
					Name:           "test_param_7",
					InstanceType:   "CDC",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "info"},
				},
			},
		})
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.NoError(t, err)
	})

	t.Run("no cdc instance modify parameter", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mockCDCApiService := mockutilcdc.NewMockCDCApiService(ctrl)
		cdc.ApiService = mockCDCApiService

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, &meta.ClusterMeta{Cluster: mockCluster()})
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "7",
					Name:           "test_param_7",
					InstanceType:   "CDC",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "info"},
				},
			},
		})
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.Error(t, err)
	})

	t.Run("no tiflash instance apply parameter", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mockCDCApiService := mockutilcdc.NewMockCDCApiService(ctrl)
		cdc.ApiService = mockCDCApiService

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, &meta.ClusterMeta{Cluster: mockCluster()})
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "7",
					Name:           "test_param_7",
					InstanceType:   "TiFlash",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "info"},
				},
			},
		})
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.NoError(t, err)
	})

	t.Run("no tiflash instance modify parameter", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mockCDCApiService := mockutilcdc.NewMockCDCApiService(ctrl)
		cdc.ApiService = mockCDCApiService

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, &meta.ClusterMeta{Cluster: mockCluster()})
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "7",
					Name:           "test_param_7",
					InstanceType:   "TiFlash",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "info"},
				},
			},
		})
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.Error(t, err)
	})

	t.Run("read only parameter not modify", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextHasApplyParameter, false)
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Reboot: false,
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "TiDB",
					UpdateSource:   0,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"0", "1024"},
					RangeType:      1,
					ReadOnly:       1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
			Nodes: []string{"172.16.1.12:9000"},
		})
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.Error(t, err)
	})

	t.Run("no tiflash apply parameter", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mock2rdService.EXPECT().EditConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("1", nil)
		mock2rdService.EXPECT().ShowConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Reboot: false,
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "TiFlash",
					UpdateSource:   0,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"0", "1024"},
					RangeType:      1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
			Nodes: []string{"172.16.1.12:9000"},
		})
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.NoError(t, err)
	})

	t.Run("no tiflash modify parameter", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mock2rdService.EXPECT().EditConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("1", nil)
		mock2rdService.EXPECT().ShowConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Reboot: false,
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "TiFlash",
					UpdateSource:   0,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"0", "1024"},
					RangeType:      1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
			Nodes: []string{"172.16.1.12:9000"},
		})
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.NoError(t, err)
	})

	t.Run("no cdc apply parameter", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mockCDCApiService := mockutilcdc.NewMockCDCApiService(ctrl)
		cdc.ApiService = mockCDCApiService

		mockCDCApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, nil)

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Reboot: false,
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "CDC",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"0", "1024"},
					RangeType:      1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
			Nodes: []string{"172.16.1.12:9000"},
		})
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.NoError(t, err)
	})

	t.Run("no cdc modify parameter", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mockCDCApiService := mockutilcdc.NewMockCDCApiService(ctrl)
		cdc.ApiService = mockCDCApiService

		mockCDCApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, nil)

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Reboot: false,
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "CDC",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"0", "1024"},
					RangeType:      1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
			Nodes: []string{"172.16.1.12:9000"},
		})
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.NoError(t, err)
	})

	t.Run("tidb api modify fail", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService
		mockTiDBApiService := mockutiltidbhttp.NewMockTiDBApiService(ctrl)
		http.ApiService = mockTiDBApiService

		mockTiDBApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, errors.New("update fail"))

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Reboot: false,
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "TiDB",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"0", "1024"},
					RangeType:      1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
			Nodes: []string{"172.16.1.12:9000"},
		})
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.Error(t, err)
	})

	t.Run("tikv api modify fail", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService
		mockTiKVApiService := mockutiltikv.NewMockTiKVApiService(ctrl)
		tikv.ApiService = mockTiKVApiService

		mockTiKVApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, errors.New("update fail"))

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Reboot: false,
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "TiKV",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"0", "1024"},
					RangeType:      1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
			Nodes: []string{"172.16.1.12:9000"},
		})
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.Error(t, err)
	})

	t.Run("pd api modify fail", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService
		mockPDApiService := mockutilpd.NewMockPDApiService(ctrl)
		pd.ApiService = mockPDApiService

		mockPDApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, errors.New("update fail"))

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Reboot: false,
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "PD",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"0", "1024"},
					RangeType:      1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
			Nodes: []string{"172.16.1.12:9000"},
		})
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.Error(t, err)
	})

	t.Run("cdc api modify fail", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService
		mockCDCApiService := mockutilcdc.NewMockCDCApiService(ctrl)
		cdc.ApiService = mockCDCApiService

		mockCDCApiService.EXPECT().EditConfig(gomock.Any(), gomock.Any()).Return(true, errors.New("update fail"))

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, &ModifyParameter{
			Reboot: false,
			Params: []*ModifyClusterParameterInfo{
				{
					ParamId:        "1",
					Name:           "test_param_1",
					InstanceType:   "CDC",
					UpdateSource:   3,
					HasApply:       1,
					SystemVariable: "",
					Type:           0,
					Range:          []string{"0", "1024"},
					RangeType:      1,
					RealValue:      structs.ParameterRealValue{ClusterValue: "1"},
				},
			},
			Nodes: []string{"172.16.1.12:9000"},
		})
		err := modifyParameters(&workflowModel.WorkFlowNode{}, modifyCtx)
		assert.Error(t, err)
	})
}

func TestDefaultFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock2rdService := mock_deployment.NewMockInterface(ctrl)
	deployment.M = mock2rdService

	t.Run("success", func(t *testing.T) {
		mock2rdService.EXPECT().EditConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("124", nil)

		refreshCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		refreshCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyParameter := mockModifyParameter()
		modifyParameter.Reboot = true
		refreshCtx.SetData(contextModifyParameters, modifyParameter)
		refreshCtx.SetData(contextClusterConfigStr, "user: tiem\ntiem_version: v1.0.0-beta.7\ntopology:\n  global:\n    user: tiem\n    group: tiem\n")
		err := parameterFail(&workflowModel.WorkFlowNode{}, refreshCtx)
		assert.NoError(t, err)
	})

	t.Run("reload fail rollback", func(t *testing.T) {
		mock2rdService.EXPECT().EditConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return("124", errors.New("edit config fail"))

		refreshCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		refreshCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyParameter := mockModifyParameter()
		modifyParameter.Reboot = true
		refreshCtx.SetData(contextModifyParameters, modifyParameter)
		refreshCtx.SetData(contextClusterConfigStr, "user: tiem\ntiem_version: v1.0.0-beta.7\ntopology:\n  global:\n    user: tiem\n    group: tiem\n")
		err := parameterFail(&workflowModel.WorkFlowNode{}, refreshCtx)
		assert.Error(t, err)
	})
}

func TestExecutor_refreshParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock2rdService := mock_deployment.NewMockInterface(ctrl)
	deployment.M = mock2rdService

	t.Run("success", func(t *testing.T) {
		mock2rdService.EXPECT().Reload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("123", nil)

		refreshCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		refreshCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyParameter := mockModifyParameter()
		modifyParameter.Reboot = true
		refreshCtx.SetData(contextModifyParameters, modifyParameter)
		err := refreshParameter(&workflowModel.WorkFlowNode{}, refreshCtx)
		assert.NoError(t, err)
	})

	t.Run("success2", func(t *testing.T) {
		refreshCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		refreshCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyParameter := mockModifyParameter()
		modifyParameter.Reboot = false
		refreshCtx.SetData(contextModifyParameters, modifyParameter)
		err := refreshParameter(&workflowModel.WorkFlowNode{}, refreshCtx)
		assert.NoError(t, err)
	})
}

func TestExecutor_persistParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		clusterParameterRW.EXPECT().ApplyClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, parameterGroupId string, clusterId string, param []*parameter.ClusterParameterMapping) error {
				return nil
			})

		applyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		applyCtx.SetData(contextClusterMeta, mockClusterMeta())
		applyCtx.SetData(contextModifyParameters, mockModifyParameter())
		applyCtx.SetData(contextHasApplyParameter, true)
		err := persistParameter(&workflowModel.WorkFlowNode{}, applyCtx)
		assert.NoError(t, err)
	})
}

func TestExecutor_persistParameter2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		clusterParameterRW.EXPECT().UpdateClusterParameter(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		applyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		applyCtx.SetData(contextClusterMeta, mockClusterMeta())
		applyCtx.SetData(contextModifyParameters, mockModifyParameter())
		applyCtx.SetData(contextHasApplyParameter, false)
		err := persistParameter(&workflowModel.WorkFlowNode{}, applyCtx)
		assert.NoError(t, err)
	})
}

func TestManager_fillParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("fill parameter success", func(t *testing.T) {
		mockTiDBApiService := mockutiltidbhttp.NewMockTiDBApiService(ctrl)
		http.ApiService = mockTiDBApiService
		mockTiKVApiService := mockutiltikv.NewMockTiKVApiService(ctrl)
		tikv.ApiService = mockTiKVApiService
		mockPDApiService := mockutilpd.NewMockPDApiService(ctrl)
		pd.ApiService = mockPDApiService
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mockTiDBApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, nil)
		mockTiKVApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, nil)
		mockPDApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, nil)
		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(string(tomlConfigContent), nil)

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := fillParameters(modifyCtx, map[interface{}][]*ModifyClusterParameterInfo{
			"TiDB": {
				{
					ParamId:      "1",
					Category:     "coprocessor",
					Name:         "region-max-size",
					InstanceType: "TiDB",
					Type:         1,
					RealValue:    structs.ParameterRealValue{ClusterValue: "144MiB"},
				},
			},
			"TiKV": {
				{
					ParamId:      "2",
					Category:     "security",
					Name:         "enable-sem",
					InstanceType: "TiKV",
					HasApply:     1,
					Type:         2,
					RealValue:    structs.ParameterRealValue{ClusterValue: "true"},
				},
			},
			"PD": {
				{
					ParamId:      "3",
					Category:     "log.file",
					Name:         "max-size",
					InstanceType: "PD",
					HasApply:     1,
					Type:         0,
					RealValue:    structs.ParameterRealValue{ClusterValue: "102400"},
				},
				{
					ParamId:      "5",
					Category:     "basic",
					Name:         "quota-backend-bytes",
					InstanceType: "PD",
					HasApply:     1,
					Type:         0,
					RealValue:    structs.ParameterRealValue{ClusterValue: ""},
				},
			},
			"TiFlash": {
				{
					ParamId:      "4",
					Category:     "basic",
					Name:         "quota-backend-bytes",
					InstanceType: "TiFlash",
					HasApply:     1,
					Type:         0,
					RealValue:    structs.ParameterRealValue{ClusterValue: "8589934592"},
				},
			},
		})
		assert.NoError(t, err)
	})

	t.Run("request tidb api error", func(t *testing.T) {
		mockTiDBApiService := mockutiltidbhttp.NewMockTiDBApiService(ctrl)
		http.ApiService = mockTiDBApiService

		mockTiDBApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, errors.New("request tidb api error"))

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := fillParameters(modifyCtx, map[interface{}][]*ModifyClusterParameterInfo{
			"TiDB": {
				{
					ParamId:      "1",
					Category:     "basic",
					Name:         "oom-action",
					InstanceType: "TiDB",
					Type:         1,
					RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("request tikv api error", func(t *testing.T) {
		mockTiKVApiService := mockutiltikv.NewMockTiKVApiService(ctrl)
		tikv.ApiService = mockTiKVApiService

		mockTiKVApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, errors.New("request tikv api error"))

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := fillParameters(modifyCtx, map[interface{}][]*ModifyClusterParameterInfo{
			"TiKV": {
				{
					ParamId:      "1",
					Category:     "basic",
					Name:         "oom-action",
					InstanceType: "TiDB",
					Type:         1,
					RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("request pd api error", func(t *testing.T) {
		mockPDApiService := mockutilpd.NewMockPDApiService(ctrl)
		pd.ApiService = mockPDApiService

		mockPDApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return(apiContent, errors.New("request pd api error"))

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := fillParameters(modifyCtx, map[interface{}][]*ModifyClusterParameterInfo{
			"PD": {
				{
					ParamId:      "1",
					Category:     "basic",
					Name:         "oom-action",
					InstanceType: "TiDB",
					Type:         1,
					RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("request tiflash config error", func(t *testing.T) {
		mock2rdService := mock_deployment.NewMockInterface(ctrl)
		deployment.M = mock2rdService

		mock2rdService.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(string(tomlConfigContent), errors.New("request tiflash config error"))

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := fillParameters(modifyCtx, map[interface{}][]*ModifyClusterParameterInfo{
			"TiFlash": {
				{
					ParamId:      "1",
					Category:     "basic",
					Name:         "oom-action",
					InstanceType: "TiFlash",
					Type:         1,
					RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("json decode error", func(t *testing.T) {
		mockTiDBApiService := mockutiltidbhttp.NewMockTiDBApiService(ctrl)
		http.ApiService = mockTiDBApiService

		mockTiDBApiService.EXPECT().ShowConfig(gomock.Any(), gomock.Any()).Return([]byte("json decode error"), nil)

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: make(map[string]string),
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextHasApplyParameter, true)
		err := fillParameters(modifyCtx, map[interface{}][]*ModifyClusterParameterInfo{
			"TiDB": {
				{
					ParamId:      "1",
					Category:     "basic",
					Name:         "oom-action",
					InstanceType: "TiDB",
					Type:         1,
					RealValue:    structs.ParameterRealValue{ClusterValue: "cancel"},
				},
			},
		})
		assert.Error(t, err)
	})
}
