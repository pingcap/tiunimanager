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

package changefeed

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/util/api/cdc"
)

type Service interface {
	//
	// CreateBetweenClusters
	// @Description: createExecutor a change feed task for replicating the incremental data of source cluster to target cluster
	// @param ctx
	// @param sourceClusterID
	// @param targetClusterID
	// @param relationType
	// @return ID
	// @return err
	//
	CreateBetweenClusters(ctx context.Context, sourceClusterID string, targetClusterID string, relationType constants.ClusterRelationType) (ID string, err error)

	//
	// ReverseBetweenClusters reverse change feed task
	// @Description: it will delete all change feed tasks from source to target, then create a new one from the target
	// @param ctx
	// @param sourceClusterID
	// @param targetClusterID
	// @param relationType
	// @return ID
	// @return err
	//
	ReverseBetweenClusters(ctx context.Context, sourceClusterID string, targetClusterID string, relationType constants.ClusterRelationType) (ID string, err error)
	//
	// Detail
	// @Description: query change feed task detail info
	// @param ctx
	// @param request
	// @return resp
	// @return err
	//
	Detail(ctx context.Context, request cluster.DetailChangeFeedTaskReq) (resp cluster.DetailChangeFeedTaskResp, err error)
}

func GetChangeFeedService() Service {
	serviceOnce.Do(func() {
		if service == nil {
			service = &Manager{}
		}
	})
	return service
}

func MockChangeFeedService(s Service) {
	service = s
}

func (p *Manager) CreateBetweenClusters(ctx context.Context, sourceClusterID string, targetClusterID string, relationType constants.ClusterRelationType) (ID string, err error) {
	sourceCluster, err := meta.Get(ctx, sourceClusterID)
	if err != nil {
		return
	}
	targetCluster, err := meta.Get(ctx, targetClusterID)
	if err != nil {
		return
	}
	address := targetCluster.GetClusterConnectAddresses()[0]

	user, err := targetCluster.GetDBUserNamePassword(ctx, constants.DBUserCDCDataSync)
	if err != nil {
		return
	}

	task := &changefeed.ChangeFeedTask{
		Entity: dbCommon.Entity{
			TenantId: framework.GetTenantIDFromContext(ctx),
			Status:   string(constants.ChangeFeedStatusInitial),
		},
		Name:        fmt.Sprintf("from-%s-to-%s", sourceCluster.Cluster.Name, targetCluster.Cluster.Name),
		ClusterId:   sourceClusterID,
		StartTS:     0,
		FilterRules: []string{},
		Type:        constants.DownstreamTypeTiDB,
		Downstream: &changefeed.TiDBDownstream{
			Ip:              address.IP,
			Port:            address.Port,
			Username:        user.Name,
			Password:        string(user.Password),
			TargetClusterId: targetClusterID,
		},
	}

	task, err = models.GetChangeFeedReaderWriter().Create(ctx, task)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("init change feed task failed, %s", err.Error())
		return
	} else {
		ID = task.ID
	}

	err = p.createExecutor(ctx, sourceCluster, task)
	if err != nil {
		models.GetChangeFeedReaderWriter().Delete(ctx, ID)
	}

	return
}

func (p *Manager) ReverseBetweenClusters(ctx context.Context, sourceClusterID string, targetClusterID string, relationType constants.ClusterRelationType) (ID string, err error) {
	sourceCluster, err := meta.Get(ctx, sourceClusterID)
	if err != nil {
		return
	}

	tasks, _, queryErr := models.GetChangeFeedReaderWriter().Query(ctx, sourceClusterID, []constants.DownstreamType{constants.DownstreamTypeTiDB}, constants.UnfinishedChangeFeedStatus(), 0, 1000)
	if queryErr != nil {
		err = queryErr
		return
	}
	for _, task := range tasks {
		// check target, type, status
		if task.Type == constants.DownstreamTypeTiDB &&
			!constants.ChangeFeedStatus(task.Status).IsFinal() &&
			task.Downstream.(*changefeed.TiDBDownstream).TargetClusterId == targetClusterID {
			result, deleteError := cdc.CDCService.DeleteChangeFeedTask(ctx, cdc.ChangeFeedDeleteReq{
				CDCAddress:   sourceCluster.GetCDCClientAddresses()[0].ToString(),
				ChangeFeedID: task.ID,
			})

			if deleteError != nil || !result.Accepted {
				framework.LogWithContext(ctx).Errorf("failed to delete change feed task, err = %v, result = %v", err, result)
			}

			err = models.GetChangeFeedReaderWriter().Delete(ctx, task.ID)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("failed to delete change feed task, %s", err.Error())
				return
			}
		}
	}

	return p.CreateBetweenClusters(ctx, targetClusterID, sourceClusterID, relationType)
}

func (p *Manager) Detail(ctx context.Context, request cluster.DetailChangeFeedTaskReq) (resp cluster.DetailChangeFeedTaskResp, err error) {
	task, err := models.GetChangeFeedReaderWriter().Get(ctx, request.ID)
	if err != nil {
		return
	}

	if task.Status == constants.ChangeFeedStatusInitial.ToString() {
		err = errors.NewError(errors.TIEM_CHANGE_FEED_NOT_FOUND, "change feed task has not been created")
	}

	clusterMeta, err := meta.Get(ctx, task.ClusterId)
	if err != nil {
		return
	}
	cdcAddress := clusterMeta.GetCDCClientAddresses()
	if len(cdcAddress) == 0 {
		err = errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "CDC components required, cluster %s", clusterMeta.Cluster.ID)
		return
	}
	resp.ChangeFeedTaskInfo = parse(*task)

	taskDetail, detailError := cdc.CDCService.DetailChangeFeedTask(ctx, cdc.ChangeFeedDetailReq{
		CDCAddress:   clusterMeta.GetCDCClientAddresses()[0].ToString(),
		ChangeFeedID: task.ID,
	})

	if detailError == nil {
		resp.ChangeFeedTaskInfo.AcceptDownstreamSyncTS(taskDetail.CheckPointTSO)
		resp.ChangeFeedTaskInfo.AcceptUpstreamUpdateTS(currentTSO())
	} else {
		framework.LogWithContext(ctx).Errorf("detail change feed task err = %s", err)
	}

	return
}
