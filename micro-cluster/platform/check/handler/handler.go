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

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/deployment"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	hostInspector "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/inspect"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"math"
	"strconv"
	"strings"
)

type Report struct {
	Info *structs.CheckReportInfo
}

type Replication struct {
	MaxReplicas int32 `json:"max-replicas"`
}

func (p *Report) ParseFrom(ctx context.Context, checkID string) error {
	rw := models.GetReportReaderWriter()
	reportInfo, err := rw.GetReport(ctx, checkID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get report %s error: %s", checkID, err.Error())
		return err
	}
	p.Info = &reportInfo

	return nil
}

func (p *Report) CheckTenants(ctx context.Context) error {
	// query all tenants
	tenants, err := models.GetAccountReaderWriter().QueryTenants(ctx)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query tenants error: %s", err.Error())
		return err
	}

	for _, tenant := range tenants {
		err = p.CheckTenant(ctx, tenant.ID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf(
				"check tenant %s error: %s", tenant.ID, err.Error())
			return err
		}
	}

	return nil
}

func (p *Report) GetClusterAllocatedResource(ctx context.Context, meta *management.Result) (allocatedCPUCores int32,
	allocatedMemory int32, allocatedStorage int32) {
	for _, instance := range meta.Instances {
		allocatedCPUCores += int32(instance.CpuCores)
		allocatedMemory += int32(instance.Memory)
		allocatedStorage += instance.DiskCapacity
	}
	return allocatedCPUCores, allocatedMemory, allocatedStorage
}

func (p *Report) GetClusterCopies(ctx context.Context, clusterID string) (int32, error) {
	// get cluster meta
	clusterMeta, err := meta.Get(ctx, clusterID)
	if err != nil {
		return 0, err
	}

	pdAddress := clusterMeta.GetPDClientAddresses()
	if len(pdAddress) <= 0 {
		return 0, errors.NewError(errors.TIEM_PD_NOT_FOUND_ERROR, "cluster not found pd instance")
	}

	pdID := strings.Join([]string{pdAddress[0].IP, strconv.Itoa(pdAddress[0].Port)}, ":")

	config, err := deployment.M.Ctl(ctx, deployment.TiUPComponentTypeCtrl, clusterMeta.Cluster.Version, spec.ComponentPD,
		"/home/tiem/.tiup", []string{"-u", pdID, "config", "show", "replication"}, meta.DefaultTiupTimeOut)
	if err != nil {
		return 0, err
	}

	replication := &Replication{}
	if err = json.Unmarshal([]byte(config), replication); err != nil {
		return 0, errors.WrapError(errors.TIEM_UNMARSHAL_ERROR,
			fmt.Sprintf("parse placement rules error: %s", err.Error()), err)
	}

	return replication.MaxReplicas, nil
}

func (p *Report) GetClusterAccountStatus(ctx context.Context, clusterID string) (structs.CheckStatus, error) {
	accountStatus := structs.CheckStatus{}
	clusterMeta, err := meta.Get(ctx, clusterID)
	if err != nil {
		return accountStatus, err
	}
	_, err = meta.CreateSQLLink(ctx, clusterMeta)
	if err != nil {
		accountStatus.Health = false
		accountStatus.Message = err.Error()
	} else {
		accountStatus.Health = true
	}

	return accountStatus, nil
}

func (p *Report) GetClusterTopology(ctx context.Context, clusterID string) (string, error) {
	return "", nil
}

func (p *Report) GetClusterRegionStatus(ctx context.Context, clusterID string) (structs.CheckStatus, error) {
	return structs.CheckStatus{}, nil
}

func (p *Report) GetInstanceStatus(ctx context.Context, instanceID string) (structs.CheckStatus, error) {
	return structs.CheckStatus{}, nil
}

func (p *Report) CheckInstanceParameters(ctx context.Context, instanceID string) (map[string]structs.CheckAny, error) {
	return nil, nil
}

func (p *Report) CheckInstances(ctx context.Context, instances []*management.ClusterInstance) ([]structs.InstanceCheck, error) {
	instanceChecks := make([]structs.InstanceCheck, 0)

	for _, instance := range instances {
		status, err := p.GetInstanceStatus(ctx, instance.ID)
		if err != nil {
			return instanceChecks, err
		}
		parameters, err := p.CheckInstanceParameters(ctx, instance.ID)
		if err != nil {
			return instanceChecks, err
		}
		instanceChecks = append(instanceChecks, structs.InstanceCheck{
			ID:         instance.ID,
			Status:     status,
			Parameters: parameters,
		})
	}
	return instanceChecks, nil
}

func (p *Report) CheckClusters(ctx context.Context, clusterMetas []*management.Result) ([]structs.ClusterCheck, error) {
	clusterChecks := make([]structs.ClusterCheck, 0)

	for _, meta := range clusterMetas {
		allocatedCPUCores, allocatedMemory, allocatedStorage := p.GetClusterAllocatedResource(ctx, meta)
		copies, err := p.GetClusterCopies(ctx, meta.Cluster.ID)
		if err != nil {
			return clusterChecks, err
		}
		accountStatus, err := p.GetClusterAccountStatus(ctx, meta.Cluster.ID)
		if err != nil {
			return clusterChecks, err
		}
		topology, err := p.GetClusterTopology(ctx, meta.Cluster.ID)
		if err != nil {
			return clusterChecks, err
		}
		regionStatus, err := p.GetClusterRegionStatus(ctx, meta.Cluster.ID)
		if err != nil {
			return clusterChecks, err
		}
		instanceChecks, err := p.CheckInstances(ctx, meta.Instances)
		if err != nil {
			return clusterChecks, err
		}
		clusterChecks = append(clusterChecks, structs.ClusterCheck{
			ID:                meta.Cluster.ID,
			MaintenanceStatus: meta.Cluster.MaintenanceStatus,
			CPU:               allocatedCPUCores,
			Memory:            allocatedMemory,
			Storage:           allocatedStorage,
			Copies: structs.CheckInt32{
				Valid:         copies == int32(meta.Cluster.Copies),
				RealValue:     copies,
				ExpectedValue: int32(meta.Cluster.Copies),
			},
			AccountStatus: accountStatus,
			Topology:      topology,
			RegionStatus:  regionStatus,
			Instances:     instanceChecks,
		})
	}
	return clusterChecks, nil
}

func (p *Report) CheckTenant(ctx context.Context, tenantID string) error {
	// get clusters from tenantID
	clusterMetas, err := models.GetClusterReaderWriter().QueryClusters(ctx, tenantID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"get cluster metas from tenant %s error: %s", tenantID, err.Error())
		return err
	}

	// check clusters
	clusterChecks, err := p.CheckClusters(ctx, clusterMetas)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"check clusters from tenant %s error: %s", tenantID, err.Error())
		return err
	}

	// get allocated resources from specified tenant
	allocatedCPUCores := 0
	allocatedMemory := 0
	allocatedStorage := 0
	for _, meta := range clusterMetas {
		cpuCores, memory, storage := p.GetClusterAllocatedResource(ctx, meta)
		allocatedCPUCores += int(cpuCores)
		allocatedMemory += int(memory)
		allocatedStorage += int(storage)
	}

	// get tenant info
	tenantInfo, err := models.GetAccountReaderWriter().GetTenant(ctx, tenantID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get tenant %s error: %s", tenantID, err.Error())
		return err
	}
	if len(p.Info.Tenants) == 0 {
		p.Info.Tenants = make(map[string]structs.TenantCheck)
	}
	if _, ok := p.Info.Tenants[tenantID]; !ok {
		p.Info.Tenants[tenantID] = structs.TenantCheck{
			ClusterCount: structs.CheckRangeInt32{
				Valid:         len(clusterMetas) >= 0 && int32(len(clusterMetas)) <= tenantInfo.MaxCluster,
				RealValue:     int32(len(clusterMetas)),
				ExpectedRange: []int32{0, tenantInfo.MaxCluster},
			},
			CPURatio:     float32(allocatedCPUCores) / float32(tenantInfo.MaxCPU),
			MemoryRatio:  float32(allocatedMemory) / float32(tenantInfo.MaxMemory),
			StorageRatio: float32(allocatedStorage) / float32(tenantInfo.MaxStorage),
			Clusters:     clusterChecks,
		}
	}

	return nil
}

func (p *Report) CheckHostAllocatedResource(ctx context.Context, hostInfos []structs.HostInfo) (map[string]*structs.CheckInt32,
	map[string]*structs.CheckInt32, map[string]map[string]*structs.CheckString, error) {

	cpuAllocated, err := hostInspector.GetHostInspector().CheckCpuAllocated(ctx, hostInfos)
	if err != nil {
		return nil, nil, nil, err
	}

	memAllocated, err := hostInspector.GetHostInspector().CheckMemAllocated(ctx, hostInfos)
	if err != nil {
		return nil, nil, nil, err
	}

	diskAllocated, err := hostInspector.GetHostInspector().CheckDiskAllocated(ctx, hostInfos)
	if err != nil {
		return nil, nil, nil, err
	}

	return cpuAllocated, memAllocated, diskAllocated, nil
}

func (p *Report) CheckHosts(ctx context.Context) error {
	// query all hosts
	rw := models.GetResourceReaderWriter()
	hosts, _, err := rw.Query(ctx, &structs.Location{}, &structs.HostFilter{}, 0, math.MaxInt32)
	if err != nil {
		return err
	}

	checkHosts := make(map[string]structs.HostCheck)
	hostInfos := make([]structs.HostInfo, 0)
	for _, host := range hosts {
		hostInfos = append(hostInfos, structs.HostInfo{ID: host.ID})
	}

	cpu, memory, disk, err := p.CheckHostAllocatedResource(ctx, hostInfos)
	if err != nil {
		return err
	}

	for key, _ := range cpu {
		diskAllocated := make(map[string]structs.CheckString)
		for path, value := range disk[key] {
			if _, ok := diskAllocated[path]; !ok {
				diskAllocated[path] = *value
			}
		}
		if _, ok := checkHosts[key]; !ok {
			checkHosts[key] = structs.HostCheck{
				CPUAllocated:    *cpu[key],
				MemoryAllocated: *memory[key],
				DiskAllocated:   diskAllocated,
			}
		}
	}

	p.Info.Hosts = structs.HostsCheck{
		Hosts: checkHosts,
	}
	return nil
}

func (p *Report) Serialize(ctx context.Context) (string, error) {
	report, err := json.Marshal(p.Info)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("serialize report info error: %s", err.Error())
		return "", err
	}
	return string(report), nil
}
