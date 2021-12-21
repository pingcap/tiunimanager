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

package handler

import (
	"bytes"
	"context"
	"github.com/pingcap-inc/tiem/message/cluster"
	"text/template"

	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	newConstants "github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	resource "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
)

type ClusterMeta struct {
	Cluster              *management.Cluster
	Instances            map[string][]*management.ClusterInstance
	NodeExporterPort     int32
	BlackboxExporterPort int32
}

// BuildCluster
// @Description: build cluster from structs.CreateClusterParameter
// @Receiver p
// @Parameter ctx
// @Parameter param
// @return error
func (p *ClusterMeta) BuildCluster(ctx context.Context, param structs.CreateClusterParameter) error {
	p.Cluster = &management.Cluster{
		Entity: dbCommon.Entity{
			TenantId: framework.GetTenantIDFromContext(ctx),
			Status:   string(constants.ClusterInitializing),
		},
		Name:              param.Name,
		DBPassword:        param.DBPassword,
		Type:              param.Type,
		Version:           param.Version,
		TLS:               param.TLS,
		Tags:              param.Tags,
		OwnerId:           framework.GetUserIDFromContext(ctx),
		ParameterGroupID:  param.ParameterGroupID,
		Copies:            param.Copies,
		Exclusive:         param.Exclusive,
		Region:            param.Region,
		CpuArchitecture:   constants.ArchType(param.CpuArchitecture),
		MaintenanceStatus: constants.ClusterMaintenanceNone,
		MaintainWindow:    "",
	}
	// default user
	if len(param.DBUser) == 0 {
		p.Cluster.DBUser = "root"
	} else {
		p.Cluster.DBUser = param.DBUser
	}
	_, err := models.GetClusterReaderWriter().Create(ctx, p.Cluster)
	if err == nil {
		framework.LogWithContext(ctx).Infof("create cluster [%s] succeed", p.Cluster.Name)
	} else {
		framework.LogWithContext(ctx).Errorf("create cluster [%s] failed, err : %s", p.Cluster.Name, err.Error())
	}
	return err
}

// AddInstances
// @Description add new instances into cluster topology, then alloc host ip, port and disk for these instances
// @Parameter	computes
// @Return		error
func (p *ClusterMeta) AddInstances(ctx context.Context, computes []structs.ClusterResourceParameterCompute) error {
	if len(computes) <= 0 {
		return framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "parameter is invalid!")
	}

	if p.Cluster == nil {
		return framework.NewTiEMError(common.TIEM_UNRECOGNIZED_ERROR, "cluster is nil!")
	}

	if len(p.Instances) == 0 {
		p.Instances = make(map[string][]*management.ClusterInstance)
	}

	for _, compute := range computes {
		for _, item := range compute.Resource {
			for i := 0; i < item.Count; i++ {
				instance := &management.ClusterInstance {
					Entity: dbCommon.Entity{
						TenantId: p.Cluster.TenantId,
						Status:   string(constants.ClusterInstanceInitializing),
					},
					Type:         compute.Type,
					Version:      p.Cluster.Version,
					ClusterID:    p.Cluster.ID,
					CpuCores:     int8(knowledge.ParseCpu(item.Spec)),
					Memory:       int8(knowledge.ParseMemory(item.Spec)),
					Zone:         structs.GetDomainNameFromCode(item.Zone),
					DiskType:     item.DiskType,
					DiskCapacity: int32(item.DiskCapacity),
				}
				p.Instances[compute.Type] = append(p.Instances[compute.Type], instance)
			}
		}
	}
	framework.LogWithContext(ctx).Infof("add new instances into cluster[%s] topology", p.Cluster.ID)
	return nil
}

func (p *ClusterMeta) AddDefaultInstances(ctx context.Context) error {
	if string(newConstants.EMProductIDTiDB) == p.Cluster.Type {
		for _, t := range newConstants.ParasiteComponentIDs {
			instance := &management.ClusterInstance{
				Entity: dbCommon.Entity {
					TenantId: p.Cluster.TenantId,
					Status:   string(constants.ClusterInstanceInitializing),
				},
				Type:      string(t),
				Version:   p.Cluster.Version,
				ClusterID: p.Cluster.ID,
			}
			p.Instances[string(t)] = append(p.Instances[string(t)], instance)
		}

	}

	return nil
}

func (p *ClusterMeta) GenerateInstanceResourceRequirements(ctx context.Context) ([]resource.AllocRequirement, []*management.ClusterInstance, error) {
	instances := p.GetInstanceByStatus(ctx, constants.ClusterInstanceInitializing)
	requirements := make([]resource.AllocRequirement, 0)

	allocInstances := make([]*management.ClusterInstance, 0)
	for _, instance := range instances {
		if Contain(newConstants.ParasiteComponentIDs, newConstants.EMProductComponentIDType(instance.Type)) {
			continue
		}
		portRange := knowledge.GetComponentPortRange(p.Cluster.Type, p.Cluster.Version, instance.Type)
		requirements = append(requirements, resource.AllocRequirement{
			Location: structs.Location{
				Region: p.Cluster.Region,
				Zone:   instance.Zone,
			},
			Require: resource.Requirement{
				Exclusive: p.Cluster.Exclusive,
				PortReq: []resource.PortRequirement{
					{
						Start:   int32(portRange.Start),
						End:     int32(portRange.End),
						PortCnt: int32(portRange.Count),
					},
				},
				DiskReq: resource.DiskRequirement{
					NeedDisk: true,
					Capacity: instance.DiskCapacity,
					DiskType: instance.DiskType,
				},
				ComputeReq: resource.ComputeRequirement{
					ComputeResource: resource.ComputeResource{
						CpuCores: int32(instance.CpuCores),
						Memory:   int32(instance.Memory),
					},
				},
			},
			Count: 1,
			HostFilter: resource.Filter{
				Arch: string(p.Cluster.CpuArchitecture),
			},
			Strategy: resource.RandomRack,
		})
		allocInstances = append(allocInstances, instance)
	}
	return requirements, allocInstances, nil
}

func (p *ClusterMeta) GenerateGlobalPortRequirements(ctx context.Context) ([]resource.AllocRequirement, error) {
	requirements := make([]resource.AllocRequirement, 0)

	if p.Cluster.Status != string(constants.ClusterInitializing) {
		framework.LogWithContext(ctx).Infof("cluster [%s] is not initializing, no need to alloc global port resource", p.Cluster.Name)
		return requirements, nil
	}

	portRange := knowledge.GetClusterPortRange(p.Cluster.Type, p.Cluster.Version)
	requirements = append(requirements, resource.AllocRequirement{
		Location: structs.Location{Region: p.Cluster.Region},
		Require: resource.Requirement{
			PortReq: []resource.PortRequirement{
				{
					Start:   int32(portRange.Start),
					End:     int32(portRange.End),
					PortCnt: int32(portRange.Count),
				},
			},
			DiskReq: resource.DiskRequirement{NeedDisk: false},
			ComputeReq: resource.ComputeRequirement{
				ComputeResource: resource.ComputeResource{},
			},
		},
		Count: 1,
		HostFilter: resource.Filter{
			Arch: string(p.Cluster.CpuArchitecture),
		},
		Strategy: resource.ClusterPorts,
	})

	return requirements, nil
}

func (p *ClusterMeta) ApplyGlobalPortResource(nodeExporterPort, blackboxExporterPort int32) {
	p.NodeExporterPort = nodeExporterPort
	p.BlackboxExporterPort = blackboxExporterPort
}

func (p *ClusterMeta) ApplyInstanceResource(resource *resource.AllocRsp, instances []*management.ClusterInstance) {
	for i, instance := range instances {
		instance.HostID = resource.Results[i].HostId
		instance.HostIP = append(instance.HostIP, resource.Results[i].HostIp)
		instance.Ports = resource.Results[i].PortRes[0].Ports
		instance.DiskID = resource.Results[i].DiskRes.DiskId
		instance.DiskPath = resource.Results[i].DiskRes.Path
	}
	if p.Cluster.Status == string(constants.ClusterInstanceInitializing) {
		pd := p.Instances[string(newConstants.ComponentIDPD)][0]
		for _, t := range newConstants.ParasiteComponentIDs {
			instance := p.Instances[string(t)][0]
			instance.HostID = pd.HostID
			instance.HostIP = pd.HostIP
			instance.DiskID = pd.DiskID
			instance.DiskPath = pd.DiskPath
			switch t {
			case newConstants.ComponentIDGrafana:
				instance.Ports = pd.Ports[3:4]
				continue
			case newConstants.ComponentIDPrometheus:
				instance.Ports = pd.Ports[2:3]
				continue
			case newConstants.ComponentIDAlertManger:
				instance.Ports = pd.Ports[4:6]
				continue
			default:
			}
		}
	}
}

func (p *ClusterMeta) GetInstanceByStatus(ctx context.Context, status constants.ClusterInstanceRunningStatus) []*management.ClusterInstance {
	instances := make([]*management.ClusterInstance, 0)
	for _, components := range p.Instances {
		for _, instance := range components {
			if instance.Status != string(status) {
				continue
			}
			instances = append(instances, instance)
		}
	}
	return instances
}

// GenerateTopologyConfig
// @Description generate yaml config based on cluster topology
// @Return		yaml config
// @Return		error
func (p *ClusterMeta) GenerateTopologyConfig(ctx context.Context) (string, error) {
	if p.Cluster == nil || len(p.Instances) == 0 {
		return "", framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "cluster topology is empty, please check it!")
	}

	t, err := template.New("cluster_topology.yaml").ParseFiles("template/cluster_topology.yaml")
	if err != nil {
		return "", framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, err.Error())
	}

	topology := new(bytes.Buffer)
	if err = t.Execute(topology, p); err != nil {
		return "", framework.NewTiEMError(common.TIEM_UNRECOGNIZED_ERROR, err.Error())
	}
	framework.LogWithContext(ctx).Infof("generate topology config: %s", topology.String())

	return topology.String(), nil
}

// UpdateClusterStatus
// @Description update cluster status
// @Return		error
func (p *ClusterMeta) UpdateClusterStatus(ctx context.Context, status constants.ClusterRunningStatus) error {
	err := models.GetClusterReaderWriter().UpdateStatus(ctx, p.Cluster.ID, status)

	if err != nil {
		framework.LogWithContext(ctx).Infof("update cluster[%s] status into %s failed", p.Cluster.Name, status)
	} else {
		p.Cluster.Status = string(status)
		framework.LogWithContext(ctx).Errorf("update cluster[%s] status into %s succeed", p.Cluster.Name, status)
	}
	return err
}

// GetInstance
// @Description get instance based on instanceID
// @Parameter	instance id
// @Return		instance
// @Return		error
func (p *ClusterMeta) GetInstance(ctx context.Context, instanceID string) (*management.ClusterInstance, error) {
	for _, components := range p.Instances {
		for _, instance := range components {
			if instance.ID == instanceID {
				return instance, nil
			}
		}
	}
	return nil, framework.NewTiEMError(common.TIEM_INSTANCE_NOT_FOUND, "instance not found")
}

// IsComponentRequired
// @Description judge whether component is required
// @Parameter	component type
// @Return		bool
func (p *ClusterMeta) IsComponentRequired(ctx context.Context, componentType string) bool {
	return knowledge.GetComponentSpec(p.Cluster.Type,
		p.Cluster.Version, componentType).ComponentConstraint.ComponentRequired
}

// DeleteInstance
// @Description delete instance from cluster topology based on instance id
// @Parameter	instance id
// @Return		error
func (p *ClusterMeta) DeleteInstance(ctx context.Context, instanceID string) (*management.ClusterInstance, error) {
	instance, err := p.GetInstance(ctx, instanceID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(
			"get instance %s error: %s", instanceID, err.Error())
		return nil, err
	}

	if err = models.GetClusterReaderWriter().DeleteInstance(ctx, instance.ID); err != nil {
		return nil, err
	}

	// delete instance from cluster topology
	for componentType, components := range p.Instances {
		for index, item := range components {
			if item.ID == instance.ID {
				components = append(components[:index], components[index+1:]...)
				p.Instances[componentType] = components
			}
		}
	}

	return instance, nil
}

// CloneMeta
// @Description: clone meta info from cluster based on create cluster parameter
// @Receiver p
// @Parameter ctx
// @return *ClusterMeta
func (p *ClusterMeta) CloneMeta(ctx context.Context, parameter structs.CreateClusterParameter) (*ClusterMeta, error) {
	meta := &ClusterMeta{}
	// clone cluster info
	meta.Cluster = &management.Cluster{
		Entity: dbCommon.Entity{
			TenantId: framework.GetTenantIDFromContext(ctx),
			Status:   string(constants.ClusterInitializing),
		},
		Name:              parameter.Name,       // user specify (required)
		DBUser:            parameter.DBUser,     // user specify (required)
		DBPassword:        parameter.DBPassword, // user specify (required)
		Region:            parameter.Region,     // user specify (required)
		Type:              p.Cluster.Type,       // user not specify
		Version:           p.Cluster.Version,    // user specify (option)
		Tags:              p.Cluster.Tags,       // user specify (option)
		TLS:               p.Cluster.TLS,        // user specify (option)
		OwnerId:           framework.GetUserIDFromContext(ctx),
		ParameterGroupID:  p.Cluster.ParameterGroupID, // user specify (option)
		Copies:            p.Cluster.Copies,           // user specify (option)
		Exclusive:         p.Cluster.Exclusive,        // user specify (option)
		CpuArchitecture:   p.Cluster.CpuArchitecture,  // user specify (option)
		MaintenanceStatus: constants.ClusterMaintenanceNone,
		MaintainWindow:    p.Cluster.MaintainWindow,
	}
	// if user specify cluster version
	if len(parameter.Version) > 0 {
		if parameter.Version < p.Cluster.Version {
			return nil, framework.NewTiEMError(common.TIEM_CHECK_CLUSTER_VERSION_ERROR,
				"the specified cluster version is less than source cluster version")
		}
		meta.Cluster.Version = parameter.Version
	}
	// if user specify cluster tags
	if len(parameter.Tags) > 0 {
		meta.Cluster.Tags = parameter.Tags
	}
	// if user specify tls
	if parameter.TLS != p.Cluster.TLS {
		meta.Cluster.TLS = parameter.TLS
	}
	// if user specify parameter group id
	if len(parameter.ParameterGroupID) > 0 {
		meta.Cluster.ParameterGroupID = parameter.ParameterGroupID
	}
	// if user specify copies
	if parameter.Copies > 0 {
		meta.Cluster.Copies = parameter.Copies
	}
	// if user specify exclusive
	if parameter.Exclusive != p.Cluster.Exclusive {
		meta.Cluster.Exclusive = parameter.Exclusive
	}
	// if user specify cpu arch
	if len(parameter.CpuArchitecture) > 0 {
		meta.Cluster.CpuArchitecture = constants.ArchType(parameter.CpuArchitecture)
	}

	// write cluster into db
	_, err := models.GetClusterReaderWriter().Create(ctx, meta.Cluster)
	if err != nil {
		return nil, err
	}

	// clone instances
	meta.Instances = make(map[string][]*management.ClusterInstance)
	for componentType, components := range p.Instances {
		for _, instance := range components {
			newInstance := &management.ClusterInstance{
				Entity: dbCommon.Entity{
					TenantId: p.Cluster.TenantId,
					Status:   string(constants.ClusterInstanceInitializing),
				},
				Type:         instance.Type,
				ClusterID:    meta.Cluster.ID,
				Zone:         instance.Zone,
				Version:      meta.Cluster.Version,
				CpuCores:     instance.CpuCores,
				Memory:       instance.Memory,
				DiskType:     instance.DiskType,
				DiskCapacity: instance.DiskCapacity,
			}
			meta.Instances[componentType] = append(meta.Instances[componentType], newInstance)
		}
	}

	return meta, nil
}

// StartMaintenance
// @Description: try to start a maintenance
// @Receiver p
// @Parameter ctx
// @Parameter maintenanceStatus
// @return error
func (p *ClusterMeta) StartMaintenance(ctx context.Context, maintenanceStatus constants.ClusterMaintenanceStatus) error {
	// deleting will end all maintenance
	err := models.GetClusterReaderWriter().SetMaintenanceStatus(ctx, p.Cluster.ID, maintenanceStatus)

	if err == nil {
		p.Cluster.MaintenanceStatus = maintenanceStatus
	}
	return err
}

// EndMaintenance
// @Description: clear maintenance status after maintenance finished or failed
// @Receiver p
// @Parameter ctx
// @Parameter maintenanceStatus
// @return error
func (p *ClusterMeta) EndMaintenance(ctx context.Context, originStatus constants.ClusterMaintenanceStatus) error {
	err := models.GetClusterReaderWriter().ClearMaintenanceStatus(ctx, p.Cluster.ID, originStatus)

	if err == nil {
		p.Cluster.MaintenanceStatus = constants.ClusterMaintenanceNone
	}
	return err
}

type ComponentAddress struct {
	IP   string
	Port int
}

// GetClusterConnectAddresses
// @Description: Access the TiDB cluster
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetClusterConnectAddresses() []ComponentAddress {
	// got all tidb instances, then get connect addresses
	instances := p.Instances[string(newConstants.ComponentIDTiDB)]
	address := make([]ComponentAddress, 0)

	for _, instance := range instances {
		if instance.Status == string(constants.ClusterInstanceRunning) || instance.Status == string(constants.ClusterInstanceInitializing) {
			address = append(address, ComponentAddress{
				IP:   instance.HostIP[0],
				Port: int(instance.Ports[0]),
			})
		}
	}
	return address
}

// GetClusterStatusAddress
// @Description: TiDB Server status information reporting.
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetClusterStatusAddress() []ComponentAddress {
	instances := p.Instances[string(newConstants.ComponentIDTiDB)]
	address := make([]ComponentAddress, 0)

	for _, instance := range instances {
		if instance.Status == string(constants.ClusterInstanceRunning) {
			address = append(address, ComponentAddress{
				IP:   instance.HostIP[0],
				Port: int(instance.Ports[1]),
			})
		}
	}
	return address
}

// GetTiKVStatusAddress
// @Description: TiKV Server status information reporting.
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetTiKVStatusAddress() []ComponentAddress {
	instances := p.Instances[string(newConstants.ComponentIDTiKV)]
	address := make([]ComponentAddress, 0)

	for _, instance := range instances {
		if instance.Status == string(constants.ClusterInstanceRunning) {
			address = append(address, ComponentAddress{
				IP:   instance.HostIP[0],
				Port: int(instance.Ports[1]),
			})
		}
	}
	return address
}

// GetPDClientAddresses
// @Description: communication address for PD Servers to connect.
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetPDClientAddresses() []ComponentAddress {
	instances := p.Instances[string(newConstants.ComponentIDPD)]
	address := make([]ComponentAddress, 0)

	for _, instance := range instances {
		if instance.Status == string(constants.ClusterInstanceRunning) {
			address = append(address, ComponentAddress{
				IP:   instance.HostIP[0],
				Port: int(instance.Ports[0]),
			})
		}
	}
	return address
}

// GetMonitorAddresses
// @Description: Prometheus Service communication port
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetMonitorAddresses() []ComponentAddress {
	instances := p.Instances[string(newConstants.ComponentIDPrometheus)]
	address := make([]ComponentAddress, 0)

	for _, instance := range instances {
		if instance.Status == string(constants.ClusterInstanceRunning) {
			address = append(address, ComponentAddress{
				IP:   instance.HostIP[0],
				Port: int(instance.Ports[0]),
			})
		}
	}
	return address
}

// GetGrafanaAddresses
// @Description: Grafana Service communication port
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetGrafanaAddresses() []ComponentAddress {
	instances := p.Instances[string(newConstants.ComponentIDGrafana)]
	address := make([]ComponentAddress, 0)

	for _, instance := range instances {
		if instance.Status == string(constants.ClusterInstanceRunning) {
			address = append(address, ComponentAddress{
				IP:   instance.HostIP[0],
				Port: int(instance.Ports[0]),
			})
		}
	}
	return address
}

// GetAlertManagerAddresses
// @Description: AlertManager Service communication port
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetAlertManagerAddresses() []ComponentAddress {
	instances := p.Instances[string(newConstants.ComponentIDAlertManager)]
	address := make([]ComponentAddress, 0)

	for _, instance := range instances {
		if instance.Status == string(constants.ClusterInstanceRunning) {
			address = append(address, ComponentAddress{
				IP:   instance.HostIP[0],
				Port: int(instance.Ports[0]),
			})
		}
	}
	return address
}

type TiDBUserInfo struct {
	ClusterID string
	UserName  string
	Password  string
}

// GetClusterUserNamePasswd
// @Description: get tidb cluster username and password
// @Receiver p
// @return []TiDBUserInfo
func (p *ClusterMeta) GetClusterUserNamePasswd() *TiDBUserInfo {
	return &TiDBUserInfo{
		ClusterID: p.Cluster.ID,
		UserName:  p.Cluster.DBUser,
		Password:  p.Cluster.DBPassword,
	}
}

// UpdateMeta
// @Description: update cluster meta, include cluster and all instances
// @Receiver p
// @Parameter ctx
// @return error
func (p *ClusterMeta) UpdateMeta(ctx context.Context) error {
	instances := make([]*management.ClusterInstance, 0)
	if p.Instances != nil {
		for _, v := range p.Instances {
			instances = append(instances, v...)
		}
	}
	return models.GetClusterReaderWriter().UpdateMeta(ctx, p.Cluster, instances)
}

// Delete
// @Description: delete cluster
// @Receiver p
// @Parameter ctx
// @return error
func (p *ClusterMeta) Delete(ctx context.Context) error {
	return models.GetClusterReaderWriter().Delete(ctx, p.Cluster.ID)
}

func Get(ctx context.Context, clusterID string) (*ClusterMeta, error) {
	cluster, instances, err := models.GetClusterReaderWriter().GetMeta(ctx, clusterID)

	if err != nil {
		return nil, err
	}
	return buildMeta(cluster, instances), err
}

func buildMeta(cluster *management.Cluster, instances []*management.ClusterInstance) *ClusterMeta {
	instancesMap := make(map[string][]*management.ClusterInstance)

	if instances != nil && len(instances) > 0 {
		for _, instance := range instances {
			if existed, ok := instancesMap[instance.Type]; ok {
				instancesMap[instance.Type] = append(existed, instance)
			} else {
				instancesMap[instance.Type] = append(make([]*management.ClusterInstance, 0), instance)
			}
		}
	}
	return &ClusterMeta{
		Cluster:   cluster,
		Instances: instancesMap,
	}
}

// Query
// @Description: query cluster
// @Parameter ctx
// @Parameter req
// @return resp
// @return total
// @return err
func Query(ctx context.Context, req cluster.QueryClustersReq) (resp cluster.QueryClusterResp, total int, err error) {
	filters := management.Filters {
		TenantId: framework.GetTenantIDFromContext(ctx),
		NameLike: req.Name,
		Type: req.Type,
		Tag: req.Tag,
	}
	if req.ClusterID != "" {
		filters.ClusterIDs = []string{req.ClusterID}
	}
	if req.Status != "" {
		filters.StatusFilters = []newConstants.ClusterRunningStatus{newConstants.ClusterRunningStatus(req.Status)}
	}

	result, page, err := models.GetClusterReaderWriter().QueryMetas(ctx, filters, req.PageRequest)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query clusters failed, request = %v, errors = %s", req, err)
		return
	} else {
		total = page.Total
	}

	// build cluster info
	resp.Clusters = make([]structs.ClusterInfo, 0)
	for _, v := range result {
		meta := buildMeta(v.Cluster, v.Instances)
		resp.Clusters = append(resp.Clusters, meta.DisplayClusterInfo(ctx))
	}

	return
}

func (p *ClusterMeta) DisplayClusterInfo(ctx context.Context) structs.ClusterInfo {
	cluster := p.Cluster
	clusterInfo := &structs.ClusterInfo{
		ID:              cluster.ID,
		UserID:          cluster.OwnerId,
		Name:            cluster.Name,
		Type:            cluster.Type,
		Version:         cluster.Version,
		DBUser:          cluster.DBUser,
		Tags:            cluster.Tags,
		TLS:             cluster.TLS,
		Status:          cluster.Status,
		Copies:          cluster.Copies,
		Exclusive:       cluster.Exclusive,
		CpuArchitecture: string(cluster.CpuArchitecture),
		MaintainStatus:  string(cluster.MaintenanceStatus),
		Whitelist:       []string{},
		MaintainWindow:  cluster.MaintainWindow,
		CreateTime:      cluster.CreatedAt,
		UpdateTime:      cluster.UpdatedAt,
	}

	// component address
	address := p.GetClusterConnectAddresses()
	for _, a := range address {
		clusterInfo.IntranetConnectAddresses = append(clusterInfo.IntranetConnectAddresses, fmt.Sprintf("%s:%d", a.IP, a.Port))
	}
	clusterInfo.ExtranetConnectAddresses = clusterInfo.IntranetConnectAddresses

	clusterInfo.AlertUrl = fmt.Sprintf(fmt.Sprintf("%s:%d", p.GetAlertManagerAddresses()[0].IP, p.GetAlertManagerAddresses()[0].Port))
	clusterInfo.GrafanaUrl = fmt.Sprintf(fmt.Sprintf("%s:%d", p.GetGrafanaAddresses()[0].IP, p.GetGrafanaAddresses()[0].Port))

	mockUsage := func() structs.Usage {
		return structs.Usage{
			Total:     100,
			Used:      50,
			UsageRate: 0.5,
		}
	}
	clusterInfo.CpuUsage = mockUsage()
	clusterInfo.MemoryUsage = mockUsage()
	clusterInfo.StorageUsage = mockUsage()
	clusterInfo.BackupSpaceUsage = mockUsage()

	return *clusterInfo
}

func (p *ClusterMeta) DisplayInstanceInfo(ctx context.Context) (structs.ClusterTopologyInfo, structs.ClusterResourceInfo) {
	topologyInfo := new(structs.ClusterTopologyInfo)
	resourceInfo := new(structs.ClusterResourceInfo)

	if len(p.Instances) == 0 {
		return *topologyInfo, *resourceInfo
	}

	for k, v := range p.Instances {
		if Contain(newConstants.ParasiteComponentIDs, newConstants.EMProductComponentIDType(k)) {
			continue
		}
		instanceResource := structs.ClusterResourceParameterCompute{
			Type:  k,
			Count: 0,
		}
		for _, instance := range v {
			// append topology
			instanceInfo := structs.ClusterInstanceInfo{
				ID:        instance.ID,
				Type:      instance.Type,
				Role:      instance.Role,
				Version:   instance.Version,
				Status:    instance.Status,
				HostID:    instance.HostID,
				Addresses: instance.HostIP,
				Ports:     instance.Ports,
				Spec: structs.ProductSpecInfo{
					ID:   knowledge.GenSpecCode(int32(instance.CpuCores), int32(instance.Memory)),
					Name: knowledge.GenSpecCode(int32(instance.CpuCores), int32(instance.Memory)),
				},
				Zone: structs.ZoneInfo{
					ID:   instance.Zone,
					Name: instance.Zone,
				},
			}
			topologyInfo.Topology = append(topologyInfo.Topology, instanceInfo)

			spec := knowledge.GenSpecCode(int32(instance.CpuCores), int32(instance.Memory))

			newResourceSpec := true
			for i, resource := range instanceResource.Resource {
				if resource.Equal(instance.Zone, spec, instance.DiskType, int(instance.DiskCapacity)) {
					newResourceSpec = false
					instanceResource.Resource[i].Count = resource.Count + 1
				}
			}
			if newResourceSpec {
				instanceResource.Resource = append(instanceResource.Resource, structs.ClusterResourceParameterComputeResource{
					Zone:         instance.Zone,
					Spec:         spec,
					DiskType:     instance.DiskType,
					DiskCapacity: int(instance.DiskCapacity),
					Count:        1,
				})
			}
			instanceResource.Count = instanceResource.Count + 1
		}
		resourceInfo.InstanceResource = append(resourceInfo.InstanceResource, instanceResource)
	}
	return *topologyInfo, *resourceInfo
}
