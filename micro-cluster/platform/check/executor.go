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
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	workflow "github.com/pingcap-inc/tiem/workflow2"
)

func checkCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	var checkID string
	err := context.GetData(ContextCheckID, &checkID)
	if err != nil {
		return err
	}
	var clusterID string
	err = context.GetData(ContextClusterID, &clusterID)
	if err != nil {
		return err
	}

	report := GetReportService()

	err = report.ParseFrom(context.Context, checkID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf("parse from report %s error: %s", checkID, err.Error())
		return err
	}

	err = report.CheckCluster(context.Context, clusterID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf("check cluster %s error: %s", clusterID, err.Error())
		return err
	}

	info, err := report.Serialize(context.Context)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf("serialize check report %s error: %s", checkID, err.Error())
		return err
	}

	err = models.GetReportReaderWriter().UpdateReport(context.Context, checkID, info)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update check report %s error: %s", checkID, err.Error())
		return err
	}

	return nil
}

func checkTenants(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	var checkID string
	err := context.GetData(ContextCheckID, &checkID)
	if err != nil {
		return err
	}

	report := GetReportService()

	err = report.ParseFrom(context.Context, checkID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"parse from report %s error: %s", checkID, err.Error())
		return err
	}

	err = report.CheckTenants(context.Context)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"check platform tenants error: %s", err.Error())
		return err
	}

	info, err := report.Serialize(context.Context)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"serialize check report %s error: %s", checkID, err.Error())
		return err
	}

	err = models.GetReportReaderWriter().UpdateReport(context.Context, checkID, info)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update check report %s error: %s", checkID, err.Error())
		return err
	}

	return nil
}

func checkHosts(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	var checkID string
	err := context.GetData(ContextCheckID, &checkID)
	if err != nil {
		return err
	}

	report := GetReportService()

	err = report.ParseFrom(context.Context, checkID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"parse from report %s error: %s", checkID, err.Error())
		return err
	}

	err = report.CheckHosts(context.Context)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"check platform hosts error: %s", err.Error())
		return err
	}

	info, err := report.Serialize(context.Context)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"serialize check report %s error: %s", checkID, err.Error())
		return err
	}

	err = models.GetReportReaderWriter().UpdateReport(context.Context, checkID, info)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update check report %s error: %s", checkID, err.Error())
		return err
	}
	node.Record(fmt.Sprintf("Check report: %s", info))

	return nil
}

// endCheck
// @Description: end to check platform
func endCheck(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	var checkID string
	err := context.GetData(ContextCheckID, &checkID)
	if err != nil {
		return err
	}

	if err := models.GetReportReaderWriter().UpdateStatus(context.Context,
		checkID, string(constants.CheckCompleted)); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update check report %s status into %s", checkID, constants.CheckCompleted)
		return err
	}

	return nil
}

func handleFail(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	var checkID string
	err := context.GetData(ContextCheckID, &checkID)
	if err != nil {
		return err
	}

	if err := models.GetReportReaderWriter().UpdateStatus(context.Context,
		checkID, string(constants.CheckFailure)); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update check report %s status into %s", checkID, constants.CheckFailure)
		return err
	}

	return nil
}
