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
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/models"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	mock_check "github.com/pingcap-inc/tiem/test/mockcheck"
	mock_report "github.com/pingcap-inc/tiem/test/mockreport"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckCluster(gomock.Any(), "123").Return(nil)
		reportCheck.EXPECT().Serialize(gomock.Any()).Return("", nil)
		MockReportService(reportCheck)

		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateReport(gomock.Any(), "111", gomock.Any()).Return(nil)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		flowContext.SetData(ContextClusterID, "123")
		err := checkCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("parse from error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(errors.New("parse from error"))
		MockReportService(reportCheck)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		flowContext.SetData(ContextClusterID, "123")
		err := checkCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("check error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckCluster(gomock.Any(), "123").Return(errors.New("check error"))
		MockReportService(reportCheck)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		flowContext.SetData(ContextClusterID, "123")
		err := checkCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("serialize error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckCluster(gomock.Any(), "123").Return(nil)
		reportCheck.EXPECT().Serialize(gomock.Any()).Return("", errors.New("serialize error"))
		MockReportService(reportCheck)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		flowContext.SetData(ContextClusterID, "123")
		err := checkCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("update report error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckCluster(gomock.Any(), "123").Return(nil)
		reportCheck.EXPECT().Serialize(gomock.Any()).Return("", nil)
		MockReportService(reportCheck)

		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateReport(gomock.Any(), "111", gomock.Any()).Return(errors.New("update report error"))

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		flowContext.SetData(ContextClusterID, "123")
		err := checkCluster(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestCheckTenants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckTenants(gomock.Any()).Return(nil)
		reportCheck.EXPECT().Serialize(gomock.Any()).Return("", nil)
		MockReportService(reportCheck)

		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateReport(gomock.Any(), "111", gomock.Any()).Return(nil)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkTenants(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("parse from error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(errors.New("parse from error"))
		MockReportService(reportCheck)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkTenants(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("check error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckTenants(gomock.Any()).Return(errors.New("check error"))
		MockReportService(reportCheck)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkTenants(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("serialize error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckTenants(gomock.Any()).Return(nil)
		reportCheck.EXPECT().Serialize(gomock.Any()).Return("", errors.New("serialize error"))
		MockReportService(reportCheck)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkTenants(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("update report error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckTenants(gomock.Any()).Return(nil)
		reportCheck.EXPECT().Serialize(gomock.Any()).Return("", nil)
		MockReportService(reportCheck)

		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateReport(gomock.Any(), "111", gomock.Any()).Return(errors.New("update report error"))

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkTenants(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestCheckHosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckHosts(gomock.Any()).Return(nil)
		reportCheck.EXPECT().Serialize(gomock.Any()).Return("", nil)
		MockReportService(reportCheck)

		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateReport(gomock.Any(), "111", gomock.Any()).Return(nil)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkHosts(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("parse from error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(errors.New("parse from error"))
		MockReportService(reportCheck)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkHosts(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("check error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckHosts(gomock.Any()).Return(errors.New("check error"))
		MockReportService(reportCheck)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkHosts(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("serialize error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckHosts(gomock.Any()).Return(nil)
		reportCheck.EXPECT().Serialize(gomock.Any()).Return("", errors.New("serialize error"))
		MockReportService(reportCheck)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkHosts(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})

	t.Run("update report error", func(t *testing.T) {
		reportCheck := mock_report.NewMockReportService(ctrl)
		reportCheck.EXPECT().ParseFrom(gomock.Any(), "111").Return(nil)
		reportCheck.EXPECT().CheckHosts(gomock.Any()).Return(nil)
		reportCheck.EXPECT().Serialize(gomock.Any()).Return("", nil)
		MockReportService(reportCheck)

		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateReport(gomock.Any(), "111", gomock.Any()).Return(errors.New("update report error"))

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := checkHosts(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestEndCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateStatus(gomock.Any(), "111", string(constants.CheckCompleted)).Return(nil)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := endCheck(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateStatus(gomock.Any(), "111", string(constants.CheckCompleted)).Return(errors.New("end check error"))

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := endCheck(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}

func TestHandleFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateStatus(gomock.Any(), "111", string(constants.CheckFailure)).Return(nil)

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := handleFail(&workflowModel.WorkFlowNode{}, flowContext)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		reportRW := mock_check.NewMockReaderWriter(ctrl)
		models.SetReportReaderWriter(reportRW)

		reportRW.EXPECT().UpdateStatus(gomock.Any(), "111", string(constants.CheckFailure)).Return(errors.New("handle fail error"))

		flowContext := workflow.NewFlowContext(context.TODO())
		flowContext.SetData(ContextCheckID, "111")
		err := handleFail(&workflowModel.WorkFlowNode{}, flowContext)
		assert.Error(t, err)
	})
}
