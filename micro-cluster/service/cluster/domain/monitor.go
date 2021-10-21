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
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: monitor
 * @Description: Get TiDB cluster monitoring link
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/10/20 14:56
*******************************************************************************/

package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
)

// Monitor
// @Description: monitoring link object
type Monitor struct {
	ClusterId  string `json:"clusterId"`
	AlertUrl   string `json:"alertUrl"`
	GrafanaUrl string `json:"grafanaUrl"`
}

// DescribeMonitor
// @Description: describe monitoring link
// @Parameter ctx
// @Parameter ope
// @Parameter clusterId
// @return *Monitor
// @return error
func DescribeMonitor(ctx context.Context, ope *clusterpb.OperatorDTO, clusterId string) (*Monitor, error) {
	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil || clusterAggregation == nil || clusterAggregation.Cluster == nil {
		return nil, errors.New("load cluster aggregation failed")
	}
	monitor, err := getMonitorUrl(clusterAggregation)
	return monitor, err
}

func getMonitorUrl(clusterAggregation *ClusterAggregation) (*Monitor, error) {
	configModel := clusterAggregation.CurrentTopologyConfigRecord.ConfigModel

	if len(configModel.Alertmanagers) <= 0 {
		return nil, errors.New("get cluster alert manager component failed")
	}
	alertServer := configModel.Alertmanagers[0]
	alertUrl := fmt.Sprintf("http://%s:%d", alertServer.Host, alertServer.WebPort)

	if len(configModel.Grafanas) <= 0 {
		return nil, errors.New("get cluster grafana component failed")
	}
	grafanaServer := configModel.Grafanas[0]
	grafanaUrl := fmt.Sprintf("http://%s:%d", grafanaServer.Host, grafanaServer.Port)

	monitor := &Monitor{
		ClusterId:  clusterAggregation.Cluster.Id,
		AlertUrl:   alertUrl,
		GrafanaUrl: grafanaUrl,
	}
	return monitor, nil
}
