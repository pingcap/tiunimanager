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
 * @File: metabuilder.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/11
*******************************************************************************/

package meta

import (
	"context"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

// CloneMeta
// @Description: clone meta info from cluster based on create cluster parameter
// @Receiver p
// @Parameter ctx
// @return *ClusterMeta
func (p *ClusterMeta) CloneMeta(ctx context.Context, parameter structs.CreateClusterParameter, computes []structs.ClusterResourceParameterCompute) (*ClusterMeta, error) {
	meta := &ClusterMeta{}
	// clone cluster info
	meta.Cluster = &management.Cluster{
		Entity: dbCommon.Entity{
			TenantId: framework.GetTenantIDFromContext(ctx),
			Status:   string(constants.ClusterInitializing),
		},
		Name:              parameter.Name,    // user specify (required)
		Region:            parameter.Region,  // user specify (required)
		Type:              p.Cluster.Type,    // user not specify
		Version:           p.Cluster.Version, // user specify (option)
		Tags:              p.Cluster.Tags,    // user specify (option)
		TLS:               p.Cluster.TLS,     // user specify (option)
		OwnerId:           framework.GetUserIDFromContext(ctx),
		Vendor:            parameter.Vendor,
		ParameterGroupID:  p.Cluster.ParameterGroupID, // user specify (option)
		Copies:            p.Cluster.Copies,           // user specify (option)
		Exclusive:         p.Cluster.Exclusive,        // user specify (option)
		CpuArchitecture:   p.Cluster.CpuArchitecture,  // user specify (option)
		MaintenanceStatus: constants.ClusterMaintenanceNone,
		MaintainWindow:    p.Cluster.MaintainWindow,
	}

	// if user specify cluster version
	if len(parameter.Version) > 0 {
		cmp, err := CompareTiDBVersion(parameter.Version, p.Cluster.Version)
		if err != nil {
			return nil, err
		}
		if !cmp {
			return nil, errors.NewError(errors.TIEM_CHECK_CLUSTER_VERSION_ERROR,
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
	got, err := models.GetClusterReaderWriter().Create(ctx, meta.Cluster)
	if err != nil {
		return nil, err
	}
	// set root user
	meta.DBUsers = make(map[string]*management.DBUser)
	meta.DBUsers[string(constants.Root)] = &management.DBUser{
		ClusterID: got.ID,
		Name:      constants.DBUserName[constants.Root],
		Password:  dbCommon.Password(parameter.DBPassword),
		RoleType:  string(constants.Root),
	}
	// add instances
	err = meta.AddInstances(ctx, computes)
	if err != nil {
		return nil, err
	}

	// add default instances
	if err = meta.AddDefaultInstances(ctx); err != nil {
		return nil, err
	}

	return meta, nil
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
		Type:              param.Type,
		Version:           param.Version,
		TLS:               param.TLS,
		Tags:              param.Tags,
		OwnerId:           framework.GetUserIDFromContext(ctx),
		ParameterGroupID:  param.ParameterGroupID,
		Copies:            param.Copies,
		Exclusive:         param.Exclusive,
		Region:            param.Region,
		Vendor:            param.Vendor,
		CpuArchitecture:   constants.ArchType(param.CpuArchitecture),
		MaintenanceStatus: constants.ClusterMaintenanceNone,
		MaintainWindow:    "",
	}
	got, err := models.GetClusterReaderWriter().Create(ctx, p.Cluster)
	if err == nil {
		framework.LogWithContext(ctx).Infof("create cluster %s succeed, id = %s", p.Cluster.Name, got.ID)
		// set root user in clusterMeta
		p.DBUsers = make(map[string]*management.DBUser)
		p.DBUsers[string(constants.Root)] = &management.DBUser{
			ClusterID: got.ID,
			Name:      constants.DBUserName[constants.Root],
			Password:  dbCommon.Password(param.DBPassword),
			RoleType:  string(constants.Root),
		}
		//fmt.Println("got: ",got.ID, "user: ", p.DBUsers[string(constants.Root)].ClusterID)
	} else {
		framework.LogWithContext(ctx).Errorf("create cluster %s failed, err : %s", p.Cluster.Name, err.Error())
	}
	return err
}

var TagTakeover = "takeover"

func (p *ClusterMeta) BuildForTakeover(ctx context.Context, name string, dbPassword string) error {
	p.Cluster = &management.Cluster{
		Entity: dbCommon.Entity{
			ID:       name,
			TenantId: framework.GetTenantIDFromContext(ctx),
			Status:   string(constants.ClusterRunning),
		},
		// todo get from takeover request
		Vendor:         "Local",
		Name:           name,
		Tags:           []string{TagTakeover},
		OwnerId:        framework.GetUserIDFromContext(ctx),
		MaintainWindow: "",
	}
	got, err := models.GetClusterReaderWriter().Create(ctx, p.Cluster)
	if err == nil {
		// set root user in clusterMeta
		p.DBUsers = make(map[string]*management.DBUser)
		rootUser := &management.DBUser{
			ClusterID: got.ID,
			Name:      constants.DBUserName[constants.Root],
			Password:  dbCommon.Password(dbPassword),
			RoleType:  string(constants.Root),
		}
		p.DBUsers[string(constants.Root)] = rootUser

		err = models.GetClusterReaderWriter().CreateDBUser(ctx, p.DBUsers[string(constants.Root)])
	}

	if err != nil {
		framework.LogWithContext(ctx).Errorf("takeover cluster %s failed, err : %s", p.Cluster.Name, err.Error())
	}
	return err
}

// ParseTopologyFromConfig
// @Description: parse topology from yaml config
// @Receiver p
// @Parameter ctx
// @Parameter specs
// @return []*management.ClusterInstance
// @return error
func (p *ClusterMeta) ParseTopologyFromConfig(ctx context.Context, specs *spec.Specification) error {
	if specs == nil {
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, "cannot parse empty specification")
	}
	instances := make([]*management.ClusterInstance, 0)
	if len(specs.PDServers) > 0 {
		for _, server := range specs.PDServers {
			instances = append(instances, parseInstanceFromSpec(p.Cluster, constants.ComponentIDPD, func() string {
				return server.Host
			}, func() []int32 {
				return []int32{
					int32(server.ClientPort),
					int32(server.PeerPort),
				}
			}).SetPresetDir(server.DeployDir, server.DataDir, server.LogDir))
		}
	}
	if len(specs.TiDBServers) > 0 {
		for _, server := range specs.TiDBServers {
			instances = append(instances, parseInstanceFromSpec(p.Cluster, constants.ComponentIDTiDB, func() string {
				return server.Host
			}, func() []int32 {
				return []int32{
					int32(server.Port),
					int32(server.StatusPort),
				}
			}).SetPresetDir(server.DeployDir, "", server.LogDir))
		}
	}
	if len(specs.TiKVServers) > 0 {
		for _, server := range specs.TiKVServers {
			instances = append(instances, parseInstanceFromSpec(p.Cluster, constants.ComponentIDTiKV, func() string {
				return server.Host
			}, func() []int32 {
				return []int32{
					int32(server.Port),
					int32(server.StatusPort),
				}
			}).SetPresetDir(server.DeployDir, server.DataDir, server.LogDir))
		}
	}
	if len(specs.TiFlashServers) > 0 {
		for _, server := range specs.TiFlashServers {
			instances = append(instances, parseInstanceFromSpec(p.Cluster, constants.ComponentIDTiFlash, func() string {
				return server.Host
			}, func() []int32 {
				return []int32{
					int32(server.TCPPort),
					int32(server.HTTPPort),
					int32(server.FlashServicePort),
					int32(server.FlashProxyPort),
					int32(server.FlashProxyStatusPort),
					int32(server.StatusPort),
				}
			}).SetPresetDir(server.DeployDir, server.DataDir, server.LogDir))
		}
	}
	if len(specs.CDCServers) > 0 {
		for _, server := range specs.CDCServers {
			instances = append(instances, parseInstanceFromSpec(p.Cluster, constants.ComponentIDCDC, func() string {
				return server.Host
			}, func() []int32 {
				return []int32{
					int32(server.Port),
				}
			}).SetPresetDir(server.DeployDir, server.DataDir, server.LogDir))
		}
	}
	if len(specs.Grafanas) > 0 {
		for _, server := range specs.Grafanas {
			instances = append(instances, parseInstanceFromSpec(p.Cluster, constants.ComponentIDGrafana, func() string {
				return server.Host
			}, func() []int32 {
				return []int32{
					int32(server.Port),
				}
			}).SetPresetDir(server.DeployDir, "", ""))
		}
	}
	if len(specs.Alertmanagers) > 0 {
		for _, server := range specs.Alertmanagers {
			instances = append(instances, parseInstanceFromSpec(p.Cluster, constants.ComponentIDAlertManger, func() string {
				return server.Host
			}, func() []int32 {
				return []int32{
					int32(server.WebPort),
					int32(server.ClusterPort),
				}
			}).SetPresetDir(server.DeployDir, server.DataDir, server.LogDir))
		}
	}
	if len(specs.Monitors) > 0 {
		for _, server := range specs.Monitors {
			instances = append(instances, parseInstanceFromSpec(p.Cluster, constants.ComponentIDPrometheus, func() string {
				return server.Host
			}, func() []int32 {
				return []int32{
					int32(server.Port),
				}
			}).SetPresetDir(server.DeployDir, server.DataDir, server.LogDir))
		}
	}

	p.acceptInstances(instances)

	err := models.GetClusterReaderWriter().UpdateInstance(ctx, instances...)
	if err != nil {
		return err
	}
	return nil
}

func parseInstanceFromSpec(cluster *management.Cluster, componentType constants.EMProductComponentIDType, getHost func() string, getPort func() []int32) *management.ClusterInstance {
	return &management.ClusterInstance{
		Entity: dbCommon.Entity{
			TenantId: cluster.TenantId,
			Status:   string(constants.ClusterInstanceRunning),
		},
		Type:      string(componentType),
		Version:   cluster.Version,
		ClusterID: cluster.ID,
		HostIP:    []string{getHost()},
		Ports:     getPort(),
	}
}

func buildMeta(cluster *management.Cluster, instances []*management.ClusterInstance, users []*management.DBUser) *ClusterMeta {
	meta := &ClusterMeta{
		Cluster: cluster,
	}
	meta.acceptInstances(instances)
	meta.acceptDBUsers(users)
	return meta
}

func (p *ClusterMeta) acceptInstances(instances []*management.ClusterInstance) {
	instancesMap := make(map[string][]*management.ClusterInstance)

	if len(instances) > 0 {
		for _, instance := range instances {
			if existed, ok := instancesMap[instance.Type]; ok {
				instancesMap[instance.Type] = append(existed, instance)
			} else {
				instancesMap[instance.Type] = append(make([]*management.ClusterInstance, 0), instance)
			}
		}
	}
	p.Instances = instancesMap
}

func (p *ClusterMeta) acceptDBUsers(users []*management.DBUser) {
	usersMap := make(map[string]*management.DBUser)

	if len(users) > 0 {
		for _, user := range users {
			usersMap[user.RoleType] = user
		}
	}
	p.DBUsers = usersMap
}
