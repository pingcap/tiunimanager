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

	"github.com/pingcap-inc/tiem/apimodels/cluster/upgrade"

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

func (p *Manager) QueryUpdatePath(ctx context.Context, clusterID string) ([]*upgrade.Path, error) {
	cluster, err := domain.GetClusterDetail(ctx, clusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to query update path, %s", err.Error())
		return []*upgrade.Path{}, framework.WrapError(common.TIEM_UPGRADE_QUERY_PATH_FAILED, "failed to query update path", err)
	}

	version := cluster.Cluster.ClusterVersion
	productUpgradePaths, err := models.GetUpgradeReaderWriter().QueryBySrcVersion(ctx, version.Name)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to query update path, %s", err.Error())
		return []*upgrade.Path{}, framework.WrapError(common.TIEM_UPGRADE_QUERY_PATH_FAILED, "failed to query update path", err)
	}

	pathMap := make(map[string][]string)
	for _, productUpgradePath := range productUpgradePaths {
		if versions, ok := pathMap[productUpgradePath.Type]; ok {
			versions = append(versions, productUpgradePath.DstVersion)
		} else {
			versions = []string{productUpgradePath.DstVersion}
		}
	}

	var paths []*upgrade.Path
	for k, v := range pathMap {
		path := upgrade.Path{
			Type:     k,
			Versions: v,
		}
		paths = append(paths, &path)
	}

	return paths, nil
}
