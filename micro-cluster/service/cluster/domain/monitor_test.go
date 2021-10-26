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
 * @File: monitor_test
 * @Description: Test Get TiDB cluster monitoring link
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/10/20 17:51
*******************************************************************************/

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pingcap/tiup/pkg/cluster/spec"
)

func TestDescribeMonitor(t *testing.T) {
	setupMockAdapter()
	monitor, err := DescribeMonitor(nil, nil, "testCluster")
	assert.Nil(t, err)
	assert.NotNil(t, monitor)
}

func TestGetMonitorUrlFromCluster(t *testing.T) {

	tests := []struct {
		name    string
		args    *ClusterAggregation
		wantErr bool
		asserts []func(args *ClusterAggregation, m *Monitor) bool
	}{
		{
			"normal",
			&ClusterAggregation{
				Cluster: &Cluster{
					Id: "abc",
				},
				CurrentTopologyConfigRecord: &TopologyConfigRecord{
					ConfigModel: &spec.Specification{
						Alertmanagers: []*spec.AlertmanagerSpec{
							{
								Host:    "127.0.0.1",
								WebPort: 9093,
							},
						},
						Grafanas: []*spec.GrafanaSpec{
							{
								Host: "127.0.0.1",
								Port: 3000,
							},
						},
					},
				},
			},
			false,
			[]func(args *ClusterAggregation, m *Monitor) bool{
				func(args *ClusterAggregation, m *Monitor) bool {
					return m.ClusterId == "abc"
				},
				func(args *ClusterAggregation, m *Monitor) bool {
					return m.AlertUrl == "http://127.0.0.1:9093"
				},
				func(args *ClusterAggregation, m *Monitor) bool {
					return m.GrafanaUrl == "http://127.0.0.1:3000"
				},
			},
		},
		{
			"normal by no default port",
			&ClusterAggregation{
				Cluster: &Cluster{
					Id: "abc",
				},
				CurrentTopologyConfigRecord: &TopologyConfigRecord{
					ConfigModel: &spec.Specification{
						Alertmanagers: []*spec.AlertmanagerSpec{
							{
								Host: "127.0.0.1",
							},
						},
						Grafanas: []*spec.GrafanaSpec{
							{
								Host: "127.0.0.1",
							},
						},
					},
				},
			},
			false,
			[]func(args *ClusterAggregation, m *Monitor) bool{
				func(args *ClusterAggregation, m *Monitor) bool {
					return m.ClusterId == "abc"
				},
				func(args *ClusterAggregation, m *Monitor) bool {
					return m.AlertUrl == "http://127.0.0.1:9093"
				},
				func(args *ClusterAggregation, m *Monitor) bool {
					return m.GrafanaUrl == "http://127.0.0.1:3000"
				},
			},
		},
		{
			"define alert manager not existed",
			&ClusterAggregation{
				Cluster: &Cluster{
					Id: "abc",
				},
				CurrentTopologyConfigRecord: &TopologyConfigRecord{
					ConfigModel: &spec.Specification{
						Grafanas: []*spec.GrafanaSpec{
							{
								Host: "127.0.0.1",
								Port: 3000,
							},
						},
					},
				},
			},
			true,
			[]func(args *ClusterAggregation, m *Monitor) bool{},
		},
		{
			"define grafana not existed",
			&ClusterAggregation{
				Cluster: &Cluster{
					Id: "abc",
				},
				CurrentTopologyConfigRecord: &TopologyConfigRecord{
					ConfigModel: &spec.Specification{
						Alertmanagers: []*spec.AlertmanagerSpec{
							{
								Host:    "127.0.0.1",
								WebPort: 9093,
							},
						},
					},
				},
			},
			true,
			[]func(args *ClusterAggregation, m *Monitor) bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMonitorUrl(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMonitorUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(tt.args, got) {
					t.Errorf("getMonitorUrl() assert got false, index = %v, got = %v", i, got)
				}
			}
		})
	}
}
