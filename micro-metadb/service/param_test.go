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
 * @Description: param service test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/23 14:27
*******************************************************************************/

package service

import (
	"context"
	"strconv"
	"testing"

	"github.com/pingcap-inc/tiem/micro-metadb/models"

	"github.com/pingcap-inc/tiem/library/util/uuidutil"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/framework"
)

func TestDBServiceHandler_CreateParamGroup(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestDBServiceHandler_CreateParamGroup() test error, build params err.")
		return
	}

	type args struct {
		req *dbpb.DBCreateParamGroupRequest
		rsp *dbpb.DBCreateParamGroupResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, rsp *dbpb.DBCreateParamGroupResponse) bool
	}{
		{
			"normal",
			args{
				req: &dbpb.DBCreateParamGroupRequest{
					Name:       "test_param_group",
					DbType:     1,
					HasDefault: 1,
					Version:    "5.0",
					Spec:       "TiKV",
					GroupType:  1,
					Note:       "test param group",
					Params: []*dbpb.DBSubmitParamDTO{
						{
							ParamId:      int64(params[0].ID),
							DefaultValue: "1",
							Note:         "test param 1",
						},
						{
							ParamId:      int64(params[1].ID),
							DefaultValue: "2",
							Note:         "test param 2",
						},
					},
				},
				rsp: &dbpb.DBCreateParamGroupResponse{
					ParamGroupId: 0,
					Status:       nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBCreateParamGroupResponse) bool{
				func(a args, rsp *dbpb.DBCreateParamGroupResponse) bool { return rsp.ParamGroupId > 0 },
				func(a args, rsp *dbpb.DBCreateParamGroupResponse) bool { return rsp.Status.Code == 0 },
			},
		},
		{
			"nameIsEmpty",
			args{
				req: &dbpb.DBCreateParamGroupRequest{
					Name:       "",
					DbType:     1,
					HasDefault: 1,
					Version:    "4.0",
					Spec:       "PD",
					GroupType:  1,
					Note:       "test param group",
					Params: []*dbpb.DBSubmitParamDTO{
						{
							ParamId:      int64(params[0].ID),
							DefaultValue: "1",
							Note:         "test param 1",
						},
					},
				},
				rsp: &dbpb.DBCreateParamGroupResponse{
					ParamGroupId: 0,
					Status:       nil,
				},
			},
			true,
			[]func(a args, rsp *dbpb.DBCreateParamGroupResponse) bool{
				func(a args, rsp *dbpb.DBCreateParamGroupResponse) bool { return rsp.ParamGroupId == 0 },
				func(a args, rsp *dbpb.DBCreateParamGroupResponse) bool { return rsp.Status.Code > 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.CreateParamGroup(context.TODO(), tt.args.req, tt.args.rsp)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("CreateParamGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.rsp) {
					t.Errorf("CreateParamGroup() test error, testname = %v, assert %v, args = %v, got rsp = %v", tt.name, i, tt.args, tt.args.rsp)
				}
			}
		})
	}
}

func TestDBServiceHandler_UpdateParamGroup(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestDBServiceHandler_UpdateParamGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(1, params)
	if err != nil {
		t.Errorf("TestDBServiceHandler_UpdateParamGroup() test error, build param group err.")
		return
	}

	type args struct {
		req *dbpb.DBUpdateParamGroupRequest
		rsp *dbpb.DBUpdateParamGroupResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, rsp *dbpb.DBUpdateParamGroupResponse) bool
	}{
		{
			"normal",
			args{
				req: &dbpb.DBUpdateParamGroupRequest{
					ParamGroupId: int64(groups[0].ID),
					Name:         "update_param_group_1",
					Version:      "5.0.2",
					Spec:         "16C32G",
					Note:         "update param group 1",
					Params:       nil,
				},
				rsp: &dbpb.DBUpdateParamGroupResponse{
					ParamGroupId: 0,
					Status:       nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBUpdateParamGroupResponse) bool{
				func(a args, rsp *dbpb.DBUpdateParamGroupResponse) bool { return rsp.ParamGroupId > 0 },
				func(a args, rsp *dbpb.DBUpdateParamGroupResponse) bool { return rsp.Status.Code == 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.UpdateParamGroup(context.TODO(), tt.args.req, tt.args.rsp)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("UpdateParamGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.rsp) {
					t.Errorf("UpdateParamGroup() test error, testname = %v, assert %v, args = %v, got rsp = %v", tt.name, i, tt.args, tt.args.rsp)
				}
			}
		})
	}
}

func TestDBServiceHandler_DeleteParamGroup(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestDBServiceHandler_DeleteParamGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestDBServiceHandler_DeleteParamGroup() test error, build param group err.")
		return
	}

	type args struct {
		req *dbpb.DBDeleteParamGroupRequest
		rsp *dbpb.DBDeleteParamGroupResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, rsp *dbpb.DBDeleteParamGroupResponse) bool
	}{
		{
			"normal",
			args{
				req: &dbpb.DBDeleteParamGroupRequest{
					ParamGroupId: int64(groups[0].ID),
				},
				rsp: &dbpb.DBDeleteParamGroupResponse{
					ParamGroupId: 0,
					Status:       nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBDeleteParamGroupResponse) bool{
				func(a args, rsp *dbpb.DBDeleteParamGroupResponse) bool { return rsp.ParamGroupId > 0 },
				func(a args, rsp *dbpb.DBDeleteParamGroupResponse) bool { return rsp.Status.Code == 0 },
			},
		},
		{
			"normal2",
			args{
				req: &dbpb.DBDeleteParamGroupRequest{
					ParamGroupId: int64(groups[1].ID),
				},
				rsp: &dbpb.DBDeleteParamGroupResponse{
					ParamGroupId: 0,
					Status:       nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBDeleteParamGroupResponse) bool{
				func(a args, rsp *dbpb.DBDeleteParamGroupResponse) bool { return rsp.ParamGroupId > 0 },
				func(a args, rsp *dbpb.DBDeleteParamGroupResponse) bool { return rsp.Status.Code == 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.DeleteParamGroup(context.TODO(), tt.args.req, tt.args.rsp)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("DeleteParamGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.rsp) {
					t.Errorf("DeleteParamGroup() test error, testname = %v, assert %v, args = %v, got rsp = %v", tt.name, i, tt.args, tt.args.rsp)
				}
			}
		})
	}
}

func TestDBServiceHandler_ListParamGroup(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestDBServiceHandler_ListParamGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(2, params)
	if err != nil {
		t.Errorf("TestDBServiceHandler_ListParamGroup() test error, build param group err.")
		return
	}

	type args struct {
		req *dbpb.DBListParamGroupRequest
		rsp *dbpb.DBListParamGroupResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, rsp *dbpb.DBListParamGroupResponse) bool
	}{
		{
			"normal",
			args{
				req: &dbpb.DBListParamGroupRequest{
					Name:      "group",
					DbType:    int32(groups[0].DbType),
					HasDetail: false,
					Page: &dbpb.DBParamsPageDTO{
						Page:     1,
						PageSize: 10,
						Total:    10,
					},
				},
				rsp: &dbpb.DBListParamGroupResponse{
					ParamGroups: nil,
					Page:        nil,
					Status:      nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBListParamGroupResponse) bool{
				func(a args, rsp *dbpb.DBListParamGroupResponse) bool { return len(rsp.ParamGroups) > 0 },
				func(a args, rsp *dbpb.DBListParamGroupResponse) bool { return rsp.Status.Code == 0 },
				func(a args, rsp *dbpb.DBListParamGroupResponse) bool { return rsp.Page.Total > 0 },
				func(a args, rsp *dbpb.DBListParamGroupResponse) bool { return rsp.ParamGroups[0].Params == nil },
			},
		},
		{
			"param detail",
			args{
				req: &dbpb.DBListParamGroupRequest{
					Name:      "param",
					DbType:    int32(groups[0].DbType),
					HasDetail: true,
					Page: &dbpb.DBParamsPageDTO{
						Page:     1,
						PageSize: 10,
						Total:    10,
					},
				},
				rsp: &dbpb.DBListParamGroupResponse{
					ParamGroups: nil,
					Page:        nil,
					Status:      nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBListParamGroupResponse) bool{
				func(a args, rsp *dbpb.DBListParamGroupResponse) bool { return len(rsp.ParamGroups) > 0 },
				func(a args, rsp *dbpb.DBListParamGroupResponse) bool { return rsp.Status.Code == 0 },
				func(a args, rsp *dbpb.DBListParamGroupResponse) bool { return rsp.Page.Total > 0 },
				func(a args, rsp *dbpb.DBListParamGroupResponse) bool { return len(rsp.ParamGroups[0].Params) > 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ListParamGroup(context.TODO(), tt.args.req, tt.args.rsp)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("ListParamGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.rsp) {
					t.Errorf("ListParamGroup() test error, testname = %v, assert %v, args = %v, got rsp = %v", tt.name, i, tt.args, tt.args.rsp)
				}
			}
		})
	}
}

func TestDBServiceHandler_FindParamGroupByID(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestDBServiceHandler_FindParamGroupByID() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(1, params)
	if err != nil {
		t.Errorf("TestDBServiceHandler_FindParamGroupByID() test error, build param group err.")
		return
	}

	type args struct {
		req *dbpb.DBFindParamGroupByIDRequest
		rsp *dbpb.DBFindParamGroupByIDResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, rsp *dbpb.DBFindParamGroupByIDResponse) bool
	}{
		{
			"normal",
			args{
				req: &dbpb.DBFindParamGroupByIDRequest{
					ParamGroupId: int64(groups[0].ID),
				},
				rsp: &dbpb.DBFindParamGroupByIDResponse{
					ParamGroup: nil,
					Status:     nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBFindParamGroupByIDResponse) bool{
				func(a args, rsp *dbpb.DBFindParamGroupByIDResponse) bool { return rsp.ParamGroup.Name != "" },
				func(a args, rsp *dbpb.DBFindParamGroupByIDResponse) bool { return rsp.Status.Code == 0 },
				func(a args, rsp *dbpb.DBFindParamGroupByIDResponse) bool { return len(rsp.ParamGroup.Params) > 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.FindParamGroupByID(context.TODO(), tt.args.req, tt.args.rsp)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("FindParamGroupByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.rsp) {
					t.Errorf("FindParamGroupByID() test error, testname = %v, assert %v, args = %v, got rsp = %v", tt.name, i, tt.args, tt.args.rsp)
				}
			}
		})
	}
}

func TestDBServiceHandler_ApplyParamGroup(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestDBServiceHandler_ApplyParamGroup() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(1, params)
	if err != nil {
		t.Errorf("TestDBServiceHandler_ApplyParamGroup() test error, build param group err.")
		return
	}
	cluster, err := buildCluster()
	if err != nil {
		t.Errorf("TestDBServiceHandler_ApplyParamGroup() test error, build cluster err.")
		return
	}

	type args struct {
		req *dbpb.DBApplyParamGroupRequest
		rsp *dbpb.DBApplyParamGroupResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, rsp *dbpb.DBApplyParamGroupResponse) bool
	}{
		{
			"normal",
			args{
				req: &dbpb.DBApplyParamGroupRequest{
					ParamGroupId: int64(groups[0].ID),
					ClusterId:    cluster.ID,
					Params: []*dbpb.DBApplyParamDTO{
						{
							ParamId: int64(params[0].ID),
							RealValue: &dbpb.DBParamRealValueDTO{
								Cluster: "123",
								Instances: []*dbpb.DBParamInstanceRealValueDTO{
									{
										Instance: "172.16.1.1",
										Value:    "100",
									},
								},
							},
						},
						{
							ParamId: int64(params[1].ID),
							RealValue: &dbpb.DBParamRealValueDTO{
								Cluster:   "1024",
								Instances: nil,
							},
						},
					},
				},
				rsp: &dbpb.DBApplyParamGroupResponse{
					ParamGroupId: 0,
					ClusterId:    "",
					Status:       nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBApplyParamGroupResponse) bool{
				func(a args, rsp *dbpb.DBApplyParamGroupResponse) bool { return rsp.ParamGroupId > 0 },
				func(a args, rsp *dbpb.DBApplyParamGroupResponse) bool { return rsp.Status.Code == 0 },
				func(a args, rsp *dbpb.DBApplyParamGroupResponse) bool { return rsp.ClusterId != "" },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ApplyParamGroup(context.TODO(), tt.args.req, tt.args.rsp)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("ApplyParamGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.rsp) {
					t.Errorf("ApplyParamGroup() test error, testname = %v, assert %v, args = %v, got rsp = %v", tt.name, i, tt.args, tt.args.rsp)
				}
			}
		})
	}
}

func TestDBServiceHandler_FindParamsByClusterId(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestDBServiceHandler_FindParamsByClusterId() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(1, params)
	if err != nil {
		t.Errorf("TestDBServiceHandler_FindParamsByClusterId() test error, build param group err.")
		return
	}
	cluster, err := buildCluster()
	if err != nil {
		t.Errorf("TestDBServiceHandler_FindParamsByClusterId() test error, build cluster err.")
		return
	}

	type args struct {
		req *dbpb.DBFindParamsByClusterIdRequest
		rsp *dbpb.DBFindParamsByClusterIdResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, rsp *dbpb.DBFindParamsByClusterIdResponse) bool
	}{
		{
			"normal",
			args{
				req: &dbpb.DBFindParamsByClusterIdRequest{
					ClusterId: cluster.ID,
					Page: &dbpb.DBParamsPageDTO{
						Page:     1,
						PageSize: 10,
						Total:    0,
					},
				},
				rsp: &dbpb.DBFindParamsByClusterIdResponse{
					ParamGroupId: 0,
					Params:       nil,
					Page:         nil,
					Status:       nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBFindParamsByClusterIdResponse) bool{
				func(a args, rsp *dbpb.DBFindParamsByClusterIdResponse) bool { return rsp.ParamGroupId > 0 },
				func(a args, rsp *dbpb.DBFindParamsByClusterIdResponse) bool { return rsp.Status.Code == 0 },
				func(a args, rsp *dbpb.DBFindParamsByClusterIdResponse) bool { return len(rsp.Params) > 0 },
				func(a args, rsp *dbpb.DBFindParamsByClusterIdResponse) bool { return rsp.Page.Total > 0 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ApplyParamGroup(context.TODO(), &dbpb.DBApplyParamGroupRequest{
				ParamGroupId: int64(groups[0].ID),
				ClusterId:    cluster.ID,
				Params: []*dbpb.DBApplyParamDTO{
					{
						ParamId: int64(params[0].ID),
						RealValue: &dbpb.DBParamRealValueDTO{
							Cluster: "123",
							Instances: []*dbpb.DBParamInstanceRealValueDTO{
								{
									Instance: "172.16.1.1",
									Value:    "100",
								},
							},
						},
					},
					{
						ParamId: int64(params[1].ID),
						RealValue: &dbpb.DBParamRealValueDTO{
							Cluster:   "1024",
							Instances: nil,
						},
					},
				},
			}, &dbpb.DBApplyParamGroupResponse{
				ParamGroupId: 0,
				ClusterId:    "",
				Status:       nil,
			})
			if err != nil {
				t.Errorf("FindParamsByClusterId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = handler.FindParamsByClusterId(context.TODO(), tt.args.req, tt.args.rsp)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("FindParamsByClusterId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.rsp) {
					t.Errorf("FindParamsByClusterId() test error, testname = %v, assert %v, args = %v, got rsp = %v", tt.name, i, tt.args, tt.args.rsp)
				}
			}
		})
	}
}

func TestDBServiceHandler_UpdateClusterParams(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	params, err := buildParams(2)
	if err != nil {
		t.Errorf("TestDBServiceHandler_UpdateClusterParams() test error, build params err.")
		return
	}
	groups, err := buildParamGroup(1, params)
	if err != nil {
		t.Errorf("TestDBServiceHandler_UpdateClusterParams() test error, build param group err.")
		return
	}
	cluster, err := buildCluster()
	if err != nil {
		t.Errorf("TestDBServiceHandler_UpdateClusterParams() test error, build cluster err.")
		return
	}

	type args struct {
		req *dbpb.DBUpdateClusterParamsRequest
		rsp *dbpb.DBUpdateClusterParamsResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, rsp *dbpb.DBUpdateClusterParamsResponse) bool
	}{
		{
			"normal",
			args{
				req: &dbpb.DBUpdateClusterParamsRequest{
					ClusterId: cluster.ID,
					Params: []*dbpb.DBApplyParamDTO{
						{
							ParamId: int64(params[0].ID),
							RealValue: &dbpb.DBParamRealValueDTO{
								Cluster: "123",
								Instances: []*dbpb.DBParamInstanceRealValueDTO{
									{
										Instance: "172.16.2.1",
										Value:    "200",
									},
								},
							},
						},
					},
				},
				rsp: &dbpb.DBUpdateClusterParamsResponse{
					ClusterId: "",
					Status:    nil,
				},
			},
			false,
			[]func(a args, rsp *dbpb.DBUpdateClusterParamsResponse) bool{
				func(a args, rsp *dbpb.DBUpdateClusterParamsResponse) bool { return rsp.Status.Code == 0 },
				func(a args, rsp *dbpb.DBUpdateClusterParamsResponse) bool { return rsp.ClusterId != "" },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ApplyParamGroup(context.TODO(), &dbpb.DBApplyParamGroupRequest{
				ParamGroupId: int64(groups[0].ID),
				ClusterId:    cluster.ID,
				Params: []*dbpb.DBApplyParamDTO{
					{
						ParamId: int64(params[0].ID),
						RealValue: &dbpb.DBParamRealValueDTO{
							Cluster: "10",
							Instances: []*dbpb.DBParamInstanceRealValueDTO{
								{
									Instance: "172.16.1.1",
									Value:    "100",
								},
							},
						},
					},
					{
						ParamId: int64(params[1].ID),
						RealValue: &dbpb.DBParamRealValueDTO{
							Cluster:   "1024",
							Instances: nil,
						},
					},
				},
			}, &dbpb.DBApplyParamGroupResponse{
				ParamGroupId: 0,
				ClusterId:    "",
				Status:       nil,
			})
			if err != nil {
				t.Errorf("UpdateClusterParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = handler.UpdateClusterParams(context.TODO(), tt.args.req, tt.args.rsp)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("UpdateClusterParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, tt.args.rsp) {
					t.Errorf("UpdateClusterParams() test error, testname = %v, assert %v, args = %v, got rsp = %v", tt.name, i, tt.args, tt.args.rsp)
				}
			}
		})
	}
}

func buildParamGroup(count uint, params []*models.ParamDO) (pgs []*models.ParamGroupDO, err error) {
	pgs = make([]*models.ParamGroupDO, count)
	for i := range pgs {
		pgs[i] = &models.ParamGroupDO{
			Name:       "test_param_group_" + uuidutil.GenerateID(),
			Spec:       "8C16G",
			HasDefault: 1,
			DbType:     1,
			GroupType:  1,
			Version:    "5.0",
			Note:       "test param group " + strconv.Itoa(i),
		}
		pgm := make([]*models.ParamGroupMapDO, len(params))
		for j := range params {
			pgm[j] = &models.ParamGroupMapDO{
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

func buildParams(count uint) (params []*models.ParamDO, err error) {
	params = make([]*models.ParamDO, count)
	for i := range params {
		params[i] = &models.ParamDO{
			Name:          "test_param_" + uuidutil.GenerateID(),
			ComponentType: "TiKV",
			Type:          0,
			Unit:          "kb",
			Range:         "[\"0\", \"10\"]",
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

func buildCluster() (*models.Cluster, error) {
	cluster, err := Dao.ClusterManager().CreateCluster(context.TODO(), models.Cluster{
		Entity: models.Entity{
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
