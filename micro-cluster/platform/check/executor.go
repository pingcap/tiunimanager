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
	"encoding/json"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/models"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

func checkTenants(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	checkID := context.GetData(ContextCheckID).(string)
	info := structs.CheckReportInfo{
		Tenants: map[string]structs.TenantCheck{
			"tenant01": {
				ClusterCount: 1,
			},
		},
	}
	reportInfo, err := json.Marshal(info)
	if err != nil {
		return err
	}
	err = models.GetReportReaderWriter().UpdateReport(context.Context, checkID, string(reportInfo))
	if err != nil {
		return err
	}

	return nil
}

func endCheck(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	return nil
}

func handleFail(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	return nil
}
