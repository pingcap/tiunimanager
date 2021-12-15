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
 * @File: manager
 * @Description: cluster parameter service implements
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 10:01
*******************************************************************************/

package parameter

import (
	"context"
	"encoding/json"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/pingcap-inc/tiem/models/cluster/parameter"

	"github.com/pingcap-inc/tiem/common/structs"

	"github.com/pingcap-inc/tiem/library/common"

	"github.com/pingcap-inc/tiem/models"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) QueryClusterParameters(ctx context.Context, req cluster.QueryClusterParametersReq) (resp cluster.QueryClusterParametersResp, page *clusterpb.RpcPage, err error) {
	offset := (req.Page - 1) * req.PageSize
	pgId, params, total, err := models.GetClusterParameterReaderWriter().QueryClusterParameter(ctx, req.ClusterID, offset, req.PageSize)
	if err != nil {
		return resp, page, framework.WrapError(common.TIEM_CLUSTER_PARAMETER_QUERY_ERROR, common.TIEM_CLUSTER_PARAMETER_QUERY_ERROR.Explain(), err)
	}

	resp = cluster.QueryClusterParametersResp{ParamGroupId: pgId}
	resp.Params = make([]structs.ClusterParameterInfo, len(params))
	for i, param := range params {
		// convert range
		ranges := make([]string, 0)
		if len(param.Range) > 0 {
			err = json.Unmarshal([]byte(param.Range), &ranges)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("failed to convert parameter range. req: %v, err: %v", req, err)
				return resp, page, framework.WrapError(common.TIEM_CONVERT_OBJ_FAILED, common.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
			}
		}
		// convert realValue
		realValue := structs.ParameterRealValue{}
		if len(param.RealValue) > 0 {
			err = json.Unmarshal([]byte(param.RealValue), &realValue)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("failed to convert parameter realValue. req: %v, err: %v", req, err)
				return resp, page, framework.WrapError(common.TIEM_CONVERT_OBJ_FAILED, common.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
			}
		}
		resp.Params[i] = structs.ClusterParameterInfo{
			ParamId:        param.ID,
			Category:       param.Category,
			Name:           param.Name,
			InstanceType:   param.InstanceType,
			SystemVariable: param.SystemVariable,
			Type:           param.Type,
			Unit:           param.Unit,
			Range:          ranges,
			HasReboot:      param.HasReboot,
			HasApply:       param.HasApply,
			UpdateSource:   param.UpdateSource,
			DefaultValue:   param.DefaultValue,
			RealValue:      realValue,
			Description:    param.Description,
			Note:           param.Note,
			CreatedAt:      param.CreatedAt.Unix(),
			UpdatedAt:      param.UpdatedAt.Unix(),
		}
	}

	page = &clusterpb.RpcPage{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Total:    int32(total),
	}
	return resp, page, nil
}

func (m *Manager) UpdateClusterParameters(ctx context.Context, req cluster.UpdateClusterParametersReq) (resp cluster.UpdateClusterParametersResp, err error) {
	params := make([]*parameter.ClusterParameterMapping, len(req.Params))
	for i, param := range req.Params {
		b, err := json.Marshal(param.RealValue)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to convert parameter realValue. req: %v, err: %v", req, err)
			return resp, framework.WrapError(common.TIEM_CONVERT_OBJ_FAILED, common.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
		}
		params[i] = &parameter.ClusterParameterMapping{
			ClusterID:   req.ClusterID,
			ParameterID: param.ParamId,
			RealValue:   string(b),
		}
	}
	err = models.GetClusterParameterReaderWriter().UpdateClusterParameter(ctx, req.ClusterID, params)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update cluster parameter req: %v, err: %v", req, err)
		return resp, framework.WrapError(common.TIEM_CLUSTER_PARAMETER_UPDATE_ERROR, common.TIEM_CLUSTER_PARAMETER_UPDATE_ERROR.Explain(), err)
	}
	resp = cluster.UpdateClusterParametersResp{ClusterID: req.ClusterID}
	return resp, nil
}

func (m *Manager) InspectClusterParameters(ctx context.Context, req cluster.InspectClusterParametersReq) (resp cluster.InspectClusterParametersResp, err error) {
	// todo: Reliance on parameter source query implementation
	return
}
