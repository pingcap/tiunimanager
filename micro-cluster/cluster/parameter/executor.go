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
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap-inc/tiem/common/constants"

	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/parameter"

	"github.com/pingcap-inc/tiem/common/structs"

	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	secondparty2 "github.com/pingcap-inc/tiem/models/workflow/secondparty"

	"github.com/pingcap-inc/tiem/library/secondparty"
	spec2 "github.com/pingcap-inc/tiem/library/spec"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/pingcap/tiup/pkg/cluster/spec"
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
func asyncMaintenance(ctx context.Context, meta *handler.ClusterMeta, data map[string]interface{}, status constants.ClusterMaintenanceStatus, flowName string) (flowID string, err error) {
	if err = meta.StartMaintenance(ctx, status); err != nil {
		framework.LogWithContext(ctx).Errorf("start maintenance failed, clusterID = %s, status = %s,error = %s", meta.Cluster.ID, status, err.Error())
		return
	}

	if flow, flowError := workflow.GetWorkFlowService().CreateWorkFlow(ctx, meta.Cluster.ID, flowName); flowError != nil {
		framework.LogWithContext(ctx).Errorf("create flow %s failed, clusterID = %s, error = %s", flow.Flow.Name, meta.Cluster.ID, err.Error())
		err = flowError
		return
	} else {
		flowID = flow.Flow.ID
		flow.Context.SetData(contextClusterMeta, &meta)
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

// endMaintenance
// @Description: clear maintenance status after maintenance finished or failed
func endMaintenance(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(contextClusterMeta).(*handler.ClusterMeta)
	return clusterMeta.EndMaintenance(context, clusterMeta.Cluster.MaintenanceStatus)
}

// setClusterFailure
// @Description: set cluster running status to constants.ClusterFailure
func setClusterFailure(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(contextClusterMeta).(*handler.ClusterMeta)
	if err := clusterMeta.UpdateClusterStatus(context.Context, constants.ClusterFailure); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update cluster[%s] instances status into failure error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"set cluster[%s] status into failure successfully", clusterMeta.Cluster.Name)
	return nil
}

// setClusterOnline
// @Description: set cluster running status to constants.ClusterRunning
func setClusterOnline(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(contextClusterMeta).(*handler.ClusterMeta)
	if clusterMeta.Cluster.Status == string(constants.ClusterInitializing) {
		if err := clusterMeta.UpdateClusterStatus(context.Context, constants.ClusterRunning); err != nil {
			framework.LogWithContext(context.Context).Errorf(
				"update cluster[%s] status into running error: %s", clusterMeta.Cluster.Name, err.Error())
			return err
		}
	}
	if err := clusterMeta.UpdateClusterStatus(context.Context, constants.ClusterRunning); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update cluster[%s] instances status into running error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"set cluster[%s]  status into running successfully", clusterMeta.Cluster.Name)
	return nil
}

// persistUpdateParameter
// @Description: persist update parameter
// @Parameter node
// @Parameter ctx
// @return error
func persistUpdateParameter(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	req := ctx.GetData(contextUpdateParameterInfo).(*cluster.UpdateClusterParametersReq)

	params := make([]*parameter.ClusterParameterMapping, len(req.Params))
	for i, param := range req.Params {
		b, err := json.Marshal(param.RealValue)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("failed to convert parameter realValue. req: %v, err: %v", req, err)
			return framework.SimpleError(common.TIEM_CONVERT_OBJ_FAILED)
		}
		params[i] = &parameter.ClusterParameterMapping{
			ClusterID:   req.ClusterID,
			ParameterID: param.ParamId,
			RealValue:   string(b),
		}
	}
	err := models.GetClusterParameterReaderWriter().UpdateClusterParameter(ctx, req.ClusterID, params)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update cluster parameter req: %v, err: %v", req, err)
		return framework.SimpleError(common.TIEM_CLUSTER_PARAMETER_UPDATE_ERROR)
	}
	return nil
}

// persistApplyParameter
// @Description: persist apply parameter
// @Parameter node
// @Parameter ctx
// @return error
func persistApplyParameter(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	req := ctx.GetData(contextApplyParameterInfo).(*message.ApplyParameterGroupReq)
	pg, params, err := models.GetParameterGroupReaderWriter().GetParameterGroup(ctx, req.ParamGroupId)
	if err != nil || pg.ID == "" {
		framework.LogWithContext(ctx).Errorf("get parameter group req: %v, err: %v", req, err)
		return framework.WrapError(common.TIEM_PARAMETER_GROUP_DETAIL_ERROR, common.TIEM_PARAMETER_GROUP_DETAIL_ERROR.Explain(), err)
	}

	pgs := make([]*parameter.ClusterParameterMapping, len(params))
	for i, param := range params {
		realValue := structs.ParameterRealValue{ClusterValue: param.DefaultValue}
		b, err := json.Marshal(realValue)
		if err != nil {
			return framework.SimpleError(common.TIEM_PARAMETER_GROUP_APPLY_ERROR)
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
		return framework.WrapError(common.TIEM_PARAMETER_GROUP_APPLY_ERROR, common.TIEM_PARAMETER_GROUP_APPLY_ERROR.Explain(), err)
	}
	return err
}

// modifyParameters
// @Description: modify parameters
// @Parameter node
// @Parameter ctx
// @return error
func modifyParameters(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	modifyParam := ctx.GetData(contextModifyParameters).(*ModifyParameter)
	framework.LogWithContext(ctx).Debugf("got modify need reboot: %v, params size: %d", modifyParam.Reboot, len(modifyParam.Params))

	// Get the apply parameter object
	applyParameter := ctx.GetData(contextApplyParameterInfo).(*message.ApplyParameterGroupReq)

	// grouping by parameter source
	paramContainer := make(map[interface{}][]structs.ClusterParameterSampleInfo, 0)
	for i, param := range modifyParam.Params {
		// condition apply parameter and HasApply values is 0, then filter directly
		if applyParameter != nil && param.HasApply != int(DirectApply) {
			continue
		}
		framework.LogWithContext(ctx).Debugf("loop %d modify param name: %v, cluster value: %v", i, param.Name, param.RealValue.ClusterValue)
		// condition UpdateSource values is 2, then insert tiup and sql respectively
		if param.UpdateSource == int(TiupAndSql) {
			putParameterContainer(paramContainer, int(TiUP), param)
			putParameterContainer(paramContainer, int(SQL), param)
		} else {
			putParameterContainer(paramContainer, param.UpdateSource, param)
		}
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
	return nil
}

// sqlEditConfig
// @Description: through sql edit config
// @Parameter ctx
// @Parameter params
// @return error
func sqlEditConfig(ctx *workflow.FlowContext, node *workflowModel.WorkFlowNode, params []structs.ClusterParameterSampleInfo) error {
	configs := make([]secondparty.ClusterComponentConfig, len(params))
	for i, param := range params {
		configKey := param.Name
		if param.SystemVariable != "" {
			configKey = param.SystemVariable
		}
		configs[i] = secondparty.ClusterComponentConfig{
			TiDBClusterComponent: spec2.TiDBClusterComponent(strings.ToLower(param.InstanceType)),
			ConfigKey:            configKey,
			ConfigValue:          param.RealValue.ClusterValue,
		}
	}

	clusterMeta := ctx.GetData(contextClusterMeta).(*handler.ClusterMeta)
	tidbServers := clusterMeta.GetClusterConnectAddresses()
	tidbServer := tidbServers[rand.Intn(len(tidbServers))]
	req := secondparty.ClusterEditConfigReq{
		DbConnParameter: secondparty.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: "",
			IP:       tidbServer.IP,
			Port:     strconv.Itoa(tidbServer.Port),
		},
		ComponentConfigs: configs,
	}
	err := secondparty.Manager.EditClusterConfig(ctx, req, node.ID)
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
func apiEditConfig(ctx *workflow.FlowContext, node *workflowModel.WorkFlowNode, params []structs.ClusterParameterSampleInfo) error {
	compContainer := make(map[interface{}][]structs.ClusterParameterSampleInfo, 0)
	for i, param := range params {
		framework.LogWithContext(ctx).Debugf("loop %d api componet type: %v, param name: %v", i, param.InstanceType, param.Name)
		putParameterContainer(compContainer, param.InstanceType, param)
	}
	if len(compContainer) > 0 {
		for comp, params := range compContainer {
			compStr := strings.ToLower(comp.(string))
			cm := map[string]interface{}{}
			for _, param := range params {
				clusterValue, err := convertRealParameterType(ctx, param)
				if err != nil {
					framework.LogWithContext(ctx).Errorf("convert real parameter type err = %v", err)
					return err
				}
				cm[param.Name] = clusterValue
			}
			clusterMeta := ctx.GetData(contextClusterMeta).(*handler.ClusterMeta)

			// Get the instance host and port of the component based on the topology
			servers := make(map[string]uint, 0)
			switch strings.ToLower(compStr) {
			case spec.ComponentTiDB:
				tidbServers := clusterMeta.GetClusterStatusAddress()
				for _, server := range tidbServers {
					servers[server.IP] = uint(server.Port)
				}
			case spec.ComponentTiKV:
				tikvServers := clusterMeta.GetTiKVStatusAddress()
				for _, server := range tikvServers {
					servers[server.IP] = uint(server.Port)
				}
			case spec.ComponentPD:
				pdServers := clusterMeta.GetPDClientAddresses()
				server := pdServers[rand.Intn(len(pdServers))]
				servers[server.IP] = uint(server.Port)
			}
			for host, port := range servers {
				hasSuc, err := secondparty.Manager.ApiEditConfig(ctx, secondparty.ApiEditConfigReq{
					TiDBClusterComponent: spec2.TiDBClusterComponent(compStr),
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
func tiupEditConfig(ctx *workflow.FlowContext, node *workflowModel.WorkFlowNode, params []structs.ClusterParameterSampleInfo) error {
	clusterMeta := ctx.GetData(contextClusterMeta).(*handler.ClusterMeta)
	configs := make([]secondparty.GlobalComponentConfig, len(params))
	for i, param := range params {
		cm := map[string]interface{}{}
		clusterValue, err := convertRealParameterType(ctx, param)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("convert real parameter type err = %s", err.Error())
			return err
		}
		cm[param.Name] = clusterValue
		configs[i] = secondparty.GlobalComponentConfig{
			TiDBClusterComponent: spec2.TiDBClusterComponent(strings.ToLower(param.InstanceType)),
			ConfigMap:            cm,
		}
	}
	framework.LogWithContext(ctx).Debugf("modify global component configs: %v", configs)
	req := secondparty.CmdEditGlobalConfigReq{
		TiUPComponent:          secondparty.ClusterComponentTypeStr,
		InstanceName:           clusterMeta.Cluster.Name,
		GlobalComponentConfigs: configs,
		TimeoutS:               0,
		Flags:                  []string{},
	}
	editConfigId, err := secondparty.Manager.ClusterEditGlobalConfig(ctx, req, node.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call secondparty tiup edit global config err = %s", err.Error())
		return err
	}
	framework.LogWithContext(ctx).Infof("got editConfigId: %v", editConfigId)
	// loop get tiup exec status
	return getTaskStatusByTaskId(ctx, node)
}

// convertRealParameterType
// @Description: convert real parameter type
// @Parameter ctx
// @Parameter param
// @return interface{}
// @return error
func convertRealParameterType(ctx *workflow.FlowContext, param structs.ClusterParameterSampleInfo) (interface{}, error) {
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
func putParameterContainer(paramContainer map[interface{}][]structs.ClusterParameterSampleInfo, key interface{}, param structs.ClusterParameterSampleInfo) {
	params := paramContainer[key]
	if params == nil {
		paramContainer[key] = []structs.ClusterParameterSampleInfo{param}
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

	clusterMeta := ctx.GetData(contextClusterMeta).(*handler.ClusterMeta)
	// need tiup reload config
	if modifyParam.Reboot {
		req := secondparty.CmdReloadConfigReq{
			TiUPComponent: secondparty.ClusterComponentTypeStr,
			InstanceName:  clusterMeta.Cluster.Name,
			TimeoutS:      0,
			Flags:         []string{},
		}
		reloadId, err := secondparty.Manager.ClusterReload(ctx, req, node.ID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("call tiup api edit global config err = %s", err.Error())
			return err
		}
		framework.LogWithContext(ctx).Infof("got reloadId: %v", reloadId)

		// loop get tiup exec status
		return getTaskStatusByTaskId(ctx, node)
	}
	return nil
}

// getTaskStatusByTaskId
// @Description: get task status by id
// @Parameter ctx
// @Parameter node
// @return error
func getTaskStatusByTaskId(ctx *workflow.FlowContext, node *workflowModel.WorkFlowNode) error {
	ticker := time.NewTicker(3 * time.Second)
	sequence := 0
	for range ticker.C {
		if sequence += 1; sequence > 200 {
			return framework.SimpleError(common.TIEM_TASK_POLLING_TIME_OUT)
		}
		framework.LogWithContext(ctx).Infof("polling node waiting, nodeId %s, nodeName %s", node.ID, node.Name)

		resp, err := secondparty.Manager.GetOperationStatusByWorkFlowNodeID(ctx, node.ID)
		if err != nil {
			framework.LogWithContext(ctx).Error(err)
			node.Fail(framework.WrapError(common.TIEM_TASK_FAILED, common.TIEM_TASK_FAILED.Explain(), err))
			return framework.WrapError(common.TIEM_TASK_FAILED, common.TIEM_TASK_FAILED.Explain(), err)
		}
		if resp.Status == secondparty2.OperationStatusError {
			node.Fail(framework.NewTiEMError(common.TIEM_TASK_FAILED, resp.ErrorStr))
			return framework.NewTiEMError(common.TIEM_TASK_FAILED, resp.ErrorStr)
		}
		if resp.Status == secondparty2.OperationStatusFinished {
			node.Success(resp.Result)
			break
		}
	}
	return nil
}
