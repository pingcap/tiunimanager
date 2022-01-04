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
	"github.com/pingcap-inc/tiem/proto/clusterservices"
	"sync"

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
		"start":       {"modifyParameter", "modifyDone", "fail", workflow.SyncFuncNode, modifyParameters},
		"modifyDone":  {"refreshParameter", "refreshDone", "fail", workflow.SyncFuncNode, refreshParameter},
		"refreshDone": {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(defaultEnd, persistParameter)},
		"fail":        {"fail", "", "", workflow.SyncFuncNode, defaultEnd},
	},
}

func (m *Manager) QueryClusterParameters(ctx context.Context, req cluster.QueryClusterParametersReq) (resp cluster.QueryClusterParametersResp, page *clusterservices.RpcPage, err error) {
	framework.LogWithContext(ctx).Infof("begin query cluster parameters, request: %+v", req)
	defer framework.LogWithContext(ctx).Infof("end query cluster parameters")

	offset := (req.Page - 1) * req.PageSize
	pgId, params, total, err := models.GetClusterParameterReaderWriter().QueryClusterParameter(ctx, req.ClusterID, offset, req.PageSize)
	if err != nil {
		return resp, page, errors.NewEMErrorf(errors.TIEM_CLUSTER_PARAMETER_QUERY_ERROR, errors.TIEM_CLUSTER_PARAMETER_QUERY_ERROR.Explain(), err)
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
				return resp, page, errors.NewEMErrorf(errors.TIEM_CONVERT_OBJ_FAILED, errors.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
			}
		}
		// convert realValue
		realValue := structs.ParameterRealValue{}
		if len(param.RealValue) > 0 {
			err = json.Unmarshal([]byte(param.RealValue), &realValue)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("failed to convert parameter realValue. req: %v, err: %v", req, err)
				return resp, page, errors.NewEMErrorf(errors.TIEM_CONVERT_OBJ_FAILED, errors.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
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

	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := handler.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluser%s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	data := make(map[string]interface{})
	data[contextModifyParameters] = &ModifyParameter{Reboot: req.Reboot, Params: req.Params}
	data[contextUpdateParameterInfo] = &req
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

func (m *Manager) ApplyParameterGroup(ctx context.Context, req message.ApplyParameterGroupReq) (resp message.ApplyParameterGroupResp, err error) {
	framework.LogWithContext(ctx).Infof("begin apply cluster parameters, request: %+v", req)
	defer framework.LogWithContext(ctx).Infof("end apply cluster parameters")

	// Get cluster info and topology from db based by clusterID
	clusterMeta, err := handler.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluser%s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	// Detail parameter group by id
	_, pgm, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("detail parameter group %s from db error: %s", req.ParamGroupId, err.Error())
		return resp, errors.NewEMErrorf(errors.TIEM_PARAMETER_GROUP_DETAIL_ERROR, errors.TIEM_PARAMETER_GROUP_DETAIL_ERROR.Explain())
	}

	// Constructing clusterParameter.ModifyParameter objects
	params := make([]structs.ClusterParameterSampleInfo, len(pgm))
	for i, param := range pgm {
		params[i] = structs.ClusterParameterSampleInfo{
			ParamId:        param.ID,
			Category:       param.Category,
			Name:           param.Name,
			InstanceType:   param.InstanceType,
			UpdateSource:   param.UpdateSource,
			SystemVariable: param.SystemVariable,
			Type:           param.Type,
			HasApply:       param.HasApply,
			RealValue:      structs.ParameterRealValue{ClusterValue: param.DefaultValue},
		}
	}

	// Get modify parameters
	data := make(map[string]interface{})
	data[contextModifyParameters] = &ModifyParameter{Reboot: req.Reboot, Params: params}
	data[contextApplyParameterInfo] = &req
	data[contextMaintenanceStatusChange] = true
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

func (m *Manager) InspectClusterParameters(ctx context.Context, req cluster.InspectClusterParametersReq) (resp cluster.InspectClusterParametersResp, err error) {
	// todo: Reliance on parameter source query implementation
	return
}
