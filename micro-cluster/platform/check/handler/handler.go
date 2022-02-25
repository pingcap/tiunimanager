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
	"github.com/fatih/color"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/deployment"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/parameter"
	hostInspector "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/inspect"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	util "github.com/pingcap-inc/tiem/util/http"
	"github.com/pingcap/kvproto/pkg/metapb"
	"github.com/pingcap/kvproto/pkg/pdpb"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
)

const GetClusterInfoCmd = "SELECT TYPE as type, count(TYPE) as count FROM information_schema.cluster_info GROUP BY TYPE;"

type TopologyInfo struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}
type Report struct {
	Info *structs.CheckReportInfo
}

type Replication struct {
	MaxReplicas int32 `json:"max-replicas"`
}

// ReplicationStatus represents the replication mode status of the region.
type ReplicationStatus struct {
	State   string `json:"state"`
	StateID uint64 `json:"state_id"`
}

// RegionInfo records detail region info for api usage.
type RegionInfo struct {
	ID          uint64              `json:"id"`
	StartKey    string              `json:"start_key"`
	EndKey      string              `json:"end_key"`
	RegionEpoch *metapb.RegionEpoch `json:"epoch,omitempty"`
	Peers       []*metapb.Peer      `json:"peers,omitempty"`

	Leader          *metapb.Peer      `json:"leader,omitempty"`
	DownPeers       []*pdpb.PeerStats `json:"down_peers,omitempty"`
	PendingPeers    []*metapb.Peer    `json:"pending_peers,omitempty"`
	WrittenBytes    uint64            `json:"written_bytes"`
	ReadBytes       uint64            `json:"read_bytes"`
	WrittenKeys     uint64            `json:"written_keys"`
	ReadKeys        uint64            `json:"read_keys"`
	ApproximateSize int64             `json:"approximate_size"`
	ApproximateKeys int64             `json:"approximate_keys"`

	ReplicationStatus *ReplicationStatus `json:"replication_status,omitempty"`
}

// RegionsInfo contains some regions with the detailed region info.
type RegionsInfo struct {
	Count   int           `json:"count"`
	Regions []*RegionInfo `json:"regions"`
}

type HealthInfo struct {
	Name     string `json:"name"`
	MemberID int64  `json:"member_id"`
	Health   bool   `json:"health"`
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
			fmt.Sprintf("parse max replicas error: %s", err.Error()), err)
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
		accountStatus.Message = "Account status are healthy."
	}

	return accountStatus, nil
}

func (p *Report) GetClusterTopology(ctx context.Context, clusterID string) (structs.CheckString, error) {
	topologyCheck := structs.CheckString{}

	clusterMeta, err := meta.Get(ctx, clusterID)
	if err != nil {
		return topologyCheck, err
	}

	db, err := meta.CreateSQLLink(ctx, clusterMeta)
	if err != nil {
		return topologyCheck, errors.WrapError(errors.TIEM_CONNECT_TIDB_ERROR, err.Error(), err)
	}
	defer db.Close()

	rows, err := db.Query(GetClusterInfoCmd)
	if err != nil {
		return topologyCheck, err
	}
	realTopology := make(map[string]int)
	realTopologyInfos := make([]TopologyInfo, 0)
	for rows.Next() {
		var topologyInfo TopologyInfo
		err = rows.Scan(&topologyInfo.Type, &topologyInfo.Count)
		if err != nil {
			return topologyCheck, err
		}
		if _, ok := realTopology[topologyInfo.Type]; !ok {
			realTopology[topologyInfo.Type] = topologyInfo.Count
			realTopologyInfos = append(realTopologyInfos, topologyInfo)
		}
	}

	topologyCheck.Valid = true
	expectedTopologyInfos := make([]TopologyInfo, 0)
	for componentType, instances := range clusterMeta.Instances {
		if meta.Contain(constants.ParasiteComponentIDs, componentType) {
			continue
		}
		if realTopology[strings.ToLower(componentType)] != len(instances) {
			topologyCheck.Valid = false
		}
		expectedTopologyInfos = append(expectedTopologyInfos, TopologyInfo{
			Type:  strings.ToLower(componentType),
			Count: len(instances),
		})
	}

	realInfos, err := json.Marshal(realTopologyInfos)
	if err != nil {
		return topologyCheck, err
	}

	expectedInfos, err := json.Marshal(expectedTopologyInfos)
	if err != nil {
		return topologyCheck, err
	}
	topologyCheck.RealValue = string(realInfos)
	topologyCheck.ExpectedValue = string(expectedInfos)

	return topologyCheck, nil
}

func (p *Report) GetClusterRegionStatus(ctx context.Context, clusterID string) (structs.CheckStatus, error) {
	regionStatus := structs.CheckStatus{}

	clusterMeta, err := meta.Get(ctx, clusterID)
	if err != nil {
		return regionStatus, err
	}

	pdAddress := clusterMeta.GetPDClientAddresses()
	if len(pdAddress) <= 0 {
		return regionStatus, errors.NewError(errors.TIEM_PD_NOT_FOUND_ERROR, "cluster not found pd instance")
	}

	pdID := strings.Join([]string{pdAddress[0].IP, strconv.Itoa(pdAddress[0].Port)}, ":")

	hasUnhealthy := false
	for _, state := range []string{"miss-peer", "pending-peer"} {
		url := fmt.Sprintf("http://%s/pd/api/v1/regions/check/%s", pdID, state)
		params := make(map[string]string)
		headers := make(map[string]string)
		resp, err := util.Get(url, params, headers)
		if err != nil {
			return regionStatus, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return regionStatus, err
		}
		regionsInfo := &RegionsInfo{}
		if err = json.Unmarshal(body, regionsInfo); err != nil {
			return regionStatus, errors.WrapError(errors.TIEM_UNMARSHAL_ERROR,
				fmt.Sprintf("parse regions info error: %s", err.Error()), err)
		}
		if regionsInfo.Count > 0 {
			regionStatus.Health = false
			regionStatus.Message = fmt.Sprintf("Regions are not fully healthy: %s",
				color.YellowString("%d %s", regionsInfo.Count, state))
			hasUnhealthy = true
		}
	}
	if !hasUnhealthy {
		regionStatus.Health = true
		regionStatus.Message = "All regions are healthy."
	}

	return regionStatus, nil
}

func (p *Report) GetClusterHealthStatus(ctx context.Context, clusterID string) (structs.CheckStatus, error) {
	healthStatus := structs.CheckStatus{}

	clusterMeta, err := meta.Get(ctx, clusterID)
	if err != nil {
		return healthStatus, err
	}

	pdAddress := clusterMeta.GetPDClientAddresses()
	if len(pdAddress) <= 0 {
		return healthStatus, errors.NewError(errors.TIEM_PD_NOT_FOUND_ERROR, "cluster not found pd instance")
	}

	pdID := strings.Join([]string{pdAddress[0].IP, strconv.Itoa(pdAddress[0].Port)}, ":")
	url := fmt.Sprintf("http://%s/pd/api/v1/health", pdID)
	params := make(map[string]string)
	headers := make(map[string]string)
	resp, err := util.Get(url, params, headers)
	if err != nil {
		return healthStatus, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return healthStatus, err
	}

	healthsInfo := make([]HealthInfo, 0)
	if err = json.Unmarshal(body, healthsInfo); err != nil {
		return healthStatus, errors.WrapError(errors.TIEM_UNMARSHAL_ERROR,
			fmt.Sprintf("parse health info error: %s", err.Error()), err)
	}

	hasUnhealthy := false
	for _, info := range healthsInfo {
		if !info.Health {
			hasUnhealthy = true
			healthStatus.Health = false
			healthStatus.Message = fmt.Sprintf("Cluster node %s are not fully healthy", info.Name)
		}
	}

	if !hasUnhealthy {
		healthStatus.Health = true
		healthStatus.Message = "Cluster are fully healthy"
	}

	return healthStatus, nil
}

func (p *Report) CheckInstanceParameters(ctx context.Context, instanceID string) (map[string]structs.CheckAny, error) {
	checkParams := make(map[string]structs.CheckAny)

	resp, err := parameter.NewManager().InspectClusterParameters(ctx, cluster.InspectParametersReq{InstanceID: instanceID})
	if err != nil {
		return checkParams, err
	}

	for _, param := range resp.Params[0].ParameterInfos {
		paramName := strings.Join([]string{param.Category, param.Name}, ".")
		if _, ok := checkParams[paramName]; !ok {
			checkParams[paramName] = structs.CheckAny{
				Valid:         false,
				RealValue:     param.InspectValue,
				ExpectedValue: param.RealValue.ClusterValue,
			}
		}
	}
	return checkParams, nil
}

func (p *Report) CheckInstances(ctx context.Context, instances []*management.ClusterInstance) ([]structs.InstanceCheck, error) {
	instanceChecks := make([]structs.InstanceCheck, 0)

	for _, instance := range instances {
		parameters, err := p.CheckInstanceParameters(ctx, instance.ID)
		if err != nil {
			return instanceChecks, err
		}
		instanceChecks = append(instanceChecks, structs.InstanceCheck{
			ID:         instance.ID,
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
		topologyCheck, err := p.GetClusterTopology(ctx, meta.Cluster.ID)
		if err != nil {
			return clusterChecks, err
		}
		regionStatus, err := p.GetClusterRegionStatus(ctx, meta.Cluster.ID)
		if err != nil {
			return clusterChecks, err
		}
		healthStatus, err := p.GetClusterHealthStatus(ctx, meta.Cluster.ID)
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
			HealthStatus:  healthStatus,
			Topology:      topologyCheck,
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
