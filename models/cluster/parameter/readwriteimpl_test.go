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

/*******************************************************************************
 * @File: parameter_readwrite_test.go
 * @Description: cluster parameter unit test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/13 16:54
*******************************************************************************/

package parameter

import (
	"context"
	"testing"
)

func TestClusterParameterReadWrite_QueryClusterParameter(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestFindParamsByClusterId() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(1, params)
	if err != nil {
		t.Errorf("TestFindParamsByClusterId() test error, build param group err.")
		return
	}
	cluster, err := buildCluster()
	if err != nil {
		t.Errorf("TestFindParamsByClusterId() test error, build cluster err.")
		return
	}
	type args struct {
		clusterId string
		offset    int
		size      int
		params    []*ClusterParameterMapping
	}
	type resp struct {
		paramGroupId string
		total        int64
		params       []*ClusterParamDetail
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, ret resp) bool
	}{
		{
			"normal",
			args{
				clusterId: cluster.ID,
				offset:    0,
				size:      5,
				params: []*ClusterParameterMapping{
					{
						ParameterID: params[0].ID,
						RealValue:   "{\"cluster\": 123}",
					},
					{
						ParameterID: params[1].ID,
						RealValue:   "{\"cluster\": 1024}",
					},
				},
			},
			false,
			[]func(a args, ret resp) bool{
				func(a args, ret resp) bool { return ret.total > 0 },
				func(a args, ret resp) bool { return ret.paramGroupId != "" },
				func(a args, ret resp) bool { return len(ret.params) > 0 },
			},
		},
		{
			"normal2",
			args{
				clusterId: cluster.ID,
				offset:    0,
				size:      1,
				params: []*ClusterParameterMapping{
					{
						ParameterID: params[0].ID,
						RealValue:   "{\"cluster\": 456}",
					},
				},
			},
			false,
			[]func(a args, ret resp) bool{
				func(a args, ret resp) bool { return ret.total > 0 },
				func(a args, ret resp) bool { return ret.paramGroupId != "" },
				func(a args, ret resp) bool { return len(ret.params) > 0 },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clusterParameterRW.ApplyClusterParameter(context.TODO(), groups[0].ID, tt.args.clusterId, tt.args.params)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("ApplyClusterParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			paramGroupId, params, total, err := clusterParameterRW.QueryClusterParameter(context.TODO(), tt.args.clusterId, tt.args.offset, tt.args.size)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("QueryClusterParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ret := resp{
				paramGroupId: paramGroupId,
				total:        total,
				params:       params,
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, ret) {
					t.Errorf("QueryClusterParameter() test error, testname = %v, assert %v, args = %v, got ret = %v", tt.name, i, tt.args, ret)
				}
			}
		})
	}
}

func TestClusterParameterReadWrite_UpdateClusterParameter(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestUpdateClusterParams() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(1, params)
	if err != nil {
		t.Errorf("TestUpdateClusterParams() test error, build param group err.")
		return
	}
	cluster, err := buildCluster()
	if err != nil {
		t.Errorf("TestUpdateClusterParams() test error, build cluster err.")
		return
	}
	type args struct {
		clusterId string
		params    []*ClusterParameterMapping
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, clusterId string) bool
	}{
		{
			"normal",
			args{
				clusterId: cluster.ID,
				params: []*ClusterParameterMapping{
					{
						ParameterID: params[0].ID,
						RealValue:   "{\"cluster\": 123}",
					},
					{
						ParameterID: params[1].ID,
						RealValue:   "{\"cluster\": \"true\"}",
					},
				},
			},
			false,
			[]func(a args, clusterId string) bool{
				func(a args, clusterId string) bool { return clusterId != "" },
				func(a args, clusterId string) bool { return len(clusterId) > 0 },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clusterParameterRW.ApplyClusterParameter(context.TODO(), groups[0].ID, tt.args.clusterId, tt.args.params)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("UpdateClusterParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = clusterParameterRW.UpdateClusterParameter(context.TODO(), tt.args.clusterId, tt.args.params)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("UpdateClusterParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.clusterId) {
					t.Errorf("UpdateClusterParams() test error, testname = %v, assert %v, args = %v, got cluster id = %v", tt.name, i, tt.args, tt.args.clusterId)
				}
			}
		})
	}
}

func TestClusterParameterReadWrite_ApplyClusterParameter(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestApplyParamGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestApplyParamGroup() test error, build param group err.")
		return
	}
	cluster, err := buildCluster()
	if err != nil {
		t.Errorf("TestApplyParamGroup() test error, build cluster err.")
		return
	}
	type args struct {
		id        string
		clusterId string
		params    []*ClusterParameterMapping
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, id string) bool
	}{
		{
			"normal",
			args{
				id:        groups[0].ID,
				clusterId: cluster.ID,
				params: []*ClusterParameterMapping{
					{
						ParameterID: params[0].ID,
						RealValue:   "{\"cluster\": 123}",
					},
					{
						ParameterID: params[1].ID,
						RealValue:   "{\"cluster\": 1024}",
					},
				},
			},
			false,
			[]func(a args, id string) bool{
				func(a args, id string) bool { return id != "" },
			},
		},
		{
			"idIsEmpty",
			args{
				id:        "",
				clusterId: cluster.ID,
				params: []*ClusterParameterMapping{
					{
						ParameterID: params[0].ID,
						RealValue:   "{\"cluster\": 123}",
					},
					{
						ParameterID: params[1].ID,
						RealValue:   "{\"cluster\": 1024}",
					},
				},
			},
			true,
			[]func(a args, id string) bool{},
		},
		{
			"clusterIdIsEmpty",
			args{
				id:        groups[0].ID,
				clusterId: "",
				params: []*ClusterParameterMapping{
					{
						ParameterID: params[0].ID,
						RealValue:   "{\"cluster\": 123}",
					},
					{
						ParameterID: params[1].ID,
						RealValue:   "{\"cluster\": 1024}",
					},
				},
			},
			true,
			[]func(a args, id string) bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clusterParameterRW.ApplyClusterParameter(context.TODO(), tt.args.id, tt.args.clusterId, tt.args.params)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("ApplyParamGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.id) {
					t.Errorf("ApplyParamGroup() test error, testname = %v, assert %v, args = %v, got param group id = %v", tt.name, i, tt.args, tt.args.id)
				}
			}
		})
	}
}
