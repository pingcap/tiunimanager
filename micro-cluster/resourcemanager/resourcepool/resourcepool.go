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
	"sync"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostinitiator"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostprovider"
	"github.com/pingcap-inc/tiem/workflow"
)

type ResourcePool struct {
	hostProvider  hostprovider.HostProvider
	hostInitiator hostinitiator.HostInitiator
	// cloudHostProvider hostprovider.HostProvider
}

var resourcePool *ResourcePool
var once sync.Once

func GetResourcePool() *ResourcePool {
	once.Do(func() {
		if resourcePool == nil {
			resourcePool = new(ResourcePool)
			resourcePool.InitResourcePool()
		}
	})
	return resourcePool
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
			"start":           {Name: "start", SuccessEvent: "verifyHosts", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: authHosts},
			"verifyHosts":     {Name: "verifyHosts", SuccessEvent: "configHosts", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: verifyHosts},
			"configHosts":     {Name: "configHosts", SuccessEvent: "installSoftware", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: configHosts},
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
			"start":         {Name: "start", SuccessEvent: "joinEMCluster", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: authHosts},
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

func (p *ResourcePool) ImportHosts(ctx context.Context, hosts []structs.HostInfo, condition *structs.ImportCondition) (flowIds []string, hostIds []string, err error) {
	hostIds, err = p.hostProvider.ImportHosts(ctx, hosts)
	if err != nil {
		return flowIds, hostIds, err
	}
	var flows []*workflow.WorkFlowAggregation
	flowManager := workflow.GetWorkFlowService()
	flowName := p.selectImportFlowName(condition)
	framework.LogWithContext(ctx).Infof("import hosts select %s", flowName)
	for i, host := range hosts {
		flow, err := flowManager.CreateWorkFlow(ctx, hostIds[i], workflow.BizTypeHost, flowName)
		if err != nil {
			errMsg := fmt.Sprintf("create %s workflow failed for host %s %s, %s", flowName, host.HostName, host.IP, err.Error())
			framework.LogWithContext(ctx).Errorln(errMsg)
			return nil, hostIds, errors.WrapError(errors.TIEM_WORKFLOW_CREATE_FAILED, errMsg, err)
		}

		flowManager.AddContext(flow, rp_consts.ContextResourcePoolKey, p)
		flowManager.AddContext(flow, rp_consts.ContextHostInfoArrayKey, []structs.HostInfo{host})
		flowManager.AddContext(flow, rp_consts.ContextHostIDArrayKey, []string{hostIds[i]})

		flows = append(flows, flow)
		flowIds = append(flowIds, flow.Flow.ID)
	}
	// Sync start each flow in a goroutine: tiup-tiem/sqlite DO NOT support concurrent
	go func() {
		for i, flow := range flows {
			if err = flowManager.Start(ctx, flow); err != nil {
				errMsg := fmt.Sprintf("sync start %s workflow[%d] %s failed for host %s %s, %s", rp_consts.FlowImportHosts, i, flow.Flow.ID, hosts[i].HostName, hosts[i].IP, err.Error())
				framework.LogWithContext(ctx).Errorln(errMsg)
				continue
			} else {
				framework.LogWithContext(ctx).Infof("sync start %s workflow[%d] %s for host %s %s", rp_consts.FlowImportHosts, i, flow.Flow.ID, hosts[i].HostName, hosts[i].IP)
			}
		}
	}()
	return flowIds, hostIds, nil
}

func (p *ResourcePool) DeleteHosts(ctx context.Context, hostIds []string, force bool) (flowIds []string, err error) {
	err = p.UpdateHostStatus(ctx, hostIds, string(constants.HostDeleting))
	if err != nil {
		return nil, err
	}
	var flows []*workflow.WorkFlowAggregation
	flowManager := workflow.GetWorkFlowService()
	flowName := p.selectDeleteFlowName(force)
	framework.LogWithContext(ctx).Infof("delete hosts select %s", flowName)
	for _, hostId := range hostIds {
		flow, err := flowManager.CreateWorkFlow(ctx, hostId, workflow.BizTypeHost, flowName)
		if err != nil {
			errMsg := fmt.Sprintf("create %s workflow failed for host %s, %s", flowName, hostId, err.Error())
			framework.LogWithContext(ctx).Errorln(errMsg)
			return nil, errors.WrapError(errors.TIEM_WORKFLOW_CREATE_FAILED, errMsg, err)
		}

		flowManager.AddContext(flow, rp_consts.ContextResourcePoolKey, p)
		flowManager.AddContext(flow, rp_consts.ContextHostIDArrayKey, []string{hostId})

		flows = append(flows, flow)
		flowIds = append(flowIds, flow.Flow.ID)
	}

	go func() {
		for i, flow := range flows {
			if err = flowManager.Start(ctx, flow); err != nil {
				errMsg := fmt.Sprintf("sync start %s workflow[%d] %s failed for delete host %s, %s", rp_consts.FlowDeleteHosts, i, flow.Flow.ID, hostIds[i], err.Error())
				framework.LogWithContext(ctx).Errorln(errMsg)
			} else {
				framework.LogWithContext(ctx).Infof("sync start %s workflow[%d] %s for delete host %s", rp_consts.FlowDeleteHosts, i, flow.Flow.ID, hostIds[i])
			}
		}
	}()
	return flowIds, nil
}

func (p *ResourcePool) QueryHosts(ctx context.Context, filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, total int64, err error) {
	return p.hostProvider.QueryHosts(ctx, filter, page)
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

func (p *ResourcePool) selectImportFlowName(condition *structs.ImportCondition) (flowName string) {
	if condition.ReserveHost {
		flowName = rp_consts.FlowTakeOverHosts
	} else {
		if framework.Current.GetClientArgs().SkipHostInit || condition.SkipHostInit {
			flowName = rp_consts.FlowImportHostsWithoutInit
		} else {
			flowName = rp_consts.FlowImportHosts
		}
	}
	return
}

func (p *ResourcePool) selectDeleteFlowName(force bool) (flowName string) {
	if framework.Current.GetClientArgs().SkipHostInit || force {
		flowName = rp_consts.FlowDeleteHostsByForce
	} else {
		flowName = rp_consts.FlowDeleteHosts
	}
	return
}
