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
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/platform/check"
	workflow "github.com/pingcap/tiunimanager/workflow2"
	"sync"
)

const (
	ContextCheckID   = "CheckID"
	ContextClusterID = "ClusterID"
	DefaultCreator   = "System"
	DefaultTenantID  = "admin"
)

var checkService CheckService
var once sync.Once

func GetCheckService() CheckService {
	once.Do(func() {
		if checkService == nil {
			checkService = NewCheckManager()
		}
	})
	return checkService
}

type CheckManager struct {
	autoCheckMgr *autoCheckManager
}

func NewCheckManager() *CheckManager {
	workflowManager := workflow.GetWorkFlowService()

	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCheckPlatform, &checkDefine)
	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCheckCluster, &checkClusterDefine)

	return &CheckManager{
		autoCheckMgr: NewAutoCheckManager(),
	}
}

var checkDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowCheckPlatform,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":            {"checkTenants", "checkTenantsDone", "fail", workflow.SyncFuncNode, checkTenants},
		"checkTenantsDone": {"checkHosts", "checkHostsDone", "fail", workflow.SyncFuncNode, checkHosts},
		"checkHostsDone":   {"end", "", "", workflow.SyncFuncNode, endCheck},
		"fail":             {"fail", "", "", workflow.SyncFuncNode, handleFail},
	},
}

// Check
// @Description	check platform and generate check report
// @Parameter	request
// @Return		message.CheckPlatformRsp
// @Return		error
func (manager *CheckManager) Check(ctx context.Context, request message.CheckPlatformReq) (resp message.CheckPlatformRsp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetReportReaderWriter()

	// create and init check report
	creator := framework.GetUserIDFromContext(ctx)
	if len(creator) == 0 {
		creator = DefaultCreator
	}
	report := &check.CheckReport{
		Report:  "{}",
		Creator: creator,
		Type:    string(constants.PlatformReport),
		Status:  string(constants.CheckRunning),
	}
	report, err = rw.CreateReport(ctx, report)
	if err != nil {
		log.Errorf("create check report error: %v", err)
		return resp, err
	}

	// create workflow
	flowId, err := workflow.GetWorkFlowService().CreateWorkFlow(ctx, report.ID, workflow.BizTypePlatform, checkDefine.FlowName)
	if err != nil {
		log.Errorf("create flow failed, check report %s error: %s", report.ID, err.Error())
		return resp, err
	}

	workflow.GetWorkFlowService().InitContext(ctx, flowId, ContextCheckID, report.ID)
	if err = workflow.GetWorkFlowService().Start(ctx, flowId); err != nil {
		log.Errorf("start flow %s failed, check report %s error: %s", flowId, report.ID, err.Error())
		return resp, err
	}
	log.Infof("create flow %s succeed, check report %s", flowId, report.ID)

	resp.CheckID = report.ID
	resp.WorkFlowID = flowId

	return resp, nil
}

var checkClusterDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowCheckCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":            {"checkCluster", "checkClusterDone", "fail", workflow.SyncFuncNode, checkCluster},
		"checkClusterDone": {"end", "", "", workflow.SyncFuncNode, endCheck},
		"fail":             {"fail", "", "", workflow.SyncFuncNode, handleFail},
	},
}

// CheckCluster
// @Description	check specify cluster and generate check report
// @Parameter	request
// @Return		message.CheckClusterRsp
// @Return		error
func (manager *CheckManager) CheckCluster(ctx context.Context, request message.CheckClusterReq) (resp message.CheckClusterRsp, err error) {
	// create and init check report
	creator := framework.GetUserIDFromContext(ctx)
	if len(creator) == 0 {
		creator = DefaultCreator
	}

	report := &check.CheckReport{
		Report:  "{}",
		Creator: creator,
		Type:    string(constants.ClusterReport),
		Status:  string(constants.CheckRunning),
	}
	report, err = models.GetReportReaderWriter().CreateReport(ctx, report)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create check report error: %v", err)
		return resp, err
	}

	// create workflow
	flowId, err := workflow.GetWorkFlowService().CreateWorkFlow(ctx, report.ID, workflow.BizTypeCluster, checkClusterDefine.FlowName)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create flow failed, check report %s error: %s", report.ID, err.Error())
		return resp, err
	}

	workflow.GetWorkFlowService().InitContext(ctx, flowId, ContextCheckID, report.ID)
	workflow.GetWorkFlowService().InitContext(ctx, flowId, ContextClusterID, request.ClusterID)
	if err = workflow.GetWorkFlowService().Start(ctx, flowId); err != nil {
		framework.LogWithContext(ctx).Errorf("start flow %s failed, check report %s error: %s", flowId, report.ID, err.Error())
		return resp, err
	}
	framework.LogWithContext(ctx).Infof("create flow %s succeed, check report %s", flowId, report.ID)
	resp.CheckID = report.ID
	resp.WorkFlowID = flowId
	return resp, nil
}

// QueryCheckReports
// @Description	query check reports
// @Parameter	request
// @Return		message.QueryCheckReportsRsp
// @Return		error
func (manager *CheckManager) QueryCheckReports(ctx context.Context, request message.QueryCheckReportsReq) (resp message.QueryCheckReportsRsp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetReportReaderWriter()

	resp.ReportMetas, err = rw.QueryReports(ctx)
	if err != nil {
		log.Errorf("query all check reports error: %v", err)
		return resp, errors.NewErrorf(errors.QueryReportsScanRowError, "query all check reports error: %v", err)
	}
	return resp, nil
}

// GetCheckReport
// @Description	get check report
// @Parameter	request
// @Return		message.GetCheckReportRsp
// @Return		error
func (manager *CheckManager) GetCheckReport(ctx context.Context, request message.GetCheckReportReq) (resp message.GetCheckReportRsp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetReportReaderWriter()
	resp.ReportInfo, _, err = rw.GetReport(ctx, request.ID)
	if err != nil {
		log.Errorf("get check report %s error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.CheckReportNotExist,
			"get check report %s error: %v", request.ID, err)
	}
	return resp, err
}
