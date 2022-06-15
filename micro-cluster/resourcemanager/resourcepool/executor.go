/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	rp_consts "github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/resourcepool/constants"
	workflowModel "github.com/pingcap/tiunimanager/models/workflow"
	workflow "github.com/pingcap/tiunimanager/workflow2"
)

func validateHostInfo(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin validate host info")

	resourcePool := GetResourcePool()
	hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("validate host failed for get flow context, %v", err)
		return err
	}
	for _, host := range hosts {
		err = resourcePool.ValidateZoneInfo(ctx, &host)
		if err != nil {
			log.Errorf("validate host %s %s failed, %v", host.HostName, host.IP, err)
			return err
		}
		log.Infof("validate host %s %s on %s %s succeed", host.HostName, host.IP, host.Region, host.AZ)
	}
	node.Record(fmt.Sprintf("validate hosts %s %s on %s %s succeed", hosts[0].HostName, hosts[0].IP, hosts[0].Region, hosts[0].AZ))
	return nil
}

func authHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin authHosts")

	deployUser := framework.GetCurrentDeployUser()
	userGroup := framework.GetCurrentDeployGroup()

	resourcePool := GetResourcePool()
	hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("auth host failed for get flow context, %v", err)
		return err
	}
	for _, host := range hosts {
		err = resourcePool.hostInitiator.AuthHost(ctx, deployUser, userGroup, &host)
		if err != nil {
			log.Errorf("auth host %s %s@%s failed, %v", host.HostName, host.UserName, host.IP, err)
			return err
		}
		log.Infof("auth host %s %s succeed", host.HostName, host.IP)
	}
	node.Record(fmt.Sprintf("auth host %s %s with %s:%s succeed", hosts[0].HostName, hosts[0].IP, deployUser, userGroup))
	return nil
}

func prepare(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin prepare host env before verify")

	resourcePool := GetResourcePool()
	hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("verify host failed for get flow context, %v", err)
		return err
	}

	for _, host := range hosts {
		// install numactl and set swap off
		err = resourcePool.hostInitiator.Prepare(ctx, &host)
		if err != nil {
			log.Errorf("prepare host %s %s failed, %v", host.HostName, host.IP, err)
			return err
		}
		log.Infof("prepare host %s %s succeed", host.HostName, host.IP)
	}
	node.Record(fmt.Sprintf("prepare host %s %s succeed", hosts[0].HostName, hosts[0].IP))
	return nil
}

func verifyHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin verifyHosts")

	resourcePool := GetResourcePool()
	hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("verify host failed for get flow context, %v", err)
		return err
	}
	ignoreWarnings, err := getIgnoreWarningsFromFlowContext(ctx)
	if err != nil {
		log.Errorf("verify host failed for get flow context, %v", err)
		return err
	}
	// Store ignoreWarnings for Verify results
	wrapCtx := context.WithValue(ctx, rp_consts.ContextIgnoreWarnings, ignoreWarnings) //nolint // Use string value to identify each key, so no need to create a new type for the lint
	for _, host := range hosts {
		err = resourcePool.hostInitiator.Verify(wrapCtx, &host)
		if err != nil {
			log.Errorf("verify host %s %s failed, %v", host.HostName, host.IP, err)
			return err
		}
		log.Infof("verify host %s %s succeed", host.HostName, host.IP)
	}
	node.Record(fmt.Sprintf("verify host %s %s succeed", hosts[0].HostName, hosts[0].IP))
	return nil
}

func installSoftware(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin installSoftware")

	resourcePool := GetResourcePool()
	hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("install software failed for get flow context, %v", err)
		return err
	}

	if err = resourcePool.hostInitiator.InstallSoftware(ctx, hosts); err != nil {
		log.Errorf("install software failed for %v, %v", hosts, err)
		return err
	}
	log.Infof("install software succeed for host %s %s", hosts[0].HostName, hosts[0].IP)
	node.Record(fmt.Sprintf("install software for host %s %s succeed", hosts[0].HostName, hosts[0].IP))
	return nil
}

func joinEmCluster(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin join em cluster")

	resourcePool := GetResourcePool()
	hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("join em cluster failed for get flow context, %v", err)
		return err
	}

	// Check whether host has already install filebeat
	installed, err := resourcePool.hostInitiator.PreCheckHostInstallFilebeat(ctx, hosts)
	if err != nil {
		log.Errorf("precheck host %s %s filebeat failed, %v", hosts[0].HostName, hosts[0].IP, err)
		return err
	}
	if installed {
		msg := fmt.Sprintf("host %s %s has already installed filebeat, no need to re-join em cluster", hosts[0].HostName, hosts[0].IP)
		log.Infoln(msg)
		node.Success(msg)
		return nil
	}

	// Store nodeID for second party service
	installSoftwareCtx := context.WithValue(ctx, rp_consts.ContextWorkFlowIDKey, node.ID) //nolint // Use string value to identify each key, so no need to create a new type for the lint
	operationID, err := resourcePool.hostInitiator.JoinEMCluster(installSoftwareCtx, hosts)
	if err != nil {
		log.Errorf("join em cluster failed for host %s %s, %v", hosts[0].HostName, hosts[0].IP, err)
		return err
	}
	log.Infof("join em cluster for host %s %s in operation %s", hosts[0].HostName, hosts[0].IP, operationID)
	node.Record(fmt.Sprintf("joining em cluster for host %s %s", hosts[0].HostName, hosts[0].IP))
	node.OperationID = operationID
	return nil
}

func setHostsOnline(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin set host online")

	resourcePool := GetResourcePool()
	hostIds, err := getHostIDArrayFromFlowContext(ctx)
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

	resourcePool := GetResourcePool()
	hostIds, err := getHostIDArrayFromFlowContext(ctx)
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

	hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("delete hosts failed for get host info from flow context, %v", err)
		return err
	}

	if hosts[0].Stat != string(constants.HostLoadLoadLess) {
		errMsg := fmt.Sprintf("check before delete host failed, host %s load stat is not loadless, %s", hosts[0].ID, hosts[0].Stat)
		log.Errorln(errMsg)
		return errors.NewError(errors.TIUNIMANAGER_RESOURCE_HOST_STILL_INUSED, errMsg)
	}
	log.Infof("check host %s before delete succeed", hosts[0].ID)

	node.Record(fmt.Sprintf("check host %s before delete ", hosts[0].ID))

	return nil
}

func deleteHosts(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin delete hosts")

	resourcePool := GetResourcePool()
	hostIds, err := getHostIDArrayFromFlowContext(ctx)
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

	resourcePool := GetResourcePool()
	hosts, err := getHostInfoArrayFromFlowContext(ctx)
	if err != nil {
		log.Errorf("leave em cluster failed for get flow context, %v", err)
		return err
	}

	installed, err := resourcePool.hostInitiator.PreCheckHostInstallFilebeat(ctx, hosts)
	if err != nil {
		log.Errorf("precheck host %s %s filebeat failed, %v", hosts[0].HostName, hosts[0].IP, err)
		return err
	}
	if !installed {
		msg := fmt.Sprintf("host %s %s has already uninstalled filebeat, no need to leave em cluster", hosts[0].HostName, hosts[0].IP)
		log.Infoln(msg)
		node.Success(msg)
		return nil
	}

	// Store nodeID for second party service
	wrapCtx := context.WithValue(ctx, rp_consts.ContextWorkFlowIDKey, node.ID) //nolint // Use string value to identify each key, so no need to create a new type for the lint

	for _, host := range hosts {
		clusterNodeId := fmt.Sprintf("%s:%d", host.IP, rp_consts.HostFileBeatPort)
		operationID, err := resourcePool.hostInitiator.LeaveEMCluster(wrapCtx, clusterNodeId)
		if err != nil {
			log.Errorf("leave em cluster failed for %v, %v", host.IP, err)
			return err
		}
		node.OperationID = operationID
		log.Infof("leave em cluster for host %s %s in operation %s", host.HostName, host.IP, operationID)
		node.Record(fmt.Sprintf("leaving em cluster for host %s %s", host.HostName, host.IP))
	}

	return nil
}

func getHostInfoArrayFromFlowContext(ctx *workflow.FlowContext) (hosts []structs.HostInfo, err error) {
	err = ctx.GetData(rp_consts.ContextHostInfoArrayKey, &hosts)
	if err != nil {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextHostInfoArrayKey)
		return nil, errors.NewError(errors.TIUNIMANAGER_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	return hosts, nil
}

func getHostIDArrayFromFlowContext(ctx *workflow.FlowContext) (hostIds []string, err error) {
	err = ctx.GetData(rp_consts.ContextHostIDArrayKey, &hostIds)
	if err != nil {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextHostIDArrayKey)
		return nil, errors.NewError(errors.TIUNIMANAGER_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	return hostIds, nil
}

func getIgnoreWarningsFromFlowContext(ctx *workflow.FlowContext) (ignoreWarnings bool, err error) {
	err = ctx.GetData(rp_consts.ContextIgnoreWarnings, &ignoreWarnings)
	if err != nil {
		errMsg := fmt.Sprintf("get key %s from flow context failed", rp_consts.ContextIgnoreWarnings)
		return ignoreWarnings, errors.NewError(errors.TIUNIMANAGER_RESOURCE_EXTRACT_FLOW_CTX_ERROR, errMsg)
	}
	return ignoreWarnings, nil
}
