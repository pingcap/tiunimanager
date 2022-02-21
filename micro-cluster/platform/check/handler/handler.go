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
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
)

type Report struct {
	Info *structs.CheckReportInfo
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

func (p *Report) CheckClusters(ctx context.Context, clusterMetas []*management.Result) ([]structs.ClusterCheck, error) {

	return nil, nil
}

func (p *Report) CheckTenant(ctx context.Context, tenantID string) error {
	// get cluster count
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
		for _, instance := range meta.Instances {
			allocatedCPUCores += int(instance.CpuCores)
			allocatedMemory += int(instance.Memory)
			allocatedStorage += int(instance.DiskCapacity)
		}
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

func (p *Report) CheckHosts(ctx context.Context) error {
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
