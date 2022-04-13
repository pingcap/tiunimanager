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
	"fmt"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/pingcap-inc/tiem/deployment"

	tidbApi "github.com/pingcap-inc/tiem/util/api/tidb/http"

	"github.com/pingcap-inc/tiem/util/api/tikv"

	"github.com/pingcap-inc/tiem/util/api/pd"

	"github.com/pingcap-inc/tiem/models/cluster/management"

	"github.com/pingcap-inc/tiem/models/cluster/parameter"

	"github.com/pingcap-inc/tiem/proto/clusterservices"

	"github.com/pingcap-inc/tiem/common/errors"

	"github.com/pingcap-inc/tiem/message"

	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"

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
		"validationDone": {"modifyParameter", "modifyDone", "failParameter", workflow.PollingNode, modifyParameters},
		"modifyDone":     {"refreshParameter", "refreshDone", "failParameter", workflow.PollingNode, refreshParameter},
		"refreshDone":    {"persistParameter", "persistDone", "fail", workflow.SyncFuncNode, persistParameter},
		"persistDone":    {"end", "", "", workflow.SyncFuncNode, defaultEnd},
		"fail":           {"end", "", "", workflow.SyncFuncNode, defaultEnd},
		"failParameter":  {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(parameterFail, defaultEnd)},
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
		ranges, err := UnmarshalCovertArray(param.Range)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to convert parameter range. req: %v, err: %v", req, err)
			return resp, page, errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, errors.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
		}
		unitOptions, err := UnmarshalCovertArray(param.UnitOptions)
		if err != nil {
			return resp, page, errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, errors.TIEM_CONVERT_OBJ_FAILED.Explain(), err)
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
			UnitOptions:    unitOptions,
			Range:          ranges,
			RangeType:      param.RangeType,
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
	clusterMeta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluser%s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	params := make([]*ModifyClusterParameterInfo, 0)

	// Iterate to get the complete information of the modified parameters
	for _, detail := range paramDetails {
		for _, param := range req.Params {
			if detail.ID == param.ParamId {
				ranges, err := UnmarshalCovertArray(detail.Range)
				if err != nil {
					framework.LogWithContext(ctx).Errorf("failed to convert parameter range. req: %v, err: %v", req, err)
					return resp, err
				}
				unitOptions, err := UnmarshalCovertArray(detail.UnitOptions)
				if err != nil {
					return resp, err
				}
				params = append(params, &ModifyClusterParameterInfo{
					ParamId:        param.ParamId,
					Category:       detail.Category,
					Name:           detail.Name,
					InstanceType:   detail.InstanceType,
					UpdateSource:   detail.UpdateSource,
					ReadOnly:       detail.ReadOnly,
					SystemVariable: detail.SystemVariable,
					Type:           detail.Type,
					Range:          ranges,
					RangeType:      detail.RangeType,
					Unit:           detail.Unit,
					UnitOptions:    unitOptions,
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
	clusterMeta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluser%s meta from db error: %s", req.ClusterID, err.Error())
		return
	}

	// Detail parameter group by id
	_, pgm, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupId, "", "")
	if err != nil {
		framework.LogWithContext(ctx).Errorf("detail parameter group %s from db error: %s", req.ParamGroupId, err.Error())
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_DETAIL_ERROR, errors.TIEM_PARAMETER_GROUP_DETAIL_ERROR.Explain())
	}

	// Constructing clusterParameter.ModifyParameter objects
	params := make([]*ModifyClusterParameterInfo, len(pgm))
	for i, param := range pgm {
		ranges, err := UnmarshalCovertArray(param.Range)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to convert parameter range. req: %v, err: %v", req, err)
			return resp, err
		}
		unitOptions, err := UnmarshalCovertArray(param.UnitOptions)
		if err != nil {
			return resp, err
		}
		params[i] = &ModifyClusterParameterInfo{
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
			RangeType:      param.RangeType,
			Unit:           param.Unit,
			UnitOptions:    unitOptions,
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
	pg, params, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupId, "", "")
	if err != nil || pg.ID == "" {
		framework.LogWithContext(ctx).Errorf("get parameter group req: %v, err: %v", req, err)
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_DETAIL_ERROR, err.Error())
	}

	// Define the list of prepared component IDs
	preCompIDs := []constants.EMProductComponentIDType{constants.ComponentIDTiDB, constants.ComponentIDTiKV,
		constants.ComponentIDPD, constants.ComponentIDCDC, constants.ComponentIDTiFlash}
	// Define the list of component IDs that exist
	existsCompIDs := make([]string, 0)

	// Define modify param grouping by instance type
	modifyParamContainer := make(map[interface{}][]*ModifyClusterParameterInfo)
	// Defines the parameters for storing components that are not included in the current cluster
	otherParamContainer := make(map[interface{}][]*ModifyClusterParameterInfo)

	// Define the cluster mapping parameter object used to store
	pgs := make([]*parameter.ClusterParameterMapping, 0)

	// If hasEmptyValue is true, get the running value of the cluster by parameter inspection
	if hasEmptyValue {
		// Fill by cluster parameter inspect
		clusterMeta, err := meta.Get(ctx, req.ClusterID)
		if err != nil {
			return resp, err
		}
		// Determine which deployment components are included in the takeover cluster
		for _, compID := range preCompIDs {
			if _, ok := clusterMeta.Instances[string(compID)]; ok {
				existsCompIDs = append(existsCompIDs, string(compID))
			}
		}

		// Parameter grouping based on existing components
		existsCompIDStr := strings.Join(existsCompIDs, ",")
		for _, param := range params {
			clusterParamInfo := &ModifyClusterParameterInfo{
				ParamId:      param.ID,
				Category:     param.Category,
				Name:         param.Name,
				InstanceType: param.InstanceType,
				Type:         param.Type,
				RealValue:    structs.ParameterRealValue{ClusterValue: ""},
			}
			if strings.Contains(existsCompIDStr, param.InstanceType) {
				putParameterContainer(modifyParamContainer, param.InstanceType, clusterParamInfo)
			} else {
				putParameterContainer(otherParamContainer, param.InstanceType, clusterParamInfo)
			}
		}

		// Iterate through filled patrol parameter values
		for compID, modifyParams := range modifyParamContainer {
			if err := fillParameter(ctx, compID.(string), clusterMeta, modifyParams); err != nil {
				return resp, err
			}
			// Adding cluster parameter mapping object data
			for _, param := range modifyParams {
				if err = appendClusterParamMapping(req.ClusterID, param.ParamId, param.RealValue, pgs); err != nil {
					return resp, err
				}
			}
		}

		// Add cluster parameters other mapped object relationships
		for _, modifyParams := range otherParamContainer {
			for _, param := range modifyParams {
				if err = appendClusterParamMapping(req.ClusterID, param.ParamId, param.RealValue, pgs); err != nil {
					return resp, err
				}
			}
		}
	} else {
		// Fill by default with the parameter group
		for _, param := range params {
			realValue := structs.ParameterRealValue{ClusterValue: param.DefaultValue}
			if err = appendClusterParamMapping(req.ClusterID, param.ID, realValue, pgs); err != nil {
				return resp, err
			}
		}
	}
	// Persist cluster parameter mappings record.
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

// appendClusterParamMapping
// @Description: append cluster parameter mappings
// @Parameter clusterId
// @Parameter paramId
// @Parameter realValue
// @Parameter pgs
// @return error
func appendClusterParamMapping(clusterId, paramId string, realValue structs.ParameterRealValue, pgs []*parameter.ClusterParameterMapping) error {
	b, err := json.Marshal(realValue)
	if err != nil {
		return errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, err.Error())
	}
	pgs = append(pgs, &parameter.ClusterParameterMapping{ // nolint
		ClusterID:   clusterId,
		ParameterID: paramId,
		RealValue:   string(b),
	})
	return nil
}

func (m *Manager) InspectClusterParameters(ctx context.Context, req cluster.InspectParametersReq) (resp cluster.InspectParametersResp, err error) {
	if req.InstanceID == "" && req.ClusterID == "" {
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "both clusterID and instanceID are empty")
	}

	// Define cluster component instance
	var clusterInstance *management.ClusterInstance
	clusterID := req.ClusterID
	instanceType := ""

	if req.InstanceID != "" {
		clusterInstance, err = models.GetClusterReaderWriter().GetInstance(ctx, req.InstanceID)
		if err != nil {
			return resp, err
		}
		clusterID = clusterInstance.ClusterID
		instanceType = clusterInstance.Type
	}

	// Query cluster parameters by meta
	clusterParams, _, err := m.QueryClusterParameters(ctx, cluster.QueryClusterParametersReq{ClusterID: clusterID, InstanceType: instanceType})
	if err != nil {
		return resp, err
	}

	// Define the component instance parameter container
	compContainer := make(map[string][]structs.ClusterParameterInfo)
	for _, param := range clusterParams.Params {
		// If the cluster value is empty then it is skipped directly
		if param.RealValue.ClusterValue == "" {
			continue
		}
		if param.HasApply == int(ModifyApply) {
			continue
		}

		instParams := compContainer[param.InstanceType]
		if instParams == nil {
			compContainer[param.InstanceType] = []structs.ClusterParameterInfo{param}
		} else {
			instParams = append(instParams, param)
			compContainer[param.InstanceType] = instParams
		}
	}

	// Define the set that holds the values of the inspection parameters
	inspectAllParameters := make([]cluster.InspectParameters, 0)

	// Instance level inspection
	if req.InstanceID != "" {
		compParams := compContainer[clusterInstance.Type]
		switch clusterInstance.Type {
		case string(constants.ComponentIDTiDB):
			inspectParams, err := inspectTiDBComponentInstance(ctx, clusterInstance, compParams)
			if err != nil {
				return resp, err
			}
			inspectAllParameters = append(inspectAllParameters, inspectParams)
		case string(constants.ComponentIDTiKV):
			inspectParams, err := inspectTiKVComponentInstance(ctx, clusterInstance, compParams)
			if err != nil {
				return resp, err
			}
			inspectAllParameters = append(inspectAllParameters, inspectParams)
		case string(constants.ComponentIDPD):
			inspectParams, err := inspectPDComponentInstance(ctx, clusterInstance, compParams)
			if err != nil {
				return resp, err
			}
			inspectAllParameters = append(inspectAllParameters, inspectParams)
		case string(constants.ComponentIDCDC), string(constants.ComponentIDTiFlash):
			inspectParams, err := inspectInstanceByConfig(ctx, clusterInstance, compParams)
			if err != nil {
				return resp, err
			}
			inspectAllParameters = append(inspectAllParameters, inspectParams)
		default:
			return resp, fmt.Errorf(fmt.Sprintf("Component [%s] type is not supported", clusterInstance.Type))
		}
		return cluster.InspectParametersResp{Params: inspectAllParameters}, nil
	}

	// Cluster level inspection
	clusterMeta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		return resp, err
	}
	for comp, compParams := range compContainer {
		// Get component cluster instances
		instances := clusterMeta.Instances[comp]
		switch comp {
		case string(constants.ComponentIDTiDB):
			// Iterating over instances to get connection information
			for _, instance := range instances {
				inspectParams, err := inspectTiDBComponentInstance(ctx, instance, compParams)
				if err != nil {
					return resp, err
				}
				inspectAllParameters = append(inspectAllParameters, inspectParams)
			}
		case string(constants.ComponentIDTiKV):
			for _, instance := range instances {
				inspectParams, err := inspectTiKVComponentInstance(ctx, instance, compParams)
				if err != nil {
					return resp, err
				}
				inspectAllParameters = append(inspectAllParameters, inspectParams)
			}
		case string(constants.ComponentIDPD):
			for _, instance := range instances {
				inspectParams, err := inspectPDComponentInstance(ctx, instance, compParams)
				if err != nil {
					return resp, err
				}
				inspectAllParameters = append(inspectAllParameters, inspectParams)
			}
		case string(constants.ComponentIDCDC), string(constants.ComponentIDTiFlash):
			for _, instance := range instances {
				inspectParams, err := inspectInstanceByConfig(ctx, instance, compParams)
				if err != nil {
					return resp, err
				}
				inspectAllParameters = append(inspectAllParameters, inspectParams)
			}
		default:
			return resp, fmt.Errorf(fmt.Sprintf("Component [%s] type is not supported", comp))
		}
	}
	return cluster.InspectParametersResp{Params: inspectAllParameters}, nil
}

// inspectPDComponentInstance
// @Description: inspect pd component instance
// @Parameter ctx
// @Parameter instance
// @Parameter compParams
// @return inspectParams
// @return err
func inspectPDComponentInstance(ctx context.Context, instance *management.ClusterInstance, compParams []structs.ClusterParameterInfo) (inspectParams cluster.InspectParameters, err error) {
	if instance.Status == string(constants.ClusterInstanceRunning) {
		// pd api get instance parameters
		apiContent, err := pd.ApiService.ShowConfig(ctx, cluster.ApiShowConfigReq{
			InstanceHost: instance.HostIP[0],
			InstancePort: uint(instance.Ports[0]),
			Headers:      map[string]string{},
			Params:       map[string]string{},
		})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to call pd api show config, err = %s", err)
			return inspectParams, err
		}
		inspectParams, err = inspectInstancesByApiAndConfig(ctx, instance, apiContent, compParams)
		if err != nil {
			return inspectParams, err
		}
	}
	return inspectParams, nil
}

// inspectTiKVComponentInstance
// @Description: inspect tikv component instance
// @Parameter ctx
// @Parameter instance
// @Parameter compParams
// @return inspectParams
// @return err
func inspectTiKVComponentInstance(ctx context.Context, instance *management.ClusterInstance, compParams []structs.ClusterParameterInfo) (inspectParams cluster.InspectParameters, err error) {
	if instance.Status == string(constants.ClusterInstanceRunning) {
		// tikv api get instance parameters
		apiContent, err := tikv.ApiService.ShowConfig(ctx, cluster.ApiShowConfigReq{
			InstanceHost: instance.HostIP[0],
			InstancePort: uint(instance.Ports[1]),
			Headers:      map[string]string{},
			Params:       map[string]string{},
		})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to call tikv api show config, err = %s", err)
			return inspectParams, err
		}
		inspectParams, err = inspectInstancesByApiAndConfig(ctx, instance, apiContent, compParams)
		if err != nil {
			return inspectParams, err
		}
	}
	return inspectParams, nil
}

// inspectTiDBComponentInstance
// @Description: inspect tidb component instance
// @Parameter ctx
// @Parameter instance
// @Parameter compParams
// @return inspectParams
// @return err
func inspectTiDBComponentInstance(ctx context.Context, instance *management.ClusterInstance, compParams []structs.ClusterParameterInfo) (inspectParams cluster.InspectParameters, err error) {
	if instance.Status == string(constants.ClusterInstanceRunning) {
		// tidb api get instance parameters
		apiContent, err := tidbApi.ApiService.ShowConfig(ctx, cluster.ApiShowConfigReq{
			InstanceHost: instance.HostIP[0],
			InstancePort: uint(instance.Ports[1]),
			Headers:      map[string]string{},
			Params:       map[string]string{},
		})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to call %s api show config, err = %s", instance.Type, err)
			return inspectParams, err
		}
		inspectParams, err = inspectInstancesByApiAndConfig(ctx, instance, apiContent, compParams)
		if err != nil {
			return inspectParams, err
		}
	}
	return inspectParams, nil
}

// inspectInstanceByConfig
// @Description: inspect instances by config file
// @Parameter ctx
// @Parameter instance
// @Parameter compParams
// @return inspectParameter
// @return err
func inspectInstanceByConfig(ctx context.Context, instance *management.ClusterInstance, compParams []structs.ClusterParameterInfo) (inspectParameter cluster.InspectParameters, err error) {
	if instance.Status == string(constants.ClusterInstanceRunning) {
		// Define the set of parameters to store the inspection
		inspectParameterInfos := make([]cluster.InspectParameterInfo, 0)

		// Inspect diff config file parameter
		inspectConfigParams, err := inspectConfigParameter(ctx, instance, compParams)
		if err != nil {
			return inspectParameter, err
		}
		inspectParameterInfos = append(inspectParameterInfos, inspectConfigParams...)

		// filling data
		inspectParameter = cluster.InspectParameters{
			InstanceID:     instance.ID,
			InstanceType:   instance.Type,
			ParameterInfos: inspectParameterInfos,
		}
	}
	return inspectParameter, nil
}

// inspectInstancesByApiAndConfig
// @Description: inspect instances by api and config api
// @Parameter ctx
// @Parameter instance
// @Parameter apiContent
// @Parameter compParams
// @return inspectParams
// @return err
func inspectInstancesByApiAndConfig(ctx context.Context, instance *management.ClusterInstance, apiContent []byte, compParams []structs.ClusterParameterInfo) (inspectParams cluster.InspectParameters, err error) {
	// Define the set of parameters to store the inspection
	inspectParameterInfos := make([]cluster.InspectParameterInfo, 0)

	// Inspect api parameter
	framework.LogWithContext(ctx).Infof("request %s api show config response: %v", instance.Type, string(apiContent))
	inspectApiParams, instDiffParams, err := inspectApiParameter(ctx, instance, apiContent, compParams)
	if err != nil {
		return inspectParams, err
	}
	inspectParameterInfos = append(inspectParameterInfos, inspectApiParams...)

	// If the diff params are not empty, then read the parsed configuration file and determine if the values are equal
	if len(instDiffParams) > 0 {
		// Inspect diff config file parameter
		inspectConfigParams, err := inspectConfigParameter(ctx, instance, instDiffParams)
		if err != nil {
			return inspectParams, err
		}
		inspectParameterInfos = append(inspectParameterInfos, inspectConfigParams...)
	}
	// filling data
	inspectParams = cluster.InspectParameters{
		InstanceID:     instance.ID,
		InstanceType:   instance.Type,
		ParameterInfos: inspectParameterInfos,
	}
	return inspectParams, nil
}

// inspectApiParameter
// @Description: inspect api parameter
// @Parameter ctx
// @Parameter instance
// @Parameter apiContent
// @Parameter compParams
// @return inspectParams
// @return instDiffParams
// @return err
func inspectApiParameter(ctx context.Context, instance *management.ClusterInstance, apiContent []byte, compParams []structs.ClusterParameterInfo) (
	inspectParams []cluster.InspectParameterInfo, instDiffParams []structs.ClusterParameterInfo, err error) {
	reqApiParams := map[string]interface{}{}
	if err = json.Unmarshal(apiContent, &reqApiParams); err != nil {
		framework.LogWithContext(ctx).Errorf("failed to convert %s api parameters, err = %v", instance.Type, err)
		return inspectParams, instDiffParams, errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, "failed to convert %s api parameters, err = %v", instance.Type, err)
	}
	// Get api flattened parameter set
	flattenedApiParams, err := FlattenedParameters(reqApiParams)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to flattened %s api parameters, err = %v", instance.Type, err)
		return inspectParams, instDiffParams, errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, "failed to flattened %s api parameters, err = %v", instance.Type, err)
	}

	// Define the set of parameters to store that are not included in the api return
	instDiffParams = make([]structs.ClusterParameterInfo, 0)
	inspectParams = make([]cluster.InspectParameterInfo, 0)

	// Traversing metadata cluster parameters
	for _, param := range compParams {
		fullName := DisplayFullParameterName(param.Category, param.Name)
		// Determine whether the parameters of the api query contain the current metadata cluster parameters
		if _, ok := flattenedApiParams[fullName]; ok {
			inspectParameterInfos, err := inspectParameterValue(ctx, flattenedApiParams, param)
			if err != nil {
				return inspectParams, instDiffParams, err
			}
			inspectParams = append(inspectParams, inspectParameterInfos...)

		} else {
			// Parameters that do not include the api result set are then compared by the configuration file
			instDiffParams = append(instDiffParams, param)
		}
	}
	return inspectParams, instDiffParams, nil
}

// inspectConfigParameter
// @Description: inspect config parameter
// @Parameter ctx
// @Parameter instance
// @Parameter instDiffParams
// @return inspectParams
// @return err
func inspectConfigParameter(ctx context.Context, instance *management.ClusterInstance, instDiffParams []structs.ClusterParameterInfo) (inspectParams []cluster.InspectParameterInfo, err error) {
	inspectParams = make([]cluster.InspectParameterInfo, 0)
	// pull config
	configContentStr, err := pullConfig(ctx, instance.ClusterID, instance.Type, instance.GetDeployDir(), instance.HostIP[0])
	if err != nil {
		return inspectParams, err
	}
	reqConfigParams := map[string]interface{}{}
	if err := toml.Unmarshal([]byte(configContentStr), &reqConfigParams); err != nil {
		framework.LogWithContext(ctx).Errorf("failed to convert %s config file parameters, err = %v", instance.Type, err)
		return inspectParams, errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, "failed to convert %s config file parameters, err = %v", instance.Type, err)
	}
	// Get config file flattened parameter set
	flattenedConfigParams, err := FlattenedParameters(reqConfigParams)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to flattened %s config file parameters, err = %v", instance.Type, err)
		return inspectParams, errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, "failed to flattened %s config file parameters, err = %v", instance.Type, err)
	}

	for _, param := range instDiffParams {
		fullName := DisplayFullParameterName(param.Category, param.Name)
		// Determine whether the parameters of the config file query contain the current metadata cluster parameters
		if _, ok := flattenedConfigParams[fullName]; ok {
			inspectParameterInfos, err := inspectParameterValue(ctx, flattenedConfigParams, param)
			if err != nil {
				return inspectParams, err
			}
			inspectParams = append(inspectParams, inspectParameterInfos...)
		}
	}
	return inspectParams, nil
}

// pullConfig
// @Description: pull component config by tiup
// @Parameter ctx
// @Parameter clusterID
// @Parameter instanceType
// @Parameter deployDir
// @Parameter remoteHostIP
// @return string
// @return error
func pullConfig(ctx context.Context, clusterID, instanceType, deployDir, remoteHostIP string) (string, error) {
	// Calling the tiup pull interface
	configPath := fmt.Sprintf("%s/conf/%s.toml", deployDir, strings.ToLower(instanceType))
	framework.LogWithContext(ctx).Debugf("current instance type %s config path is %s", instanceType, configPath)

	// Call the pull command to get the instance configuration
	tiupHomeForTidb := framework.GetTiupHomePathForTidb()
	configContentStr, err := deployment.M.Pull(ctx, deployment.TiUPComponentTypeCluster, clusterID, configPath,
		tiupHomeForTidb, []string{"-N", remoteHostIP}, 0)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call secondparty tiup pull %s config err = %s", instanceType, err.Error())
		return "", err
	}
	return configContentStr, nil
}

// inspectParameterValue
// @Description: inspect parameter value
// @Parameter ctx
// @Parameter flattenedParams
// @Parameter param
// @return inspectParamInfos
// @return err
func inspectParameterValue(ctx context.Context, flattenedParams map[string]string, param structs.ClusterParameterInfo) (inspectParamInfos []cluster.InspectParameterInfo, err error) {
	inspectParamInfos = make([]cluster.InspectParameterInfo, 0)
	fullName := DisplayFullParameterName(param.Category, param.Name)
	instValue := flattenedParams[fullName]
	// If the value contains the unit, need to determine whether the units need to be replaced
	for srcUnit, replaceUnit := range replaceUnits {
		if strings.HasSuffix(instValue, srcUnit) {
			instValue = strings.ReplaceAll(instValue, srcUnit, replaceUnit)
			break
		}
	}

	// If the value is numeric, inst value if it is with units, then convert the units
	if param.Type == int(Integer) || param.Type == int(Float) {
		for srcUnit := range units {
			if strings.HasSuffix(instValue, srcUnit) {
				if cvtInstValue, ok := convertUnitValue([]string{srcUnit}, instValue); ok {
					instValue = fmt.Sprintf("%d", cvtInstValue)
					break
				}
			}
		}
	}
	// If the api parameter value is not equal to the metadata parameter set, add the result set
	if instValue != param.RealValue.ClusterValue {
		framework.LogWithContext(ctx).Debugf("inspect %s parameter `%s` cluster value: %s, inst value: %s", param.InstanceType, fullName, param.RealValue.ClusterValue, instValue)
		// If it is a string with units, then convert the units to determine if they are equal
		if param.Type == int(String) && param.RangeType == int(ContinuousRange) {
			cvtClusterValue, ok1 := convertUnitValue(param.UnitOptions, param.RealValue.ClusterValue)
			cvtInstValue, ok2 := convertUnitValue(param.UnitOptions, param.RealValue.ClusterValue)
			if ok1 && ok2 && cvtClusterValue == cvtInstValue {
				return inspectParamInfos, nil
			}
		}
		// If integer or float type, compatible with scientific notation, e.g.: 1.048576e+07
		if param.Type == int(Integer) || param.Type == int(Float) {
			cvtClusterValue, err1 := convertRealParameterType(ctx, param.Type, param.RealValue.ClusterValue)
			cvtInstValue, err2 := convertRealParameterType(ctx, param.Type, instValue)
			if err1 == nil && err2 == nil && cvtClusterValue == cvtInstValue {
				return inspectParamInfos, nil
			}
		}
		framework.LogWithContext(ctx).Infof("inspect %s parameter `%s` cluster value: %s, inst value: %s", param.InstanceType, fullName, param.RealValue.ClusterValue, instValue)
		inspectValue, err := convertRealParameterType(ctx, param.Type, instValue)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("convert real parameter type err = %s", err.Error())
			return inspectParamInfos, err
		}
		inspectParamInfos = append(inspectParamInfos, convertInspectParameterInfo(param, inspectValue))
	}
	return inspectParamInfos, nil
}

// convertInspectParameterInfo
// @Description: convert inspect parameter info
// @Parameter param
// @Parameter inspectValue
// @return cluster.InspectParameterInfo
func convertInspectParameterInfo(param structs.ClusterParameterInfo, inspectValue interface{}) cluster.InspectParameterInfo {
	return cluster.InspectParameterInfo{
		ParamID:        param.ParamId,
		Category:       param.Category,
		Name:           param.Name,
		SystemVariable: param.SystemVariable,
		Type:           param.Type,
		Unit:           param.Unit,
		UnitOptions:    param.UnitOptions,
		Range:          param.Range,
		RangeType:      param.RangeType,
		HasReboot:      param.HasReboot,
		ReadOnly:       param.ReadOnly,
		Description:    param.Description,
		Note:           param.Note,
		RealValue:      param.RealValue,
		InspectValue:   inspectValue,
	}
}
