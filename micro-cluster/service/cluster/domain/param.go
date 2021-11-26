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
 * @File: param_group
 * @Description: param group and cluster param management service
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/24 14:11
*******************************************************************************/

package domain

import (
	"context"
	"encoding/json"

	"github.com/pingcap-inc/tiem/library/client"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
)

const (
	defaultParamGroup = 1
	customParamGroup  = 2
)

func CreateParamGroup(ctx context.Context, req *clusterpb.CreateParamGroupRequest, resp *clusterpb.CreateParamGroupResponse) error {
	dbReq := dbpb.DBCreateParamGroupRequest{}
	err := convertObj(req, &dbReq)
	if err != nil {
		return err
	}

	dbRsp, err := client.DBClient.CreateParamGroup(ctx, &dbReq)
	if err != nil {
		return err
	}
	resp.RespStatus = convertRespStatus(dbRsp.Status)
	resp.ParamGroupId = dbRsp.ParamGroupId
	return nil
}

func UpdateParamGroup(ctx context.Context, req *clusterpb.UpdateParamGroupRequest, resp *clusterpb.UpdateParamGroupResponse) error {
	dbReq := dbpb.DBUpdateParamGroupRequest{}
	err := convertObj(req, &dbReq)
	if err != nil {
		return err
	}

	dbRsp, err := client.DBClient.UpdateParamGroup(ctx, &dbReq)
	if err != nil {
		return err
	}
	resp.RespStatus = convertRespStatus(dbRsp.Status)
	resp.ParamGroupId = dbRsp.ParamGroupId
	return nil
}

func DeleteParamGroup(ctx context.Context, req *clusterpb.DeleteParamGroupRequest, resp *clusterpb.DeleteParamGroupResponse) error {
	group, err := client.DBClient.FindParamGroupByID(ctx, &dbpb.DBFindParamGroupByIDRequest{ParamGroupId: req.ParamGroupId})
	if err != nil {
		return err
	}
	if group.ParamGroup.HasDefault == defaultParamGroup {
		resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: 533, Message: "The default param group cannot be deleted."}
		return nil
	}
	dbRsp, err := client.DBClient.DeleteParamGroup(ctx, &dbpb.DBDeleteParamGroupRequest{ParamGroupId: req.ParamGroupId})
	if err != nil {
		return err
	}
	resp.RespStatus = convertRespStatus(dbRsp.Status)
	resp.ParamGroupId = dbRsp.ParamGroupId
	return nil
}

func ListParamGroup(ctx context.Context, req *clusterpb.ListParamGroupRequest, resp *clusterpb.ListParamGroupResponse) error {
	var dbReq dbpb.DBListParamGroupRequest
	err := convertObj(req, &dbReq)
	if err != nil {
		return err
	}

	dbRsp, err := client.DBClient.ListParamGroup(ctx, &dbReq)
	if err != nil {
		return err
	}

	resp.Page = convertPage(dbRsp.Page)
	resp.RespStatus = convertRespStatus(dbRsp.Status)
	if dbRsp.ParamGroups != nil {
		pgs := make([]*clusterpb.ParamGroupDTO, len(dbRsp.ParamGroups))
		err := convertObj(dbRsp.ParamGroups, &pgs)
		if err != nil {
			return err
		}
		resp.ParamGroups = pgs
	}
	return nil
}

func DetailParamGroup(ctx context.Context, req *clusterpb.DetailParamGroupRequest, resp *clusterpb.DetailParamGroupResponse) error {
	dbRsp, err := client.DBClient.FindParamGroupByID(ctx, &dbpb.DBFindParamGroupByIDRequest{ParamGroupId: req.ParamGroupId})
	if err != nil {
		return err
	}
	resp.RespStatus = convertRespStatus(dbRsp.Status)
	if dbRsp.ParamGroup != nil {
		pg := clusterpb.ParamGroupDTO{}
		err = convertObj(dbRsp.ParamGroup, &pg)
		if err != nil {
			return err
		}
		resp.ParamGroup = &pg
	}
	return nil
}

func ApplyParamGroup(ctx context.Context, req *clusterpb.ApplyParamGroupRequest, resp *clusterpb.ApplyParamGroupResponse) error {
	// query param group by id
	group, err := client.DBClient.FindParamGroupByID(ctx, &dbpb.DBFindParamGroupByIDRequest{ParamGroupId: req.ParamGroupId})
	if err != nil {
		return err
	}
	params := make([]*dbpb.DBApplyParamDTO, len(group.ParamGroup.Params))
	for i, param := range group.ParamGroup.Params {
		params[i] = &dbpb.DBApplyParamDTO{
			ParamId:   param.ParamId,
			RealValue: &dbpb.DBParamRealValueDTO{Cluster: param.DefaultValue},
		}
	}

	dbRsp, err := client.DBClient.ApplyParamGroup(ctx, &dbpb.DBApplyParamGroupRequest{
		ParamGroupId: req.ParamGroupId,
		ClusterId:    req.ClusterId,
		Params:       params,
	})
	if err != nil {
		return err
	}
	resp.RespStatus = convertRespStatus(dbRsp.Status)
	resp.ParamGroupId = dbRsp.ParamGroupId
	resp.ClusterId = dbRsp.ClusterId
	return nil

}

func CopyParamGroup(ctx context.Context, req *clusterpb.CopyParamGroupRequest, resp *clusterpb.CopyParamGroupResponse) error {
	// query param group by id
	group, err := client.DBClient.FindParamGroupByID(ctx, &dbpb.DBFindParamGroupByIDRequest{ParamGroupId: req.ParamGroupId})
	if err != nil {
		return err
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
		HasDefault: customParamGroup,
		Version:    group.ParamGroup.Version,
		Spec:       group.ParamGroup.Spec,
		GroupType:  group.ParamGroup.GroupType,
		ParentId:   group.ParamGroup.ParamGroupId,
		Params:     params,
	})
	if err != nil {
		return err
	}
	resp.RespStatus = convertRespStatus(dbRsp.Status)
	resp.ParamGroupId = dbRsp.ParamGroupId
	return nil
}

func ListClusterParams(ctx context.Context, req *clusterpb.ListClusterParamsRequest, resp *clusterpb.ListClusterParamsResponse) error {
	var dbReq *dbpb.DBFindParamsByClusterIdRequest
	err := convertObj(req, &dbReq)
	if err != nil {
		return err
	}

	dbRsp, err := client.DBClient.FindParamsByClusterId(ctx, dbReq)
	if err != nil {
		return err
	}

	resp.Page = convertPage(dbRsp.Page)
	resp.RespStatus = convertRespStatus(dbRsp.Status)
	if dbRsp.Params != nil {
		ps := make([]*clusterpb.ClusterParamDTO, len(dbRsp.Params))
		err := convertObj(dbRsp.Params, &ps)
		if err != nil {
			return err
		}
		resp.Params = ps
	}
	return nil
}

func UpdateClusterParams(ctx context.Context, req *clusterpb.UpdateClusterParamsRequest, resp *clusterpb.UpdateClusterParamsResponse) error {
	var dbReq *dbpb.DBUpdateClusterParamsRequest
	err := convertObj(req, &dbReq)
	if err != nil {
		return err
	}

	dbRsp, err := client.DBClient.UpdateClusterParams(ctx, dbReq)
	if err != nil {
		return err
	}
	resp.RespStatus = &clusterpb.ResponseStatusDTO{Code: dbRsp.Status.Code, Message: dbRsp.Status.Message}
	resp.ClusterId = dbRsp.ClusterId
	return nil
}

func InspectClusterParams(ctx context.Context, req *clusterpb.InspectClusterParamsRequest, resp *clusterpb.InspectClusterParamsResponse) error {
	// todo: Reliance on parameter source update implementation
	return nil
}

func convertRespStatus(status *dbpb.DBParamResponseStatus) *clusterpb.ResponseStatusDTO {
	return &clusterpb.ResponseStatusDTO{Code: status.Code, Message: status.Message}
}

func convertPage(page *dbpb.DBParamsPageDTO) *clusterpb.PageDTO {
	return &clusterpb.PageDTO{Page: page.Page, PageSize: page.PageSize, Total: page.Total}
}

func convertObj(src interface{}, dst interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, dst)
	if err != nil {
		return err
	}
	return nil
}
