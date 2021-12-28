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
	flowManager.RegisterWorkFlow(context.TODO(), rp_consts.FlowImportHosts, &workflow.WorkFlowDefine{
		FlowName: rp_consts.FlowImportHosts,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start":           {Name: "start", SuccessEvent: "configHosts", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: verifyHosts},
			"configHosts":     {Name: "configHosts", SuccessEvent: "installSoftware", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: configHosts},
			"installSoftware": {Name: "installSoftware", SuccessEvent: "succeed", FailEvent: "fail", ReturnType: workflow.SyncFuncNode, Executor: installSoftware},
			"succeed":         {Name: "succeed", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: importHostSucceed},
			"fail":            {Name: "fail", SuccessEvent: "", FailEvent: "", ReturnType: workflow.SyncFuncNode, Executor: importHostsFail},
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

func (p *ResourcePool) ImportHosts(ctx context.Context, hosts []structs.HostInfo) (flowIds []string, hostIds []string, err error) {
	for _, host := range hosts {
		hostId, err := p.hostProvider.ImportHosts(ctx, []structs.HostInfo{host})
		if err != nil {
			return flowIds, hostIds, err
		}
		flowManager := workflow.GetWorkFlowService()
		flow, err := flowManager.CreateWorkFlow(ctx, hostId[0], rp_consts.FlowImportHosts)
		if err != nil {
			errMsg := fmt.Sprintf("create %s workflow failed for host %s %s, %s", rp_consts.FlowImportHosts, host.HostName, host.IP, err.Error())
			framework.LogWithContext(ctx).Errorln(errMsg)
			return flowIds, hostIds, errors.WrapError(errors.TIEM_WORKFLOW_CREATE_FAILED, errMsg, err)
		}

		flowManager.AddContext(flow, rp_consts.ContextResourcePoolKey, p)
		flowManager.AddContext(flow, rp_consts.ContextImportHostInfoKey, hosts)
		flowManager.AddContext(flow, rp_consts.ContextImportHostIDsKey, hostIds)
		if err = flowManager.AsyncStart(ctx, flow); err != nil {
			errMsg := fmt.Sprintf("async start %s workflow failed for host %s %s, %s", host.HostName, host.IP, rp_consts.FlowImportHosts, err.Error())
			framework.LogWithContext(ctx).Errorln(errMsg)
			return flowIds, hostIds, errors.WrapError(errors.TIEM_WORKFLOW_START_FAILED, errMsg, err)
		}
		flowIds = append(flowIds, flow.Flow.ID)
		hostIds = append(hostIds, hostId...)
	}
	return flowIds, hostIds, nil
}

func (p *ResourcePool) DeleteHosts(ctx context.Context, hostIds []string) (err error) {
	return p.hostProvider.DeleteHosts(ctx, hostIds)
}

func (p *ResourcePool) QueryHosts(ctx context.Context, filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, err error) {
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

func (p *ResourcePool) GetStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks *structs.Stocks, err error) {
	return p.hostProvider.GetStocks(ctx, location, hostFilter, diskFilter)
}
