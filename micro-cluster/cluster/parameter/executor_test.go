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
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"

	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockconfig"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"

	"github.com/pingcap-inc/tiem/common/structs"

	"github.com/pingcap-inc/tiem/message/cluster"

	"github.com/pingcap-inc/tiem/models/cluster/parameter"

	"github.com/pingcap-inc/tiem/test/mockmodels/mockclusterparameter"

	"github.com/pingcap-inc/tiem/models/parametergroup"

	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockparametergroup"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message"
	secondparty2 "github.com/pingcap-inc/tiem/models/workflow/secondparty"
	mock_secondparty_v2 "github.com/pingcap-inc/tiem/test/mocksecondparty_v2"
	"github.com/pingcap-inc/tiem/workflow"
)

func TestExecutor_asyncMaintenance(t *testing.T) {
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

	clusterManagementRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, bizId string, flowName string) (*workflow.WorkFlowAggregation, error) {
			return mockWorkFlowAggregation(), nil
		})
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).AnyTimes()
	configRW.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).AnyTimes()

	data := map[string]interface{}{}
	data[contextModifyParameters] = mockModifyParameter()
	resp, err := asyncMaintenance(context.TODO(), mockClusterMeta(), data, constants.ClusterMaintenanceModifyParameterAndRestarting, modifyParametersDefine.FlowName)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

func TestExecutor_endMaintenance(t *testing.T) {
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
			FlowData: map[string]interface{}{},
		}
		ctx.SetData(contextClusterMeta, mockClusterMeta())
		ctx.SetData(contextModifyParameters, mockModifyParameter())
		err := endMaintenance(mockWorkFlowAggregation().CurrentNode, ctx)
		assert.NoError(t, err)
	})
}

func TestExecutor_setClusterFailure_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		ctx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		ctx.SetData(contextClusterMeta, mockClusterMeta())
		ctx.SetData(contextModifyParameters, mockModifyParameter())
		err := setClusterFailure(mockWorkFlowAggregation().CurrentNode, ctx)
		assert.NoError(t, err)
	})
}
func TestExecutor_setClusterFailure_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("set cluster status failed"))

		ctx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		ctx.SetData(contextClusterMeta, mockClusterMeta())
		ctx.SetData(contextModifyParameters, mockModifyParameter())
		err := setClusterFailure(mockWorkFlowAggregation().CurrentNode, ctx)
		assert.Error(t, err)
	})
}

func TestExecutor_setClusterOnline_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		ctx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		ctx.SetData(contextClusterMeta, mockClusterMeta())
		ctx.SetData(contextModifyParameters, mockModifyParameter())
		err := setClusterOnline(mockWorkFlowAggregation().CurrentNode, ctx)
		assert.NoError(t, err)
	})
}

func TestExecutor_setClusterOnline_Error1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("set cluster status failed"))

		ctx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		clusterMeta := mockClusterMeta()
		clusterMeta.Cluster.Status = string(constants.ClusterInitializing)
		ctx.SetData(contextClusterMeta, clusterMeta)
		ctx.SetData(contextModifyParameters, mockModifyParameter())
		err := setClusterOnline(mockWorkFlowAggregation().CurrentNode, ctx)
		assert.Error(t, err)
	})
}

func TestExecutor_setClusterOnline_Error2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterManagementRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterManagementRW)
	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		clusterManagementRW.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("set cluster status failed"))

		ctx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		ctx.SetData(contextClusterMeta, mockClusterMeta())
		ctx.SetData(contextModifyParameters, mockModifyParameter())
		err := setClusterOnline(mockWorkFlowAggregation().CurrentNode, ctx)
		assert.Error(t, err)
	})
}

func TestExecutor_convertRealParameterType(t *testing.T) {
	applyCtx := &workflow.FlowContext{
		Context:  context.TODO(),
		FlowData: map[string]interface{}{},
	}

	v, err := convertRealParameterType(applyCtx, structs.ClusterParameterSampleInfo{
		ParamId:   "1",
		Name:      "param1",
		Type:      0,
		RealValue: structs.ParameterRealValue{ClusterValue: "1"},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, v)

	v, err = convertRealParameterType(applyCtx, structs.ClusterParameterSampleInfo{
		ParamId:   "2",
		Name:      "param2",
		Type:      1,
		RealValue: structs.ParameterRealValue{ClusterValue: "debug"},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, "debug", v)

	v, err = convertRealParameterType(applyCtx, structs.ClusterParameterSampleInfo{
		ParamId:   "3",
		Name:      "param3",
		Type:      2,
		RealValue: structs.ParameterRealValue{ClusterValue: "true"},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, true, v)

	v, err = convertRealParameterType(applyCtx, structs.ClusterParameterSampleInfo{
		ParamId:   "4",
		Name:      "param4",
		Type:      3,
		RealValue: structs.ParameterRealValue{ClusterValue: "3.14"},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 3.14, v)

	v, err = convertRealParameterType(applyCtx, structs.ClusterParameterSampleInfo{
		ParamId:   "5",
		Name:      "param5",
		Type:      4,
		RealValue: structs.ParameterRealValue{ClusterValue: "[\"debug\",\"info\"]"},
	})
	assert.NoError(t, err)
	expect := []interface{}{"debug", "info"}
	assert.EqualValues(t, expect, v)

}

func TestExecutor_modifyParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock2rdService := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mock2rdService

	t.Run("success", func(t *testing.T) {
		mock2rdService.EXPECT().ClusterEditGlobalConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return("1", nil)
		mock2rdService.EXPECT().GetOperationStatusByWorkFlowNodeID(gomock.Any(), gomock.Any()).Return(secondparty.GetOperationStatusResp{
			Status: secondparty2.OperationStatus_Finished, Result: "success", ErrorStr: "",
		}, nil)

		mock2rdService.EXPECT().ApiEditConfig(gomock.Any(), gomock.Any()).Return(true, nil)
		mock2rdService.EXPECT().ApiEditConfig(gomock.Any(), gomock.Any()).Return(true, nil)
		mock2rdService.EXPECT().ApiEditConfig(gomock.Any(), gomock.Any()).Return(true, nil)
		mock2rdService.EXPECT().EditClusterConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		modifyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		modifyCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyCtx.SetData(contextModifyParameters, mockModifyParameter())
		modifyCtx.SetData(contextApplyParameterInfo, &message.ApplyParameterGroupReq{
			ParamGroupId: "1",
			ClusterID:    "123",
			Reboot:       true,
		})
		err := modifyParameters(mockWorkFlowAggregation().CurrentNode, modifyCtx)
		assert.NoError(t, err)
	})
}

func TestExecutor_refreshParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock2rdService := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mock2rdService

	t.Run("success", func(t *testing.T) {
		mock2rdService.EXPECT().ClusterReload(gomock.Any(), gomock.Any(), gomock.Any()).Return("123", nil)
		mock2rdService.EXPECT().GetOperationStatusByWorkFlowNodeID(gomock.Any(), gomock.Any()).Return(secondparty.GetOperationStatusResp{
			Status: secondparty2.OperationStatus_Finished, Result: "success", ErrorStr: "",
		}, nil)

		refreshCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		refreshCtx.SetData(contextClusterMeta, mockClusterMeta())
		modifyParameter := mockModifyParameter()
		modifyParameter.Reboot = true
		refreshCtx.SetData(contextModifyParameters, modifyParameter)
		err := refreshParameter(mockWorkFlowAggregation().CurrentNode, refreshCtx)
		assert.NoError(t, err)
	})
}

func TestExecutor_persistApplyParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parameterGroupRW := mockparametergroup.NewMockReaderWriter(ctrl)
	models.SetParameterGroupReaderWriter(parameterGroupRW)
	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		parameterGroupRW.EXPECT().GetParameterGroup(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, parameterGroupId string) (group *parametergroup.ParameterGroup, params []*parametergroup.ParamDetail, err error) {
				return &parametergroup.ParameterGroup{ID: "1"}, []*parametergroup.ParamDetail{
					{
						Parameter:    parametergroup.Parameter{ID: "1"},
						DefaultValue: "10",
						Note:         "param1",
					},
				}, nil
			})
		clusterParameterRW.EXPECT().ApplyClusterParameter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, parameterGroupId string, clusterId string, param []*parameter.ClusterParameterMapping) error {
				return nil
			})

		applyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		applyCtx.SetData(contextClusterMeta, mockClusterMeta())
		applyCtx.SetData(contextModifyParameters, mockModifyParameter())
		applyCtx.SetData(contextApplyParameterInfo, &message.ApplyParameterGroupReq{
			ParamGroupId: "1",
			ClusterID:    "123",
		})
		err := persistApplyParameter(mockWorkFlowAggregation().CurrentNode, applyCtx)
		assert.NoError(t, err)
	})
}

func TestExecutor_persistUpdateParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterParameterRW := mockclusterparameter.NewMockReaderWriter(ctrl)
	models.SetClusterParameterReaderWriter(clusterParameterRW)

	t.Run("success", func(t *testing.T) {
		clusterParameterRW.EXPECT().UpdateClusterParameter(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, clusterId string, params []*parameter.ClusterParameterMapping) (err error) {
				return nil
			})

		applyCtx := &workflow.FlowContext{
			Context:  context.TODO(),
			FlowData: map[string]interface{}{},
		}
		applyCtx.SetData(contextClusterMeta, mockClusterMeta())
		applyCtx.SetData(contextModifyParameters, mockModifyParameter())
		applyCtx.SetData(contextUpdateParameterInfo, &cluster.UpdateClusterParametersReq{
			ClusterID: "123",
			Params: []structs.ClusterParameterSampleInfo{
				{
					ParamId:      "1",
					Name:         "param1",
					InstanceType: "TiDB",
				},
			},
		})
		err := persistUpdateParameter(mockWorkFlowAggregation().CurrentNode, applyCtx)
		assert.NoError(t, err)
	})
}