/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
 * @File: executor.go
 * @Description: flow task executor
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/15 17:08
*******************************************************************************/

package parameter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/shopspring/decimal"

	"github.com/pingcap/tiunimanager/deployment"
	"gopkg.in/yaml.v2"

	"github.com/pingcap/tiunimanager/util/api/cdc"

	"github.com/pingcap/tiunimanager/util/api/pd"

	"github.com/pingcap/tiunimanager/util/api/tikv"

	"github.com/pingcap/tiunimanager/message/cluster"

	tidbApi "github.com/pingcap/tiunimanager/util/api/tidb/http"

	"github.com/pingcap/tiunimanager/util/api/tidb/sql"

	"github.com/pingcap/tiunimanager/common/errors"

	"github.com/pingcap/tiunimanager/common/constants"

	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/cluster/parameter"

	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"

	"github.com/pingcap/tiunimanager/library/framework"

	spec2 "github.com/pingcap/tiunimanager/library/spec"
	workflowModel "github.com/pingcap/tiunimanager/models/workflow"
	workflow "github.com/pingcap/tiunimanager/workflow2"
	tiupSpec "github.com/pingcap/tiup/pkg/cluster/spec"
)

// asyncMaintenance
// @Description: asynchronous process for cluster maintenance
// @Parameter ctx
// @Parameter meta
// @Parameter data
// @Parameter status
// @Parameter flowName
// @return flowID
// @return err
func asyncMaintenance(ctx context.Context, meta *meta.ClusterMeta, data map[string]interface{}, status constants.ClusterMaintenanceStatus, flowName string) (flowID string, err error) {
	// condition maintenance status change
	if data[contextMaintenanceStatusChange].(bool) {
		if err = meta.StartMaintenance(ctx, status); err != nil {
			framework.LogWithContext(ctx).Errorf("start maintenance failed, clusterID = %s, status = %s,error = %s", meta.Cluster.ID, status, err.Error())
			return
		}
	}

	if flowId, flowError := workflow.GetWorkFlowService().CreateWorkFlow(ctx, meta.Cluster.ID, workflow.BizTypeCluster, flowName); flowError != nil {
		framework.LogWithContext(ctx).Errorf("create flow failed, clusterID = %s, error = %s", meta.Cluster.ID, err)
		err = flowError
		return
	} else {
		flowID = flowId
		workflow.GetWorkFlowService().InitContext(ctx, flowId, contextClusterMeta, meta)
		for k, v := range data {
			workflow.GetWorkFlowService().InitContext(ctx, flowId, k, v)
		}
		if err = workflow.GetWorkFlowService().Start(ctx, flowId); err != nil {
			framework.LogWithContext(ctx).Errorf("start flow %s failed, clusterID = %s, error = %s", flowName, meta.Cluster.ID, err.Error())
			return
		}
		framework.LogWithContext(ctx).Infof("create flow %s succeed, clusterID = %s", flowName, meta.Cluster.ID)
	}
	return
}

// defaultEnd
// @Description: clear maintenance status after maintenance finished or failed
func defaultEnd(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin default end executor method")
	defer framework.LogWithContext(ctx).Info("end default end executor method")

	var clusterMeta meta.ClusterMeta
	err := ctx.GetData(contextClusterMeta, &clusterMeta)
	if err != nil {
		return err
	}
	var maintenanceStatusChange bool
	err = ctx.GetData(contextMaintenanceStatusChange, &maintenanceStatusChange)
	if err != nil {
		return err
	}
	if maintenanceStatusChange {
		if err := clusterMeta.EndMaintenance(ctx, clusterMeta.Cluster.MaintenanceStatus); err != nil {
			framework.LogWithContext(ctx).Errorf("end cluster %s maintenance status failed, %s", clusterMeta.Cluster.ID, err.Error())
			return err
		}
		ctx.SetData(contextClusterMeta, &clusterMeta)
	}
	return nil
}

// parameterFail
// @Description: Rollback logic for default failures
func parameterFail(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin parameter fail executor method")
	defer framework.LogWithContext(ctx).Info("end parameter fail executor method")

	// Get tiup show-config result
	var clusterConfigStr string
	err := ctx.GetData(contextClusterConfigStr, &clusterConfigStr)
	if err != nil {
		return err
	}
	if clusterConfigStr == "" {
		return nil
	}

	var clusterMeta meta.ClusterMeta
	err = ctx.GetData(contextClusterMeta, &clusterMeta)
	if err != nil {
		return err
	}
	tiupHomeForTidb := framework.GetTiupHomePathForTidb()
	// If the reload fails, then do a meta rollback
	taskId, err := deployment.M.EditConfig(ctx, deployment.TiUPComponentTypeCluster, clusterMeta.Cluster.ID,
		clusterConfigStr, tiupHomeForTidb, node.ParentID, []string{}, 0)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call secondparty tiup rollback global config task id = %v, err = %s", taskId, err.Error())
		return err
	}
	return nil
}

// persistParameter
// @Description: persist parameter
// @Parameter node
// @Parameter ctx
// @return error
func persistParameter(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin persist parameter executor method")
	defer framework.LogWithContext(ctx).Info("end persist parameter executor method")

	var modifyParam ModifyParameter
	err := ctx.GetData(contextModifyParameters, &modifyParam)
	if err != nil {
		return err
	}
	params := make([]*parameter.ClusterParameterMapping, len(modifyParam.Params))
	for i, param := range modifyParam.Params {
		b, err := json.Marshal(param.RealValue)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to convert parameter real value err: %v", err)
			return errors.NewErrorf(errors.TIUNIMANAGER_CONVERT_OBJ_FAILED, errors.TIUNIMANAGER_CONVERT_OBJ_FAILED.Explain())
		}
		params[i] = &parameter.ClusterParameterMapping{
			ClusterID:   modifyParam.ClusterID,
			ParameterID: param.ParamId,
			RealValue:   string(b),
		}
	}

	// Get the apply parameter object
	var hasApplyParameter bool
	err = ctx.GetData(contextHasApplyParameter, &hasApplyParameter)
	if err != nil {
		return err
	}
	if hasApplyParameter {
		framework.LogWithContext(ctx).Infof("current has apply parameter: %v", hasApplyParameter)
		// persist apply parameter
		err := models.GetClusterParameterReaderWriter().ApplyClusterParameter(ctx, modifyParam.ParamGroupId, modifyParam.ClusterID, params)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("apply parameter group convert resp err: %v", err)
			return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_GROUP_APPLY_ERROR, err.Error())
		}
	} else {
		// persist update parameter
		err := models.GetClusterParameterReaderWriter().UpdateClusterParameter(ctx, modifyParam.ClusterID, params)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("update cluster parameter err: %v", err)
			return errors.NewErrorf(errors.TIUNIMANAGER_CLUSTER_PARAMETER_UPDATE_ERROR, errors.TIUNIMANAGER_CLUSTER_PARAMETER_UPDATE_ERROR.Explain(), err)
		}
	}
	return nil
}

// validationParameter
// @Description: validation parameters
// @Parameter node
// @Parameter ctx
// @return error
func validationParameter(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin validation parameters executor method")
	defer framework.LogWithContext(ctx).Info("end validation parameters executor method")

	var modifyParam ModifyParameter
	err := ctx.GetData(contextModifyParameters, &modifyParam)
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Debugf("got validation parameters size: %d", len(modifyParam.Params))

	for _, param := range modifyParam.Params {
		// validate parameter value by range field
		if !ValidateRange(param, true) {
			if param.RangeType == int(ContinuousRange) && len(param.Range) == 2 {
				return fmt.Errorf(fmt.Sprintf("Validation parameter `%s` failed, update value: %s, can take a range of values: %v",
					DisplayFullParameterName(param.Category, param.Name), param.RealValue.ClusterValue, param.Range))
			} else {
				return fmt.Errorf(fmt.Sprintf("Validation parameter `%s` failed, update value: %s, optional values: %v",
					DisplayFullParameterName(param.Category, param.Name), param.RealValue.ClusterValue, param.Range))
			}
		}
	}
	node.Record("validate parameters")
	return nil
}

// modifyParameters
// @Description: modify parameters
// @Parameter node
// @Parameter ctx
// @return error
func modifyParameters(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin modify parameters executor method")
	defer framework.LogWithContext(ctx).Info("end modify parameters executor method")
	var clusterMeta meta.ClusterMeta
	err := ctx.GetData(contextClusterMeta, &clusterMeta)
	if err != nil {
		return err
	}

	var modifyParam ModifyParameter
	err = ctx.GetData(contextModifyParameters, &modifyParam)
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Debugf("got modify need reboot: %v, parameters size: %d", modifyParam.Reboot, len(modifyParam.Params))
	var maintenanceStatusChange bool
	err = ctx.GetData(contextMaintenanceStatusChange, &maintenanceStatusChange)
	if err != nil {
		return err
	}

	// Get the apply parameter object
	var applyParameter bool
	err = ctx.GetData(contextHasApplyParameter, &applyParameter)
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Debugf("modify parameter get apply parameter: %v", applyParameter)

	// Define variables to determine if polling for results is required
	hasPolling := false

	// fill param grouping by instance type
	fillParamContainer := make(map[interface{}][]*ModifyClusterParameterInfo)
	// grouping by parameter source
	modifyParamContainer := make(map[interface{}][]*ModifyClusterParameterInfo)
	for i, param := range modifyParam.Params {
		if param.InstanceType == string(constants.ComponentIDCDC) && len(clusterMeta.GetCDCClientAddresses()) == 0 {
			// If it is a parameter of CDC, apply the parameter without installing CDC, then skip directly
			if applyParameter {
				// The real value is set to an unknown empty value
				param.RealValue.ClusterValue = ""
				continue
			} else {
				return fmt.Errorf("get %s address from meta failed, empty address", constants.ComponentIDCDC)
			}
		}
		if param.InstanceType == string(constants.ComponentIDTiFlash) && len(clusterMeta.GetTiFlashClientAddresses()) == 0 {
			// If it is a parameter of TiFlash, apply the parameter without installing TiFlash, then skip directly
			if applyParameter {
				// The real value is set to an unknown empty value
				param.RealValue.ClusterValue = ""
				continue
			} else {
				return fmt.Errorf("get %s address from meta failed, empty address", constants.ComponentIDTiFlash)
			}
		}
		// condition apply parameter and HasApply values is 0, then filter directly
		if applyParameter && param.HasApply != int(DirectApply) {
			putParameterContainer(fillParamContainer, param.InstanceType, param)
			continue
		}
		// If it is an apply parameter with an empty parameter value, it is skipped directly
		if applyParameter && strings.TrimSpace(param.RealValue.ClusterValue) == "" {
			continue
		}
		// If the parameter is modified and is triggered by another workflow and the parameter value is empty, then skip directly
		if !applyParameter && !maintenanceStatusChange && strings.TrimSpace(param.RealValue.ClusterValue) == "" {
			continue
		}

		// If the parameters are modified, read-only parameters are not allowed to be modified
		if !applyParameter && param.ReadOnly == int(ReadOnly) {
			return fmt.Errorf(fmt.Sprintf("Read-only parameters `%s` are not allowed to be modified", DisplayFullParameterName(param.Category, param.Name)))
		}
		framework.LogWithContext(ctx).Debugf("loop %d modify param name: %v, cluster value: %v", i, param.Name, param.RealValue.ClusterValue)
		if param.UpdateSource == int(TiUPAndSQL) {
			// condition UpdateSource values is 2, then insert tiup and sql respectively
			hasPolling = true
			putParameterContainer(modifyParamContainer, int(TiUP), param)
			putParameterContainer(modifyParamContainer, int(SQL), param)
		} else if param.UpdateSource == int(TiUPAndAPI) {
			// condition UpdateSource values is 4, then insert tiup and api respectively
			hasPolling = true
			putParameterContainer(modifyParamContainer, int(TiUP), param)
			putParameterContainer(modifyParamContainer, int(API), param)
		} else {
			if param.UpdateSource == int(TiUP) {
				hasPolling = true
			}
			putParameterContainer(modifyParamContainer, param.UpdateSource, param)
		}
		node.Record(fmt.Sprintf("modify parameter `%s` in %s to %s; ", DisplayFullParameterName(param.Category, param.Name), param.InstanceType, param.RealValue.ClusterValue))
	}

	// If polling is not needed, call node.Success() to terminate workflow polling
	if !hasPolling {
		node.Success()
	}

	for source, params := range modifyParamContainer {
		framework.LogWithContext(ctx).Debugf("loop current param container source: %v, params size: %d", source, len(params))
		switch source.(int) {
		case int(TiUP):
			if err := tiupEditConfig(ctx, node, params); err != nil {
				return err
			}
		case int(SQL):
			if err := sqlEditConfig(ctx, node, params); err != nil {
				return err
			}
		case int(API):
			if err := apiEditConfig(ctx, node, params); err != nil {
				return err
			}
		}
	}

	// If it is an apply parameter, get the parameter value that is not directly applied to populate the meta database
	// Mainly some parameter sets generated according to the example environment
	if applyParameter {
		if err := fillParameters(ctx, fillParamContainer); err != nil {
			return err
		}
	}
	ctx.SetData(contextModifyParameters, &modifyParam)
	node.Record("modify parameters")
	return nil
}

// fillParameters
// @Description: fill parameters. Mainly some parameter sets generated according to the example environment
// @Parameter ctx
// @Parameter fillParamContainer
// @return error
func fillParameters(ctx *workflow.FlowContext, fillParamContainer map[interface{}][]*ModifyClusterParameterInfo) error {
	var clusterMeta meta.ClusterMeta
	err := ctx.GetData(contextClusterMeta, &clusterMeta)
	if err != nil {
		return err
	}

	for instanceType, params := range fillParamContainer {
		if err := fillParameter(ctx, instanceType.(string), &clusterMeta, params); err != nil {
			return err
		}
	}
	return nil
}

func fillParameter(ctx context.Context, componentID string, clusterMeta *meta.ClusterMeta, params []*ModifyClusterParameterInfo) error {
	switch componentID {
	case string(constants.ComponentIDTiDB):
		tidbServers := clusterMeta.GetClusterStatusAddress()
		if len(tidbServers) == 0 {
			return fmt.Errorf("get tidb status address from meta failed, empty address")
		}
		// api edit config
		apiContent, err := tidbApi.ApiService.ShowConfig(ctx, cluster.ApiShowConfigReq{
			InstanceHost: tidbServers[0].IP,
			InstancePort: uint(tidbServers[0].Port),
			Headers:      map[string]string{},
		})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to call %s api show config, err = %s", componentID, err)
			return err
		}
		// handle fill parameter value
		if err := handleApiFillParams(ctx, apiContent, componentID, params); err != nil {
			return err
		}
	case string(constants.ComponentIDTiKV):
		tikvServers := clusterMeta.GetTiKVStatusAddress()
		if len(tikvServers) == 0 {
			return fmt.Errorf("get tikv status address from meta failed, empty address")
		}
		// api edit config
		apiContent, err := tikv.ApiService.ShowConfig(ctx, cluster.ApiShowConfigReq{
			InstanceHost: tikvServers[0].IP,
			InstancePort: uint(tikvServers[0].Port),
			Headers:      map[string]string{},
		})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to call %s api show config, err = %s", componentID, err)
			return err
		}
		// handle fill parameter value
		if err := handleApiFillParams(ctx, apiContent, componentID, params); err != nil {
			return err
		}
	case string(constants.ComponentIDPD):
		pdServers := clusterMeta.GetPDClientAddresses()
		if len(pdServers) == 0 {
			return fmt.Errorf("get pd status address from meta failed, empty address")
		}
		// api edit config
		apiContent, err := pd.ApiService.ShowConfig(ctx, cluster.ApiShowConfigReq{
			InstanceHost: pdServers[0].IP,
			InstancePort: uint(pdServers[0].Port),
			Headers:      map[string]string{},
		})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to call %s api show config, err = %s", componentID, err)
			return err
		}
		// handle fill parameter value
		if err := handleApiFillParams(ctx, apiContent, componentID, params); err != nil {
			return err
		}
	case string(constants.ComponentIDCDC), string(constants.ComponentIDTiFlash):
		// Get component cluster instances
		instances := clusterMeta.Instances[componentID]
		if len(instances) > 0 {
			// pull config
			configContentStr, err := pullConfig(ctx, instances[0].ClusterID, instances[0].Type, instances[0].GetDeployDir(), instances[0].HostIP[0])
			if err != nil {
				framework.LogWithContext(ctx).Errorf("failed to call %s pull show config, err = %s", componentID, err)
				return err
			}
			// handle fill parameter value
			if err := handleConfigFillParams(ctx, []byte(configContentStr), componentID, params); err != nil {
				return err
			}
		}
	}
	return nil
}

// handleApiFillParams
// @Description: handle api fill parameters value
// @Parameter ctx
// @Parameter content
// @Parameter instanceType
// @Parameter params
// @return err
func handleApiFillParams(ctx context.Context, content []byte, instanceType string, params []*ModifyClusterParameterInfo) (err error) {
	reqApiParams := map[string]interface{}{}
	d := json.NewDecoder(bytes.NewReader(content))
	d.UseNumber()
	if err = d.Decode(&reqApiParams); err != nil {
		framework.LogWithContext(ctx).Errorf("failed to convert %s api parameters, err = %v", instanceType, err)
		return errors.NewErrorf(errors.TIUNIMANAGER_CONVERT_OBJ_FAILED, "failed to convert %s api parameters, err = %v", instanceType, err)
	}
	// Get api flattened parameter set
	return handleFillParamResult(ctx, reqApiParams, instanceType, params)
}

// handleConfigFillParams
// @Description: handle config fill parameters value
// @Parameter ctx
// @Parameter content
// @Parameter instanceType
// @Parameter params
// @return err
func handleConfigFillParams(ctx context.Context, content []byte, instanceType string, params []*ModifyClusterParameterInfo) (err error) {
	reqConfigParams := map[string]interface{}{}
	if err = toml.Unmarshal(content, &reqConfigParams); err != nil {
		framework.LogWithContext(ctx).Errorf("failed to convert %s config parameters, err = %v", instanceType, err)
		return errors.NewErrorf(errors.TIUNIMANAGER_CONVERT_OBJ_FAILED, "failed to convert %s config parameters, err = %v", instanceType, err)
	}
	// Get config flattened parameter set
	return handleFillParamResult(ctx, reqConfigParams, instanceType, params)
}

// handleFillParamResult
// @Description: handle fill parameter result
// @Parameter ctx
// @Parameter reqConfigParams
// @Parameter instanceType
// @Parameter params
// @return err
func handleFillParamResult(ctx context.Context, reqConfigParams map[string]interface{}, instanceType string, params []*ModifyClusterParameterInfo) (err error) {
	flattenedParams, err := FlattenedParameters(reqConfigParams)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to flattened %s parameters, err = %v", instanceType, err)
		return errors.NewErrorf(errors.TIUNIMANAGER_CONVERT_OBJ_FAILED, "failed to flattened %s parameters, err = %v", instanceType, err)
	}
	for _, param := range params {
		fullName := DisplayFullParameterName(param.Category, param.Name)
		if _, ok := flattenedParams[fullName]; ok {
			instValue := flattenedParams[fullName]
			// If the value contains the unit, need to determine whether the units need to be replaced
			for srcUnit, replaceUnit := range replaceUnits {
				if strings.HasSuffix(instValue, srcUnit) {
					instValue = strings.ReplaceAll(instValue, srcUnit, replaceUnit)
					break
				}
			}
			if param.Type == int(Integer) || param.Type == int(Float) {
				// If the value of the numeric type contains units, the units need to be converted to the int value of the base unit
				for srcUnit := range units {
					if strings.HasSuffix(instValue, srcUnit) {
						if cvtInstValue, ok := convertUnitValue([]string{srcUnit}, instValue); ok {
							instValue = fmt.Sprintf("%d", cvtInstValue)
							break
						}
					}
				}
				// If integer or float type, compatible with scientific notation, e.g.: 1.048576e+07
				cvtInstValue, err := convertRealParameterType(ctx, param.Type, instValue)
				if err != nil {
					return err
				}
				instValue = fmt.Sprintf("%v", cvtInstValue)
			}
			param.RealValue.ClusterValue = instValue
		}
	}
	return nil
}

// sqlEditConfig
// @Description: through sql edit config
// @Parameter ctx
// @Parameter params
// @return error
func sqlEditConfig(ctx *workflow.FlowContext, node *workflowModel.WorkFlowNode, params []*ModifyClusterParameterInfo) error {
	framework.LogWithContext(ctx).Info("begin sql edit config executor method")
	defer framework.LogWithContext(ctx).Info("end sql edit config executor method")

	configs := make([]sql.ClusterComponentConfig, len(params))
	for i, param := range params {
		configKey := param.Name
		// set config key from system variable
		if param.SystemVariable != "" {
			configKey = param.SystemVariable
		}
		configs[i] = sql.ClusterComponentConfig{
			TiDBClusterComponent: spec2.TiDBClusterComponent(strings.ToLower(param.InstanceType)),
			ConfigKey:            configKey,
			ConfigValue:          param.RealValue.ClusterValue,
		}
	}

	var clusterMeta meta.ClusterMeta
	err := ctx.GetData(contextClusterMeta, &clusterMeta)
	if err != nil {
		return err
	}
	tidbServers := clusterMeta.GetClusterConnectAddresses()
	if len(tidbServers) == 0 {
		framework.LogWithContext(ctx).Errorf("get tidb connect address from meta failed, empty address")
		return fmt.Errorf("get tidb connect address from meta failed, empty address")
	}
	tidbServer := tidbServers[rand.Intn(len(tidbServers))]
	framework.LogWithContext(ctx).Infof("get cluster [%s] tidb server from meta, %+v", clusterMeta.Cluster.ID, tidbServers)

	tidbUserInfo, err := clusterMeta.GetDBUserNamePassword(ctx, constants.DBUserParameterManagement)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get cluster %s user info from meta falied, %s ", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	if tidbUserInfo == nil {
		framework.LogWithContext(ctx).Errorf("get cluster [%s] user info from meta", clusterMeta.Cluster.ID)
		return fmt.Errorf("get cluster user name from meta failed, empty address")
	}

	req := sql.ClusterEditConfigReq{
		DbConnParameter: sql.DbConnParam{
			Username: tidbUserInfo.Name,
			Password: tidbUserInfo.Password.Val,
			IP:       tidbServer.IP,
			Port:     strconv.Itoa(tidbServer.Port),
		},
		ComponentConfigs: configs,
	}
	err = sql.SqlService.EditClusterConfig(ctx, req, node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call secondparty sql edit cluster config err = %s", err.Error())
		return err
	}
	return nil
}

// apiEditConfig
// @Description: through cluster component api edit config
// @Parameter ctx
// @Parameter params
// @return error
func apiEditConfig(ctx *workflow.FlowContext, node *workflowModel.WorkFlowNode, params []*ModifyClusterParameterInfo) error {
	framework.LogWithContext(ctx).Info("begin api edit config executor method")
	defer framework.LogWithContext(ctx).Info("end api edit config executor method")

	compContainer := make(map[interface{}][]*ModifyClusterParameterInfo)
	for i, param := range params {
		framework.LogWithContext(ctx).Debugf("loop %d api componet type: %v, param name: %v", i, param.InstanceType, param.Name)
		putParameterContainer(compContainer, param.InstanceType, param)
	}
	if len(compContainer) > 0 {
		for comp, params := range compContainer {
			cm := map[string]interface{}{}
			for _, param := range params {
				clusterValue, err := convertRealParameterType(ctx, param.Type, param.RealValue.ClusterValue)
				if err != nil {
					framework.LogWithContext(ctx).Errorf("convert real parameter type err = %v", err)
					return err
				}
				configKey := DisplayFullParameterName(param.Category, param.Name)
				// If system variable not empty, set config key from system variable.
				if param.SystemVariable != "" {
					configKey = param.SystemVariable
				}
				// display full parameter name
				cm[configKey] = clusterValue
			}
			var clusterMeta meta.ClusterMeta
			err := ctx.GetData(contextClusterMeta, &clusterMeta)
			if err != nil {
				return err
			}

			switch comp.(string) {
			case string(constants.ComponentIDTiDB):
				tidbServers := clusterMeta.GetClusterStatusAddress()
				if len(tidbServers) == 0 {
					framework.LogWithContext(ctx).Errorf("get tidb status address from meta failed, empty address")
					return fmt.Errorf("get tidb status address from meta failed, empty address")
				}
				for _, server := range tidbServers {
					// api edit config
					hasSuc, err := tidbApi.ApiService.EditConfig(ctx, cluster.ApiEditConfigReq{
						InstanceHost: server.IP,
						InstancePort: uint(server.Port),
						Headers:      map[string]string{},
						ConfigMap:    cm,
					})
					if err != nil || !hasSuc {
						framework.LogWithContext(ctx).Errorf("call secondparty api edit config is %v, err = %s", hasSuc, err)
						return err
					}
				}
			case string(constants.ComponentIDTiKV):
				tikvServers := clusterMeta.GetTiKVStatusAddress()
				if len(tikvServers) == 0 {
					framework.LogWithContext(ctx).Errorf("get tikv address from meta failed, empty address")
					return fmt.Errorf("get tikv address from meta failed, empty address")
				}
				for _, server := range tikvServers {
					// api edit config
					hasSuc, err := tikv.ApiService.EditConfig(ctx, cluster.ApiEditConfigReq{
						InstanceHost: server.IP,
						InstancePort: uint(server.Port),
						Headers:      map[string]string{},
						ConfigMap:    cm,
					})
					if err != nil || !hasSuc {
						framework.LogWithContext(ctx).Errorf("call secondparty api edit config is %v, err = %s", hasSuc, err)
						return err
					}
				}
			case string(constants.ComponentIDPD):
				pdServers := clusterMeta.GetPDClientAddresses()
				if len(pdServers) == 0 {
					framework.LogWithContext(ctx).Errorf("get pd address from meta failed, empty address")
					return fmt.Errorf("get pd address from meta failed, empty address")
				}
				server := pdServers[rand.Intn(len(pdServers))]
				// api edit config
				hasSuc, err := pd.ApiService.EditConfig(ctx, cluster.ApiEditConfigReq{
					InstanceHost: server.IP,
					InstancePort: uint(server.Port),
					Headers:      map[string]string{},
					ConfigMap:    cm,
				})
				if err != nil || !hasSuc {
					framework.LogWithContext(ctx).Errorf("call secondparty api edit config is %v, err = %s", hasSuc, err)
					return err
				}
			case string(constants.ComponentIDCDC):
				cdcServers := clusterMeta.GetCDCClientAddresses()
				if len(cdcServers) == 0 {
					framework.LogWithContext(ctx).Errorf("get cdc address from meta failed, empty address")
					return fmt.Errorf("get cdc address from meta failed, empty address")
				}
				server := cdcServers[rand.Intn(len(cdcServers))]
				// api edit config
				hasSuc, err := cdc.ApiService.EditConfig(ctx, cluster.ApiEditConfigReq{
					InstanceHost: server.IP,
					InstancePort: uint(server.Port),
					Headers:      map[string]string{},
					ConfigMap:    cm,
				})
				if err != nil || !hasSuc {
					framework.LogWithContext(ctx).Errorf("call secondparty api edit config is %v, err = %s", hasSuc, err)
					return err
				}
			default:
				return fmt.Errorf(fmt.Sprintf("Component [%s] type modification is not supported", comp.(string)))
			}
		}
	}
	return nil
}

// tiupEditConfig
// @Description: through tiup edit config
// @Parameter ctx
// @Parameter node
// @Parameter params
// @return error
func tiupEditConfig(ctx *workflow.FlowContext, node *workflowModel.WorkFlowNode, params []*ModifyClusterParameterInfo) error {
	framework.LogWithContext(ctx).Info("begin tiup edit config executor method")
	defer framework.LogWithContext(ctx).Info("end tiup edit config executor method")

	var clusterMeta meta.ClusterMeta
	err := ctx.GetData(contextClusterMeta, &clusterMeta)
	if err != nil {
		return err
	}
	configs := make([]GlobalComponentConfig, len(params))
	for i, param := range params {
		cm := map[string]interface{}{}
		clusterValue, err := convertRealParameterType(ctx, param.Type, param.RealValue.ClusterValue)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("convert real parameter type err = %s", err.Error())
			return err
		}
		// display full parameter name
		cm[DisplayFullParameterName(param.Category, param.Name)] = clusterValue
		configs[i] = GlobalComponentConfig{
			TiDBClusterComponent: spec2.TiDBClusterComponent(strings.ToLower(param.InstanceType)),
			ConfigMap:            cm,
		}
	}
	framework.LogWithContext(ctx).Debugf("modify global component configs: %v", configs)

	// invoke tiup show-config
	tiupHomeForTidb := framework.GetTiupHomePathForTidb()
	topoStr, err := deployment.M.ShowConfig(ctx, deployment.TiUPComponentTypeCluster, clusterMeta.Cluster.ID,
		tiupHomeForTidb, []string{}, meta.DefaultTiupTimeOut)
	if err != nil {
		return err
	}
	ctx.SetData(contextClusterConfigStr, topoStr)
	topo := &tiupSpec.Specification{}
	if err = yaml.UnmarshalStrict([]byte(topoStr), topo); err != nil {
		framework.LogWithContext(ctx).Errorf("parse original config(%s) error: %+v", topoStr, err)
		return err
	}
	yamlConfig, err := generateNewYamlConfig(configs, topo)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("generate new yaml config err = %s", err.Error())
		return err
	}
	editConfigId, err := deployment.M.EditConfig(ctx, deployment.TiUPComponentTypeCluster, clusterMeta.Cluster.ID,
		yamlConfig, tiupHomeForTidb, node.ParentID, []string{}, 0)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call secondparty tiup edit global config err = %s", err.Error())
		return err
	}
	framework.LogWithContext(ctx).Infof("got editConfigId: %v", editConfigId)
	node.OperationID = editConfigId
	return nil
}

func generateNewYamlConfig(configs []GlobalComponentConfig, topo *tiupSpec.Specification) (string, error) {
	var componentServerConfigs map[string]interface{}

	for _, globalComponentConfig := range configs {
		switch globalComponentConfig.TiDBClusterComponent {
		case spec2.TiDBClusterComponent_TiDB:
			componentServerConfigs = topo.ServerConfigs.TiDB
		case spec2.TiDBClusterComponent_TiKV:
			componentServerConfigs = topo.ServerConfigs.TiKV
		case spec2.TiDBClusterComponent_PD:
			componentServerConfigs = topo.ServerConfigs.PD
		case spec2.TiDBClusterComponent_TiFlash:
			componentServerConfigs = topo.ServerConfigs.TiFlash
		case spec2.TiDBClusterComponent_TiFlashLearner:
			componentServerConfigs = topo.ServerConfigs.TiFlashLearner
		case spec2.TiDBClusterComponent_Pump:
			componentServerConfigs = topo.ServerConfigs.Pump
		case spec2.TiDBClusterComponent_Drainer:
			componentServerConfigs = topo.ServerConfigs.Drainer
		case spec2.TiDBClusterComponent_CDC:
			componentServerConfigs = topo.ServerConfigs.CDC
		}
		if componentServerConfigs == nil {
			componentServerConfigs = make(map[string]interface{})
		}
		for k, v := range globalComponentConfig.ConfigMap {
			componentServerConfigs[k] = v
		}
		switch globalComponentConfig.TiDBClusterComponent {
		case spec2.TiDBClusterComponent_TiDB:
			topo.ServerConfigs.TiDB = componentServerConfigs
		case spec2.TiDBClusterComponent_TiKV:
			topo.ServerConfigs.TiKV = componentServerConfigs
		case spec2.TiDBClusterComponent_PD:
			topo.ServerConfigs.PD = componentServerConfigs
		case spec2.TiDBClusterComponent_TiFlash:
			topo.ServerConfigs.TiFlash = componentServerConfigs
		case spec2.TiDBClusterComponent_TiFlashLearner:
			topo.ServerConfigs.TiFlashLearner = componentServerConfigs
		case spec2.TiDBClusterComponent_Pump:
			topo.ServerConfigs.Pump = componentServerConfigs
		case spec2.TiDBClusterComponent_Drainer:
			topo.ServerConfigs.Drainer = componentServerConfigs
		case spec2.TiDBClusterComponent_CDC:
			topo.ServerConfigs.CDC = componentServerConfigs
		}
	}

	newData, err := yaml.Marshal(topo)
	if err != nil {
		return "", err
	}
	return string(newData), nil
}

// convertRealParameterType
// @Description: convert real parameter type
// @Parameter ctx
// @Parameter paramType
// @Parameter value
// @return interface{}
// @return error
func convertRealParameterType(ctx context.Context, paramType int, value string) (interface{}, error) {
	switch paramType {
	case int(Integer):
		// Compatible with scientific notation, e.g.: 1.44e+06
		decimalNum, err := decimal.NewFromString(value)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("decimal.NewFromString error, numStr:%s, err:%v", value, err)
			return nil, err
		}
		c, err := strconv.ParseInt(decimalNum.String(), 10, 64)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("strconv realvalue type int fail, err = %s", err.Error())
			return nil, err
		}
		return c, nil
	case int(Boolean):
		c, err := strconv.ParseBool(value)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("strconv realvalue type bool fail, err = %s", err.Error())
			return nil, err
		}
		return c, nil
	case int(Float):
		c, err := strconv.ParseFloat(value, 64)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("strconv realvalue type float fail, err = %s", err.Error())
			return nil, err
		}
		// Retains floating precision and is not converted to integer
		valStr := strings.Split(value, ".")
		if len(valStr) == 2 {
			num, err := strconv.Atoi(valStr[1])
			if err != nil {
				return nil, err
			}
			if num == 0 {
				c += 1e-8
			}
		}
		return c, nil
	case int(Array):
		var c interface{}
		err := json.Unmarshal([]byte(value), &c)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("strconv realvalue type array fail, err = %s", err.Error())
			return nil, err
		}
		return c, nil
	default:
		return value, nil
	}
}

// putParameterContainer
// @Description: put parameter container
// @Parameter paramContainer
// @Parameter key
// @Parameter param
func putParameterContainer(paramContainer map[interface{}][]*ModifyClusterParameterInfo, key interface{}, param *ModifyClusterParameterInfo) {
	params := paramContainer[key]
	if params == nil {
		paramContainer[key] = []*ModifyClusterParameterInfo{param}
	} else {
		params = append(params, param)
		paramContainer[key] = params
	}
}

// refreshParameter
// @Description: refresh cluster parameter
// @Parameter node
// @Parameter ctx
// @return error
func refreshParameter(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	var modifyParam ModifyParameter
	err := ctx.GetData(contextModifyParameters, &modifyParam)
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Debugf("got modify need reboot: %v, params size: %d", modifyParam.Reboot, len(modifyParam.Params))

	var clusterMeta meta.ClusterMeta
	err = ctx.GetData(contextClusterMeta, &clusterMeta)
	if err != nil {
		return err
	}
	// need tiup reload config
	if modifyParam.Reboot {
		flags := make([]string, 0)
		// Check for partial node instances
		if modifyParam.Nodes != nil && len(modifyParam.Nodes) > 0 {
			flags = append(flags, "-N")
			flags = append(flags, strings.Join(modifyParam.Nodes, ","))
		}
		tiupHomeForTidb := framework.GetTiupHomePathForTidb()
		reloadId, err := deployment.M.Reload(ctx, deployment.TiUPComponentTypeCluster, clusterMeta.Cluster.ID, tiupHomeForTidb, node.ParentID, flags, 0)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("call tiup api edit global config err = %s", err.Error())
			return err
		}
		framework.LogWithContext(ctx).Infof("got reloadId: %v", reloadId)
		node.OperationID = reloadId
	} else {
		node.Success()
	}
	node.Record(fmt.Sprintf("refresh cluster %s parameters ", clusterMeta.Cluster.ID))
	return nil
}
