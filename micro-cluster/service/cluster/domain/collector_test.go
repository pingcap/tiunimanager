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
 * @File: collector_test.go
 * @Description: collector test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/9 14:16
*******************************************************************************/

package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCollectorTiDBLogConfig(t *testing.T) {
	cluster := defaultCluster()
	clusters := make([]*ClusterAggregation, 1)
	clusters[0] = cluster
	configs, err := buildCollectorTiDBLogConfig(context.TODO(), "127.0.0.1", clusters)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(configs))
	assert.Equal(t, "127.0.0.1", configs[0].TiDB.Input.Fields.Ip)
	assert.Equal(t, "127.0.0.1", configs[0].PD.Input.Fields.Ip)

	configs, err = buildCollectorTiDBLogConfig(context.TODO(), "127.0.0.2", clusters)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(configs))
	assert.Equal(t, "127.0.0.2", configs[0].TiFlash.Input.Fields.Ip)
	assert.Equal(t, "127.0.0.2", configs[0].TiCDC.Input.Fields.Ip)
}

func TestBuildCollectorModuleDetail(t *testing.T) {
	aggregation := defaultCluster()
	host := "127.0.0.1"
	logPath := "/mnt/sda/testCluster/tikv-deploy/testCluster/tidb-log/tikv.log"
	collectorModuleDetail := buildCollectorModuleDetail(aggregation, host, logPath)
	assert.Equal(t, aggregation.Cluster.Id, collectorModuleDetail.Input.Fields.ClusterId)
	assert.Equal(t, host, collectorModuleDetail.Input.Fields.Ip)
	assert.Equal(t, logPath, collectorModuleDetail.Var.Paths[0])
}

func TestListClusterHosts(t *testing.T) {
	aggregation := defaultCluster()
	hosts := listClusterHosts(aggregation)
	assert.Equal(t, 2, len(hosts))
}
