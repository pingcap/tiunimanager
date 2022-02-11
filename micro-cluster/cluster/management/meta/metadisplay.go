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

/*******************************************************************************
 * @File: metadisplay.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/11
*******************************************************************************/

package meta

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"sort"
	"strings"
)

// GetClusterConnectAddresses
// @Description: Access the TiDB cluster
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetClusterConnectAddresses() []ComponentAddress {
	// got all tidb instances, then get connect addresses
	instances := p.Instances[string(constants.ComponentIDTiDB)]
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

// GetClusterStatusAddress
// @Description: TiDB Server status information reporting.
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetClusterStatusAddress() []ComponentAddress {
	instances := p.Instances[string(constants.ComponentIDTiDB)]
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
	instances := p.Instances[string(constants.ComponentIDTiKV)]
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
	instances := p.Instances[string(constants.ComponentIDPD)]
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

// GetCDCClientAddresses
// @Description: communication address for CDC Servers to connect.
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetCDCClientAddresses() []ComponentAddress {
	instances := p.Instances[string(constants.ComponentIDCDC)]
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

// GetTiFlashClientAddresses
// @Description: communication address for TiFlash Servers to connect.
// @Receiver p
// @return []ComponentAddress
func (p *ClusterMeta) GetTiFlashClientAddresses() []ComponentAddress {
	instances := p.Instances[string(constants.ComponentIDTiFlash)]
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
	instances := p.Instances[string(constants.ComponentIDPrometheus)]
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
	instances := p.Instances[string(constants.ComponentIDGrafana)]
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
	instances := p.Instances[string(constants.ComponentIDAlertManger)]
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

func (p *ClusterMeta) GetClusterComponentProperties(ctx context.Context) ([]structs.ProductComponentProperty, error) {
	detail, err := GetProductDetail(ctx, p.Cluster.Vendor, p.Cluster.Region, p.Cluster.Type)
	if err != nil {
		return nil, err
	}

	if v, versionOK := detail.Versions[p.Cluster.Version]; !versionOK {
		return nil, errors.NewErrorf(errors.TIEM_UNSUPPORT_PRODUCT, "version %s is not supported", p.Cluster.Version)
	} else if properties, archOK := v.Arch[string(p.Cluster.CpuArchitecture)]; !archOK {
		return nil, errors.NewErrorf(errors.TIEM_UNSUPPORT_PRODUCT, "arch %s is not supported", p.Cluster.CpuArchitecture)
	} else {
		return properties, nil
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

// GetDBUserNamePassword
// @Description: get username and password of the different type user
// @Receiver p
// @return BDUser
func (p *ClusterMeta) GetDBUserNamePassword(ctx context.Context, roleType constants.DBUserRoleType) (*management.DBUser, error) {
	user := p.DBUsers[string(roleType)]
	if user == nil {
		msg := fmt.Sprintf("get %s user from cluser %s failed, empty user", roleType, p.Cluster.ID)
		framework.LogWithContext(ctx).Errorf(msg)
		return user, errors.NewError(errors.TIEM_USER_NOT_FOUND, msg)
	}
	return user, nil
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
	return nil, errors.NewError(errors.TIEM_INSTANCE_NOT_FOUND, "instance not found")
}

// IsComponentRequired
// @Description judge whether component is required
// @Parameter	component type
// @Return		bool
func (p *ClusterMeta) IsComponentRequired(ctx context.Context, componentType string) bool {
	properties, err := p.GetClusterComponentProperties(ctx)
	if err != nil {
		framework.LogWithContext(ctx).Errorf(err.Error())
		return false
	}

	for _, property := range properties {
		if property.ID == componentType && property.MinInstance > 0 {
			return true
		}
	}
	return false
}

func (p *ClusterMeta) IsTakenOver() bool {
	if len(p.Cluster.Tags) > 0 {
		for _, tag := range p.Cluster.Tags {
			if tag == TagTakeover {
				return true
			}
		}
	}
	return false
}

func (p *ClusterMeta) GetRelations(ctx context.Context) ([]*management.ClusterRelation, error) {
	return models.GetClusterReaderWriter().GetRelations(ctx, p.Cluster.ID)
}

func (p *ClusterMeta) GetDBUsers(ctx context.Context) ([]*management.DBUser, error) {
	return models.GetClusterReaderWriter().GetDBUser(ctx, p.Cluster.ID)
}

func (p *ClusterMeta) GetMajorVersion() string {
	return strings.Split(p.Cluster.Version, ".")[0]

}

func (p *ClusterMeta) GetMinorVersion() string {
	numbers := strings.Split(p.Cluster.Version, ".")
	if len(numbers) > 1 {
		return fmt.Sprintf("%s.%s", numbers[0], numbers[1])
	} else {
		return fmt.Sprintf(numbers[0])
	}
}

func (p *ClusterMeta) GetRevision() string {
	return p.Cluster.Version
}

func (p *ClusterMeta) DisplayClusterInfo(ctx context.Context) structs.ClusterInfo {
	cluster := p.Cluster
	clusterInfo := &structs.ClusterInfo{
		ID:              cluster.ID,
		UserID:          cluster.OwnerId,
		Name:            cluster.Name,
		Type:            cluster.Type,
		Version:         cluster.Version,
		Tags:            cluster.Tags,
		TLS:             cluster.TLS,
		Vendor:          cluster.Vendor,
		Region:          cluster.Region,
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
	// todo: display users?
	// component address
	address := p.GetClusterConnectAddresses()
	for _, a := range address {
		clusterInfo.IntranetConnectAddresses = append(clusterInfo.IntranetConnectAddresses, fmt.Sprintf("%s:%d", a.IP, a.Port))
	}
	clusterInfo.ExtranetConnectAddresses = clusterInfo.IntranetConnectAddresses

	if alertAddress := p.GetAlertManagerAddresses(); len(alertAddress) > 0 {
		clusterInfo.AlertUrl = fmt.Sprintf("%s:%d", alertAddress[0].IP, alertAddress[0].Port)
	}

	if grafanaAddress := p.GetGrafanaAddresses(); len(grafanaAddress) > 0 {
		clusterInfo.GrafanaUrl = fmt.Sprintf("%s:%d", grafanaAddress[0].IP, grafanaAddress[0].Port)
	}

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
		if Contain(constants.ParasiteComponentIDs, constants.EMProductComponentIDType(k)) {
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
					ID:   structs.GenSpecCode(int32(instance.CpuCores), int32(instance.Memory)),
					Name: structs.GenSpecCode(int32(instance.CpuCores), int32(instance.Memory)),
				},
				Zone: structs.ZoneInfo{
					ZoneID:   structs.GenDomainCodeByName(p.Cluster.Region, instance.Zone),
					ZoneName: instance.Zone,
				},
			}
			topologyInfo.Topology = append(topologyInfo.Topology, instanceInfo)

			spec := structs.GenSpecCode(int32(instance.CpuCores), int32(instance.Memory))

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

	instanceWrapper := InstanceWrapper{topologyInfo.Topology, func(p, q *structs.ClusterInstanceInfo) bool {
		return constants.EMProductComponentIDType(p.Type).SortWeight() > constants.EMProductComponentIDType(q.Type).SortWeight()
	}}
	sort.Sort(instanceWrapper)

	topologyInfo.Topology = instanceWrapper.infos
	return *topologyInfo, *resourceInfo
}

type InstanceWrapper struct {
	infos []structs.ClusterInstanceInfo
	by    func(p, q *structs.ClusterInstanceInfo) bool
}

func (pw InstanceWrapper) Len() int {
	return len(pw.infos)
}
func (pw InstanceWrapper) Swap(i, j int) {
	pw.infos[i], pw.infos[j] = pw.infos[j], pw.infos[i]
}
func (pw InstanceWrapper) Less(i, j int) bool {
	return pw.by(&pw.infos[i], &pw.infos[j])
}
