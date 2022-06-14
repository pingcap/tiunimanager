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
 * @Description: parameter group service implements
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/9 15:27
*******************************************************************************/

package parametergroup

import (
	"context"
	"fmt"

	"github.com/pingcap/tiunimanager/micro-cluster/cluster/parameter"

	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/parametergroup"
	"github.com/pingcap/tiunimanager/proto/clusterservices"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) CreateParameterGroup(ctx context.Context, req message.CreateParameterGroupReq) (resp message.CreateParameterGroupResp, err error) {
	// validate parameter
	if req.GroupType != int(Cluster) && req.GroupType != int(Instance) {
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "The groupType value can only be 1 or 2")
	}
	if req.DBType != int(TiDB) && req.DBType != int(DM) {
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "The dbType value can only be 1 or 2")
	}
	if req.Params == nil || len(req.Params) == 0 {
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "The params cannot be empty")
	}

	// validation parameter range
	if err = validateParameterRange(ctx, req.Params); err != nil {
		return resp, err
	}
	pg := &parametergroup.ParameterGroup{
		Name:           req.Name,
		ClusterSpec:    req.ClusterSpec,
		HasDefault:     int(CUSTOM), // The created parameter group is a custom parameter group
		DBType:         req.DBType,
		GroupType:      req.GroupType,
		ClusterVersion: req.ClusterVersion,
		Note:           req.Note,
	}
	pgm := make([]*parametergroup.ParameterGroupMapping, len(req.Params))
	for i, param := range req.Params {
		pgm[i] = &parametergroup.ParameterGroupMapping{
			ParameterID:  param.ID,
			DefaultValue: param.DefaultValue,
			Note:         param.Note,
		}
	}

	// Check for the existence of extended parameters added when creating a parameter group
	if err = checkAddParametersExists(ctx, req.AddParams); err != nil {
		return resp, err
	}

	// invoke database reader writer.
	parameterGroup, err := models.GetParameterGroupReaderWriter().CreateParameterGroup(ctx, pg, pgm, req.AddParams)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create parameter group req: %v, err: %v", req, err)
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_CREATE_ERROR, errors.TIUNIMANAGER_PARAMETER_GROUP_CREATE_ERROR.Explain(), err)
	}
	resp = message.CreateParameterGroupResp{ParamGroupID: parameterGroup.ID}
	return resp, nil
}

func (m *Manager) UpdateParameterGroup(ctx context.Context, req message.UpdateParameterGroupReq) (resp message.UpdateParameterGroupResp, err error) {
	// validation parameters range
	if err = validateParameterRange(ctx, req.Params); err != nil {
		return resp, err
	}
	group, _, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupID, "", "")
	if err != nil || group.ID == "" {
		framework.LogWithContext(ctx).Errorf("get parameter group req: %v, err: %v", req, err)
		err = errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DETAIL_ERROR, errors.TIUNIMANAGER_PARAMETER_GROUP_DETAIL_ERROR.Explain(), err)
		return
	}

	// default parameter group not be modify.
	if group.HasDefault == int(DEFAULT) {
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_DEFAULT_PARAM_GROUP_NOT_MODIFY, errors.TIUNIMANAGER_DEFAULT_PARAM_GROUP_NOT_MODIFY.Explain())
	}

	// Check for the existence of extended parameters added when creating a parameter group
	if err = checkAddParametersExists(ctx, req.AddParams); err != nil {
		return resp, err
	}

	pg := &parametergroup.ParameterGroup{
		ID:             req.ParamGroupID,
		Name:           req.Name,
		ClusterSpec:    req.ClusterSpec,
		ClusterVersion: req.ClusterVersion,
		Note:           req.Note,
	}
	pgm := make([]*parametergroup.ParameterGroupMapping, len(req.Params))
	for i, param := range req.Params {
		pgm[i] = &parametergroup.ParameterGroupMapping{
			ParameterID:  param.ID,
			DefaultValue: param.DefaultValue,
			Note:         param.Note,
		}
	}

	err = models.GetParameterGroupReaderWriter().UpdateParameterGroup(ctx, pg, pgm, req.AddParams, req.DelParams)

	if err != nil {
		framework.LogWithContext(ctx).Errorf("update parameter group invoke metadb err: %v", err)
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_UPDATE_ERROR, errors.TIUNIMANAGER_PARAMETER_GROUP_UPDATE_ERROR.Explain(), err)
	}
	resp = message.UpdateParameterGroupResp{ParamGroupID: req.ParamGroupID}
	return resp, nil
}

func (m *Manager) DeleteParameterGroup(ctx context.Context, req message.DeleteParameterGroupReq) (resp message.DeleteParameterGroupResp, err error) {
	pg, _, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupID, "", "")
	if err != nil || pg.ID == "" {
		framework.LogWithContext(ctx).Errorf("get parameter group req: %v, err: %v", req, err)
		err = errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DETAIL_ERROR, errors.TIUNIMANAGER_PARAMETER_GROUP_DETAIL_ERROR.Explain(), err)
		return
	}

	// default parameter group not be deleted.
	if pg.HasDefault == int(DEFAULT) {
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_DEFAULT_PARAM_GROUP_NOT_DEL, errors.TIUNIMANAGER_DEFAULT_PARAM_GROUP_NOT_DEL.Explain())
	}
	err = models.GetParameterGroupReaderWriter().DeleteParameterGroup(ctx, pg.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("delete parameter group invoke metadb err: %v", err)
		err = errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DELETE_ERROR, errors.TIUNIMANAGER_PARAMETER_GROUP_DELETE_ERROR.Explain(), err)
		return
	}
	resp = message.DeleteParameterGroupResp{ParamGroupID: pg.ID}
	return resp, nil
}

func (m *Manager) QueryParameterGroup(ctx context.Context, req message.QueryParameterGroupReq) (resp []message.QueryParameterGroupResp, page *clusterservices.RpcPage, err error) {
	offset := (req.Page - 1) * req.PageSize
	pgs, total, err := models.GetParameterGroupReaderWriter().QueryParameterGroup(ctx, req.Name, req.ClusterSpec, req.ClusterVersion, req.DBType, req.HasDefault, offset, req.PageSize)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query parameter group req: %v, err: %v", req, err)
		return resp, page, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_QUERY_ERROR, errors.TIUNIMANAGER_PARAMETER_GROUP_QUERY_ERROR.Explain(), err)
	}

	resp = make([]message.QueryParameterGroupResp, len(pgs))
	for i, pg := range pgs {
		resp[i] = message.QueryParameterGroupResp{ParameterGroupInfo: convertParameterGroupInfo(pg)}

		// condition load parameter details
		if req.HasDetail {
			pgm, err := models.GetParameterGroupReaderWriter().QueryParametersByGroupId(ctx, pg.ID, "", "")
			if err != nil {
				framework.LogWithContext(ctx).Errorf("query parameter group req: %v, err: %v", req, err)
				return resp, page, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_QUERY_ERROR, errors.TIUNIMANAGER_PARAMETER_QUERY_ERROR.Explain(), err)
			}
			params := make([]structs.ParameterGroupParameterInfo, len(pgm))
			for j, param := range pgm {
				pgi, err := convertParameterGroupParameterInfo(param)
				if err != nil {
					framework.LogWithContext(ctx).Errorf("failed to convert parameter group. req: %v, err: %v", req, err)
					return resp, page, errors.NewErrorf(errors.TIUNIMANAGER_CONVERT_OBJ_FAILED, errors.TIUNIMANAGER_CONVERT_OBJ_FAILED.Explain(), err)
				}
				params[j] = pgi
			}
			resp[i].Params = params
		}
	}

	page = &clusterservices.RpcPage{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Total:    int32(total),
	}
	return resp, page, nil
}

func (m *Manager) DetailParameterGroup(ctx context.Context, req message.DetailParameterGroupReq) (resp message.DetailParameterGroupResp, err error) {
	pg, pgm, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupID, req.ParamName, req.InstanceType)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get parameter group req: %v, err: %v", req, err)
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DETAIL_ERROR, errors.TIUNIMANAGER_PARAMETER_GROUP_DETAIL_ERROR.Explain(), err)
	}
	resp = message.DetailParameterGroupResp{ParameterGroupInfo: convertParameterGroupInfo(pg)}

	params := make([]structs.ParameterGroupParameterInfo, len(pgm))
	for i, param := range pgm {
		pgi, err := convertParameterGroupParameterInfo(param)
		if err != nil {
			return resp, errors.NewErrorf(errors.TIUNIMANAGER_CONVERT_OBJ_FAILED, errors.TIUNIMANAGER_CONVERT_OBJ_FAILED.Explain(), err)
		}
		params[i] = pgi
	}
	resp.Params = params
	return resp, nil
}

func (m *Manager) CopyParameterGroup(ctx context.Context, req message.CopyParameterGroupReq) (resp message.CopyParameterGroupResp, err error) {
	// get parameter group by id
	pg, params, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupID, "", "")
	if err != nil || pg.ID == "" {
		framework.LogWithContext(ctx).Errorf("get parameter group req: %v, err: %v", req, err)
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_DETAIL_ERROR, errors.TIUNIMANAGER_PARAMETER_GROUP_DETAIL_ERROR.Explain(), err)
	}

	// Determine if the name of the copy parameter group is modified
	if pg.Name == req.Name {
		framework.LogWithContext(ctx).Errorf("Parameter group name %s already exists", req.Name)
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_NAME_ALREADY_EXISTS, errors.TIUNIMANAGER_PARAMETER_GROUP_NAME_ALREADY_EXISTS.Explain())
	}

	pgm := make([]*parametergroup.ParameterGroupMapping, len(params))
	for i, param := range params {
		pgm[i] = &parametergroup.ParameterGroupMapping{
			ParameterID:  param.ID,
			DefaultValue: param.DefaultValue,
			Note:         param.Note,
		}
	}

	// reset parameter group object
	pg.ID = ""
	pg.Name = req.Name
	pg.Note = req.Note
	// copy parameter group HasDefault values is 2
	pg.HasDefault = int(CUSTOM)
	parameterGroup, err := models.GetParameterGroupReaderWriter().CreateParameterGroup(ctx, pg, pgm, nil)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("copy parameter group convert resp err: %v", err)
		return resp, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_COPY_ERROR, errors.TIUNIMANAGER_PARAMETER_GROUP_COPY_ERROR.Explain(), err)
	}
	resp = message.CopyParameterGroupResp{ParamGroupID: parameterGroup.ID}
	return resp, nil
}

// checkAddParametersExists
// @Description: check add parameters exists.
// @Parameter ctx
// @Parameter addParams
// @return err
func checkAddParametersExists(ctx context.Context, addParams []message.ParameterInfo) (err error) {
	if len(addParams) > 0 {
		for _, param := range addParams {
			existsParameter, err := models.GetParameterGroupReaderWriter().ExistsParameter(ctx, param.Category, param.Name, param.InstanceType)
			if err != nil {
				return err
			}
			if existsParameter != nil && existsParameter.ID != "" {
				return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_ALREADY_EXISTS,
					fmt.Sprintf("%s parameter `%s` already exists, parameter ID: %s", param.InstanceType, parameter.DisplayFullParameterName(param.Category, param.Name), existsParameter.ID))
			}
		}
	}
	return nil
}

// validateParameterRange
// @Description: validate parameter by range
// @Parameter ctx
// @Parameter reqParams
// @return err
func validateParameterRange(ctx context.Context, reqParams []structs.ParameterGroupParameterSampleInfo) (err error) {
	// query parameters list
	params, total, err := models.GetParameterGroupReaderWriter().QueryParameters(ctx, 0, 0)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query parameters list err: %v", err)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_QUERY_ERROR, errors.TIUNIMANAGER_PARAMETER_QUERY_ERROR.Explain(), err)
	}
	framework.LogWithContext(ctx).Debugf("validate parameters query count: %v", total)

	queryParamContainer := make(map[string]*parametergroup.Parameter)
	for _, queryParam := range params {
		queryParamContainer[queryParam.ID] = queryParam
	}
	// Iterate through the range of parameters to check if they are legal
	for _, reqParam := range reqParams {
		if queryParam, ok := queryParamContainer[reqParam.ID]; ok {
			ranges, err := parameter.UnmarshalCovertArray(queryParam.Range)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("failed to convert parameter range. range: %v, err: %v", queryParam.Range, err)
				return err
			}
			// convert unitOptions
			unitOptions, err := parameter.UnmarshalCovertArray(queryParam.UnitOptions)
			if err != nil {
				return err
			}

			if !parameter.ValidateRange(&parameter.ModifyClusterParameterInfo{
				ParamId:     reqParam.ID,
				Type:        queryParam.Type,
				Range:       ranges,
				RangeType:   queryParam.RangeType,
				Unit:        queryParam.Unit,
				UnitOptions: unitOptions,
				HasApply:    queryParam.HasApply,
				RealValue:   structs.ParameterRealValue{ClusterValue: reqParam.DefaultValue},
			}, false) {
				if queryParam.RangeType == int(parameter.ContinuousRange) && len(queryParam.Range) == 2 {
					return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID,
						fmt.Sprintf("Validation %s parameter `%s` failed, update value: %s, can take a range of values: %v",
							queryParam.InstanceType, parameter.DisplayFullParameterName(queryParam.Category, queryParam.Name), reqParam.DefaultValue, ranges))
				} else {
					return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID,
						fmt.Sprintf("Validation %s parameter `%s` failed, update value: %s, optional values: %v",
							queryParam.InstanceType, parameter.DisplayFullParameterName(queryParam.Category, queryParam.Name), reqParam.DefaultValue, ranges))
				}
			}
		} else {
			return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "No record found for parameter ID: %v", reqParam.ID)
		}
	}
	return nil
}

func convertParameterGroupParameterInfo(param *parametergroup.ParamDetail) (pgi structs.ParameterGroupParameterInfo, err error) {
	// convert range
	ranges, err := parameter.UnmarshalCovertArray(param.Range)
	if err != nil {
		return pgi, err
	}
	// convert unitOptions
	unitOptions, err := parameter.UnmarshalCovertArray(param.UnitOptions)
	if err != nil {
		return pgi, err
	}

	pgi = structs.ParameterGroupParameterInfo{
		ID:             param.ID,
		Category:       param.Category,
		Name:           param.Name,
		InstanceType:   param.InstanceType,
		SystemVariable: param.SystemVariable,
		Type:           param.Type,
		Unit:           param.Unit,
		UnitOptions:    unitOptions,
		Range:          ranges,
		RangeType:      param.RangeType,
		HasReboot:      param.HasReboot,
		HasApply:       param.HasApply,
		DefaultValue:   param.DefaultValue,
		UpdateSource:   param.UpdateSource,
		ReadOnly:       param.ReadOnly,
		Description:    param.Description,
		Note:           param.Note,
		CreatedAt:      param.CreatedAt.Unix(),
		UpdatedAt:      param.UpdatedAt.Unix(),
	}
	return pgi, nil
}

func convertParameterGroupInfo(pg *parametergroup.ParameterGroup) message.ParameterGroupInfo {
	return message.ParameterGroupInfo{
		ParamGroupID:   pg.ID,
		Name:           pg.Name,
		DBType:         pg.DBType,
		HasDefault:     pg.HasDefault,
		ClusterVersion: pg.ClusterVersion,
		ClusterSpec:    pg.ClusterSpec,
		GroupType:      pg.GroupType,
		Note:           pg.Note,
		CreatedAt:      pg.CreatedAt.Unix(),
		UpdatedAt:      pg.UpdatedAt.Unix(),
	}
}
