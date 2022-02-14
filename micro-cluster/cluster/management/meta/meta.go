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

package meta

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/message/cluster"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
)

type ClusterMeta struct {
	Cluster              *management.Cluster
	Instances            map[string][]*management.ClusterInstance
	DBUsers              map[string]*management.DBUser
	NodeExporterPort     int32
	BlackboxExporterPort int32
}

type ComponentAddress struct {
	IP   string
	Port int
}

func (p *ComponentAddress) ToString() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
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

// ClearClusterPhysically
// @Description: delete cluster physically, If you don't know why you should use it, then don't use it
// @Receiver p
// @Parameter ctx
// @return error
func (p *ClusterMeta) ClearClusterPhysically(ctx context.Context) error {
	return models.GetClusterReaderWriter().ClearClusterPhysically(ctx, p.Cluster.ID)
}

func Get(ctx context.Context, clusterID string) (*ClusterMeta, error) {
	cluster, instances, users, err := models.GetClusterReaderWriter().GetMeta(ctx, clusterID)

	if err != nil {
		return nil, err
	}

	return buildMeta(cluster, instances, users), err
}

// Query
// @Description: query cluster
// @Parameter ctx
// @Parameter req
// @return resp
// @return total
// @return err
func Query(ctx context.Context, req cluster.QueryClustersReq) (resp cluster.QueryClusterResp, total int, err error) {
	filters := management.Filters{
		TenantId: framework.GetTenantIDFromContext(ctx),
		NameLike: req.Name,
		Type:     req.Type,
		Tag:      req.Tag,
	}
	if len(req.ClusterID) > 0 {
		filters.ClusterIDs = []string{req.ClusterID}
	}
	if len(req.Status) > 0 {
		filters.StatusFilters = []constants.ClusterRunningStatus{constants.ClusterRunningStatus(req.Status)}
	}

	result, page, err := models.GetClusterReaderWriter().QueryMetas(ctx, filters, req.PageRequest)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query clusters failed, request = %v, errors = %s", req, err)
		return
	} else {
		framework.LogWithContext(ctx).Infof("query clusters got result = %v, page = %v", result, page)
		total = page.Total
	}

	// build cluster info
	resp.Clusters = make([]structs.ClusterInfo, 0)
	for _, v := range result {
		meta := buildMeta(v.Cluster, v.Instances, v.DBUsers)
		resp.Clusters = append(resp.Clusters, meta.DisplayClusterInfo(ctx))
	}

	return
}

type InstanceLogInfo struct {
	ClusterID    string
	InstanceType constants.EMProductComponentIDType
	IP           string
	DataDir      string
	DeployDir    string
	LogDir       string
}

func QueryInstanceLogInfo(ctx context.Context, hostId string, typeFilter []string, statusFilter []string) (infos []*InstanceLogInfo, err error) {
	instances, err := models.GetClusterReaderWriter().QueryInstancesByHost(ctx, hostId, typeFilter, statusFilter)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query instances by host failed, %s", err.Error())
		return
	}

	infos = make([]*InstanceLogInfo, 0)
	for _, instance := range instances {
		infos = append(infos, &InstanceLogInfo{
			ClusterID:    instance.ClusterID,
			InstanceType: constants.EMProductComponentIDType(instance.Type),
			IP:           instance.HostIP[0],
			DataDir:      instance.GetDataDir(),
			DeployDir:    instance.GetDeployDir(),
			LogDir:       instance.GetLogDir(),
		})
	}
	return
}

