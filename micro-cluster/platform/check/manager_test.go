/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package check

import (
	ctx "context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/platform/check"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	mock_check "github.com/pingcap-inc/tiem/test/mockcheck"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

func TestNewCheckManager(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		manager := NewCheckManager()
		assert.NotNil(t, manager)
	})
}

func TestCheckManager_CheckCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := &CheckManager{}

	t.Run("normal", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().CreateReport(gomock.Any(), gomock.Any()).Return(&check.CheckReport{ID: "111"}, nil)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, nil)
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil)
		got, err := manager.CheckCluster(ctx.TODO(), message.CheckClusterReq{ClusterID: "123"})
		assert.NoError(t, err)
		assert.Equal(t, got.CheckID, "111")
		assert.Equal(t, got.WorkFlowID, "flow01")
	})

	t.Run("create report err", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().CreateReport(gomock.Any(), gomock.Any()).Return(&check.CheckReport{ID: "111"}, errors.New("create report error"))
		_, err := manager.CheckCluster(ctx.TODO(), message.CheckClusterReq{ClusterID: "123"})
		assert.Error(t, err)
	})

	t.Run("create workflow err", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().CreateReport(gomock.Any(), gomock.Any()).Return(&check.CheckReport{ID: "111"}, nil)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, errors.New("create workflow error"))
		_, err := manager.CheckCluster(ctx.TODO(), message.CheckClusterReq{ClusterID: "123"})
		assert.Error(t, err)
	})

	t.Run("async start err", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().CreateReport(gomock.Any(), gomock.Any()).Return(&check.CheckReport{ID: "111"}, nil)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, nil)
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(errors.New("sync start error"))
		_, err := manager.CheckCluster(ctx.TODO(), message.CheckClusterReq{ClusterID: "123"})
		assert.Error(t, err)
	})
}

func TestCheckManager_Check(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := &CheckManager{}

	t.Run("normal", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().CreateReport(gomock.Any(), gomock.Any()).Return(&check.CheckReport{ID: "111"}, nil)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, nil)
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil)
		got, err := manager.Check(ctx.TODO(), message.CheckPlatformReq{})
		assert.NoError(t, err)
		assert.Equal(t, got.CheckID, "111")
		assert.Equal(t, got.WorkFlowID, "flow01")
	})

	t.Run("create report err", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().CreateReport(gomock.Any(), gomock.Any()).Return(&check.CheckReport{ID: "111"}, errors.New("create report error"))
		_, err := manager.Check(ctx.TODO(), message.CheckPlatformReq{})
		assert.Error(t, err)
	})

	t.Run("create workflow err", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().CreateReport(gomock.Any(), gomock.Any()).Return(&check.CheckReport{ID: "111"}, nil)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, errors.New("create workflow error"))
		_, err := manager.Check(ctx.TODO(), message.CheckPlatformReq{})
		assert.Error(t, err)
	})

	t.Run("async start err", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().CreateReport(gomock.Any(), gomock.Any()).Return(&check.CheckReport{ID: "111"}, nil)
		workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
		workflow.MockWorkFlowService(workflowService)
		defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
		workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
			Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
			Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
		}, nil)
		workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(errors.New("sync start error"))
		_, err := manager.Check(ctx.TODO(), message.CheckPlatformReq{})
		assert.Error(t, err)
	})
}

func TestCheckManager_GetCheckReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := &CheckManager{}

	t.Run("normal", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)
		var info interface{}
		info = &structs.CheckClusterReportInfo{
			ClusterCheck: structs.ClusterCheck{
				ID: "123",
			}}
		reportRW.EXPECT().GetReport(gomock.Any(), "111").Return(info, string(constants.ClusterReport), nil)
		got, err := manager.GetCheckReport(ctx.TODO(), message.GetCheckReportReq{ID: "111"})
		assert.NoError(t, err)
		_, err = json.Marshal(got)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)
		reportRW.EXPECT().GetReport(gomock.Any(), "222").Return(nil, "", errors.New("get report error"))
		_, err := manager.GetCheckReport(ctx.TODO(), message.GetCheckReportReq{ID: "222"})
		assert.Error(t, err)
	})
}

func TestCheckManager_QueryCheckReports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := &CheckManager{}

	t.Run("normal", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)
		reportRW.EXPECT().QueryReports(gomock.Any()).Return(make(map[string]structs.CheckReportMeta), nil)
		_, err := manager.QueryCheckReports(ctx.TODO(), message.QueryCheckReportsReq{})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)
		reportRW.EXPECT().QueryReports(gomock.Any()).Return(make(map[string]structs.CheckReportMeta), errors.New("query reports error"))
		_, err := manager.QueryCheckReports(ctx.TODO(), message.QueryCheckReportsReq{})
		assert.Error(t, err)
	})
}
