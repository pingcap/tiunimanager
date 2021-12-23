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

package resourcepool

import (
	"fmt"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

const (
	contextResourcePoolKey string = "resourcePool"
	contextImportHostsKey  string = "importHosts"
)

func verifyHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Info("begin verifyHosts")
	defer log.Infof("end verifyHosts, err: %v", err)

	resourcePool, hosts, err := getImportHostContext(ctx)
	if err != nil {
		return err
	}
	for _, host := range hosts {
		err = resourcePool.hostInitiator.Verify(ctx, &host)
		if err != nil {
			return err
		}
		log.Infof("verify host %v succeed", host)
	}
	return nil
}

func configHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin configHosts")
	defer framework.LogWithContext(ctx).Info("end configHosts")
	return nil
}

func installSoftware(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin installSoftware")
	defer framework.LogWithContext(ctx).Info("end installSoftware")

	resourcePool, hosts, err := getImportHostContext(ctx)
	if err != nil {
		return err
	}
	if err = resourcePool.hostInitiator.InstallSoftware(ctx, hosts); err != nil {
		return err
	}
	return nil
}

func createHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin createHosts")
	defer framework.LogWithContext(ctx).Info("end createHosts")
	return nil
}

func importHostSucceed(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin importHostSucceed")
	defer framework.LogWithContext(ctx).Info("end importHostSucceed")
	return nil
}

func importHostsFail(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin importHostFailed")
	defer framework.LogWithContext(ctx).Info("end importHostFailed")
	return nil
}

func getImportHostContext(ctx *workflow.FlowContext) (rp *ResourcePool, hosts []structs.HostInfo, err error) {
	var ok bool
	rp, ok = ctx.GetData(contextResourcePoolKey).(*ResourcePool)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", contextResourcePoolKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	hosts, ok = ctx.GetData(contextImportHostsKey).([]structs.HostInfo)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", contextImportHostsKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	return rp, hosts, nil
}
