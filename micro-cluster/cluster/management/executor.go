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

package management

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/common"
	resourceType "github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	resourceManagement "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
	"strconv"
	"time"
)

// prepareResource
// @Description: prepare resource for creating, scaling out
func prepareResource(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	// Alloc resource for new instances
	allocID, err := clusterMeta.AllocInstanceResource(context.Context)

	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] alloc resource error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	context.SetData(ContextAllocId, allocID)
	framework.LogWithContext(context.Context).Infof(
		"cluster[%s] alloc resource request id: %s", clusterMeta.Cluster.Name, allocID)

	return nil
}

// buildConfig
// @Description: generate topology config with cluster meta
func buildConfig(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	topology, err := clusterMeta.GenerateTopologyConfig(context.Context)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] build config error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}

	context.SetData(ContextTopology, topology)
	return nil
}

// scaleOutCluster
// @Description: execute command, scale out
func scaleOutCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster
	yamlConfig := context.GetData(ContextTopology).(string)

	framework.LogWithContext(context.Context).Infof(
		"scale out cluster[%s], version = %s, yamlConfig = %s", cluster.Name, cluster.Version, yamlConfig)
	taskId, err := secondparty.Manager.ClusterScaleOut(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.Name,
		yamlConfig, 0, []string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa"}, node.ID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] scale out error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof("get scale out cluster task id: %s", taskId)
	return nil
}

// scaleInCluster
// @Description: execute command, scale in
func scaleInCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	instanceID := context.GetData(ContextInstanceID).(string)

	instance, err := clusterMeta.GetInstance(context.Context, instanceID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] has no instance[%s]", clusterMeta.Cluster.Name, instanceID)
		return err
	}
	if clusterMeta.IsComponentRequired(context.Context, instance.Type) {
		if len(clusterMeta.Instances[instance.Type]) <= 1 {
			framework.LogWithContext(context.Context).Errorf(
				"instance: %s is unique in cluster[%s], can not delete it", instanceID, clusterMeta.Cluster.Name)
			return framework.NewTiEMError(common.TIEM_DELETE_INSTANCE_ERROR, "instance can not be deleted")
		}
	}
	framework.LogWithContext(context.Context).Infof(
		"scale in cluster %s, delete instance: %s", clusterMeta.Cluster.Name, instanceID)
	taskId, err := secondparty.Manager.ClusterScaleIn(
		context.Context, secondparty.ClusterComponentTypeStr, clusterMeta.Cluster.Name,
		instanceID, 0, []string{"--yes"}, node.ID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] scale in error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof("get scale in cluster task id: %s", taskId)
	return nil
}

// freeInstanceResource
// @Description: todo
func freeInstanceResource(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	instanceID := context.GetData(ContextInstanceID).(string)

	err := clusterMeta.DeleteInstance(context.Context, instanceID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] delete instance[%s] error: %s", clusterMeta.Cluster.Name, instanceID, err.Error())
		return err
	}

	return nil
}

func backupSourceCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	sourceClusterMeta := context.GetData(ContextSourceClusterMeta).(*handler.ClusterMeta)
	cloneStrategy := context.GetData(ContextCloneStrategy).(string)

	if cloneStrategy == string(constants.ClusterTopologyClone) {
		return nil
	}
	backupResponse, err := backuprestore.GetBRService().BackupCluster(context.Context,
		&cluster.BackupClusterDataReq{
			ClusterID:  sourceClusterMeta.Cluster.ID,
			BackupMode: string(constants.BackupModeManual),
		})
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"do backup for cluster[%s] error: %s", sourceClusterMeta.Cluster.Name, err.Error())
		return err
	}

	if err = handler.WaitWorkflow(backupResponse.WorkFlowID, 10*time.Second); err != nil {
		framework.LogWithContext(context.Context).Errorf("backup workflow error: %s", err)
		return err
	}

	context.SetData(ContextBackupID, backupResponse.BackupID)

	return nil
}

// setClusterFailure
// @Description: set cluster running status to constants.ClusterFailure
func setClusterFailure(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
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
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
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

// setClusterOffline
// @Description: set cluster running status to constants.Stopped
func setClusterOffline(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	if err := clusterMeta.UpdateClusterStatus(context.Context, constants.ClusterStopped); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update cluster[%s] status into stopped error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	return nil
}

// revertResourceAfterFailure
// @Description: revert allocated resource after creating, scaling out
func revertResourceAfterFailure(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	allocID := context.GetData(ContextAllocId)
	if allocID != nil {
		request := &resourceType.RecycleRequest{
			RecycleReqs: []resourceType.RecycleRequire{
				{
					RecycleType: resourceType.RecycleOperate,
					RequestID:   allocID.(string),
				},
			},
		}
		resourceManager := resourceManagement.NewResourceManager()
		err := resourceManager.RecycleResources(context.Context, request)
		if err != nil {
			framework.LogWithContext(context.Context).Errorf(
				"recycle request id[%s] resources error, %s", allocID.(string), err.Error())
			return err
		}
	}
	return nil
}

// endMaintenance
// @Description: clear maintenance status after maintenance finished or failed
func endMaintenance(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	return clusterMeta.EndMaintenance(context, clusterMeta.Cluster.MaintenanceStatus)
}

// persistCluster
// @Description: save cluster and instances after flow finished or failed
func persistCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	err := clusterMeta.UpdateMeta(context)
	if err != nil {
		framework.LogWithContext(context).Errorf("persist cluster error, clusterId = %s, workflowId = %s", clusterMeta.Cluster.ID, node.ParentID)
	}
	return err
}

// deployCluster
// @Description: execute command, deploy
func deployCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster
	yamlConfig := context.GetData(ContextTopology).(string)

	framework.LogWithContext(context.Context).Infof(
		"deploy cluster[%s], version = %s, yamlConfig = %s", cluster.Name, cluster.Version, yamlConfig)
	taskId, err := secondparty.Manager.ClusterDeploy(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.Name, cluster.Version,
		yamlConfig, 0, []string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa"}, node.ID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] deploy error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof("get deploy cluster task id: %s", taskId)
	return nil
}

// startCluster
// @Description: execute command, start
func startCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster

	framework.LogWithContext(context.Context).Infof(
		"start cluster[%s], version = %s", cluster.Name, cluster.Version)
	taskId, err := secondparty.Manager.ClusterStart(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.Name, 0, []string{}, node.ID,
	)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] start error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof("get start cluster task id: %s", taskId)
	return nil
}

func saveClusterMeta(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	if err := clusterMeta.Save(context.Context); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"save cluster[%s] meta into db error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	return nil
}

func syncBackupStrategy(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	sourceClusterMeta := context.GetData(ContextSourceClusterMeta).(*handler.ClusterMeta)
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	sourceStrategyRes, err := backuprestore.GetBRService().GetBackupStrategy(context.Context,
		&cluster.GetBackupStrategyReq{
			ClusterID: sourceClusterMeta.Cluster.ID,
		})
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"get cluster[%s] backup strategy error: %s", sourceClusterMeta.Cluster.Name, err.Error())
		return err
	}

	_, err = backuprestore.GetBRService().SaveBackupStrategy(context.Context,
		&cluster.SaveBackupStrategyReq{
			ClusterID: clusterMeta.Cluster.ID,
			Strategy:  sourceStrategyRes.Strategy,
		})
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"save cluster[%s] backup strategy error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}

	return nil
}

func syncParameters(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	return nil
}

func syncSystemVariables(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	//TODO: sync system variables
	return nil
}

func restoreCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cloneStrategy := context.GetData(ContextCloneStrategy).(string)
	backupID := context.GetData(ContextBackupID).(string)

	if cloneStrategy == string(constants.ClusterTopologyClone) {
		return nil
	}
	restoreResponse, err := backuprestore.GetBRService().RestoreExistCluster(context.Context,
		&cluster.RestoreExistClusterReq{
			ClusterID:  clusterMeta.Cluster.ID,
			BackupID: backupID,
		})
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"do restore for cluster[%s] by backup id[%s] error: %s", clusterMeta.Cluster.Name, backupID, err.Error())
		return err
	}

	if err = handler.WaitWorkflow(restoreResponse.WorkFlowID, 10*time.Second); err != nil {
		framework.LogWithContext(context.Context).Errorf("restore workflow error: %s", err)
		return err
	}

	return nil
}

func syncIncrData(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	return nil
}

// syncTopology
// @Description: get topology content from tiup, save it as a snapshot for comparing or recovering
// todo
func syncTopology(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	return nil
}

// stopCluster
// @Description: execute command, stop
func stopCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster

	framework.LogWithContext(context.Context).Infof(
		"stop cluster[%s], version = %s", cluster.Name, cluster.Version)
	taskId, err := secondparty.Manager.ClusterStop(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.Name, 0, []string{}, node.ID,
	)

	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] stop error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof("get stop cluster task id: %s", taskId)
	return nil
}

// destroyCluster
// @Description: execute command, destroy
func destroyCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster

	framework.LogWithContext(context.Context).Infof(
		"destroy cluster[%s], version = %s", cluster.Name, cluster.Version)
	taskId, err := secondparty.Manager.ClusterDestroy(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.Name, 0, []string{}, node.ID,
	)

	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] destroy error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof("get destroy cluster task id: %s", taskId)
	return nil
}

// deleteCluster
// @Description: delete cluster from database
func deleteCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	return clusterMeta.Delete(context)
}

// freedResource
// @Description: freed resource
func freedResource(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	err := clusterMeta.FreedInstanceResource(context)

	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] freed resource error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"cluster[%s] freed resource succeed", clusterMeta.Cluster.Name)

	return nil
}

// initDatabaseAccount
// @Description: init database account for new cluster
func initDatabaseAccount(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	framework.LogWithContext(context).Info("begin initDatabaseAccount")
	defer framework.LogWithContext(context).Info("end initDatabaseAccount")

	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster

	tidbServerHost := clusterMeta.GetClusterConnectAddresses()[0].IP
	tidbServerPort := clusterMeta.GetClusterConnectAddresses()[0].Port
	req := secondparty.ClusterSetDbPswReq{
		DbConnParameter: secondparty.DbConnParam{
			Username: "root", //todo: replace admin account
			Password: cluster.DBPassword,
			IP:       tidbServerHost,
			Port:     strconv.Itoa(tidbServerPort),
		},
	}
	err := secondparty.Manager.SetClusterDbPassword(context, req, node.ID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] init database account error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"cluster[%s] init database account succeed", clusterMeta.Cluster.Name)

	return nil
}
