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
		"start":                {"validationParameter", "validationDone", "fail", workflow.SyncFuncNode, validationParameter},
		"validationDone":       {"modifyParameter", "modifyDone", "fail", workflow.PollingNode, modifyParameters},
		"modifyDone":           {"refreshParameter", "refreshDone", "failRefreshParameter", workflow.PollingNode, refreshParameter},
		"refreshDone":          {"persistParameter", "persistDone", "fail", workflow.SyncFuncNode, persistParameter},
		"persistDone":          {"end", "", "", workflow.SyncFuncNode, defaultEnd},
		"fail":                 {"end", "", "", workflow.SyncFuncNode, defaultEnd},
		"failRefreshParameter": {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(refreshParameterFail, defaultEnd)},
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

func (m *Manager) InspectClusterParameters(ctx context.Context, req cluster.InspectParametersReq) (resp cluster.InspectParametersResp, err error) {
	// Instance level inspection
	if req.InstanceID != "" {

		return
	}
	if req.ClusterID == "" {
		return resp, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "both clusterID and instanceID are empty")
	}

	// Cluster level inspection
	params, _, err := m.QueryClusterParameters(ctx, cluster.QueryClusterParametersReq{ClusterID: req.ClusterID})
	if err != nil {
		return resp, err
	}

	// Define the component instance parameter container
	compContainer := make(map[string][]structs.ClusterParameterInfo)
	for _, param := range params.Params {
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

	clusterMeta, err := meta.Get(ctx, req.ClusterID)
	if err != nil {
		return resp, err
	}

	// Define the set that holds the values of the inspection parameters
	inspectParameters := make([]cluster.InspectParameters, 0)

	for comp, compParams := range compContainer {
		switch comp {
		case string(constants.ComponentIDTiDB):
			// Get tidb cluster instances
			instances := clusterMeta.Instances[string(constants.ComponentIDTiDB)]
			// Iterating over instances to get connection information
			for _, instance := range instances {
				if instance.Status == string(constants.ClusterInstanceRunning) {
					// tidb api get instance parameters
					apiContent, err := tidbApi.ApiService.ShowConfig(ctx, cluster.ApiShowConfigReq{
						InstanceHost: instance.HostIP[0],
						InstancePort: uint(instance.Ports[1]),
						Headers:      map[string]string{},
						Params:       map[string]string{},
					})
					if err != nil {
						framework.LogWithContext(ctx).Errorf("failed to call tidb api show config, err = %s", err)
						return resp, err
					}
					inspectParams, err := inspectInstances(ctx, instance, apiContent, compParams)
					if err != nil {
						return resp, err
					}
					inspectParameters = append(inspectParameters, inspectParams...)
				}
			}
		case string(constants.ComponentIDTiKV):
			// Get tikv cluster instances
			instances := clusterMeta.Instances[string(constants.ComponentIDTiKV)]
			// Iterating over instances to get connection information
			for _, instance := range instances {
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
						return resp, err
					}
					inspectParams, err := inspectInstances(ctx, instance, apiContent, compParams)
					if err != nil {
						return resp, err
					}
					inspectParameters = append(inspectParameters, inspectParams...)
				}
			}
		case string(constants.ComponentIDPD):
			// Get pd cluster instances
			instances := clusterMeta.Instances[string(constants.ComponentIDPD)]
			// Iterating over instances to get connection information
			for _, instance := range instances {
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
						return resp, err
					}
					inspectParams, err := inspectInstances(ctx, instance, apiContent, compParams)
					if err != nil {
						return resp, err
					}
					inspectParameters = append(inspectParameters, inspectParams...)
				}
			}
		case string(constants.ComponentIDCDC):
			// Get cdc cluster instances
			instances := clusterMeta.Instances[string(constants.ComponentIDCDC)]
			inspectParams, err := inspectInstancesByConfig(ctx, instances, compParams)
			if err != nil {
				return resp, err
			}
			inspectParameters = append(inspectParameters, inspectParams...)
		case string(constants.ComponentIDTiFlash):
			// Get tiflash cluster instances
			instances := clusterMeta.Instances[string(constants.ComponentIDTiFlash)]
			inspectParams, err := inspectInstancesByConfig(ctx, instances, compParams)
			if err != nil {
				return resp, err
			}
			inspectParameters = append(inspectParameters, inspectParams...)
		default:
			return resp, fmt.Errorf(fmt.Sprintf("Component [%s] type is not supported", comp))
		}
	}
	resp = cluster.InspectParametersResp{Params: inspectParameters}
	return resp, nil
}

func inspectInstances(ctx context.Context, instance *management.ClusterInstance, apiContent []byte, compParams []structs.ClusterParameterInfo) (inspectParams []cluster.InspectParameters, err error) {
	inspectParams = make([]cluster.InspectParameters, 0)

	// Define the set of parameters to store the inspection
	inspectParameterInfos := make([]cluster.InspectParameterInfo, 0)

	// Inspect api parameter
	inspectApiParams, instDiffParams, err := inspectApiParameter(ctx, instance, apiContent, compParams)
	if err != nil {
		return nil, err
	}
	inspectParameterInfos = append(inspectParameterInfos, inspectApiParams...)

	// If the diff params are not empty, then read the parsed configuration file and determine if the values are equal
	if len(instDiffParams) > 0 {
		// Inspect diff config file parameter
		inspectConfigParams, err := inspectConfigParameter(ctx, instance, instDiffParams)
		if err != nil {
			return nil, err
		}
		inspectParameterInfos = append(inspectParameterInfos, inspectConfigParams...)
	}
	// filling data
	inspectParams = append(inspectParams, cluster.InspectParameters{
		InstanceID:     instance.ID,
		InstanceType:   instance.Type,
		ParameterInfos: inspectParameterInfos,
	})
	return inspectParams, nil
}

func inspectInstancesByConfig(ctx context.Context, instances []*management.ClusterInstance, compParams []structs.ClusterParameterInfo) (inspectParams []cluster.InspectParameters, err error) {
	inspectParams = make([]cluster.InspectParameters, 0)
	// Iterating over instances to get connection information
	for _, instance := range instances {
		if instance.Status == string(constants.ClusterInstanceRunning) {
			// Define the set of parameters to store the inspection
			inspectParameterInfos := make([]cluster.InspectParameterInfo, 0)

			// Inspect diff config file parameter
			inspectConfigParams, err := inspectConfigParameter(ctx, instance, compParams)
			if err != nil {
				return inspectParams, err
			}
			inspectParameterInfos = append(inspectParameterInfos, inspectConfigParams...)

			inspectParams = make([]cluster.InspectParameters, 0)
			// filling data
			inspectParams = append(inspectParams, cluster.InspectParameters{
				InstanceID:     instance.ID,
				InstanceType:   instance.Type,
				ParameterInfos: inspectParameterInfos,
			})
		}
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
			instValue := flattenedApiParams[fullName]
			// If the api parameter value is not equal to the metadata parameter set, add the result set
			if instValue != param.RealValue.ClusterValue {
				inspectValue, err := convertRealParameterType(ctx, param.Type, instValue)
				if err != nil {
					framework.LogWithContext(ctx).Errorf("convert real parameter type err = %s", err.Error())
					return inspectParams, instDiffParams, err
				}
				inspectParams = append(inspectParams, convertInspectParameterInfo(param, inspectValue))
			}
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

	// Calling the tiup pull interface
	configPath := fmt.Sprintf("%s/conf/%s.toml", instance.GetDeployDir(), strings.ToLower(instance.Type))
	framework.LogWithContext(ctx).Debugf("current instance type %s config path is %s", instance.Type, configPath)

	// Call the pull command to get the instance configuration
	configContentStr, err := deployment.M.Pull(ctx, deployment.TiUPComponentTypeCluster, instance.ClusterID, configPath,
		"/home/tiem/.tiup", []string{"-N", instance.HostIP[0]}, 0)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call secondparty tiup pull %s config err = %s", instance.Type, err.Error())
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
			instValue := flattenedConfigParams[fullName]
			// If the config file parameter value is not equal to the metadata parameter set, add the result set
			if instValue != param.RealValue.ClusterValue {
				inspectValue, err := convertRealParameterType(ctx, param.Type, instValue)
				if err != nil {
					framework.LogWithContext(ctx).Errorf("convert real parameter type err = %s", err.Error())
					return inspectParams, err
				}
				inspectParams = append(inspectParams, convertInspectParameterInfo(param, inspectValue))
			}
		}
	}
	return inspectParams, nil
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
