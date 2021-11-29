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
 * @Description: param test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/26 13:01
*******************************************************************************/

package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"

	"github.com/asim/go-micro/v3/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"

	rpc_client "github.com/pingcap-inc/tiem/library/client"

	"github.com/golang/mock/gomock"

	"github.com/pingcap-inc/tiem/test/mockdb"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
)

func TestCreateParamGroup_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBCreateParamGroupRequest, opts ...client.CallOption) (*dbpb.DBCreateParamGroupResponse, error) {
			rsp := new(dbpb.DBCreateParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroupId = 1
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.CreateParamGroupRequest{
		Name:       "test_param_group",
		DbType:     1,
		HasDefault: 1,
		Version:    "5.0",
		Spec:       "8C16G",
		GroupType:  1,
		Note:       "test param group",
		Params: []*clusterpb.SubmitParamDTO{
			{
				ParamId:      1,
				DefaultValue: "10",
				Note:         "param1",
			},
		},
	}
	resp := new(clusterpb.CreateParamGroupResponse)
	err := CreateParamGroup(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("create param group %v failed, err: %v\n", req.Name, err)
	}
	if resp.RespStatus.Code != int32(codes.OK) || resp.RespStatus.Message != "" || resp.ParamGroupId != 1 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, paramGroupId: %v\n", resp.RespStatus.Code, resp.RespStatus.Message, resp.ParamGroupId)
	}
}

func TestCreateParamGroup_WithErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().CreateParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBCreateParamGroupRequest, opts ...client.CallOption) (*dbpb.DBCreateParamGroupResponse, error) {
			rsp := new(dbpb.DBCreateParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			return rsp, errors.New("param group id is 0")
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.CreateParamGroupRequest{}
	resp := new(clusterpb.CreateParamGroupResponse)
	err := CreateParamGroup(context.TODO(), req, resp)
	assert.NotNil(t, err)
}

func TestUpdateParamGroup_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBUpdateParamGroupRequest, opts ...client.CallOption) (*dbpb.DBUpdateParamGroupResponse, error) {
			rsp := new(dbpb.DBUpdateParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroupId = 1
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.UpdateParamGroupRequest{
		ParamGroupId: 1,
		Name:         "update_param_group",
		Version:      "4.0.2",
		Spec:         "16C32G",
		Note:         "update param group",
		Params: []*clusterpb.SubmitParamDTO{
			{
				ParamId:      1,
				DefaultValue: "1024",
				Note:         "param2",
			},
		},
	}
	resp := new(clusterpb.UpdateParamGroupResponse)
	err := UpdateParamGroup(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("update param group %s failed, err: %v\n", req.Name, err)
	}
	if resp.RespStatus.Code != int32(codes.OK) || resp.RespStatus.Message != "" || resp.ParamGroupId != 1 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, paramGroupId: %v\n", resp.RespStatus.Code, resp.RespStatus.Message, resp.ParamGroupId)
	}
}

func TestUpdateParamGroup_WithErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBUpdateParamGroupRequest, opts ...client.CallOption) (*dbpb.DBUpdateParamGroupResponse, error) {
			rsp := new(dbpb.DBUpdateParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			return rsp, errors.New("update param group error")
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.UpdateParamGroupRequest{
		ParamGroupId: 1,
		Name:         "update_param_group",
		Version:      "4.0.2",
		Spec:         "16C32G",
		Note:         "update param group",
		Params: []*clusterpb.SubmitParamDTO{
			{
				ParamId:      1,
				DefaultValue: "1024",
				Note:         "param2",
			},
		},
	}
	resp := new(clusterpb.UpdateParamGroupResponse)
	err := UpdateParamGroup(context.TODO(), req, resp)
	assert.NotNil(t, err)
}

func TestDeleteParamGroup_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamGroupByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamGroupByIDRequest, opts ...client.CallOption) (*dbpb.DBFindParamGroupByIDResponse, error) {
			rsp := new(dbpb.DBFindParamGroupByIDResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroup = buildDBParamGroupDTO()[1]
			return rsp, nil
		})
	mockClient.EXPECT().DeleteParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBDeleteParamGroupRequest, opts ...client.CallOption) (*dbpb.DBDeleteParamGroupResponse, error) {
			rsp := new(dbpb.DBDeleteParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroupId = 2
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.DeleteParamGroupRequest{
		ParamGroupId: 2,
	}
	resp := new(clusterpb.DeleteParamGroupResponse)
	err := DeleteParamGroup(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("delete param group %v failed, err: %v\n", req.ParamGroupId, err)
	}
	if resp.RespStatus.Code != int32(codes.OK) || resp.RespStatus.Message != "" || resp.ParamGroupId <= 0 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, paramGroupId: %v\n", resp.RespStatus.Code, resp.RespStatus.Message, resp.ParamGroupId)
	}
}

func TestDeleteParamGroup_WithErr_1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamGroupByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamGroupByIDRequest, opts ...client.CallOption) (*dbpb.DBFindParamGroupByIDResponse, error) {
			rsp := new(dbpb.DBFindParamGroupByIDResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroup = buildDBParamGroupDTO()[0]
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.DeleteParamGroupRequest{
		ParamGroupId: 1,
	}
	resp := new(clusterpb.DeleteParamGroupResponse)
	err := DeleteParamGroup(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("delete param group %v failed, err: %v\n", req.ParamGroupId, err)
	}
	assert.NotEqual(t, resp.RespStatus.Code, 0)
}

func TestDeleteParamGroup_WithErr_2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamGroupByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamGroupByIDRequest, opts ...client.CallOption) (*dbpb.DBFindParamGroupByIDResponse, error) {
			rsp := new(dbpb.DBFindParamGroupByIDResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroup = buildDBParamGroupDTO()[1]
			return rsp, nil
		})
	mockClient.EXPECT().DeleteParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBDeleteParamGroupRequest, opts ...client.CallOption) (*dbpb.DBDeleteParamGroupResponse, error) {
			rsp := new(dbpb.DBDeleteParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroupId = 2
			return rsp, errors.New("delete param group fail")
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.DeleteParamGroupRequest{
		ParamGroupId: 2,
	}
	resp := new(clusterpb.DeleteParamGroupResponse)
	err := DeleteParamGroup(context.TODO(), req, resp)
	assert.NotNil(t, err)
}

func TestListParamGroup_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().ListParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBListParamGroupRequest, opts ...client.CallOption) (*dbpb.DBListParamGroupResponse, error) {
			rsp := new(dbpb.DBListParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.Page = &dbpb.DBParamsPageDTO{Page: 1, PageSize: 10, Total: 1}
			rsp.ParamGroups = buildDBParamGroupDTO()
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.ListParamGroupRequest{
		Name:      "group",
		DbType:    1,
		HasDetail: true,
		Page:      &clusterpb.PageDTO{Page: 1, PageSize: 10},
	}
	resp := new(clusterpb.ListParamGroupResponse)
	err := ListParamGroup(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("list param group %v failed, err: %v\n", req.Name, err)
	}
	if resp.RespStatus.Code != int32(codes.OK) || resp.RespStatus.Message != "" || len(resp.ParamGroups) <= 0 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, list param len: %v\n", resp.RespStatus.Code, resp.RespStatus.Message, len(resp.ParamGroups))
	}
}

func TestListParamGroup_WithErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().ListParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBListParamGroupRequest, opts ...client.CallOption) (*dbpb.DBListParamGroupResponse, error) {
			rsp := new(dbpb.DBListParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.Page = &dbpb.DBParamsPageDTO{Page: 1, PageSize: 10, Total: 1}
			rsp.ParamGroups = buildDBParamGroupDTO()
			return rsp, errors.New("list param group fail")
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.ListParamGroupRequest{
		Name:      "group",
		DbType:    1,
		HasDetail: true,
		Page:      &clusterpb.PageDTO{Page: 1, PageSize: 10},
	}
	resp := new(clusterpb.ListParamGroupResponse)
	err := ListParamGroup(context.TODO(), req, resp)
	assert.NotNil(t, err)
}

func TestDetailParamGroup_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamGroupByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamGroupByIDRequest, opts ...client.CallOption) (*dbpb.DBFindParamGroupByIDResponse, error) {
			rsp := new(dbpb.DBFindParamGroupByIDResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroup = buildDBParamGroupDTO()[0]
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.DetailParamGroupRequest{ParamGroupId: 1}
	resp := new(clusterpb.DetailParamGroupResponse)
	err := DetailParamGroup(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("detail param group %v failed, err: %v\n", req.GetParamGroupId(), err)
	}
	if resp.RespStatus.Code != int32(codes.OK) || resp.RespStatus.Message != "" || len(resp.ParamGroup.Params) <= 0 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, list params len: %v\n", resp.RespStatus.Code, resp.RespStatus.Message, len(resp.ParamGroup.Params))
	}
}

func TestDetailParamGroup_WithErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamGroupByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamGroupByIDRequest, opts ...client.CallOption) (*dbpb.DBFindParamGroupByIDResponse, error) {
			rsp := new(dbpb.DBFindParamGroupByIDResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroup = buildDBParamGroupDTO()[0]
			return rsp, errors.New("detail param group fail")
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.DetailParamGroupRequest{ParamGroupId: 1}
	resp := new(clusterpb.DetailParamGroupResponse)
	err := DetailParamGroup(context.TODO(), req, resp)
	assert.NotNil(t, err)
}

func TestCopyParamGroup_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	paramGroupId := 1
	name := "copy_group_name"
	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamGroupByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamGroupByIDRequest, opts ...client.CallOption) (*dbpb.DBFindParamGroupByIDResponse, error) {
			rsp := new(dbpb.DBFindParamGroupByIDResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroup = buildDBParamGroupDTO()[0]
			return rsp, nil
		})

	mockClient.EXPECT().CreateParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBCreateParamGroupRequest, opts ...client.CallOption) (*dbpb.DBCreateParamGroupResponse, error) {
			rsp := new(dbpb.DBCreateParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroupId = int64(paramGroupId)
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.CopyParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
		Name:         name,
		Note:         "copy group name",
	}
	resp := new(clusterpb.CopyParamGroupResponse)
	err := CopyParamGroup(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("detail param group %v failed, err: %v\n", req.GetParamGroupId(), err)
	}
	if resp.RespStatus.Code != int32(codes.OK) || resp.RespStatus.Message != "" || resp.ParamGroupId <= 0 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, params group id: %v\n", resp.RespStatus.Code, resp.RespStatus.Message, resp.ParamGroupId)
	}
}

func TestCopyParamGroup_WithErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	paramGroupId := 1
	name := "copy_group_name"
	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamGroupByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamGroupByIDRequest, opts ...client.CallOption) (*dbpb.DBFindParamGroupByIDResponse, error) {
			rsp := new(dbpb.DBFindParamGroupByIDResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroup = buildDBParamGroupDTO()[0]
			return rsp, nil
		})

	mockClient.EXPECT().CreateParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBCreateParamGroupRequest, opts ...client.CallOption) (*dbpb.DBCreateParamGroupResponse, error) {
			rsp := new(dbpb.DBCreateParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroupId = int64(paramGroupId)
			return rsp, errors.New("copy param group fail")
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.CopyParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
		Name:         name,
		Note:         "copy group name",
	}
	resp := new(clusterpb.CopyParamGroupResponse)
	err := CopyParamGroup(context.TODO(), req, resp)
	assert.NotNil(t, err)
}

func TestApplyParamGroup_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	paramGroupId := 1
	clusterId := "123"
	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamGroupByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamGroupByIDRequest, opts ...client.CallOption) (*dbpb.DBFindParamGroupByIDResponse, error) {
			rsp := new(dbpb.DBFindParamGroupByIDResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroup = buildDBParamGroupDTO()[0]
			return rsp, nil
		})

	mockClient.EXPECT().ApplyParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBApplyParamGroupRequest, opts ...client.CallOption) (*dbpb.DBApplyParamGroupResponse, error) {
			rsp := new(dbpb.DBApplyParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroupId = int64(paramGroupId)
			rsp.ClusterId = clusterId
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.ApplyParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
		ClusterId:    clusterId,
	}
	resp := new(clusterpb.ApplyParamGroupResponse)
	err := ApplyParamGroup(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("apply param group %v failed, err: %v\n", req.GetParamGroupId(), err)
	}
	if resp.RespStatus.Code != int32(codes.OK) || resp.RespStatus.Message != "" || resp.ParamGroupId <= 0 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, params group id: %v\n", resp.RespStatus.Code, resp.RespStatus.Message, resp.ParamGroupId)
	}
}

func TestApplyParamGroup_WithErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	paramGroupId := 1
	clusterId := "123"
	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamGroupByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamGroupByIDRequest, opts ...client.CallOption) (*dbpb.DBFindParamGroupByIDResponse, error) {
			rsp := new(dbpb.DBFindParamGroupByIDResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroup = buildDBParamGroupDTO()[0]
			return rsp, nil
		})

	mockClient.EXPECT().ApplyParamGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBApplyParamGroupRequest, opts ...client.CallOption) (*dbpb.DBApplyParamGroupResponse, error) {
			rsp := new(dbpb.DBApplyParamGroupResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ParamGroupId = int64(paramGroupId)
			rsp.ClusterId = clusterId
			return rsp, errors.New("apply param group fail")
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.ApplyParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
		ClusterId:    clusterId,
	}
	resp := new(clusterpb.ApplyParamGroupResponse)
	err := ApplyParamGroup(context.TODO(), req, resp)
	assert.NotNil(t, err)
}

func TestListClusterParams_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamsByClusterId(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamsByClusterIdRequest, opts ...client.CallOption) (*dbpb.DBFindParamsByClusterIdResponse, error) {
			rsp := new(dbpb.DBFindParamsByClusterIdResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.Page = &dbpb.DBParamsPageDTO{Page: 1, PageSize: 10, Total: 1}
			rsp.ParamGroupId = 1
			rsp.Params = []*dbpb.DBClusterParamDTO{
				{
					ParamId:       1,
					Name:          "param1",
					ComponentType: "TiDB",
					Type:          0,
					Unit:          "mb",
					Range:         []string{"0", "1024"},
					HasReboot:     1,
					Source:        2,
					DefaultValue:  "256",
					Description:   "param 1",
					Note:          "param 1",
					CreateTime:    time.Now().Unix(),
					UpdateTime:    time.Now().Unix(),
					RealValue: &dbpb.DBParamRealValueDTO{
						Cluster:   "123",
						Instances: nil,
					},
				},
			}
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.ListClusterParamsRequest{
		ClusterId: "123",
		Page:      &clusterpb.PageDTO{Page: 1, PageSize: 10},
		Operator:  nil,
	}

	resp := new(clusterpb.ListClusterParamsResponse)
	err := ListClusterParams(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("list cluster %v params failed, err: %v\n", req.ClusterId, err)
	}
	if resp.RespStatus.Code != int32(codes.OK) || resp.RespStatus.Message != "" || resp.ParamGroupId <= 0 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, params group id: %v\n", resp.RespStatus.Code, resp.RespStatus.Message, resp.ParamGroupId)
	}
}

func TestListClusterParams_WithErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().FindParamsByClusterId(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBFindParamsByClusterIdRequest, opts ...client.CallOption) (*dbpb.DBFindParamsByClusterIdResponse, error) {
			rsp := new(dbpb.DBFindParamsByClusterIdResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.Page = &dbpb.DBParamsPageDTO{Page: 1, PageSize: 10, Total: 1}
			rsp.ParamGroupId = 1
			rsp.Params = []*dbpb.DBClusterParamDTO{
				{
					ParamId:       1,
					Name:          "param1",
					ComponentType: "TiDB",
					Type:          0,
					Unit:          "mb",
					Range:         []string{"0", "1024"},
					HasReboot:     1,
					Source:        2,
					DefaultValue:  "256",
					Description:   "param 1",
					Note:          "param 1",
					CreateTime:    time.Now().Unix(),
					UpdateTime:    time.Now().Unix(),
					RealValue: &dbpb.DBParamRealValueDTO{
						Cluster:   "123",
						Instances: nil,
					},
				},
			}
			return rsp, errors.New("list cluster params fail")
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.ListClusterParamsRequest{
		ClusterId: "123",
		Page:      &clusterpb.PageDTO{Page: 1, PageSize: 10},
		Operator:  nil,
	}

	resp := new(clusterpb.ListClusterParamsResponse)
	err := ListClusterParams(context.TODO(), req, resp)
	assert.NotNil(t, err)
}

func TestUpdateClusterParams_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateClusterParams(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBUpdateClusterParamsRequest, opts ...client.CallOption) (*dbpb.DBUpdateClusterParamsResponse, error) {
			rsp := new(dbpb.DBUpdateClusterParamsResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ClusterId = "123"
			return rsp, nil
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.UpdateClusterParamsRequest{
		ClusterId: "123",
		Params: []*clusterpb.UpdateClusterParamDTO{
			{
				ParamId: 1,
				RealValue: &clusterpb.ParamRealValueDTO{
					Cluster:   "1024",
					Instances: nil,
				},
			},
		},
		Operator: nil,
	}

	resp := new(clusterpb.UpdateClusterParamsResponse)
	err := UpdateClusterParams(context.TODO(), req, resp)
	if err != nil {
		t.Errorf("update cluster %v params failed, err: %v\n", req.ClusterId, err)
	}
	if resp.RespStatus.Code != int32(codes.OK) || resp.RespStatus.Message != "" || resp.ClusterId == "" {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, cluster id: %v\n", resp.RespStatus.Code, resp.RespStatus.Message, resp.ClusterId)
	}
}

func TestUpdateClusterParams_WithErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().UpdateClusterParams(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, in *dbpb.DBUpdateClusterParamsRequest, opts ...client.CallOption) (*dbpb.DBUpdateClusterParamsResponse, error) {
			rsp := new(dbpb.DBUpdateClusterParamsResponse)
			rsp.Status = new(dbpb.DBParamResponseStatus)
			rsp.ClusterId = "123"
			return rsp, errors.New("update cluster params fail")
		})
	rpc_client.DBClient = mockClient

	req := &clusterpb.UpdateClusterParamsRequest{
		ClusterId: "123",
		Params: []*clusterpb.UpdateClusterParamDTO{
			{
				ParamId: 1,
				RealValue: &clusterpb.ParamRealValueDTO{
					Cluster:   "1024",
					Instances: nil,
				},
			},
		},
		Operator: nil,
	}

	resp := new(clusterpb.UpdateClusterParamsResponse)
	err := UpdateClusterParams(context.TODO(), req, resp)
	assert.NotNil(t, err)
}

func buildDBParamGroupDTO() []*dbpb.DBParamGroupDTO {
	return []*dbpb.DBParamGroupDTO{
		{
			ParamGroupId: 1,
			Name:         "param_group_1",
			DbType:       1,
			HasDefault:   1,
			Version:      "5.0",
			Spec:         "8C16G",
			GroupType:    1,
			Note:         "default param group 1",
			CreateTime:   time.Now().Unix(),
			UpdateTime:   time.Now().Unix(),
			Params: []*dbpb.DBParamDTO{
				{
					ParamId:       1,
					Name:          "param1",
					ComponentType: "TiDB",
					Type:          0,
					Unit:          "mb",
					Range:         []string{"0", "1024"},
					HasReboot:     1,
					Source:        2,
					DefaultValue:  "256",
					Description:   "param 1",
					Note:          "param 1",
					CreateTime:    time.Now().Unix(),
					UpdateTime:    time.Now().Unix(),
				},
			},
		},
		{
			ParamGroupId: 2,
			Name:         "param_group_2",
			DbType:       1,
			HasDefault:   2,
			Version:      "4.0.2",
			Spec:         "16C16G",
			GroupType:    1,
			Note:         "param group 2",
			CreateTime:   time.Now().Unix(),
			UpdateTime:   time.Now().Unix(),
			Params: []*dbpb.DBParamDTO{
				{
					ParamId:       1,
					Name:          "param1",
					ComponentType: "TiDB",
					Type:          0,
					Unit:          "mb",
					Range:         []string{"0", "1024"},
					HasReboot:     1,
					Source:        2,
					DefaultValue:  "256",
					Description:   "param 1",
					Note:          "param 1",
					CreateTime:    time.Now().Unix(),
					UpdateTime:    time.Now().Unix(),
				},
				{
					ParamId:       2,
					Name:          "param2",
					ComponentType: "TiKV",
					Type:          1,
					Unit:          "",
					Range:         []string{"true", "false"},
					HasReboot:     1,
					Source:        2,
					DefaultValue:  "false",
					Description:   "param 2",
					Note:          "param 2",
					CreateTime:    time.Now().Unix(),
					UpdateTime:    time.Now().Unix(),
				},
			},
		},
	}
}
