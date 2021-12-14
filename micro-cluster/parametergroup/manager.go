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
 * @File: manager.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/9 15:27
*******************************************************************************/

package parametergroup

import (
	"context"
	"strconv"

	"github.com/pingcap-inc/tiem/message"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/convert"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

type ParamGroupType int32

const (
	DEFAULT ParamGroupType = 1
	CUSTOM  ParamGroupType = 2
)

type ParamSource int32

const (
	TiUP ParamSource = iota
	SQL
	TiupAndSql
	API
)

type ParamValueType int32

const (
	Integer ParamValueType = iota
	String
	Boolean
	Float
)

type ModifyParam struct {
	Reboot bool
	Params []*ApplyParam
}

type ApplyParam struct {
	ParamId       int64
	Name          string
	ComponentType string
	HasReboot     int32
	Source        int32
	Type          int32
	RealValue     clusterpb.ParamRealValueDTO
}

func (m *Manager) CreateParameterGroup(ctx context.Context, req message.CreateParameterGroupReq) (resp message.CreateParameterGroupResp, err error) {
	dbReq := dbpb.DBCreateParamGroupRequest{}
	err = convert.ConvertObj(req, &dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create param group req: %v, err: %v", req, err)
		return
	}
	dbRsp, err := client.DBClient.CreateParamGroup(ctx, &dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create param group req: %v, err: %v", req, err)
		err = framework.WrapError(common.TIEM_PARAMETER_GROUP_CREATE_ERROR, "failed to create parameter group", err)
		return
	}
	resp = message.CreateParameterGroupResp{ParamGroupID: strconv.FormatInt(dbRsp.ParamGroupId, 10)}
	return resp, nil
}

func (m *Manager) UpdateParameterGroup(ctx context.Context, req message.UpdateParameterGroupReq) (resp message.UpdateParameterGroupResp, err error) {
	dbReq := dbpb.DBUpdateParamGroupRequest{}
	err = convert.ConvertObj(req, &dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update param group req: %v, err: %v", req, err)
		return
	}
	dbRsp, err := client.DBClient.UpdateParamGroup(ctx, &dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update param group invoke metadb err: %v", err)
		err = framework.WrapError(common.TIEM_PARAMETER_GROUP_UPDATE_ERROR, "failed to update parameter group", err)
		return
	}
	resp = message.UpdateParameterGroupResp{ParamGroupID: strconv.FormatInt(dbRsp.ParamGroupId, 10)}
	return resp, nil
}

func (m *Manager) DeleteParameterGroup(ctx context.Context, req message.DeleteParameterGroupReq) (resp message.DeleteParameterGroupResp, err error) {
	paramGroupId, err := strconv.Atoi(req.ID)
	if err != nil {
		return resp, err
	}
	group, err := client.DBClient.FindParamGroupByID(ctx, &dbpb.DBFindParamGroupByIDRequest{ParamGroupId: int64(paramGroupId)})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("delete param group req: %v, err: %v", req, err)
		return
	}
	if group.ParamGroup.HasDefault == int32(DEFAULT) {
		err = framework.WrapError(common.TIEM_DEFAULT_PARAM_GROUP_NOT_DEL, common.TIEM_DEFAULT_PARAM_GROUP_NOT_DEL.Explain(), err)
		return
	}
	dbRsp, err := client.DBClient.DeleteParamGroup(ctx, &dbpb.DBDeleteParamGroupRequest{ParamGroupId: int64(paramGroupId)})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("delete param group invoke metadb err: %v", err)
		return
	}
	resp = message.DeleteParameterGroupResp{ParamGroupID: strconv.FormatInt(dbRsp.ParamGroupId, 10)}
	return resp, nil
}

func (m *Manager) QueryParameterGroup(ctx context.Context, req message.QueryParameterGroupReq) (resp []message.QueryParameterGroupResp, page clusterpb.RpcPage, err error) {
	var dbReq dbpb.DBListParamGroupRequest
	err = convert.ConvertObj(req, &dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("list param group req: %v, err: %v", req, err)
		return
	}

	_, err = client.DBClient.ListParamGroup(ctx, &dbReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("list param group invoke metadb err: %v", err)
		return
	}

	//resp.Page = convertPage(dbRsp.Page)
	//resp.RespStatus = convertRespStatus(dbRsp.Status)
	//if dbRsp.ParamGroups != nil {
	//	pgs := make([]*clusterpb.ParamGroupDTO, len(dbRsp.ParamGroups))
	//	err = convert.ConvertObj(dbRsp.ParamGroups, &pgs)
	//	if err != nil {
	//		framework.LogWithContext(ctx).Errorf("list param group convert resp err: %v", err)
	//		return err
	//	}
	//	resp.ParamGroups = pgs
	//}
	return resp, clusterpb.RpcPage{Page: 1, PageSize: 10, Total: 10}, nil
}

func (m *Manager) DetailParameterGroup(ctx context.Context, req message.DetailParameterGroupReq) (resp message.DetailParameterGroupResp, err error) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		return
	}
	_, err = client.DBClient.FindParamGroupByID(ctx, &dbpb.DBFindParamGroupByIDRequest{ParamGroupId: int64(id)})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("detail param group invoke metadb err: %v", err)
		return
	}
	//if dbRsp.ParamGroup != nil {
	//	pg := clusterpb.ParamGroupDTO{}
	//	err = convert.ConvertObj(dbRsp.ParamGroup, &pg)
	//	if err != nil {
	//		framework.LogWithContext(ctx).Errorf("detail param group convert resp err: %v", err)
	//		return err
	//	}
	//	resp.ParamGroup = &pg
	//}
	return resp, nil
}

func (m *Manager) ApplyParameterGroup(ctx context.Context, req message.ApplyParameterGroupReq) (resp message.ApplyParameterGroupResp, err error) {
	// query param group by id
	paramGroupId, err := strconv.Atoi(req.ID)
	if err != nil {
		return
	}
	group, err := client.DBClient.FindParamGroupByID(ctx, &dbpb.DBFindParamGroupByIDRequest{ParamGroupId: int64(paramGroupId)})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("apply param group invoke metadb err: %v", err)
		return
	}
	params := make([]*dbpb.DBApplyParamDTO, len(group.ParamGroup.Params))
	for i, param := range group.ParamGroup.Params {
		params[i] = &dbpb.DBApplyParamDTO{
			ParamId:   param.ParamId,
			RealValue: &dbpb.DBParamRealValueDTO{Cluster: param.DefaultValue},
		}
	}

	dbRsp, err := client.DBClient.ApplyParamGroup(ctx, &dbpb.DBApplyParamGroupRequest{
		ParamGroupId: int64(paramGroupId),
		ClusterId:    req.ClusterID,
		Params:       params,
	})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("apply param group convert resp err: %v", err)
		return
	}
	resp.ParamGroupId = strconv.FormatInt(dbRsp.ParamGroupId, 10)
	return resp, nil

}

func (m *Manager) CopyParameterGroup(ctx context.Context, req message.CopyParameterGroupReq) (resp message.CopyParameterGroupResp, err error) {
	// query param group by id
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		return
	}
	group, err := client.DBClient.FindParamGroupByID(ctx, &dbpb.DBFindParamGroupByIDRequest{ParamGroupId: int64(id)})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("copy param group invoke metadb err: %v", err)
		return
	}
	params := make([]*dbpb.DBSubmitParamDTO, len(group.ParamGroup.Params))
	for i, param := range group.ParamGroup.Params {
		params[i] = &dbpb.DBSubmitParamDTO{
			ParamId:      param.ParamId,
			DefaultValue: param.DefaultValue,
			Note:         param.Note,
		}
	}

	dbRsp, err := client.DBClient.CreateParamGroup(ctx, &dbpb.DBCreateParamGroupRequest{
		Name:       req.Name,
		Note:       req.Note,
		DbType:     group.ParamGroup.DbType,
		HasDefault: int32(CUSTOM),
		Version:    group.ParamGroup.Version,
		Spec:       group.ParamGroup.Spec,
		GroupType:  group.ParamGroup.GroupType,
		ParentId:   group.ParamGroup.ParamGroupId,
		Params:     params,
	})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("copy param group convert resp err: %v", err)
		return
	}
	resp.ParamGroupID = strconv.FormatInt(dbRsp.ParamGroupId, 10)
	return resp, nil
}

func convertRespStatus(status *dbpb.DBParamResponseStatus) *clusterpb.ResponseStatusDTO {
	return &clusterpb.ResponseStatusDTO{Code: status.Code, Message: status.Message}
}

func convertPage(page *dbpb.DBParamsPageDTO) *clusterpb.PageDTO {
	return &clusterpb.PageDTO{Page: page.Page, PageSize: page.PageSize, Total: page.Total}
}
