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
 * @File: param
 * @Description: param dao services
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/19 13:34
*******************************************************************************/

package service

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"github.com/pingcap-inc/tiem/micro-metadb/models"

	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
)

var ParamSuccessResponseStatus = &dbpb.DBParamResponseStatus{Code: 0}

func (handler *DBServiceHandler) CreateParamGroup(ctx context.Context, req *dbpb.DBCreateParamGroupRequest, rsp *dbpb.DBCreateParamGroupResponse) error {
	log := framework.LogWithContext(ctx)
	log.Debugln("start create param group dao service.")
	groupId, err := handler.Dao().ParamManager().AddParamGroup(ctx, parseParamGroup(req), parseSubmitParams(req.Params))
	if err != nil {
		log.Errorf("create param group err: %v", err.Error())
		st, ok := status.FromError(err)
		if ok {
			rsp.Status.Code = int32(st.Code())
			rsp.Status.Message = st.Message()
		} else {
			rsp.Status.Code = int32(codes.Internal)
			rsp.Status.Message = fmt.Sprintf("create failure param group failed, err: %v", err)
		}
	} else {
		rsp.ParamGroupId = int64(groupId)
		rsp.Status = ParamSuccessResponseStatus
	}
	return nil
}

func (handler *DBServiceHandler) UpdateParamGroup(ctx context.Context, req *dbpb.DBUpdateParamGroupRequest, rsp *dbpb.DBUpdateParamGroupResponse) error {
	log := framework.LogWithContext(ctx)
	log.Debugln("start update param group dao service.")
	err := handler.Dao().ParamManager().UpdateParamGroup(ctx, uint(req.ParamGroupId),
		req.Name, req.Spec, req.Version, req.Note, parseSubmitParams(req.Params))
	if err != nil {
		log.Errorf("update param group err: %v", err.Error())
		st, ok := status.FromError(err)
		if ok {
			rsp.Status.Code = int32(st.Code())
			rsp.Status.Message = st.Message()
		} else {
			rsp.Status.Code = int32(codes.Internal)
			rsp.Status.Message = fmt.Sprintf("update failure param group failed, err: %v", err)
		}
	} else {
		rsp.ParamGroupId = req.ParamGroupId
		rsp.Status = ParamSuccessResponseStatus
	}
	return nil
}

func (handler *DBServiceHandler) DeleteParamGroup(ctx context.Context, req *dbpb.DBDeleteParamGroupRequest, rsp *dbpb.DBDeleteParamGroupResponse) error {
	log := framework.LogWithContext(ctx)
	log.Debugln("start delete param group dao service.")
	err := handler.Dao().ParamManager().DeleteParamGroup(ctx, uint(req.GetParamGroupId()))
	if err != nil {
		log.Errorf("delete param group err: %v", err.Error())
		st, ok := status.FromError(err)
		if ok {
			rsp.Status.Code = int32(st.Code())
			rsp.Status.Message = st.Message()
		} else {
			rsp.Status.Code = int32(codes.Internal)
			rsp.Status.Message = fmt.Sprintf("delete failure param group failed, err: %v", err)
		}
	} else {
		rsp.ParamGroupId = req.ParamGroupId
		rsp.Status = ParamSuccessResponseStatus
	}
	return nil
}

func (handler *DBServiceHandler) ListParamGroup(ctx context.Context, req *dbpb.DBListParamGroupRequest, rsp *dbpb.DBListParamGroupResponse) error {
	log := framework.LogWithContext(ctx)
	log.Debugln("start list param group dao service.")
	groups, total, err := handler.Dao().ParamManager().ListParamGroup(ctx, req.Name, req.Spec, req.Version, req.DbType, req.HasDefault,
		int(req.Page.Page-1)*int(req.Page.PageSize), int(req.Page.PageSize))
	if err != nil {
		log.Errorf("list param group err: %v", err.Error())
		st, ok := status.FromError(err)
		if ok {
			rsp.Status.Code = int32(st.Code())
			rsp.Status.Message = st.Message()
		} else {
			rsp.Status.Code = int32(codes.Internal)
			rsp.Status.Message = fmt.Sprintf("list failure param group failed, err: %v", err)
		}
	} else {
		rsp.Status = ParamSuccessResponseStatus
		rsp.Page = &dbpb.DBParamsPageDTO{
			Page:     req.Page.Page,
			PageSize: req.Page.PageSize,
			Total:    int32(total),
		}
		rsp.ParamGroups = make([]*dbpb.DBParamGroupDTO, len(groups))
		for i, group := range groups {
			// condition load parameter details
			if req.HasDetail {
				params, err := handler.Dao().ParamManager().LoadParamsByGroupId(ctx, group.ID)
				if err != nil {
					log.Errorf("list param group by id err: %v", err.Error())
					rsp.Status.Code = int32(codes.Internal)
					rsp.Status.Message = fmt.Sprintf("list failure param group failed, err: %v", err)
					break
				}
				rsp.ParamGroups[i] = parseParamDetail(group, params)
			} else {
				rsp.ParamGroups[i] = parseParamDetail(group, nil)
			}
		}
	}
	return nil
}

func (handler *DBServiceHandler) FindParamGroupByID(ctx context.Context, req *dbpb.DBFindParamGroupByIDRequest, rsp *dbpb.DBFindParamGroupByIDResponse) error {
	log := framework.LogWithContext(ctx)
	log.Debugln("start find param group by id dao service.")
	group, params, err := handler.Dao().ParamManager().LoadParamGroup(ctx, uint(req.ParamGroupId))
	if err != nil {
		log.Errorf("find param group by id err: %v", err.Error())
		st, ok := status.FromError(err)
		if ok {
			rsp.Status.Code = int32(st.Code())
			rsp.Status.Message = st.Message()
		} else {
			rsp.Status.Code = int32(codes.Internal)
			rsp.Status.Message = fmt.Sprintf("find failure param group failed, err: %v", err)
		}
	} else {
		rsp.ParamGroup = parseParamDetail(group, params)
		rsp.Status = ParamSuccessResponseStatus
	}
	return nil
}

func (handler *DBServiceHandler) ApplyParamGroup(ctx context.Context, req *dbpb.DBApplyParamGroupRequest, rsp *dbpb.DBApplyParamGroupResponse) error {
	log := framework.LogWithContext(ctx)
	log.Debugln("start apply param group dao service.")
	params, err := convertClusterParam(req.Params)
	if err != nil {
		log.Errorf("apply param group err: %v", err.Error())
		rsp.Status.Code = int32(codes.Internal)
		rsp.Status.Message = fmt.Sprintf("apply failure param group failed, err: %v", err)
	}
	err = handler.Dao().ParamManager().ApplyParamGroup(ctx, uint(req.ParamGroupId), req.ClusterId, params)
	if err != nil {
		log.Errorf("apply param group err: %v", err.Error())
		st, ok := status.FromError(err)
		if ok {
			rsp.Status.Code = int32(st.Code())
			rsp.Status.Message = st.Message()
		} else {
			rsp.Status.Code = int32(codes.Internal)
			rsp.Status.Message = fmt.Sprintf("apply failure param group failed, err: %v", err)
		}
	} else {
		rsp.Status = ParamSuccessResponseStatus
		rsp.ParamGroupId = req.ParamGroupId
		rsp.ClusterId = req.ClusterId
	}
	return nil
}

func (handler *DBServiceHandler) FindParamsByClusterId(ctx context.Context, req *dbpb.DBFindParamsByClusterIdRequest, rsp *dbpb.DBFindParamsByClusterIdResponse) error {
	log := framework.LogWithContext(ctx)
	log.Debugln("start find params by cluster id dao service.")
	paramGroupId, total, params, err := handler.Dao().ParamManager().FindParamsByClusterId(ctx, req.ClusterId, int(req.Page.Page-1)*int(req.Page.PageSize), int(req.Page.PageSize))
	if err != nil {
		log.Errorf("find params by cluster err: %v", err.Error())
		st, ok := status.FromError(err)
		if ok {
			rsp.Status.Code = int32(st.Code())
			rsp.Status.Message = st.Message()
		} else {
			rsp.Status.Code = int32(codes.Internal)
			rsp.Status.Message = fmt.Sprintf("find failure cluster params failed, err: %v", err)
		}
	} else {
		rsp.Status = ParamSuccessResponseStatus
		rsp.ParamGroupId = int64(paramGroupId)
		rsp.Page = &dbpb.DBParamsPageDTO{
			Page:     req.Page.Page,
			PageSize: req.Page.PageSize,
			Total:    int32(total),
		}
		rsp.Params, err = parseClusterParam(params)
		if err != nil {
			rsp.Status.Code = int32(codes.Internal)
			rsp.Status.Message = fmt.Sprintf("find failure cluster params failed, err: %v", err)
		}
	}
	return nil
}

func (handler *DBServiceHandler) UpdateClusterParams(ctx context.Context, req *dbpb.DBUpdateClusterParamsRequest, rsp *dbpb.DBUpdateClusterParamsResponse) error {
	log := framework.LogWithContext(ctx)
	log.Debugln("start update cluster params dao service.")
	params, err := convertClusterParam(req.Params)
	if err != nil {
		log.Errorf("update cluster params err: %v", err.Error())
		rsp.Status.Code = int32(codes.Internal)
		rsp.Status.Message = fmt.Sprintf("update failure cluster params failed, err: %v", err)
	}
	err = handler.Dao().ParamManager().UpdateClusterParams(ctx, req.ClusterId, params)
	if err != nil {
		log.Errorf("update cluster params err: %v", err.Error())
		st, ok := status.FromError(err)
		if ok {
			rsp.Status.Code = int32(st.Code())
			rsp.Status.Message = st.Message()
		} else {
			rsp.Status.Code = int32(codes.Internal)
			rsp.Status.Message = fmt.Sprintf("update failure cluster params failed, err: %v", err)
		}
	} else {
		rsp.Status = ParamSuccessResponseStatus
		rsp.ClusterId = req.ClusterId
	}
	return nil
}

func parseClusterParam(params []*models.ClusterParamDetail) (dtos []*dbpb.DBClusterParamDTO, err error) {
	dtos = make([]*dbpb.DBClusterParamDTO, len(params))
	for i, param := range params {
		realValueDto := &dbpb.DBParamRealValueDTO{}
		err := json.Unmarshal([]byte(param.RealValue), realValueDto)
		if err != nil {
			return nil, err
		}
		dtos[i] = &dbpb.DBClusterParamDTO{
			ParamId:       int64(param.Id),
			Name:          param.Name,
			ComponentType: param.ComponentType,
			Type:          int32(param.Type),
			Unit:          param.Unit,
			Range:         param.Range,
			HasReboot:     int32(param.HasReboot),
			Source:        int32(param.Source),
			DefaultValue:  param.DefaultValue,
			Description:   param.Description,
			Note:          param.Note,
			CreateTime:    param.CreatedAt.Unix(),
			UpdateTime:    param.UpdatedAt.Unix(),
			RealValue:     realValueDto,
		}
	}
	return dtos, nil
}

func convertClusterParam(params []*dbpb.DBApplyParamDTO) (dtos []*models.ClusterParamMapDO, err error) {
	dtos = make([]*models.ClusterParamMapDO, len(params))
	for i, param := range params {
		b, err := json.Marshal(param.RealValue)
		if err != nil {
			return dtos, err
		}
		dtos[i] = &models.ClusterParamMapDO{
			ParamId:   uint(param.ParamId),
			RealValue: string(b),
		}
	}
	return dtos, nil
}

func parseParamDetail(group *models.ParamGroupDO, params []*models.ParamDetail) (dto *dbpb.DBParamGroupDTO) {
	dto = &dbpb.DBParamGroupDTO{
		ParamGroupId: int64(group.ID),
		Name:         group.Name,
		DbType:       int32(group.DbType),
		HasDefault:   int32(group.HasDefault),
		Version:      group.Version,
		Spec:         group.Spec,
		GroupType:    int32(group.GroupType),
		Note:         group.Note,
		CreateTime:   group.CreatedAt.Unix(),
		UpdateTime:   group.UpdatedAt.Unix(),
	}
	if params == nil {
		return
	}
	dto.Params = make([]*dbpb.DBParamDTO, len(params))
	for i, param := range params {
		dto.Params[i] = &dbpb.DBParamDTO{
			ParamId:       int64(param.Id),
			Name:          param.Name,
			ComponentType: param.ComponentType,
			Type:          int32(param.Type),
			Unit:          param.Unit,
			Range:         param.Range,
			HasReboot:     int32(param.HasReboot),
			Source:        int32(param.Source),
			DefaultValue:  param.DefaultValue,
			Description:   param.Description,
			Note:          param.Note,
			CreateTime:    param.CreatedAt.Unix(),
			UpdateTime:    param.UpdatedAt.Unix(),
		}
	}
	return dto
}

func parseParamGroup(req *dbpb.DBCreateParamGroupRequest) (p *models.ParamGroupDO) {
	if req == nil {
		return
	}
	return &models.ParamGroupDO{
		Name:       req.Name,
		ParentId:   uint(req.ParentId),
		Spec:       req.Spec,
		HasDefault: int(req.HasDefault),
		DbType:     int(req.DbType),
		GroupType:  int(req.GroupType),
		Version:    req.Version,
		Note:       req.Note,
	}
}

func parseSubmitParams(req []*dbpb.DBSubmitParamDTO) (ps []*models.ParamGroupMapDO) {
	if req == nil {
		return
	}
	params := make([]*models.ParamGroupMapDO, len(req))
	for i, p := range req {
		params[i] = &models.ParamGroupMapDO{
			ParamId:      uint(p.ParamId),
			DefaultValue: p.DefaultValue,
			Note:         p.Note,
		}
	}
	return params
}
