/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: manager
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/7
*******************************************************************************/

package upgrade

import (
	"context"

	"github.com/pingcap-inc/tiem/message/cluster"

	"github.com/pingcap-inc/tiem/common/structs"

	"github.com/pingcap-inc/tiem/library/secondparty"

	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
)

type Manager struct {
}

func NewManager() *Manager {
	return &Manager{}
}

func (p *Manager) QueryProductUpdatePath(ctx context.Context, clusterID string) ([]*structs.ProductUpgradePathItem, error) {
	cluster, err := domain.GetClusterDetail(ctx, clusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to query update path, %s", err.Error())
		return []*structs.ProductUpgradePathItem{}, framework.WrapError(common.TIEM_UPGRADE_QUERY_PATH_FAILED, "failed to query upgrade path", err)
	}

	version := cluster.Cluster.ClusterVersion
	productUpgradePaths, err := models.GetUpgradeReaderWriter().QueryBySrcVersion(ctx, version.Name)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to query update path, %s", err.Error())
		return []*structs.ProductUpgradePathItem{}, framework.WrapError(common.TIEM_UPGRADE_QUERY_PATH_FAILED, "failed to query upgrade path", err)
	}

	pathMap := make(map[string][]string)
	for _, productUpgradePath := range productUpgradePaths {
		if versions, ok := pathMap[productUpgradePath.Type]; ok {
			versions = append(versions, productUpgradePath.DstVersion)
		} else {
			versions = []string{productUpgradePath.DstVersion}
		}
	}

	var paths []*structs.ProductUpgradePathItem
	for k, v := range pathMap {
		path := structs.ProductUpgradePathItem{
			Type:     k,
			Versions: v,
		}
		paths = append(paths, &path)
	}

	return paths, nil
}

func (p *Manager) QueryUpgradeVersionDiffInfo(ctx context.Context, clusterID string, version string) ([]*structs.ProductUpgradeVersionConfigDiffItem, error) {
	cluster, err := domain.GetClusterDetail(ctx, clusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to query upgrade version diff, %s", err.Error())
		return []*structs.ProductUpgradeVersionConfigDiffItem{}, framework.WrapError(common.TIEM_UPGRADE_QUERY_VERSION_DIFF_FAILED, "failed to query upgrade version diff", err)
	}

	srcVersion := cluster.Cluster.ClusterVersion.Name
	// TODO: get params for clusterID and dst version and check the diffs
	framework.LogWithContext(ctx).Infof("TODO: get params for current cluster(%s:%s) and dst version(%s) and get get diffs", clusterID, srcVersion, version)

	var configDiffInfos []*structs.ProductUpgradeVersionConfigDiffItem

	return configDiffInfos, nil
}

func (p *Manager) ClusterUpgrade(ctx context.Context, req *cluster.ClusterUpgradeReq) (string, error) {
	cluster, err := domain.GetClusterDetail(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to query upgrade version diff, %s", err.Error())
		return "", framework.WrapError(common.TIEM_UPGRADE_FAILED, "failed to query upgrade version diff", err)
	}

	// TODO: get param for dst version, use parameters to apply it, and do upgrade cluster
	framework.LogWithContext(ctx).Infof("TODO: get param for dst version, use parameters to apply it, and do upgrade cluster")

	_, err = secondparty.Manager.ClusterUpgrade(
		ctx, secondparty.ClusterComponentTypeStr, cluster.Cluster.ClusterName, req.TargetVersion, 0, []string{}, "WorkflowID",
	)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to upgrade cluster, %s", err.Error())
		return "", framework.WrapError(common.TIEM_UPGRADE_FAILED, "failed to query upgrade version diff", err)
	}

	return "WorkflowID", nil
}
