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
 * @File: param_test.go
 * @Description: test params
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/22 15:11
*******************************************************************************/

package models

import (
	"context"
	"strconv"
	"testing"

	"github.com/pingcap-inc/tiem/library/util/uuidutil"
)

func TestAddParam(t *testing.T) {
	type args struct {
		p *ParamDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, id uint) bool
	}{
		{
			"normal",
			args{
				p: &ParamDO{
					Name:          "param_1",
					ComponentType: "TiDB",
					Type:          0,
					Unit:          "mb",
					Range:         "[1,1024]",
					HasReboot:     0,
					Source:        0,
					Description:   "param 1",
				},
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"normal2",
			args{
				p: &ParamDO{
					Name:          "param_2",
					ComponentType: "PD",
					Type:          0,
					Unit:          "",
					Range:         "[1,10]",
					HasReboot:     0,
					Source:        1,
					Description:   "param 2",
				},
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"normal3",
			args{
				p: &ParamDO{
					Name:          "param_3",
					ComponentType: "TiKV",
					Type:          1,
					Unit:          "",
					Range:         "[true,false]",
					HasReboot:     1,
					Source:        1,
					Description:   "param 3",
				},
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
	}
	paramManager := Dao.ParamManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := paramManager.AddParam(context.TODO(), tt.args.p)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("AddParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, id) {
					t.Errorf("AddParam() test error, testname = %v, assert %v, args = %v, got param id = %v", tt.name, i, tt.args, id)
				}
			}
		})
	}
}

func TestDeleteParam(t *testing.T) {
	paramManager := Dao.ParamManager()
	params, err := buildParams(3)
	if err != nil {
		t.Errorf("TestDeleteParam() test error, build params err.")
		return
	}

	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, id uint) bool
	}{
		{
			"normal",
			args{
				id: params[0].ID,
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"normal2",
			args{
				id: params[1].ID,
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"normal3",
			args{
				id: params[2].ID,
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paramManager.DeleteParam(context.TODO(), tt.args.id)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("DeleteParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.id) {
					t.Errorf("DeleteParam() test error, testname = %v, assert %v, args = %v, got param id = %v", tt.name, i, tt.args, tt.args.id)
				}
			}
		})
	}
}

func buildParams(count uint) (params []*ParamDO, err error) {
	params = make([]*ParamDO, count)
	for i := range params {
		params[i] = &ParamDO{
			Name:          "test_param_" + uuidutil.GenerateID(),
			ComponentType: "TiKV",
			Type:          0,
			Unit:          "kb",
			Range:         "[0,10]",
			HasReboot:     1,
			Source:        1,
			Description:   "test param name order " + strconv.Itoa(i),
		}
		paramId, err := Dao.ParamManager().AddParam(context.TODO(), params[i])
		if err != nil {
			return params, err
		}
		params[i].ID = paramId
	}
	return params, nil
}

func TestUpdateParam(t *testing.T) {
	paramManager := Dao.ParamManager()
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestUpdateParam() test error, build params err.")
		return
	}

	type args struct {
		id uint
		p  *ParamDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, id uint) bool
	}{
		{
			"normal",
			args{
				id: params[0].ID,
				p: &ParamDO{
					Name:          "update_param_1",
					ComponentType: "TiDB",
					Description:   "update param 1",
				},
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"normal2",
			args{
				id: params[1].ID,
				p: &ParamDO{
					Name:          "update_param_2",
					ComponentType: "TiKV",
					Description:   "update param 2",
				},
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paramManager.UpdateParam(context.TODO(), tt.args.id, tt.args.p)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("UpdateParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.id) {
					t.Errorf("UpdateParam() test error, testname = %v, assert %v, args = %v, got param id = %v", tt.name, i, tt.args, tt.args.id)
				}
			}
		})
	}
}

func TestFindParam(t *testing.T) {
	paramManager := Dao.ParamManager()
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestFindParam() test error, build params err.")
		return
	}

	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, p ParamDO) bool
	}{
		{
			"normal",
			args{
				id: params[0].ID,
			},
			false,
			[]func(a args, p ParamDO) bool{
				func(a args, p ParamDO) bool { return a.id == p.ID },
				func(a args, p ParamDO) bool { return len(p.Name) > 0 },
				func(a args, p ParamDO) bool { return len(p.ComponentType) > 0 },
			},
		},
		{
			"normal2",
			args{
				id: params[1].ID,
			},
			false,
			[]func(a args, p ParamDO) bool{
				func(a args, p ParamDO) bool { return a.id == p.ID },
				func(a args, p ParamDO) bool { return len(p.Name) > 0 },
				func(a args, p ParamDO) bool { return len(p.ComponentType) > 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := paramManager.LoadParamById(context.TODO(), tt.args.id)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("FindParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, param) {
					t.Errorf("FindParam() test error, testname = %v, assert %v, args = %v, got param = %v", tt.name, i, tt.args, param)
				}
			}
		})
	}
}

func TestAddParamGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestAddParamGroup() test error, build params err.")
		return
	}
	type args struct {
		pg  *ParamGroupDO
		pgm []*ParamGroupMapDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, id uint) bool
	}{
		{
			"normal",
			args{
				pg: &ParamGroupDO{
					Name:       "test_param_group",
					Spec:       "8C16G",
					HasDefault: 1,
					DbType:     1,
					GroupType:  1,
					Version:    "5.0",
					Note:       "test param group",
				},
				pgm: []*ParamGroupMapDO{
					{
						ParamId:      params[0].ID,
						DefaultValue: "1",
						Note:         "test param 1",
					},
					{
						ParamId:      params[1].ID,
						DefaultValue: "2",
						Note:         "test param 2",
					},
				},
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"nameIsEmpty",
			args{
				pg: &ParamGroupDO{
					Name:       "",
					Spec:       "8C16G",
					HasDefault: 1,
					DbType:     1,
					GroupType:  1,
					Version:    "5.0",
					Note:       "test param group 2",
				},
				pgm: []*ParamGroupMapDO{
					{
						ParamId:      1,
						DefaultValue: "1",
						Note:         "test param 1",
					},
				},
			},
			true,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"specIsEmpty",
			args{
				pg: &ParamGroupDO{
					Name:       "test_param_group2",
					Spec:       "",
					HasDefault: 1,
					DbType:     1,
					GroupType:  1,
					Version:    "5.0",
					Note:       "test param group 2",
				},
				pgm: []*ParamGroupMapDO{
					{
						ParamId:      1,
						DefaultValue: "1",
						Note:         "test param 1",
					},
				},
			},
			true,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"versionIsEmpty",
			args{
				pg: &ParamGroupDO{
					Name:       "test_param_group3",
					Spec:       "8C16G",
					HasDefault: 1,
					DbType:     1,
					GroupType:  1,
					Version:    "",
					Note:       "test param group 2",
				},
				pgm: []*ParamGroupMapDO{
					{
						ParamId:      1,
						DefaultValue: "1",
						Note:         "test param 1",
					},
				},
			},
			true,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
	}
	paramManager := Dao.ParamManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := paramManager.AddParamGroup(context.TODO(), tt.args.pg, tt.args.pgm)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("AddParamGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, id) {
					t.Errorf("AddParamGroup() test error, testname = %v, assert %v, args = %v, got param group id = %v", tt.name, i, tt.args, id)
				}
			}
		})
	}
}

func buildParamGroup(count uint, params []*ParamDO) (pgs []*ParamGroupDO, err error) {
	pgs = make([]*ParamGroupDO, count)
	for i := range pgs {
		pgs[i] = &ParamGroupDO{
			Name:       "test_param_group_" + uuidutil.GenerateID(),
			Spec:       "8C16G",
			HasDefault: 1,
			DbType:     1,
			GroupType:  1,
			Version:    "5.0",
			Note:       "test param group " + strconv.Itoa(i),
		}
		pgm := make([]*ParamGroupMapDO, len(params))
		for j := range params {
			pgm[j] = &ParamGroupMapDO{
				ParamId:      params[j].ID,
				DefaultValue: strconv.Itoa(j),
				Note:         "test param " + strconv.Itoa(j),
			}
		}
		paramId, err := Dao.ParamManager().AddParamGroup(context.TODO(), pgs[i], pgm)
		if err != nil {
			return pgs, err
		}
		pgs[i].ID = paramId
	}
	return pgs, nil
}

func TestUpdateParamGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestUpdateParamGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestUpdateParamGroup() test error, build param group err.")
		return
	}
	type args struct {
		id      uint
		name    string
		spec    string
		version string
		note    string
		pgm     []*ParamGroupMapDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, id uint) bool
	}{
		{
			"normal",
			args{
				id:      groups[0].ID,
				name:    "update_param_group_1",
				spec:    "16C32G",
				version: "5.1",
				note:    "update param group 1",
				pgm: []*ParamGroupMapDO{
					{
						ParamId:      params[0].ID,
						DefaultValue: "11",
						Note:         "test param 1",
					},
					{
						ParamId:      params[1].ID,
						DefaultValue: "21",
						Note:         "test param 2",
					},
				},
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"normal2",
			args{
				id:      groups[1].ID,
				name:    "update_param_group_2",
				spec:    "8C32G",
				version: "5.2",
				note:    "update param group 2",
				pgm: []*ParamGroupMapDO{
					{
						ParamId:      params[0].ID,
						DefaultValue: "22",
						Note:         "test param 1",
					},
					{
						ParamId:      params[1].ID,
						DefaultValue: "32",
						Note:         "test param 2",
					},
				},
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
	}
	paramManager := Dao.ParamManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paramManager.UpdateParamGroup(context.TODO(), tt.args.id, tt.args.name,
				tt.args.spec, tt.args.version, tt.args.note, tt.args.pgm)
			if (err != nil) && tt.wantErr {
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.id) {
					t.Errorf("UpdateParamGroup() test error, testname = %v, assert %v, args = %v, got param group id = %v", tt.name, i, tt.args, tt.args.id)
				}
			}
		})
	}
}

func TestDeleteParamGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestDeleteParamGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestDeleteParamGroup() test error, build param group err.")
		return
	}
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, id uint) bool
	}{
		{
			"normal",
			args{
				id: groups[0].ID,
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
		{
			"normal2",
			args{
				id: groups[1].ID,
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
	}
	paramManager := Dao.ParamManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paramManager.DeleteParamGroup(context.TODO(), tt.args.id)
			if (err != nil) && tt.wantErr {
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.id) {
					t.Errorf("DeleteParamGroup() test error, testname = %v, assert %v, args = %v, got param group id = %v", tt.name, i, tt.args, tt.args.id)
				}
			}
		})
	}
}

func TestListParamGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestListParamGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestListParamGroup() test error, build param group err.")
		return
	}

	type args struct {
		name       string
		spec       string
		version    string
		dbType     int32
		hasDefault int32
		offset     int
		size       int
	}
	type resp struct {
		groups []*ParamGroupDO
		total  int64
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
				name:       "param",
				spec:       groups[0].Spec,
				version:    groups[0].Version,
				dbType:     1,
				hasDefault: 1,
				offset:     0,
				size:       10,
			},
			false,
			[]func(a args, ret resp) bool{
				func(a args, ret resp) bool { return len(ret.groups) > 0 },
				func(a args, ret resp) bool { return ret.total > 0 },
			},
		},
	}
	paramManager := Dao.ParamManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, total, err := paramManager.ListParamGroup(context.TODO(), tt.args.name, tt.args.spec,
				tt.args.version, tt.args.dbType, tt.args.hasDefault, tt.args.offset, tt.args.size)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("LoadParamGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ret := resp{
				groups: groups,
				total:  total,
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, ret) {
					t.Errorf("LoadParamGroup() test error, testname = %v, assert %v, args = %v, got ret = %v", tt.name, i, tt.args, ret)
				}
			}
		})
	}
}

func TestLoadParamGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestLoadParamGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestLoadParamGroup() test error, build param group err.")
		return
	}

	type args struct {
		id uint
	}
	type resp struct {
		group  *ParamGroupDO
		params []*ParamDetail
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
				id: groups[0].ID,
			},
			false,
			[]func(a args, ret resp) bool{
				func(a args, ret resp) bool { return len(ret.params) > 0 },
				func(a args, ret resp) bool { return ret.group.Name != "" },
			},
		},
		{
			"normal2",
			args{
				id: groups[1].ID,
			},
			false,
			[]func(a args, ret resp) bool{
				func(a args, ret resp) bool { return len(ret.params) > 0 },
				func(a args, ret resp) bool { return ret.group.Name != "" },
			},
		},
	}
	paramManager := Dao.ParamManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, params, err := paramManager.LoadParamGroup(context.TODO(), tt.args.id)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("LoadParamGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ret := resp{
				group:  group,
				params: params,
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, ret) {
					t.Errorf("LoadParamGroup() test error, testname = %v, assert %v, args = %v, got ret = %v", tt.name, i, tt.args, ret)
				}
			}
		})
	}
}

func TestApplyParamGroup(t *testing.T) {
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
		id        uint
		clusterId string
		params    []*ClusterParamMapDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, id uint) bool
	}{
		{
			"normal",
			args{
				id:        groups[0].ID,
				clusterId: cluster.ID,
				params: []*ClusterParamMapDO{
					{
						ParamId:   params[0].ID,
						RealValue: "{\"cluster\": 123}",
					},
					{
						ParamId:   params[1].ID,
						RealValue: "{\"cluster\": 1024}",
					},
				},
			},
			false,
			[]func(a args, id uint) bool{
				func(a args, id uint) bool { return id > 0 },
			},
		},
	}
	paramManager := Dao.ParamManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paramManager.ApplyParamGroup(context.TODO(), tt.args.id, tt.args.clusterId, tt.args.params)
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

func TestFindParamsByClusterId(t *testing.T) {
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
		params    []*ClusterParamMapDO
	}
	type resp struct {
		paramGroupId uint
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
				params: []*ClusterParamMapDO{
					{
						ParamId:   params[0].ID,
						RealValue: "{\"cluster\": 123}",
					},
					{
						ParamId:   params[1].ID,
						RealValue: "{\"cluster\": 1024}",
					},
				},
			},
			false,
			[]func(a args, ret resp) bool{
				func(a args, ret resp) bool { return ret.total > 0 },
				func(a args, ret resp) bool { return ret.paramGroupId > 0 },
				func(a args, ret resp) bool { return len(ret.params) > 0 },
			},
		},
		{
			"normal2",
			args{
				clusterId: cluster.ID,
				offset:    0,
				size:      1,
				params: []*ClusterParamMapDO{
					{
						ParamId:   params[0].ID,
						RealValue: "{\"cluster\": 456}",
					},
				},
			},
			false,
			[]func(a args, ret resp) bool{
				func(a args, ret resp) bool { return ret.total > 0 },
				func(a args, ret resp) bool { return ret.paramGroupId > 0 },
				func(a args, ret resp) bool { return len(ret.params) > 0 },
			},
		},
	}
	paramManager := Dao.ParamManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paramManager.ApplyParamGroup(context.TODO(), groups[0].ID, tt.args.clusterId, tt.args.params)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("FindParamsByClusterId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			paramGroupId, total, params, err := paramManager.FindParamsByClusterId(context.TODO(), tt.args.clusterId, tt.args.offset, tt.args.size)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("FindParamsByClusterId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ret := resp{
				paramGroupId: paramGroupId,
				total:        total,
				params:       params,
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, ret) {
					t.Errorf("FindParamsByClusterId() test error, testname = %v, assert %v, args = %v, got ret = %v", tt.name, i, tt.args, ret)
				}
			}
		})
	}
}

func TestUpdateClusterParams(t *testing.T) {
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
		params    []*ClusterParamMapDO
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
				params: []*ClusterParamMapDO{
					{
						ParamId:   params[0].ID,
						RealValue: "{\"cluster\": 123}",
					},
					{
						ParamId:   params[1].ID,
						RealValue: "{\"cluster\": \"true\"}",
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
	paramManager := Dao.ParamManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := paramManager.ApplyParamGroup(context.TODO(), groups[0].ID, tt.args.clusterId, tt.args.params)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("UpdateClusterParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = paramManager.UpdateClusterParams(context.TODO(), tt.args.clusterId, tt.args.params)
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

func buildCluster() (*Cluster, error) {
	cluster, err := Dao.ClusterManager().CreateCluster(context.TODO(), Cluster{
		Entity: Entity{
			TenantId: "111",
		},
		Name:    "build_test_cluster_" + uuidutil.GenerateID(),
		OwnerId: "111",
	})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}
