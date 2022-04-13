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
	"github.com/pingcap-inc/tiem/message"
	"strconv"
	"sync"
	"time"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostinitiator"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostprovider"
	"github.com/pingcap-inc/tiem/models"
	workflow "github.com/pingcap-inc/tiem/workflow2"
)

type ResourcePool struct {
	hostProvider  hostprovider.HostProvider
	hostInitiator hostinitiator.HostInitiator
	// cloudHostProvider hostprovider.HostProvider
}

var globalResourcePool *ResourcePool
var once sync.Once

func GetResourcePool() *ResourcePool {
	once.Do(func() {
		if globalResourcePool == nil {
			globalResourcePool = new(ResourcePool)
			globalResourcePool.InitResourcePool()
		}
	})
	return globalResourcePool
}

func (p *ResourcePool) InitResourcePool() {
	p.hostProvider = hostprovider.NewFileHostProvider()
	p.hostInitiator = hostinitiator.NewFileHostInitiator()

	flowManager := workflow.GetWorkFlowService()
	p.registerImportHostsWorkFlow(context.TODO(), flowManager)
	p.registerDeleteHostsWorkFlow(context.TODO(), flowManager)
}

func (p *ResourcePool) registerImportHostsWorkFlow(ctx context.Context, flowManager workflow.WorkFlowService) {
	log := framework.LogWithContext(ctx)

	log.Infoln("register import hosts workflow with host init")
	flowManager.RegisterWorkFlow(ctx, rp_consts.FlowImportHosts, &workflow.WorkFlowDefine{
		FlowName: rp_consts.FlowImportHosts,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":           {Name: "start", SuccessEvent: "authhosts", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: validateHostInfo},
			"authhosts":       {Name: "authhosts", SuccessEvent: "prepare", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: authHosts},
			"prepare":         {Name: "prepare", SuccessEvent: "verifyHosts", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: prepare},
			"verifyHosts":     {Name: "verifyHosts", SuccessEvent: "installSoftware", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: verifyHosts},
			"installSoftware": {Name: "installSoftware", SuccessEvent: "joinEMCluster", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: installSoftware},
			"joinEMCluster":   {Name: "joinEMCluster", SuccessEvent: "succeed", FailEvent: "fail", ReturnType: workflow.PollingNode, Executor: joinEmCluster},
			"succeed":         {Name: "succeed", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: setHostsOnline},
			"fail":            {Name: "fail", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: setHostsFail},
		},
	})

	log.Infoln("register import hosts workflow without host init")
	flowManager.RegisterWorkFlow(ctx, rp_consts.FlowImportHostsWithoutInit, &workflow.WorkFlowDefine{
		FlowName: rp_consts.FlowImportHostsWithoutInit,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start": {Name: "start", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: setHostsOnline},
		},
	})

	log.Infoln("register take over hosts workflow")
	flowManager.RegisterWorkFlow(ctx, rp_consts.FlowTakeOverHosts, &workflow.WorkFlowDefine{
		FlowName: rp_consts.FlowTakeOverHosts,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":         {Name: "start", SuccessEvent: "authhosts", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: validateHostInfo},
			"authhosts":     {Name: "authhosts", SuccessEvent: "joinEMCluster", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: authHosts},
			"joinEMCluster": {Name: "joinEMCluster", SuccessEvent: "succeed", FailEvent: "fail", ReturnType: workflow.PollingNode, Executor: joinEmCluster},
			"succeed":       {Name: "succeed", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: setHostsOnline},
			"fail":          {Name: "fail", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: setHostsFail},
		},
	})
}

func (p *ResourcePool) registerDeleteHostsWorkFlow(ctx context.Context, flowManager workflow.WorkFlowService) {
	log := framework.LogWithContext(ctx)

	log.Infoln("register delete hosts workflow with host init")
	flowManager.RegisterWorkFlow(ctx, rp_consts.FlowDeleteHosts, &workflow.WorkFlowDefine{
		FlowName: rp_consts.FlowDeleteHosts,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":          {Name: "start", SuccessEvent: "leaveEMCluster", FailEvent: "recover", ReturnType: workflow.SyncFuncNode, Executor: checkHostBeforeDelete},
			"leaveEMCluster": {Name: "leaveEMCluster", SuccessEvent: "succeed", FailEvent: "fail", ReturnType: workflow.PollingNode, Executor: leaveEmCluster},
			"succeed":        {Name: "succeed", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: deleteHosts},
			"recover":        {Name: "recover", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: setHostsOnline},
			"fail":           {Name: "fail", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: setHostsFail},
		},
	})

	log.Infoln("register delete hosts workflow without host init")
	flowManager.RegisterWorkFlow(ctx, rp_consts.FlowDeleteHostsByForce, &workflow.WorkFlowDefine{
		FlowName: rp_consts.FlowDeleteHostsByForce,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":   {Name: "start", SuccessEvent: "succeed", FailEvent: "recover", ReturnType: workflow.SyncFuncNode, Executor: checkHostBeforeDelete},
			"succeed": {Name: "succeed", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: deleteHosts},
			"recover": {Name: "recover", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: setHostsOnline},
		},
	})
}

func (p *ResourcePool) GetHostProvider() hostprovider.HostProvider {
	return p.hostProvider
}

func (p *ResourcePool) SetHostProvider(provider hostprovider.HostProvider) {
	p.hostProvider = provider
}

func (p *ResourcePool) SetHostInitiator(initiator hostinitiator.HostInitiator) {
	p.hostInitiator = initiator
}

func (p *ResourcePool) getSSHConfigPort(ctx context.Context) int {
	sshConfigPort, err := models.GetConfigReaderWriter().GetConfig(ctx, constants.ConfigKeyDefaultSSHPort)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get config ConfigKeyDefaultSSHPort failed, %v", err)
		// return a default ssh port. If can not connect host using the default port, will fail in host initiator
		return rp_consts.HostSSHPort
	}
	portStr := sshConfigPort.ConfigValue
	if portStr == "" {
		framework.LogWithContext(ctx).Warnln("get config ConfigKeyDefaultSSHPort is null")
		return rp_consts.HostSSHPort
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get config ConfigKeyDefaultSSHPort invalid string %s to atoi, %v", portStr, err)
		return rp_consts.HostSSHPort
	}
	return port
}

func (p *ResourcePool) ImportHosts(ctx context.Context, hosts []structs.HostInfo, condition *structs.ImportCondition) (flowIds []string, hostIds []string, err error) {
	hostIds, err = p.hostProvider.ImportHosts(ctx, hosts)
	if err != nil {
		return flowIds, hostIds, err
	}
	flowManager := workflow.GetWorkFlowService()
	flowName := p.selectImportFlowName(condition)
	hostSSHPort := p.getSSHConfigPort(ctx)
	framework.LogWithContext(ctx).Infof("import hosts select %s, with ssh port %d", flowName, hostSSHPort)
	for i, host := range hosts {
		flowId, err := flowManager.CreateWorkFlow(ctx, hostIds[i], workflow.BizTypeHost, flowName)
		if err != nil {
			errMsg := fmt.Sprintf("create %s workflow failed for host %s %s, %s", flowName, host.HostName, host.IP, err.Error())
			framework.LogWithContext(ctx).Errorln(errMsg)
			return nil, hostIds, errors.WrapError(errors.TIEM_WORKFLOW_CREATE_FAILED, errMsg, err)
		}

		host.SSHPort = int32(hostSSHPort)
		flowManager.InitContext(ctx, flowId, rp_consts.ContextHostInfoArrayKey, []structs.HostInfo{host})
		flowManager.InitContext(ctx, flowId, rp_consts.ContextHostIDArrayKey, []string{hostIds[i]})
		// Whether ignore warnings when verify host
		ignoreWarnings := condition.IgnoreWarings || framework.Current.GetClientArgs().IgnoreHostWarns
		flowManager.InitContext(ctx, flowId, rp_consts.ContextIgnoreWarnings, ignoreWarnings)

		flowIds = append(flowIds, flowId)
	}
	// Sync start each flow in a goroutine: tiup-tiem/sqlite DO NOT support concurrent
	operationName := fmt.Sprintf(
		"%s.%s BackgroundTask",
		framework.GetMicroServiceNameFromContext(ctx),
		framework.GetMicroEndpointNameFromContext(ctx),
	)
	framework.StartBackgroundTask(ctx, operationName, func(ctx context.Context) error {
		// Unset EmTiup Using flag after all flows in this task done
		defer framework.UnsetInEmTiupProcess()
		for i, flowId := range flowIds {
			if err = flowManager.Start(ctx, flowId); err != nil {
				errMsg := fmt.Sprintf("async start %s workflow[%d] %s failed for host %s %s, %s", rp_consts.FlowImportHosts, i, flowId, hosts[i].HostName, hosts[i].IP, err.Error())
				framework.LogWithContext(ctx).Errorln(errMsg)
				continue
			} else {
				framework.LogWithContext(ctx).Infof("sync start %s workflow[%d] %s for host %s %s", rp_consts.FlowImportHosts, i, flowId, hosts[i].HostName, hosts[i].IP)
			}
			ticker := time.NewTicker(5 * time.Second)
			sequence := int32(0)
			for range ticker.C {
				sequence++
				if sequence > 1000 {
					break
				}
				resp, err := flowManager.DetailWorkFlow(ctx, message.QueryWorkFlowDetailReq{WorkFlowID: flowId})
				if err != nil {
					framework.LogWithContext(ctx).Errorln(err)
					break
				}
				if constants.WorkFlowStatusError == resp.Info.Status {
					framework.LogWithContext(ctx).Errorln(fmt.Sprintf("%s workflow %s end failed", flowName, flowId))
					break
				}
				if constants.WorkFlowStatusFinished == resp.Info.Status {
					framework.LogWithContext(ctx).Infof("%s workflow %s end success", flowName, flowId)
					break
				}
			}
		}
		return nil
	})

	return flowIds, hostIds, nil
}

func (p *ResourcePool) DeleteHosts(ctx context.Context, hostIds []string, force bool) (flowIds []string, err error) {
	flowManager := workflow.GetWorkFlowService()
	for _, hostId := range hostIds {
		hosts, count, err := p.QueryHosts(ctx, &structs.Location{}, &structs.HostFilter{HostID: hostId}, &structs.PageRequest{})
		if err != nil {
			errMsg := fmt.Sprintf("query host %v failed, %v", hostId, err)
			framework.LogWithContext(ctx).Errorln(errMsg)
			return nil, errors.WrapError(errors.TIEM_RESOURCE_DELETE_HOST_ERROR, errMsg, err)
		}
		if count == 0 {
			errMsg := fmt.Sprintf("deleting host %s is not found", hostId)
			framework.LogWithContext(ctx).Errorln(errMsg)
			return nil, errors.NewError(errors.TIEM_RESOURCE_DELETE_HOST_ERROR, errMsg)
		}
		flowName := p.selectDeleteFlowName(&hosts[0], force)
		framework.LogWithContext(ctx).Infof("delete host %s select %s", hostId, flowName)

		flowId, err := flowManager.CreateWorkFlow(ctx, hostId, workflow.BizTypeHost, flowName)
		if err != nil {
			errMsg := fmt.Sprintf("create %s workflow failed for host %s, %s", flowName, hostId, err.Error())
			framework.LogWithContext(ctx).Errorln(errMsg)
			return nil, errors.WrapError(errors.TIEM_WORKFLOW_CREATE_FAILED, errMsg, err)
		}

		flowManager.InitContext(ctx, flowId, rp_consts.ContextHostIDArrayKey, []string{hostId})
		flowManager.InitContext(ctx, flowId, rp_consts.ContextHostInfoArrayKey, hosts)

		flowIds = append(flowIds, flowId)
	}

	err = p.UpdateHostStatus(ctx, hostIds, string(constants.HostDeleting))
	if err != nil {
		return nil, err
	}
	operationName := fmt.Sprintf(
		"%s.%s BackgroundTask",
		framework.GetMicroServiceNameFromContext(ctx),
		framework.GetMicroEndpointNameFromContext(ctx),
	)
	framework.StartBackgroundTask(ctx, operationName, func(ctx context.Context) error {
		// Unset EmTiup Using flag after all flows in this task done
		defer framework.UnsetInEmTiupProcess()
		for i, flowId := range flowIds {
			if err = flowManager.Start(ctx, flowId); err != nil {
				errMsg := fmt.Sprintf("sync start %s workflow[%d] %s failed for delete host %s, %s", rp_consts.FlowDeleteHosts, i, flowId, hostIds[i], err.Error())
				framework.LogWithContext(ctx).Errorln(errMsg)
			} else {
				framework.LogWithContext(ctx).Infof("sync start %s workflow[%d] %s for delete host %s", rp_consts.FlowDeleteHosts, i, flowId, hostIds[i])
			}
			ticker := time.NewTicker(5 * time.Second)
			sequence := int32(0)
			for range ticker.C {
				sequence++
				if sequence > 1000 {
					break
				}
				resp, err := flowManager.DetailWorkFlow(ctx, message.QueryWorkFlowDetailReq{WorkFlowID: flowId})
				if err != nil {
					framework.LogWithContext(ctx).Errorln(err)
					break
				}
				if constants.WorkFlowStatusError == resp.Info.Status {
					framework.LogWithContext(ctx).Errorln(fmt.Sprintf("deleteHosts workflow %s end failed", flowId))
					break
				}
				if constants.WorkFlowStatusFinished == resp.Info.Status {
					framework.LogWithContext(ctx).Infof("deleteHosts workflow %s end success", flowId)
					break
				}
			}
		}
		return nil
	})

	return flowIds, nil
}

func (p *ResourcePool) ValidateZoneInfo(ctx context.Context, host *structs.HostInfo) (err error) {
	return p.hostProvider.ValidateZoneInfo(ctx, host)
}

func (p *ResourcePool) QueryHosts(ctx context.Context, location *structs.Location, filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, total int64, err error) {
	return p.hostProvider.QueryHosts(ctx, location, filter, page)
}

func (p *ResourcePool) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	return p.hostProvider.UpdateHostStatus(ctx, hostIds, status)
}

func (p *ResourcePool) UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error) {
	return p.hostProvider.UpdateHostReserved(ctx, hostIds, reserved)
}

func (p *ResourcePool) GetHierarchy(ctx context.Context, filter *structs.HostFilter, level int, depth int) (root *structs.HierarchyTreeNode, err error) {
	return p.hostProvider.GetHierarchy(ctx, filter, level, depth)
}

func (p *ResourcePool) GetStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks map[string]*structs.Stocks, err error) {
	return p.hostProvider.GetStocks(ctx, location, hostFilter, diskFilter)
}

func (p *ResourcePool) UpdateHostInfo(ctx context.Context, host structs.HostInfo) (err error) {
	return p.hostProvider.UpdateHostInfo(ctx, host)
}

func (p *ResourcePool) CreateDisks(ctx context.Context, hostId string, disks []structs.DiskInfo) (diskIds []string, err error) {
	return p.hostProvider.CreateDisks(ctx, hostId, disks)
}

func (p *ResourcePool) DeleteDisks(ctx context.Context, diskIds []string) (err error) {
	return p.hostProvider.DeleteDisks(ctx, diskIds)
}

func (p *ResourcePool) UpdateDisk(ctx context.Context, disk structs.DiskInfo) (err error) {
	return p.hostProvider.UpdateDisk(ctx, disk)
}

func (p *ResourcePool) selectImportFlowName(condition *structs.ImportCondition) (flowName string) {
	if framework.Current.GetClientArgs().SkipHostInit || condition.SkipHostInit {
		flowName = rp_consts.FlowImportHostsWithoutInit
	} else {
		if condition.ReserveHost {
			flowName = rp_consts.FlowTakeOverHosts
		} else {
			flowName = rp_consts.FlowImportHosts
		}
	}
	return
}

func (p *ResourcePool) selectDeleteFlowName(host *structs.HostInfo, force bool) (flowName string) {
	if framework.Current.GetClientArgs().SkipHostInit || force || host.Status == string(constants.HostFailed) {
		flowName = rp_consts.FlowDeleteHostsByForce
	} else {
		flowName = rp_consts.FlowDeleteHosts
	}
	return
}
