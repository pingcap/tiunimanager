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
 * @File: parametergroup_readwrite_test.go
 * @Description: parameter group unit test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/13 15:00
*******************************************************************************/

package parametergroup

import (
	"context"
	"testing"
)

func TestParameterGroupReadWrite_CreateParameterGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_CreateParameterGroup() test error, build params err.")
		return
	}
	type args struct {
		pg  *ParameterGroup
		pgm []*ParameterGroupMapping
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, pg *ParameterGroup) bool
	}{
		{
			"normal",
			args{
				pg: &ParameterGroup{
					Name:           "test_param_group",
					ClusterSpec:    "8C16G",
					HasDefault:     1,
					DBType:         1,
					GroupType:      1,
					ClusterVersion: "5.0",
					Note:           "test param group",
				},
				pgm: []*ParameterGroupMapping{
					{
						ParameterID:  params[0].ID,
						DefaultValue: "1",
						Note:         "test param 1",
					},
					{
						ParameterID:  params[1].ID,
						DefaultValue: "2",
						Note:         "test param 2",
					},
				},
			},
			false,
			[]func(a args, pg *ParameterGroup) bool{
				func(a args, pg *ParameterGroup) bool { return pg.ID != "" },
			},
		},
		{
			"nameIsEmpty",
			args{
				pg: &ParameterGroup{
					Name:           "",
					ClusterSpec:    "8C16G",
					HasDefault:     1,
					DBType:         1,
					GroupType:      1,
					ClusterVersion: "5.0",
					Note:           "test param group 2",
				},
				pgm: []*ParameterGroupMapping{
					{
						ParameterGroupID: "1",
						DefaultValue:     "1",
						Note:             "test param 1",
					},
				},
			},
			true,
			[]func(a args, pg *ParameterGroup) bool{
				func(a args, pg *ParameterGroup) bool { return pg.ID != "" },
			},
		},
		{
			"specIsEmpty",
			args{
				pg: &ParameterGroup{
					Name:           "test_param_group2",
					ClusterSpec:    "",
					HasDefault:     1,
					DBType:         1,
					GroupType:      1,
					ClusterVersion: "5.0",
					Note:           "test param group 2",
				},
				pgm: []*ParameterGroupMapping{
					{
						ParameterGroupID: "1",
						DefaultValue:     "1",
						Note:             "test param 1",
					},
				},
			},
			true,
			[]func(a args, pg *ParameterGroup) bool{
				func(a args, pg *ParameterGroup) bool { return pg.ID != "" },
			},
		},
		{
			"versionIsEmpty",
			args{
				pg: &ParameterGroup{
					Name:           "test_param_group3",
					ClusterSpec:    "8C16G",
					HasDefault:     1,
					DBType:         1,
					GroupType:      1,
					ClusterVersion: "",
					Note:           "test param group 2",
				},
				pgm: []*ParameterGroupMapping{
					{
						ParameterGroupID: "1",
						DefaultValue:     "1",
						Note:             "test param 1",
					},
				},
			},
			true,
			[]func(a args, pg *ParameterGroup) bool{
				func(a args, pg *ParameterGroup) bool { return pg.ID != "" },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parameterGroup, err := testRW.CreateParameterGroup(context.TODO(), tt.args.pg, tt.args.pgm)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("CreateParameterGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, parameterGroup) {
					t.Errorf("CreateParameterGroup() test error, testname = %v, assert %v, args = %v, got param group = %v", tt.name, i, tt.args, parameterGroup)
				}
			}
		})
	}
}

func TestParameterGroupReadWrite_DeleteParameterGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_DeleteParameterGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_DeleteParameterGroup() test error, build param group err.")
		return
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args) bool
	}{
		{
			"normal",
			args{
				id: groups[0].ID,
			},
			false,
			[]func(a args) bool{
				func(a args) bool { return a.id != "" },
			},
		},
		{
			"normal2",
			args{
				id: groups[1].ID,
			},
			false,
			[]func(a args) bool{
				func(a args) bool { return a.id != "" },
			},
		},
		{
			"idIsEmpty",
			args{
				id: "",
			},
			true,
			[]func(a args) bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRW.DeleteParameterGroup(context.TODO(), tt.args.id)
			if (err != nil) && tt.wantErr {
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args) {
					t.Errorf("DeleteParameterGroup() test error, testname = %v, assert %v, args = %v, got param group id = %v", tt.name, i, tt.args, tt.args.id)
				}
			}
		})
	}
}

func TestParameterGroupReadWrite_UpdateParameterGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_UpdateParameterGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_UpdateParameterGroup() test error, build param group err.")
		return
	}
	type args struct {
		pg  *ParameterGroup
		pgm []*ParameterGroupMapping
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args) bool
	}{
		{
			"normal",
			args{
				pg: &ParameterGroup{
					ID:             groups[0].ID,
					Name:           "update_param_group_1",
					ClusterSpec:    "16C32G",
					ClusterVersion: "5.1",
					Note:           "update param group 1",
				},
				pgm: []*ParameterGroupMapping{
					{
						ParameterID:  params[0].ID,
						DefaultValue: "11",
						Note:         "test param 1",
					},
					{
						ParameterID:  params[1].ID,
						DefaultValue: "21",
						Note:         "test param 2",
					},
				},
			},
			false,
			[]func(a args) bool{
				func(a args) bool { return a.pg.ID != "" },
			},
		},
		{
			"normal2",
			args{
				pg: &ParameterGroup{
					ID:             groups[1].ID,
					Name:           "update_param_group_2",
					ClusterSpec:    "8C32G",
					ClusterVersion: "5.2",
					Note:           "update param group 2",
				},
				pgm: []*ParameterGroupMapping{
					{
						ParameterID:  params[0].ID,
						DefaultValue: "22",
						Note:         "test param 1",
					},
					{
						ParameterID:  params[1].ID,
						DefaultValue: "32",
						Note:         "test param 2",
					},
				},
			},
			false,
			[]func(a args) bool{
				func(a args) bool { return a.pg.ID != "" },
			},
		},
		{
			"idIsEmpty",
			args{
				pg: &ParameterGroup{
					ID:             "",
					Name:           "update_param_group_3",
					ClusterSpec:    "8C32G",
					ClusterVersion: "5.2",
					Note:           "update param group 3",
				},
				pgm: []*ParameterGroupMapping{
					{
						ParameterID:  "",
						DefaultValue: "22",
						Note:         "test param 3",
					},
					{
						ParameterID:  "",
						DefaultValue: "32",
						Note:         "test param 3",
					},
				},
			},
			true,
			[]func(a args) bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRW.UpdateParameterGroup(context.TODO(), tt.args.pg, tt.args.pgm)
			if (err != nil) && tt.wantErr {
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args) {
					t.Errorf("UpdateParamGroup() test error, testname = %v, assert %v, args = %v, got param group id = %v", tt.name, i, tt.args, tt.args.pg.ID)
				}
			}
		})
	}
}

func TestParameterGroupReadWrite_QueryParameterGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_QueryParameterGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_QueryParameterGroup() test error, build param group err.")
		return
	}

	type args struct {
		name       string
		spec       string
		version    string
		dbType     int
		hasDefault int
		offset     int
		size       int
	}
	type resp struct {
		groups []*ParameterGroup
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
				spec:       groups[0].ClusterSpec,
				version:    groups[0].ClusterVersion,
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, total, err := testRW.QueryParameterGroup(context.TODO(), tt.args.name, tt.args.spec,
				tt.args.version, tt.args.dbType, tt.args.hasDefault, tt.args.offset, tt.args.size)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("QueryParameterGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ret := resp{
				groups: groups,
				total:  total,
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, ret) {
					t.Errorf("QueryParameterGroup() test error, testname = %v, assert %v, args = %v, got ret = %v", tt.name, i, tt.args, ret)
				}
			}
		})
	}
}

func TestParameterGroupReadWrite_GetParameterGroup(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_GetParameterGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_GetParameterGroup() test error, build param group err.")
		return
	}

	type args struct {
		id string
	}
	type resp struct {
		group  *ParameterGroup
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
		{
			"idIsEmpty",
			args{
				id: "",
			},
			true,
			[]func(a args, ret resp) bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, params, err := testRW.GetParameterGroup(context.TODO(), tt.args.id)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("GetParameterGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ret := resp{
				group:  group,
				params: params,
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, ret) {
					t.Errorf("GetParameterGroup() test error, testname = %v, assert %v, args = %v, got ret = %v", tt.name, i, tt.args, ret)
				}
			}
		})
	}
}

func TestParameterGroupReadWrite_CreateParameter(t *testing.T) {
	type args struct {
		p *Parameter
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, parameter *Parameter) bool
	}{
		{
			"normal",
			args{
				p: &Parameter{
					Category:     "basic",
					Name:         "param_1",
					InstanceType: "TiDB",
					Type:         0,
					Unit:         "mb",
					Range:        "[\"1\",\"1024\"]",
					HasReboot:    0,
					UpdateSource: 0,
					Description:  "param 1",
				},
			},
			false,
			[]func(a args, parameter *Parameter) bool{
				func(a args, parameter *Parameter) bool { return parameter != nil },
				func(a args, parameter *Parameter) bool { return parameter.ID != "" },
				func(a args, parameter *Parameter) bool { return parameter.Name == "param_1" },
			},
		},
		{
			"NameIsEmpty",
			args{
				p: &Parameter{
					Category:     "basic",
					Name:         "",
					InstanceType: "PD",
					Type:         0,
					Unit:         "",
					Range:        "[\"1\",\"10\"]",
					HasReboot:    0,
					UpdateSource: 1,
					Description:  "param 2",
				},
			},
			true,
			[]func(a args, parameter *Parameter) bool{
				func(a args, parameter *Parameter) bool { return parameter != nil },
				func(a args, parameter *Parameter) bool { return parameter.ID != "" },
				func(a args, parameter *Parameter) bool { return parameter.Name == "param_2" },
			},
		},
		{
			"CategoryIsEmpty",
			args{
				p: &Parameter{
					Category:     "",
					Name:         "param_3",
					InstanceType: "TiKV",
					Type:         1,
					Unit:         "",
					HasReboot:    1,
					UpdateSource: 1,
					Description:  "param 3",
				},
			},
			true,
			[]func(a args, parameter *Parameter) bool{
				func(a args, parameter *Parameter) bool { return parameter != nil },
				func(a args, parameter *Parameter) bool { return parameter.ID != "" },
				func(a args, parameter *Parameter) bool { return parameter.Name == "param_3" },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parameter, err := testRW.CreateParameter(context.TODO(), tt.args.p)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("CreateParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, parameter) {
					t.Errorf("CreateParameter() test error, testname = %v, assert %v, args = %v, got parameter = %v", tt.name, i, tt.args, parameter)
				}
			}
		})
	}
}

func TestParameterGroupReadWrite_DeleteParameter(t *testing.T) {
	params, err := buildParams(3)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_DeleteParameter() test error, build params err.")
		return
	}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args) bool
	}{
		{
			"normal",
			args{
				id: params[0].ID,
			},
			false,
			[]func(a args) bool{
				func(a args) bool { return a.id != "" },
			},
		},
		{
			"normal2",
			args{
				id: params[1].ID,
			},
			false,
			[]func(a args) bool{
				func(a args) bool { return a.id != "" },
			},
		},
		{
			"normal3",
			args{
				id: params[2].ID,
			},
			false,
			[]func(a args) bool{
				func(a args) bool { return a.id != "" },
			},
		},
		{
			"idIsEmpty",
			args{
				id: "",
			},
			true,
			[]func(a args) bool{
				func(a args) bool { return a.id == "" },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRW.DeleteParameter(context.TODO(), tt.args.id)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("DeleteParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args) {
					t.Errorf("DeleteParameter() test error, testname = %v, assert %v, args = %v, got param id = %v", tt.name, i, tt.args, tt.args.id)
				}
			}
		})
	}
}

func TestParameterGroupReadWrite_UpdateParameter(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_UpdateParameter() test error, build params err.")
		return
	}

	type args struct {
		p *Parameter
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args) bool
	}{
		{
			"normal",
			args{
				p: &Parameter{
					ID:           params[0].ID,
					Name:         "update_param_1",
					InstanceType: "TiDB",
					Description:  "update param 1",
				},
			},
			false,
			[]func(a args) bool{
				func(a args) bool { return a.p.ID != "" },
			},
		},
		{
			"normal2",
			args{
				p: &Parameter{
					ID:           params[1].ID,
					Name:         "update_param_2",
					InstanceType: "TiKV",
					Description:  "update param 2",
				},
			},
			false,
			[]func(a args) bool{
				func(a args) bool { return a.p.ID != "" },
			},
		},
		{
			"idIsEmpty",
			args{
				p: &Parameter{
					ID: "",
				},
			},
			true,
			[]func(a args) bool{
				func(a args) bool { return a.p.ID != "" },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRW.UpdateParameter(context.TODO(), tt.args.p)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("UpdateParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args) {
					t.Errorf("UpdateParameter() test error, testname = %v, assert %v, args = %v, got parameter = %v", tt.name, i, tt.args, tt.args.p)
				}
			}
		})
	}
}

func TestParameterGroupReadWrite_GetParameter(t *testing.T) {
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestParameterGroupReadWrite_GetParameter() test error, build params err.")
		return
	}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, p *Parameter) bool
	}{
		{
			"normal",
			args{
				id: params[0].ID,
			},
			false,
			[]func(a args, p *Parameter) bool{
				func(a args, p *Parameter) bool { return a.id == p.ID },
				func(a args, p *Parameter) bool { return len(p.Name) > 0 },
				func(a args, p *Parameter) bool { return len(p.InstanceType) > 0 },
			},
		},
		{
			"normal2",
			args{
				id: params[1].ID,
			},
			false,
			[]func(a args, p *Parameter) bool{
				func(a args, p *Parameter) bool { return a.id == p.ID },
				func(a args, p *Parameter) bool { return len(p.Name) > 0 },
				func(a args, p *Parameter) bool { return len(p.InstanceType) > 0 },
			},
		},
		{
			"idIsEmpty",
			args{
				id: "",
			},
			true,
			[]func(a args, p *Parameter) bool{
				func(a args, p *Parameter) bool { return a.id == p.ID },
				func(a args, p *Parameter) bool { return len(p.Name) > 0 },
				func(a args, p *Parameter) bool { return len(p.InstanceType) > 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := testRW.GetParameter(context.TODO(), tt.args.id)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("GetParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, param) {
					t.Errorf("GetParameter() test error, testname = %v, assert %v, args = %v, got param = %v", tt.name, i, tt.args, param)
				}
			}
		})
	}
}
