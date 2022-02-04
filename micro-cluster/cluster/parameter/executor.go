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
 * @File: executor.go
 * @Description: flow task executor
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/15 17:08
*******************************************************************************/

package parameter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/deployment"
	"gopkg.in/yaml.v2"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap-inc/tiem/common/errors"

	"github.com/pingcap-inc/tiem/common/constants"

	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/parameter"

	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/library/spec"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
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

	if flow, flowError := workflow.GetWorkFlowService().CreateWorkFlow(ctx, meta.Cluster.ID, workflow.BizTypeCluster, flowName); flowError != nil {
		framework.LogWithContext(ctx).Errorf("create flow %s failed, clusterID = %s, error = %s", flow.Flow.Name, meta.Cluster.ID, err.Error())
		err = flowError
		return
	} else {
		flowID = flow.Flow.ID
		flow.Context.SetData(contextClusterMeta, meta)
		for k, v := range data {
			flow.Context.SetData(k, v)
		}
		if err = workflow.GetWorkFlowService().AsyncStart(ctx, flow); err != nil {
			framework.LogWithContext(ctx).Errorf("start flow %s failed, clusterID = %s, error = %s", flow.Flow.Name, meta.Cluster.ID, err.Error())
			return
		}
		framework.LogWithContext(ctx).Infof("create flow %s succeed, clusterID = %s", flow.Flow.Name, meta.Cluster.ID)
	}
	return
}

// defaultEnd
// @Description: clear maintenance status after maintenance finished or failed
func defaultEnd(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin default end executor method")
	defer framework.LogWithContext(ctx).Info("end default end executor method")

	clusterMeta := ctx.GetData(contextClusterMeta).(*meta.ClusterMeta)
	maintenanceStatusChange := ctx.GetData(contextMaintenanceStatusChange).(bool)
	if maintenanceStatusChange {
		if err := clusterMeta.EndMaintenance(ctx, clusterMeta.Cluster.MaintenanceStatus); err != nil {
			framework.LogWithContext(ctx).Errorf("end cluster %s maintenance status failed, %s", clusterMeta.Cluster.ID, err.Error())
			return err
		}
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

	modifyParam := ctx.GetData(contextModifyParameters).(*ModifyParameter)
	params := make([]*parameter.ClusterParameterMapping, len(modifyParam.Params))
	for i, param := range modifyParam.Params {
		b, err := json.Marshal(param.RealValue)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to convert parameter real value err: %v", err)
			return errors.NewErrorf(errors.TIEM_CONVERT_OBJ_FAILED, errors.TIEM_CONVERT_OBJ_FAILED.Explain())
		}
		params[i] = &parameter.ClusterParameterMapping{
			ClusterID:   modifyParam.ClusterID,
			ParameterID: param.ParamId,
			RealValue:   string(b),
		}
	}

	// Get the apply parameter object
	hasApplyParameter := ctx.GetData(contextHasApplyParameter)
	if hasApplyParameter != nil && hasApplyParameter.(bool) {
		framework.LogWithContext(ctx).Infof("current has apply parameter: %v", hasApplyParameter.(bool))
		// persist apply parameter
		err := models.GetClusterParameterReaderWriter().ApplyClusterParameter(ctx, modifyParam.ParamGroupId, modifyParam.ClusterID, params)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("apply parameter group convert resp err: %v", err)
			return errors.NewErrorf(errors.TIEM_PARAMETER_GROUP_APPLY_ERROR, err.Error())
		}
	} else {
		// persist update parameter
		err := models.GetClusterParameterReaderWriter().UpdateClusterParameter(ctx, modifyParam.ClusterID, params)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("update cluster parameter err: %v", err)
			return errors.NewErrorf(errors.TIEM_CLUSTER_PARAMETER_UPDATE_ERROR, errors.TIEM_CLUSTER_PARAMETER_UPDATE_ERROR.Explain(), err)
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

	modifyParam := ctx.GetData(contextModifyParameters).(*ModifyParameter)
	framework.LogWithContext(ctx).Debugf("got validation parameters size: %d", len(modifyParam.Params))

	for _, param := range modifyParam.Params {
		// validate parameter value by range field
		if !ValidateRange(param, true) {
			if len(param.Range) == 2 && (param.Type == int(Integer) || param.Type == int(Float)) {
				return fmt.Errorf(fmt.Sprintf("Validation parameter `%s` failed, update value: %s, can take a range of values: %v",
					DisplayFullParameterName(param.Category, param.Name), param.RealValue.ClusterValue, param.Range))
			} else {
				return fmt.Errorf(fmt.Sprintf("Validation parameter `%s` failed, update value: %s, optional values: %v",
					DisplayFullParameterName(param.Category, param.Name), param.RealValue.ClusterValue, param.Range))
			}
		}
	}
	node.Record("validate parameters ")
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
	clusterMeta := ctx.GetData(contextClusterMeta).(*meta.ClusterMeta)

	modifyParam := ctx.GetData(contextModifyParameters).(*ModifyParameter)
	framework.LogWithContext(ctx).Debugf("got modify need reboot: %v, parameters size: %d", modifyParam.Reboot, len(modifyParam.Params))
	maintenanceStatusChange := ctx.GetData(contextMaintenanceStatusChange)

	// Get the apply parameter object
	applyParameter := ctx.GetData(contextHasApplyParameter)
	framework.LogWithContext(ctx).Debugf("modify parameter get apply parameter: %v", applyParameter)

	// grouping by parameter source
	paramContainer := make(map[interface{}][]*ModifyClusterParameterInfo)
	for i, param := range modifyParam.Params {
		// condition apply parameter and HasApply values is 0, then filter directly
		if applyParameter != nil && param.HasApply != int(DirectApply) {
			continue
		}
		if param.InstanceType == string(constants.ComponentIDCDC) && len(clusterMeta.GetCDCClientAddresses()) == 0 {
			// If it is a parameter of CDC, apply the parameter without installing CDC, then skip directly
			if applyParameter != nil {
				// The real value is set to an unknown empty value
				param.RealValue.ClusterValue = ""
				continue
			} else {
				return fmt.Errorf("get %s address from meta failed, empty address", constants.ComponentIDCDC)
			}
		}
		if param.InstanceType == string(constants.ComponentIDTiFlash) && len(clusterMeta.GetTiFlashClientAddresses()) == 0 {
			// If it is a parameter of TiFlash, apply the parameter without installing TiFlash, then skip directly
			if applyParameter != nil {
				// The real value is set to an unknown empty value
				param.RealValue.ClusterValue = ""
				continue
			} else {
				return fmt.Errorf("get %s address from meta failed, empty address", constants.ComponentIDTiFlash)
			}
		}
		// If it is an apply parameter with an empty parameter value, it is skipped directly
		if applyParameter != nil && strings.TrimSpace(param.RealValue.ClusterValue) == "" {
			continue
		}
		// If the parameter is modified and is triggered by another workflow and the parameter value is empty, then skip directly
		if applyParameter == nil && maintenanceStatusChange != nil && !maintenanceStatusChange.(bool) && strings.TrimSpace(param.RealValue.ClusterValue) == "" {
			continue
		}

		// If the parameters are modified, read-only parameters are not allowed to be modified
		if applyParameter == nil && param.ReadOnly == int(ReadOnly) {
			return fmt.Errorf(fmt.Sprintf("Read-only parameters `%s` are not allowed to be modified", DisplayFullParameterName(param.Category, param.Name)))
		}
		framework.LogWithContext(ctx).Debugf("loop %d modify param name: %v, cluster value: %v", i, param.Name, param.RealValue.ClusterValue)
		// condition UpdateSource values is 2, then insert tiup and sql respectively
		if param.UpdateSource == int(TiupAndSql) {
			putParameterContainer(paramContainer, int(TiUP), param)
			putParameterContainer(paramContainer, int(SQL), param)
		} else {
			putParameterContainer(paramContainer, param.UpdateSource, param)
		}
		node.Record(fmt.Sprintf("modify parameter `%s` in %s to %s; ", DisplayFullParameterName(param.Category, param.Name), param.InstanceType, param.RealValue.ClusterValue))
	}

	for source, params := range paramContainer {
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
	node.Record("modify parameters ")
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

	configs := make([]secondparty.ClusterComponentConfig, len(params))
	for i, param := range params {
		configKey := param.Name
		// set config key from system variable
		if param.SystemVariable != "" {
			configKey = param.SystemVariable
		}
		configs[i] = secondparty.ClusterComponentConfig{
			TiDBClusterComponent: spec.TiDBClusterComponent(strings.ToLower(param.InstanceType)),
			ConfigKey:            configKey,
			ConfigValue:          param.RealValue.ClusterValue,
		}
	}

	clusterMeta := ctx.GetData(contextClusterMeta).(*meta.ClusterMeta)
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

	req := secondparty.ClusterEditConfigReq{
		DbConnParameter: secondparty.DbConnParam{
			Username: tidbUserInfo.Name,
			Password: string(tidbUserInfo.Password),
			IP:       tidbServer.IP,
			Port:     strconv.Itoa(tidbServer.Port),
		},
		ComponentConfigs: configs,
	}
	err = secondparty.Manager.EditClusterConfig(ctx, req, node.ID)
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
				configKey := param.Name
				// set config key from system variable
				if param.SystemVariable != "" {
					configKey = param.SystemVariable
				}
				clusterValue, err := convertRealParameterType(ctx, param)
				if err != nil {
					framework.LogWithContext(ctx).Errorf("convert real parameter type err = %v", err)
					return err
				}
				cm[configKey] = clusterValue
			}
			clusterMeta := ctx.GetData(contextClusterMeta).(*meta.ClusterMeta)

			// Get the instance host and port of the component based on the topology
			servers := make(map[string]uint)
			switch comp.(string) {
			case string(constants.ComponentIDTiDB):
				tidbServers := clusterMeta.GetClusterStatusAddress()
				if len(tidbServers) == 0 {
					framework.LogWithContext(ctx).Errorf("get tidb status address from meta failed, empty address")
					return fmt.Errorf("get tidb status address from meta failed, empty address")
				}
				for _, server := range tidbServers {
					servers[server.IP] = uint(server.Port)
				}
			case string(constants.ComponentIDTiKV):
				tikvServers := clusterMeta.GetTiKVStatusAddress()
				if len(tikvServers) == 0 {
					framework.LogWithContext(ctx).Errorf("get tikv address from meta failed, empty address")
					return fmt.Errorf("get tikv address from meta failed, empty address")
				}
				for _, server := range tikvServers {
					servers[server.IP] = uint(server.Port)
				}
			case string(constants.ComponentIDPD):
				pdServers := clusterMeta.GetPDClientAddresses()
				if len(pdServers) == 0 {
					framework.LogWithContext(ctx).Errorf("get pd address from meta failed, empty address")
					return fmt.Errorf("get pd address from meta failed, empty address")
				}
				server := pdServers[rand.Intn(len(pdServers))]
				servers[server.IP] = uint(server.Port)
			case string(constants.ComponentIDCDC):
				cdcServers := clusterMeta.GetCDCClientAddresses()
				if len(cdcServers) == 0 {
					framework.LogWithContext(ctx).Errorf("get cdc address from meta failed, empty address")
					return fmt.Errorf("get cdc address from meta failed, empty address")
				}
				server := cdcServers[rand.Intn(len(cdcServers))]
				servers[server.IP] = uint(server.Port)
			default:
				return fmt.Errorf(fmt.Sprintf("Component [%s] type modification is not supported", comp.(string)))
			}
			for host, port := range servers {
				hasSuc, err := secondparty.Manager.ApiEditConfig(ctx, secondparty.ApiEditConfigReq{
					TiDBClusterComponent: spec.TiDBClusterComponent(strings.ToLower(comp.(string))),
					InstanceHost:         host,
					InstancePort:         port,
					Headers:              map[string]string{},
					ConfigMap:            cm,
				})
				if err != nil || !hasSuc {
					framework.LogWithContext(ctx).Errorf("call secondparty api edit config is %v, err = %s", hasSuc, err)
					return err
				}
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

	clusterMeta := ctx.GetData(contextClusterMeta).(*meta.ClusterMeta)
	configs := make([]secondparty.GlobalComponentConfig, len(params))
	for i, param := range params {
		cm := map[string]interface{}{}
		clusterValue, err := convertRealParameterType(ctx, param)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("convert real parameter type err = %s", err.Error())
			return err
		}
		// display full parameter name
		cm[DisplayFullParameterName(param.Category, param.Name)] = clusterValue
		configs[i] = secondparty.GlobalComponentConfig{
			TiDBClusterComponent: spec.TiDBClusterComponent(strings.ToLower(param.InstanceType)),
			ConfigMap:            cm,
		}
	}
	framework.LogWithContext(ctx).Debugf("modify global component configs: %v", configs)

	yamlConfig, err := generateNewYamlConfig(ctx, clusterMeta.Cluster.ID, configs)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("generate new yaml config err = %s", err.Error())
		return err
	}
	editConfigId, err := deployment.M.EditConfig(ctx, deployment.TiUPComponentTypeCluster, clusterMeta.Cluster.ID,
		yamlConfig, "/home/tiem/.tiup", node.ParentID, []string{}, 0)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call secondparty tiup edit global config err = %s", err.Error())
		return err
	}
	framework.LogWithContext(ctx).Infof("got editConfigId: %v", editConfigId)
	node.OperationID = editConfigId
	// loop get tiup exec status
	return getTaskStatusByTaskId(ctx, node)
}

func generateNewYamlConfig(ctx context.Context, clusterID string, configs []secondparty.GlobalComponentConfig) (string, error) {
	topoStr, err := deployment.M.ShowConfig(ctx, deployment.TiUPComponentTypeCluster, clusterID, "/home/tiem/.tiup", []string{}, meta.DefaultTiupTimeOut)
	if err != nil {
		return "", err
	}
	topo := &tiupSpec.Specification{}
	if err = yaml.UnmarshalStrict([]byte(topoStr), topo); err != nil {
		framework.LogWithContext(ctx).Errorf("parse original config(%s) error: %+v", topoStr, err)
		return "", err
	}

	var componentServerConfigs map[string]interface{}

	for _, globalComponentConfig := range configs {
		switch globalComponentConfig.TiDBClusterComponent {
		case spec.TiDBClusterComponent_TiDB:
			componentServerConfigs = topo.ServerConfigs.TiDB
		case spec.TiDBClusterComponent_TiKV:
			componentServerConfigs = topo.ServerConfigs.TiKV
		case spec.TiDBClusterComponent_PD:
			componentServerConfigs = topo.ServerConfigs.PD
		case spec.TiDBClusterComponent_TiFlash:
			componentServerConfigs = topo.ServerConfigs.TiFlash
		case spec.TiDBClusterComponent_TiFlashLearner:
			componentServerConfigs = topo.ServerConfigs.TiFlashLearner
		case spec.TiDBClusterComponent_Pump:
			componentServerConfigs = topo.ServerConfigs.Pump
		case spec.TiDBClusterComponent_Drainer:
			componentServerConfigs = topo.ServerConfigs.Drainer
		case spec.TiDBClusterComponent_CDC:
			componentServerConfigs = topo.ServerConfigs.CDC
		}
		if componentServerConfigs == nil {
			componentServerConfigs = make(map[string]interface{})
		}
		for k, v := range globalComponentConfig.ConfigMap {
			componentServerConfigs[k] = v
		}
		switch globalComponentConfig.TiDBClusterComponent {
		case spec.TiDBClusterComponent_TiDB:
			topo.ServerConfigs.TiDB = componentServerConfigs
		case spec.TiDBClusterComponent_TiKV:
			topo.ServerConfigs.TiKV = componentServerConfigs
		case spec.TiDBClusterComponent_PD:
			topo.ServerConfigs.PD = componentServerConfigs
		case spec.TiDBClusterComponent_TiFlash:
			topo.ServerConfigs.TiFlash = componentServerConfigs
		case spec.TiDBClusterComponent_TiFlashLearner:
			topo.ServerConfigs.TiFlashLearner = componentServerConfigs
		case spec.TiDBClusterComponent_Pump:
			topo.ServerConfigs.Pump = componentServerConfigs
		case spec.TiDBClusterComponent_Drainer:
			topo.ServerConfigs.Drainer = componentServerConfigs
		case spec.TiDBClusterComponent_CDC:
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
// @Parameter param
// @return interface{}
// @return error
func convertRealParameterType(ctx *workflow.FlowContext, param *ModifyClusterParameterInfo) (interface{}, error) {
	switch param.Type {
	case int(Integer):
		c, err := strconv.ParseInt(param.RealValue.ClusterValue, 0, 64)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("strconv realvalue type int fail, err = %s", err.Error())
			return nil, err
		}
		return c, nil
	case int(Boolean):
		c, err := strconv.ParseBool(param.RealValue.ClusterValue)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("strconv realvalue type bool fail, err = %s", err.Error())
			return nil, err
		}
		return c, nil
	case int(Float):
		c, err := strconv.ParseFloat(param.RealValue.ClusterValue, 64)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("strconv realvalue type float fail, err = %s", err.Error())
			return nil, err
		}
		// Retains floating precision and is not converted to integer
		valStr := strings.Split(param.RealValue.ClusterValue, ".")
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
		err := json.Unmarshal([]byte(param.RealValue.ClusterValue), &c)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("strconv realvalue type array fail, err = %s", err.Error())
			return nil, err
		}
		return c, nil
	default:
		return param.RealValue.ClusterValue, nil
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
	modifyParam := ctx.GetData(contextModifyParameters).(*ModifyParameter)
	framework.LogWithContext(ctx).Debugf("got modify need reboot: %v, params size: %d", modifyParam.Reboot, len(modifyParam.Params))

	clusterMeta := ctx.GetData(contextClusterMeta).(*meta.ClusterMeta)
	// need tiup reload config
	if modifyParam.Reboot {
		flags := make([]string, 0)
		// Check for partial node instances
		if modifyParam.Nodes != nil && len(modifyParam.Nodes) > 0 {
			flags = append(flags, "-N")
			flags = append(flags, strings.Join(modifyParam.Nodes, ","))
		}

		reloadID, err := deployment.M.Reload(ctx, deployment.TiUPComponentTypeCluster, clusterMeta.Cluster.ID, "/home/tiem/.tiup", node.ParentID, flags, 0)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("call tiup api edit global config err = %s", err.Error())
			return err
		}
		framework.LogWithContext(ctx).Infof("got reloadID: %v", reloadID)

		// loop get tiup exec status
		node.OperationID = reloadID
		return getTaskStatusByTaskId(ctx, node)
	}
	node.Record(fmt.Sprintf("refresh cluster %s parameters ", clusterMeta.Cluster.ID))
	return nil
}

// getTaskStatusByTaskId
// @Description: get task status by id
// @Parameter ctx
// @Parameter node
// @return error
func getTaskStatusByTaskId(ctx *workflow.FlowContext, node *workflowModel.WorkFlowNode) error {
	framework.LogWithContext(ctx).Info("begin get task status")
	defer framework.LogWithContext(ctx).Info("end get task status")

	ticker := time.NewTicker(3 * time.Second)
	sequence := 0
	for range ticker.C {
		if sequence += 1; sequence > 200 {
			return errors.NewErrorf(errors.TIEM_TASK_POLLING_TIME_OUT, errors.TIEM_TASK_POLLING_TIME_OUT.Explain())
		}
		framework.LogWithContext(ctx).Infof("polling node waiting, nodeId %s, nodeName %s", node.ID, node.Name)

		op, err := deployment.M.GetStatus(ctx, node.OperationID)
		if err != nil {
			framework.LogWithContext(ctx).Error(err)
			node.Fail(errors.NewErrorf(errors.TIEM_TASK_FAILED, errors.TIEM_TASK_FAILED.Explain()))
			return errors.NewErrorf(errors.TIEM_TASK_FAILED, errors.TIEM_TASK_FAILED.Explain(), err)
		}
		if op.Status == deployment.Error {
			node.Fail(errors.NewErrorf(errors.TIEM_TASK_FAILED, op.ErrorStr))
			return errors.NewErrorf(errors.TIEM_TASK_FAILED, op.ErrorStr)
		}
		if op.Status == deployment.Finished {
			node.Success(op.Result)
			break
		}
	}
	return nil
}
