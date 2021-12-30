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
	"context"
	"fmt"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

func verifyHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin verifyHosts")

	resourcePool, hosts, err := getImportHostInfoFromFlowContext(ctx)
	if err != nil {
		log.Errorf("verify host failed for get flow context, %v", err)
		return err
	}
	for _, host := range hosts {
		err = resourcePool.hostInitiator.Verify(ctx, &host)
		if err != nil {
			log.Errorf("verify host %s %s failed, %v", host.HostName, host.IP, err)
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

func installSoftware(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin installSoftware")

	resourcePool, hosts, err := getImportHostInfoFromFlowContext(ctx)
	if err != nil {
		log.Errorf("install software failed for get flow context, %v", err)
		return err
	}

	if err = resourcePool.hostInitiator.InstallSoftware(ctx, hosts); err != nil {
		log.Errorf("install software failed for %v, %v", hosts, err)
		return err
	}
	log.Infof("install software succeed for %v", hosts)

	return nil
}

func joinEmCluster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin join em cluster")

	resourcePool, hosts, err := getImportHostInfoFromFlowContext(ctx)
	if err != nil {
		log.Errorf("join em cluster failed for get flow context, %v", err)
		return err
	}
	// Store nodeID for second party service
	installSoftwareCtx := context.WithValue(ctx, rp_consts.ContextWorkFlowNodeIDKey, node.ID)
	if err = resourcePool.hostInitiator.JoinEMCluster(installSoftwareCtx, hosts); err != nil {
		log.Errorf("join em cluster failed for %v, %v", hosts, err)
		return err
	}

	return nil
}

func importHostSucceed(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin importHostSucceed")

	resourcePool, hostIds, err := getImportHostIDsFromFlowContext(ctx)
	if err != nil {
		log.Errorf("set host online failed for get flow context, %v", err)
		return err
	}
	err = resourcePool.UpdateHostStatus(ctx, hostIds, string(constants.HostOnline))
	if err != nil {
		log.Errorf("set host %v online failed, %v", hostIds, err)
		return err
	}
	log.Infof("set host %v online succeed", hostIds)

	return nil
}

func importHostsFail(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin importHostFailed")

	resourcePool, hostIds, err := getImportHostIDsFromFlowContext(ctx)
	if err != nil {
		log.Errorf("set host failed failed for get flow context, %v", err)
		return err
	}

	err = resourcePool.UpdateHostStatus(ctx, hostIds, string(constants.HostFailed))
	if err != nil {
		log.Errorf("set host %v failed failed, %v", hostIds, err)
		return err
	}
	log.Infof("set host %v failed succeed", hostIds)

	return nil
}

func getImportHostInfoFromFlowContext(ctx *workflow.FlowContext) (rp *ResourcePool, hosts []structs.HostInfo, err error) {
	var ok bool
	rp, ok = ctx.GetData(rp_consts.ContextResourcePoolKey).(*ResourcePool)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextResourcePoolKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	hosts, ok = ctx.GetData(rp_consts.ContextImportHostInfoKey).([]structs.HostInfo)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextImportHostInfoKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	return rp, hosts, nil
}

func getImportHostIDsFromFlowContext(ctx *workflow.FlowContext) (rp *ResourcePool, hostIds []string, err error) {
	var ok bool
	rp, ok = ctx.GetData(rp_consts.ContextResourcePoolKey).(*ResourcePool)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextResourcePoolKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	hostIds, ok = ctx.GetData(rp_consts.ContextImportHostIDsKey).([]string)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextImportHostIDsKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	return rp, hostIds, nil
}
