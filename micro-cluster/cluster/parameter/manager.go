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
	"sync"

	"github.com/pingcap-inc/tiem/models/cluster/parameter"

	"github.com/pingcap-inc/tiem/proto/clusterservices"

	"github.com/pingcap-inc/tiem/common/errors"

	"github.com/pingcap-inc/tiem/message"

	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/workflow"

	"github.com/pingcap-inc/tiem/common/structs"

	"github.com/pingcap-inc/tiem/models"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
)

type Manager struct{}

var manager *Manager
var once sync.Once

func NewManager() *Manager {
	once.Do(func() {
		if manager == nil {
			workflowManager := workflow.GetWorkFlowService()
			workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowModifyParameters, &modifyParametersDefine)

			manager = &Manager{}
		}
	})
	return manager
}

var modifyParametersDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowModifyParameters,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":          {"validationParameter", "validationDone", "fail", workflow.SyncFuncNode, validationParameter},
		"validationDone": {"modifyParameter", "modifyDone", "fail", workflow.SyncFuncNode, modifyParameters},
		"modifyDone":     {"refreshParameter", "refreshDone", "fail", workflow.SyncFuncNode, refreshParameter},
		"refreshDone":    {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(defaultEnd, persistParameter)},
		"fail":           {"end", "", "", workflow.SyncFuncNode, defaultEnd},
	},
}

func (m *Manager) QueryClusterParameters(ctx context.Context, req cluster.QueryClusterParametersReq) (resp cluster.QueryClusterParametersResp, page *clusterservices.RpcPage, err error) {
	framework.LogWithContext(ctx).Infof("begin query cluster parameters, request: %+v", req)
	defer framework.LogWithContext(ctx).Infof("end query cluster parameters")

	offset := (req.Page - 1) * req.PageSize
	pgId, params, total, err := models.GetClusterParameterReaderWriter().QueryClusterParameter(ctx, req.ClusterID, req.ParamName, req.InstanceType, offset, req.PageSize)
	if err != nil {
		return resp, page, errors.NewErrorf(errors.TIEM_CLUSTER_PARAMETER_QUERY_ERROR, errors.TIEM_CLUSTER_PARAMETER_QUERY_ERROR.Explain(), err)
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
				return resp, page, errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, errors.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
			}
		}
		// convert realValue
		realValue := structs.ParameterRealValue{}
		if len(param.RealValue) > 0 {
			err = json.Unmarshal([]byte(param.RealValue), &realValue)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("failed to convert parameter realValue. req: %v, err: %v", req, err)
				return resp, page, errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, errors.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
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
			ReadOnly:       param.ReadOnly,
			DefaultValue:   param.DefaultValue,
			RealValue:      realValue,
			Description:    param.Description,
			Note:           param.Note,
			CreatedAt:      param.CreatedAt.Unix(),
			UpdatedAt:      param.UpdatedAt.Unix(),
		}
	}

	page = &clusterservices.RpcPage{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Total:    int32(total),
	}
	return resp, page, nil
}

func (m *Manager) UpdateClusterParameters(ctx context.Context, req cluster.UpdateClusterParametersReq, maintenanceStatusChange bool) (resp cluster.UpdateClusterParametersResp, err error) {
	framework.LogWithContext(ctx).Infof("begin update cluster parameters, request: %+v", req)
	defer framework.LogWithContext(ctx).Infof("end update cluster parameters")

	// query cluster parameter by cluster id
	pgId, paramDetails, total, err := models.GetClusterParameterReaderWriter().QueryClusterParameter(ctx, req.ClusterID, "", "", 0, 0)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query cluster %s parameter error: %s", req.ClusterID, err.Error())
		return
	}
	framework.LogWithContext(ctx).Infof("query cluster %s parameter group id: %s, total size: %d", req.ClusterID, pgId, total)

	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := handler.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluser%s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	params := make([]ModifyClusterParameterInfo, 0)

	// Iterate to get the complete information of the modified parameters
	for _, detail := range paramDetails {
		for _, param := range req.Params {
			if detail.ID == param.ParamId {
				ranges := make([]string, 0)
				if len(detail.Range) > 0 {
					err = json.Unmarshal([]byte(detail.Range), &ranges)
					if err != nil {
						framework.LogWithContext(ctx).Errorf("failed to convert parameter range. range: %v, err: %v", detail.Range, err)
						return
					}
				}
				params = append(params, ModifyClusterParameterInfo{
					ParamId:        param.ParamId,
					Category:       detail.Category,
					Name:           detail.Name,
					InstanceType:   detail.InstanceType,
					UpdateSource:   detail.UpdateSource,
					ReadOnly:       detail.ReadOnly,
					SystemVariable: detail.SystemVariable,
					Type:           detail.Type,
					Range:          ranges,
					HasApply:       detail.HasApply,
					RealValue:      param.RealValue,
				})
			}
		}
	}

	data := make(map[string]interface{})
	data[contextModifyParameters] = &ModifyParameter{ClusterID: req.ClusterID, Reboot: req.Reboot, Params: params, Nodes: req.Nodes}
	data[contextMaintenanceStatusChange] = maintenanceStatusChange
	workflowID, err := asyncMaintenance(ctx, clusterMeta, data, constants.ClusterMaintenanceModifyParameterAndRestarting, modifyParametersDefine.FlowName)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("cluster %s update cluster parameters aync maintenance workflow error: %s", req.ClusterID, err.Error())
		return
	}

	resp = cluster.UpdateClusterParametersResp{
		ClusterID: req.ClusterID,
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: workflowID,
		},
	}
	return resp, nil
}

func (m *Manager) ApplyParameterGroup(ctx context.Context, req message.ApplyParameterGroupReq, maintenanceStatusChange bool) (resp message.ApplyParameterGroupResp, err error) {
	framework.LogWithContext(ctx).Infof("begin apply cluster parameters, request: %+v", req)
	defer framework.LogWithContext(ctx).Infof("end apply cluster parameters")

	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := handler.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluser%s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	// Detail parameter group by id
	_, pgm, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupId, "")
	if err != nil {
		framework.LogWithContext(ctx).Errorf("detail parameter group %s from db error: %s", req.ParamGroupId, err.Error())
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_DETAIL_ERROR, errors.TIEM_PARAMETER_GROUP_DETAIL_ERROR.Explain())
	}

	// Constructing clusterParameter.ModifyParameter objects
	params := make([]ModifyClusterParameterInfo, len(pgm))
	for i, param := range pgm {
		ranges := make([]string, 0)
		if len(param.Range) > 0 {
			err = json.Unmarshal([]byte(param.Range), &ranges)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("failed to convert parameter range. range: %v, err: %v", param.Range, err)
				return
			}
		}
		params[i] = ModifyClusterParameterInfo{
			ParamId:        param.ID,
			Category:       param.Category,
			Name:           param.Name,
			InstanceType:   param.InstanceType,
			UpdateSource:   param.UpdateSource,
			ReadOnly:       param.ReadOnly,
			SystemVariable: param.SystemVariable,
			Type:           param.Type,
			HasApply:       param.HasApply,
			Range:          ranges,
			RealValue:      structs.ParameterRealValue{ClusterValue: param.DefaultValue},
		}
	}

	// Get modify parameters
	data := make(map[string]interface{})
	data[contextModifyParameters] = &ModifyParameter{ClusterID: req.ClusterID, ParamGroupId: req.ParamGroupId, Reboot: req.Reboot, Params: params, Nodes: req.Nodes}
	data[contextHasApplyParameter] = true
	data[contextMaintenanceStatusChange] = maintenanceStatusChange
	workflowID, err := asyncMaintenance(ctx, clusterMeta, data, constants.ClusterMaintenanceModifyParameterAndRestarting, modifyParametersDefine.FlowName)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("cluster %s update cluster parameters aync maintenance workflow error: %s", req.ClusterID, err.Error())
		return
	}

	resp = message.ApplyParameterGroupResp{
		ClusterID:    req.ClusterID,
		ParamGroupID: req.ParamGroupId,
		AsyncTaskWorkFlowInfo: structs.AsyncTaskWorkFlowInfo{
			WorkFlowID: workflowID,
		},
	}
	return resp, nil
}

func (m *Manager) PersistApplyParameterGroup(ctx context.Context, req message.ApplyParameterGroupReq, hasEmptyValue bool) (resp message.ApplyParameterGroupResp, err error) {
	pg, params, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupId, "")
	if err != nil || pg.ID == "" {
		framework.LogWithContext(ctx).Errorf("get parameter group req: %v, err: %v", req, err)
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_DETAIL_ERROR, err.Error())
	}

	pgs := make([]*parameter.ClusterParameterMapping, len(params))
	for i, param := range params {
		value := ""
		if !hasEmptyValue {
			value = param.DefaultValue
		}
		realValue := structs.ParameterRealValue{ClusterValue: value}
		b, err := json.Marshal(realValue)
		if err != nil {
			return resp, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_APPLY_ERROR, err.Error())
		}
		pgs[i] = &parameter.ClusterParameterMapping{
			ClusterID:   req.ClusterID,
			ParameterID: param.ID,
			RealValue:   string(b),
		}
	}
	err = models.GetClusterParameterReaderWriter().ApplyClusterParameter(ctx, req.ParamGroupId, req.ClusterID, pgs)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("apply parameter group convert resp err: %v", err)
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_APPLY_ERROR, err.Error())
	}
	resp = message.ApplyParameterGroupResp{
		ClusterID:    req.ClusterID,
		ParamGroupID: req.ParamGroupId,
	}
	return resp, nil
}

func (m *Manager) InspectClusterParameters(ctx context.Context, req cluster.InspectClusterParametersReq) (resp cluster.InspectClusterParametersResp, err error) {
	// todo: Reliance on parameter source query implementation
	return
}
