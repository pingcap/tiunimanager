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

package check

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/workflow"
)

type CheckManager struct{}

func NewCheckManager() *CheckManager {
	workflowManager := workflow.GetWorkFlowService()

	workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCheckPlatform, &checkDefine)

	return &CheckManager{}
}

var checkDefine = workflow.WorkFlowDefine{
	FlowName:  constants.FlowCheckPlatform,
	TaskNodes: map[string]*workflow.NodeDefine{},
}

// Check
// @Description	check platform and generate check report
// @Parameter	request
// @Return		message.CheckPlatformRsp
// @Return		error
func (manager *CheckManager) Check(ctx context.Context, request message.CheckPlatformReq) (resp message.CheckPlatformRsp, err error) {
	//todo: check platform
	return message.CheckPlatformRsp{}, nil
}

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

func (manager *CheckManager) GetCheckReport(ctx context.Context, request message.GetCheckReportReq) (resp message.GetCheckReportRsp, err error) {
	log := framework.LogWithContext(ctx)
	rw := models.GetReportReaderWriter()
	resp.ReportInfo, err = rw.GetReport(ctx, request.ID)
	if err != nil {
		log.Errorf("get check report %s error: %v", request.ID, err)
		return resp, errors.NewErrorf(errors.CheckReportNotExist,
			"get check report %s error: %v", request.ID, err)
	}
	return resp, err
}
