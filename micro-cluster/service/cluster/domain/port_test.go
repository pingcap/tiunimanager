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

package domain

import (
	"context"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

func setupMockAdapter() {
	TaskRepo = MockTaskRepo{}
	ClusterRepo = MockClusterRepo{}
	RemoteClusterProxy = MockInstanceRepo{}
}

type MockTaskRepo struct{}

func (m MockTaskRepo) ListFlows(ctx context.Context, bizId, keyword string, status int, page int, pageSize int) ([]*FlowWorkEntity, int, error) {
	return []*FlowWorkEntity{}, 0, nil
}

var id uint = 0

func getId() uint {
	id = id + 1
	return id
}
func (m MockTaskRepo) AddFlowWork(ctx context.Context, flowWork *FlowWorkEntity) error {
	flowWork.Id = getId()
	return nil
}

func (m MockTaskRepo) AddFlowTask(ctx context.Context, task *TaskEntity, flowId uint) error {
	task.Id = getId()
	return nil
}

func (m MockTaskRepo) AddCronTask(ctx context.Context, cronTask *CronTaskEntity) error {
	panic("implement me")
}

func (m MockTaskRepo) Persist(ctx context.Context, flowWork *FlowWorkAggregation) error {
	return nil
}

func (m MockTaskRepo) LoadFlowWork(ctx context.Context, id uint) (*FlowWorkEntity, error) {
	panic("implement me")
}

func (m MockTaskRepo) Load(ctx context.Context, id uint) (flowWork *FlowWorkAggregation, err error) {
	panic("implement me")
}

func (m MockTaskRepo) QueryCronTask(ctx context.Context, bizId string, cronTaskType int) (cronTask *CronTaskEntity, err error) {
	panic("implement me")
}

func (m MockTaskRepo) PersistCronTask(ctx context.Context, cronTask *CronTaskEntity) (err error) {
	panic("implement me")
}

type MockClusterRepo struct{}

func (m MockClusterRepo) AddCluster(ctx context.Context, cluster *Cluster) error {
	cluster.Id = "newCluster"
	return nil
}

func (m MockClusterRepo) Persist(ctx context.Context, aggregation *ClusterAggregation) error {
	return nil
}

func (m MockClusterRepo) Load(ctx context.Context, id string) (cluster *ClusterAggregation, err error) {
	return &ClusterAggregation{
		Cluster: &Cluster{
			Id:          "testCluster",
			ClusterName: "testCluster",
		},
		CurrentTopologyConfigRecord: &TopologyConfigRecord{
			ConfigModel: &spec.Specification{
				Alertmanagers: []*spec.AlertmanagerSpec{
					{
						Host:    "127.0.0.1",
						WebPort: 9091,
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
	}, nil
}

func (m MockClusterRepo) Query(ctx context.Context, clusterId, clusterName, clusterType, clusterStatus, clusterTag string, page, pageSize int) ([]*ClusterAggregation, int, error) {
	panic("implement me")
}

type MockInstanceRepo struct{}

func (m MockInstanceRepo) QueryParameterJson(ctx context.Context, clusterId string) (string, error) {
	panic("implement me")
}
