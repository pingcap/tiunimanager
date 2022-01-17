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
	"strings"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

func authHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin authHosts")

	resourcePool, hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("auth host failed for get flow context, %v", err)
		return err
	}
	for _, host := range hosts {
		err = resourcePool.hostInitiator.CopySSHID(ctx, &host)
		if err != nil {
			log.Errorf("auth host %s %s@%s failed, %v", host.HostName, host.UserName, host.IP, err)
			return err
		}
		log.Infof("auth host %v succeed", host)
	}
	node.Record("auth hosts ")
	return nil
}

func verifyHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin verifyHosts")

	resourcePool, hosts, err := getHostInfoArrayFromFlowContext(ctx)
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
	node.Record("verify hosts ")
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

	resourcePool, hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("install software failed for get flow context, %v", err)
		return err
	}

	if err = resourcePool.hostInitiator.InstallSoftware(ctx, hosts); err != nil {
		log.Errorf("install software failed for %v, %v", hosts, err)
		return err
	}
	log.Infof("install software succeed for %v", hosts)
	node.Record("install software for hosts ")
	return nil
}

func joinEmCluster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin join em cluster")

	resourcePool, hosts, err := getHostInfoArrayFromFlowContext(ctx)
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

	node.Record("join em cluster for hosts ")
	return nil
}

func setHostsOnline(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin set host online")

	resourcePool, hostIds, err := getHostIDArrayFromFlowContext(ctx)
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
	node.Record(fmt.Sprintf("set status of hosts %v to %v ", strings.Join(hostIds, ", "), constants.HostOnline))

	return nil
}

func setHostsFail(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin set host failed")

	resourcePool, hostIds, err := getHostIDArrayFromFlowContext(ctx)
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
	node.Record(fmt.Sprintf("set status of hosts %v to %v ", strings.Join(hostIds, ", "), constants.HostFailed))

	return nil
}

func checkHostBeforeDelete(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin check hosts before delete")

	resourcePool, hostIds, err := getHostIDArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("delete hosts failed for get flow context, %v", err)
		return err
	}

	hosts, _, err := resourcePool.QueryHosts(ctx, &structs.Location{}, &structs.HostFilter{HostID: hostIds[0]}, &structs.PageRequest{})
	if err != nil {
		log.Errorf("query hosts %v failed, %v", hostIds[0], err)
		return err
	}

	if hosts[0].Stat != string(constants.HostLoadLoadLess) {
		errMsg := fmt.Sprintf("check before delete host failed, host %s load stat is not loadless, %s", hosts[0].ID, hosts[0].Stat)
		log.Errorln(errMsg)
		return errors.NewError(errors.TIEM_RESOURCE_HOST_STILL_INUSED, errMsg)
	}
	log.Infof("check host %s before delete succeed", hostIds[0])

	// Set host info to context for leave em cluster executor
	ctx.SetData(rp_consts.ContextHostInfoArrayKey, hosts)
	node.Record(fmt.Sprintf("check host %s before delete ", hostIds[0]))

	return nil
}

func deleteHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin delete hosts")

	resourcePool, hostIds, err := getHostIDArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("delete hosts failed for get flow context, %v", err)
		return err
	}

	err = resourcePool.hostProvider.DeleteHosts(ctx, hostIds)
	if err != nil {
		log.Errorf("delete hosts %v failed, %v", hostIds, err)
		return err
	}
	log.Infof("delete host %v succeed", hostIds)
	node.Record(fmt.Sprintf("delete hosts %v ", strings.Join(hostIds, ", ")))

	return nil
}

func leaveEmCluster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin leave em cluster")
	resourcePool, hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("leave em cluster failed for get flow context, %v", err)
		return err
	}
	// Store nodeID for second party service
	wrapCtx := context.WithValue(ctx, rp_consts.ContextWorkFlowNodeIDKey, node.ID)
	for _, host := range hosts {
		clusterNodeId := fmt.Sprintf("%s:%d", host.IP, rp_consts.HostFileBeatPort)
		if err = resourcePool.hostInitiator.LeaveEMCluster(wrapCtx, clusterNodeId); err != nil {
			log.Errorf("leave em cluster failed for %v, %v", host.IP, err)
			return err
		}
	}
	node.Record("leave em cluster for hosts ")
	return nil
}

func getHostInfoArrayFromFlowContext(ctx *workflow.FlowContext) (rp *ResourcePool, hosts []structs.HostInfo, err error) {
	var ok bool
	rp, ok = ctx.GetData(rp_consts.ContextResourcePoolKey).(*ResourcePool)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextResourcePoolKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	hosts, ok = ctx.GetData(rp_consts.ContextHostInfoArrayKey).([]structs.HostInfo)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextHostInfoArrayKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	return rp, hosts, nil
}

func getHostIDArrayFromFlowContext(ctx *workflow.FlowContext) (rp *ResourcePool, hostIds []string, err error) {
	var ok bool
	rp, ok = ctx.GetData(rp_consts.ContextResourcePoolKey).(*ResourcePool)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextResourcePoolKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	hostIds, ok = ctx.GetData(rp_consts.ContextHostIDArrayKey).([]string)
	if !ok {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextHostIDArrayKey)
		return nil, nil, errors.NewError(errors.TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	return rp, hostIds, nil
}
