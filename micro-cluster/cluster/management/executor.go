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
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/sftp"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/log"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/parameter"
	resourceManagement "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management"
	resourceStructs "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"gopkg.in/yaml.v2"
)

// prepareResource
// @Description: prepare resource for creating, scaling out
func prepareResource(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	globalAllocId := uuidutil.GenerateID()
	instanceAllocId := uuidutil.GenerateID()

	globalRequirement := clusterMeta.GenerateGlobalPortRequirements(context)
	instanceRequirement, instances := clusterMeta.GenerateInstanceResourceRequirements(context)

	batchReq := &resourceStructs.BatchAllocRequest{
		BatchRequests: []resourceStructs.AllocReq{
			{
				Applicant: resourceStructs.Applicant{
					HolderId:          clusterMeta.Cluster.ID,
					RequestId:         instanceAllocId,
					TakeoverOperation: false,
				},
				Requires: instanceRequirement,
			},
		},
	}

	if len(globalRequirement) > 0 {
		batchReq.BatchRequests = append(batchReq.BatchRequests, resourceStructs.AllocReq{
			Applicant: resourceStructs.Applicant{
				HolderId:          clusterMeta.Cluster.ID,
				RequestId:         globalAllocId,
				TakeoverOperation: false,
			},
			Requires: globalRequirement,
		})
	}

	allocResponse, err := resourceManagement.GetManagement().GetAllocatorRecycler().AllocResources(context, batchReq)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s alloc resource error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	context.SetData(ContextAllocResource, allocResponse)

	for _, resourceResult := range allocResponse.BatchResults {
		switch resourceResult.RequestId {
		case globalAllocId:
			ports := resourceResult.Results[0].PortRes[0]
			clusterMeta.ApplyGlobalPortResource(ports.Ports[0], ports.Ports[1])
		case instanceAllocId:
			clusterMeta.ApplyInstanceResource(resourceResult, instances)
		default:
			framework.LogWithContext(context).Errorf(
				"unexpected request id in allocResponse, %v", resourceResult.Applicant)
			continue
		}
	}

	//print success information
	for _, ins := range instances {
		var hostIP []string
		if len(ins.HostIP) > 1 {
			hostIP = append(hostIP, ins.HostIP...)
			node.Record(fmt.Sprintf("type: %s, zone: %s, host IP: %s; ", ins.Type, ins.Zone, strings.Join(hostIP, ", ")))
		}else{
			node.Record(fmt.Sprintf("type: %s, zone: %s, host IP: %s; ", ins.Type, ins.Zone, ins.HostIP[0]))
		}
	}
	node.Record(fmt.Sprintf("prepare resource for cluster %s ", clusterMeta.Cluster.ID))
	return nil
}

// buildConfig
// @Description: generate topology config with cluster meta
func buildConfig(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	topology, err := clusterMeta.GenerateTopologyConfig(context.Context)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s build config error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}

	context.SetData(ContextTopology, topology)
	node.Record("build config ")
	return nil
}

// scaleOutCluster
// @Description: execute command, scale out
func scaleOutCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster
	if context.GetData(ContextTopology) == nil {
		framework.LogWithContext(context.Context).Infof(
			"when scale out cluster, not found topology yaml config")
		return nil
	}
	yamlConfig := context.GetData(ContextTopology).(string)

	framework.LogWithContext(context.Context).Infof(
		"scale out cluster %s, version %s, yamlConfig %s", cluster.ID, cluster.Version, yamlConfig)
	taskId, err := secondparty.Manager.ClusterScaleOut(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ID,
		yamlConfig, handler.DefaultTiupTimeOut, []string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa"}, node.ID, "")
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s scale out error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"get scale out cluster %s task id: %s", cluster.ID, taskId)

	node.Record(fmt.Sprintf("scale out cluster %s, version: %s, ", clusterMeta.Cluster.ID, cluster.Version))
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
			"cluster %s has no instance %s", clusterMeta.Cluster.ID, instanceID)
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"scale in cluster %s, delete instance %s", clusterMeta.Cluster.ID, instanceID)
	taskId, err := secondparty.Manager.ClusterScaleIn(
		context.Context, secondparty.ClusterComponentTypeStr, clusterMeta.Cluster.ID,
		strings.Join([]string{instance.HostIP[0], strconv.Itoa(int(instance.Ports[0]))}, ":"), handler.DefaultTiupTimeOut, []string{"--yes"}, node.ID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s scale in error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"get scale in cluster %s task id: %s", clusterMeta.Cluster.ID, taskId)

	node.Record(fmt.Sprintf("scale in cluster %s ", clusterMeta.Cluster.ID))
	return nil
}

// freeInstanceResource
// @Description: free instance resource
func freeInstanceResource(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	instanceID := context.GetData(ContextInstanceID).(string)

	instance, err := clusterMeta.DeleteInstance(context.Context, instanceID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s delete instance %s error: %s", clusterMeta.Cluster.ID, instanceID, err.Error())
		return err
	}
	// recycle instance resource
	request := &resourceStructs.RecycleRequest{
		RecycleReqs: []resourceStructs.RecycleRequire{
			{
				RecycleType: resourceStructs.RecycleHost,
				HolderID:    instance.ClusterID,
				HostID:      instance.HostID,
				ComputeReq: resourceStructs.ComputeRequirement{
					ComputeResource: resourceStructs.ComputeResource{
						CpuCores: int32(instance.CpuCores),
						Memory:   int32(instance.Memory),
					},
				},
				DiskReq: []resourceStructs.DiskResource{
					{
						DiskId: instance.DiskID,
					},
				},
				PortReq: []resourceStructs.PortResource{
					{
						Ports: instance.Ports,
					},
				},
			},
		},
	}

	err = resourceManagement.GetManagement().GetAllocatorRecycler().RecycleResources(context, request)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s recycle instance %s resource error: %s", clusterMeta.Cluster.ID, instanceID, err.Error())
		return err
	}

	node.Record(fmt.Sprintf("cluster %s recycle instance %s ", clusterMeta.Cluster.ID, instanceID))
	return nil
}

func clearBackupData(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	meta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	deleteReq := context.GetData(ContextDeleteRequest).(cluster.DeleteClusterReq)

	_, err := backuprestore.GetBRService().DeleteBackupStrategy(context.Context, cluster.DeleteBackupStrategyReq{
		ClusterID: meta.Cluster.ID,
	})
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"delete backup strategy for cluster %s error: %s", meta.Cluster.ID, err.Error())
		return err
	}
	node.Record(fmt.Sprintf("delete backup strategy for cluster %s; ", meta.Cluster.ID))

	// delete auto backup records
	_, err = backuprestore.GetBRService().DeleteBackupRecords(context.Context, cluster.DeleteBackupDataReq{
		ClusterID:  meta.Cluster.ID,
		BackupMode: string(constants.BackupModeAuto),
	})
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"delete auto backup data for cluster %s error: %s", meta.Cluster.ID, err.Error())
		return err
	}
	node.Record(fmt.Sprintf("delete auto backup data for cluster %s; ", meta.Cluster.ID))

	// delete manual backup records
	if deleteReq.KeepHistoryBackupRecords {
		framework.LogWithContext(context.Context).Infof(
			"keep manual backup data for cluster %s", meta.Cluster.ID)
		node.Record(fmt.Sprintf("keep manual backup data for cluster %s ", meta.Cluster.ID))
	} else {
		excludeBackupIDs := make([]string, 0)
		backupIdBeforeDeleting := context.GetData(ContextBackupID)

		if backupIdBeforeDeleting != nil {
			excludeBackupIDs = append(excludeBackupIDs, backupIdBeforeDeleting.(string))
		}

		_, err = backuprestore.GetBRService().DeleteBackupRecords(context.Context, cluster.DeleteBackupDataReq{
			ClusterID:        meta.Cluster.ID,
			BackupMode:       string(constants.BackupModeManual),
			ExcludeBackupIDs: excludeBackupIDs,
		})
		if err != nil {
			framework.LogWithContext(context.Context).Errorf(
				"delete manual backup data for cluster %s error: %s", meta.Cluster.ID, err.Error())
			return err
		}
		node.Record(fmt.Sprintf("delete manual backup data for cluster %s ", meta.Cluster.ID))
	}

	return nil
}

func backupBeforeDelete(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	meta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	deleteReq := context.GetData(ContextDeleteRequest).(cluster.DeleteClusterReq)

	_, err := models.GetClusterReaderWriter().GetCurrentClusterTopologySnapshot(context, meta.Cluster.ID)
	if err != nil {
		framework.LogWithContext(context).Warnf("cluster %s is not really existed", meta.Cluster.ID)
		node.Success("skip because cluster is not existed")
		return nil
	}

	if deleteReq.AutoBackup {
		backupResponse, err := backuprestore.GetBRService().BackupCluster(
			context.Context,
			cluster.BackupClusterDataReq{
				ClusterID:  meta.Cluster.ID,
				BackupMode: string(constants.BackupModeManual),
			}, false)
		if err != nil {
			framework.LogWithContext(context.Context).Errorf(
				"do backup for cluster %s error: %s", meta.Cluster.ID, err.Error())
			return err
		} else {
			context.SetData(ContextBackupID, backupResponse.BackupID)
		}
		if err = handler.WaitWorkflow(context.Context, backupResponse.WorkFlowID, 10*time.Second, 30*24*time.Hour); err != nil {
			framework.LogWithContext(context).Errorf("backup workflow error: %s", err)
			return err
		}
		node.Record(fmt.Sprintf("do backup for cluster %s ", meta.Cluster.ID))
	} else {
		node.Success("no need to backup")
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
		cluster.BackupClusterDataReq{
			ClusterID:  sourceClusterMeta.Cluster.ID,
			BackupMode: string(constants.BackupModeManual),
		}, true)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"do backup for cluster %s error: %s", sourceClusterMeta.Cluster.ID, err.Error())
		return err
	}

	context.SetData(ContextWorkflowID, backupResponse.WorkFlowID)
	context.SetData(ContextBackupID, backupResponse.BackupID)

	node.Record(fmt.Sprintf("do backup for source cluster %s, ", sourceClusterMeta.Cluster.ID),
		fmt.Sprintf("backup mode: %v ", constants.BackupModeManual))
	return nil
}

func restoreNewCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	if context.GetData(ContextBackupID) == nil {
		framework.LogWithContext(context.Context).Infof(
			"when restore new cluster, not found backup id")
		return nil
	}
	backupID := context.GetData(ContextBackupID).(string)

	restoreResponse, err := backuprestore.GetBRService().RestoreExistCluster(context.Context,
		cluster.RestoreExistClusterReq{
			ClusterID: clusterMeta.Cluster.ID,
			BackupID:  backupID,
		}, false)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf("do restore for cluster %s by backup id %s error: %s", clusterMeta.Cluster.ID, backupID, err.Error())
		return fmt.Errorf("do restore for cluster %s by backup id %s error: %s", clusterMeta.Cluster.ID, backupID, err.Error())
	}

	context.SetData(ContextWorkflowID, restoreResponse.WorkFlowID)
	node.Record(fmt.Sprintf("do restore for cluster %s, backup ID: %s ", clusterMeta.Cluster.ID, backupID))
	return nil
}

func waitWorkFlow(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	if context.GetData(ContextWorkflowID) == nil {
		framework.LogWithContext(context.Context).Infof(
			"when wait workflow, not found workflow id")
		return nil
	}
	workflowID := context.GetData(ContextWorkflowID).(string)

	if err := handler.WaitWorkflow(context.Context, workflowID, 10*time.Second, 30*24*time.Hour); err != nil {
		framework.LogWithContext(context.Context).Errorf("wait workflow %s error: %s", workflowID, err.Error())
		return err
	}

	node.Record(fmt.Sprintf("wait workflow %s done ", workflowID))
	return nil
}

// setClusterFailure
// @Description: set cluster running status to constants.ClusterFailure
func setClusterFailure(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	if err := clusterMeta.UpdateClusterStatus(context.Context, constants.ClusterFailure); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update cluster %s instances status into failure error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"set cluster %s status into failure successfully", clusterMeta.Cluster.ID)
	node.Record(fmt.Sprintf("set cluster %s status into %v ", clusterMeta.Cluster.ID, constants.ClusterFailure))
	return nil
}

// setClusterOnline
// @Description: set cluster running status to constants.ClusterRunning
func setClusterOnline(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	// set instances status into running
	for _, component := range clusterMeta.Instances {
		for _, instance := range component {
			if instance.Status != string(constants.ClusterInstanceRunning) {
				instance.Status = string(constants.ClusterInstanceRunning)
			}
		}
	}

	// set cluster status into running
	if err := clusterMeta.UpdateClusterStatus(context.Context, constants.ClusterRunning); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update cluster %s status into running error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"set cluster %s status into running successfully", clusterMeta.Cluster.ID)
	node.Record(fmt.Sprintf("set cluster %s status into %v ", clusterMeta.Cluster.ID, constants.ClusterRunning))
	return nil
}

// setClusterOffline
// @Description: set cluster running status to constants.Stopped
func setClusterOffline(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	// if instance status is running, set instances status into stopped
	for _, component := range clusterMeta.Instances {
		for _, instance := range component {
			if instance.Status == string(constants.ClusterInstanceRunning) {
				instance.Status = string(constants.ClusterStopped)
			}
		}
	}

	// set cluster status into stopped
	if err := clusterMeta.UpdateClusterStatus(context.Context, constants.ClusterStopped); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update cluster %s status into stopped error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}

	node.Record(fmt.Sprintf("set cluster %s status into %v ", clusterMeta.Cluster.ID, constants.ClusterStopped))
	return nil
}

// revertResourceAfterFailure
// @Description: revert allocated resource after creating, scaling out
func revertResourceAfterFailure(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	if context.GetData(ContextAllocResource) == nil {
		framework.LogWithContext(context.Context).Infof(
			"when recycle resource, not found alloc resource")
		return nil
	}
	resource := context.GetData(ContextAllocResource).(*resourceStructs.BatchAllocResponse)

	if resource != nil && len(resource.BatchResults) > 0 {
		request := &resourceStructs.RecycleRequest{
			RecycleReqs: []resourceStructs.RecycleRequire{},
		}

		for _, v := range resource.BatchResults {
			request.RecycleReqs = append(request.RecycleReqs, resourceStructs.RecycleRequire{
				RecycleType: resourceStructs.RecycleOperate,
				RequestID:   v.RequestId,
			})
		}

		err := resourceManagement.GetManagement().GetAllocatorRecycler().RecycleResources(context.Context, request)
		if err != nil {
			framework.LogWithContext(context.Context).Errorf(
				"recycle resources error: %s", err.Error())
			return err
		}
	}

	node.Record("recycle resource ")
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
		framework.LogWithContext(context).Errorf(
			"persist cluster error, cluster %s, workflow %s", clusterMeta.Cluster.ID, node.ParentID)
	}

	node.Record(fmt.Sprintf("persist cluster %s ", clusterMeta.Cluster.ID))
	return err
}

// deployCluster
// @Description: execute command, deploy
func deployCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster
	if context.GetData(ContextTopology) == nil {
		framework.LogWithContext(context.Context).Infof(
			"when deploy cluster, not found topology yaml config")
		return nil
	}
	yamlConfig := context.GetData(ContextTopology).(string)

	framework.LogWithContext(context.Context).Infof(
		"deploy cluster %s, version %s, yamlConfig %s", cluster.ID, cluster.Version, yamlConfig)
	taskId, err := secondparty.Manager.ClusterDeploy(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ID, cluster.Version,
		yamlConfig, handler.DefaultTiupTimeOut, []string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa"}, node.ID, "")
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s deploy error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"get deploy cluster %s task id: %s", clusterMeta.Cluster.ID, taskId)
	node.Record(fmt.Sprintf("deploy cluster %s ", clusterMeta.Cluster.ID))
	return nil
}

// startCluster
// @Description: execute command, start
func startCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster

	framework.LogWithContext(context.Context).Infof(
		"start cluster %s, version %s", cluster.ID, cluster.Version)
	taskId, err := secondparty.Manager.ClusterStart(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ID, handler.DefaultTiupTimeOut, []string{}, node.ID,
	)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s start error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"get start cluster %s task id: %s", clusterMeta.Cluster.ID, taskId)

	node.Record(fmt.Sprintf("start cluster %s, version: %s, ", clusterMeta.Cluster.ID, cluster.Version))
	return nil
}

func syncBackupStrategy(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	sourceClusterMeta := context.GetData(ContextSourceClusterMeta).(*handler.ClusterMeta)
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	sourceStrategyRes, err := backuprestore.GetBRService().GetBackupStrategy(context.Context,
		cluster.GetBackupStrategyReq{
			ClusterID: sourceClusterMeta.Cluster.ID,
		})
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"get cluster %s backup strategy error: %s", sourceClusterMeta.Cluster.ID, err.Error())
		return err
	}
	node.Record(fmt.Sprintf("get cluster %s backup strategy, ", clusterMeta.Cluster.ID))

	if len(sourceStrategyRes.Strategy.BackupDate) == 0 {
		framework.LogWithContext(context).Infof(
			"cluster %s has no backup strategy", sourceClusterMeta.Cluster.ID)
		node.Success(fmt.Sprintf("cluster %s has no backup strategy, no need to update", clusterMeta.Cluster.ID))
		return nil
	}

	_, err = backuprestore.GetBRService().SaveBackupStrategy(context.Context,
		cluster.SaveBackupStrategyReq{
			ClusterID: clusterMeta.Cluster.ID,
			Strategy:  sourceStrategyRes.Strategy,
		})
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"save cluster %s backup strategy error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}

	node.Record(fmt.Sprintf("save cluster %s backup strategy ", clusterMeta.Cluster.ID))
	return nil
}

func syncParameters(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	sourceClusterMeta := context.GetData(ContextSourceClusterMeta).(*handler.ClusterMeta)
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	sourceResponse, _, err := parameter.NewManager().QueryClusterParameters(context.Context,
		cluster.QueryClusterParametersReq{ClusterID: sourceClusterMeta.Cluster.ID})
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"query cluster %s parameters error: %s", sourceClusterMeta.Cluster.ID, err.Error())
		return err
	}

	targetParams := make([]structs.ClusterParameterSampleInfo, 0)
	reboot := false
	for _, param := range sourceResponse.Params {
		// if parameter is variable which related os(such as temp dir in os), can not update it
		if param.HasApply == int(parameter.ModifyApply) {
			continue
		}
		targetParam := structs.ClusterParameterSampleInfo{
			ParamId:   param.ParamId,
			RealValue: param.RealValue,
		}
		targetParams = append(targetParams, targetParam)
		if param.HasReboot == int(parameter.Reboot) {
			reboot = true
		}
	}
	response, err := parameter.NewManager().UpdateClusterParameters(context.Context, cluster.UpdateClusterParametersReq{
		ClusterID: clusterMeta.Cluster.ID,
		Params:    targetParams,
		Reboot:    reboot,
	}, false)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update cluster %s parameters error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	context.SetData(ContextWorkflowID, response.WorkFlowID)

	node.Record(fmt.Sprintf("update cluster %s parameters ", clusterMeta.Cluster.ID))
	return nil
}

func asyncBuildLog(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	log.GetService().BuildClusterLogConfig(context, clusterMeta.Cluster.ID)
	node.Record(fmt.Sprintf("rebuild log config for cluster %s ",clusterMeta.Cluster.ID))
	return nil
}

func restoreCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cloneStrategy := context.GetData(ContextCloneStrategy).(string)
	if context.GetData(ContextBackupID) == nil {
		framework.LogWithContext(context.Context).Infof(
			"when restore cluster, not found backup id")
		return nil
	}
	backupID := context.GetData(ContextBackupID).(string)

	if cloneStrategy == string(constants.ClusterTopologyClone) {
		return nil
	}
	restoreResponse, err := backuprestore.GetBRService().RestoreExistCluster(context.Context,
		cluster.RestoreExistClusterReq{
			ClusterID: clusterMeta.Cluster.ID,
			BackupID:  backupID,
		}, false)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"do restore for cluster %s by backup id %s error: %s", clusterMeta.Cluster.ID, backupID, err.Error())
		return err
	}
	context.SetData(ContextWorkflowID, restoreResponse.WorkFlowID)

	node.Record(fmt.Sprintf("do restore for cluster %s, backup id: %s ", clusterMeta.Cluster.ID, backupID))
	return nil
}

func syncIncrData(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	return nil
}

func getClusterSpaceInTiUP(ctx context.Context, clusterID string) string {
	tiupHome := secondparty.GetTiUPHomeForComponent(ctx, secondparty.ClusterComponentTypeStr)
	return fmt.Sprintf("%s/storage/cluster/clusters/%s/", tiupHome, clusterID)
}

// syncConnectionKey
// @Description: get private and public key from tiup
func syncConnectionKey(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	privateKey, err := readTiUPFile(context,
		getClusterSpaceInTiUP(context, clusterMeta.Cluster.ID),
		"ssh/id_rsa")
	if err != nil {
		err = errors.NewEMErrorf(errors.TIEM_CONNECT_TIDB_ERROR, "sync connection private key failed for cluster %s, err = %s", clusterMeta.Cluster.ID, err)
		framework.LogWithContext(context).Errorf(err.Error())
		return  err
	}
	publicKey, err := readTiUPFile(context,
		getClusterSpaceInTiUP(context, clusterMeta.Cluster.ID),
		"ssh/id_rsa.pub")
	if err != nil {
		err = errors.NewEMErrorf(errors.TIEM_CONNECT_TIDB_ERROR, "sync connection public key failed for cluster %s, err = %s", clusterMeta.Cluster.ID, err)
		framework.LogWithContext(context).Errorf(err.Error())
		return  err
	}
	err = models.GetClusterReaderWriter().CreateClusterTopologySnapshot(context, management.ClusterTopologySnapshot{
		ClusterID: clusterMeta.Cluster.ID,
		TenantID: clusterMeta.Cluster.TenantId,
		PrivateKey: privateKey,
		PublicKey: publicKey,
	})

	return err
}

// syncTopology
// @Description: get meta.yaml from tiup
func syncTopology(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	metaYaml, err := readTiUPFile(context,
		getClusterSpaceInTiUP(context, clusterMeta.Cluster.ID),
		"meta.yaml")
	if err != nil {
		err = errors.NewEMErrorf(errors.TIEM_CONNECT_TIDB_ERROR, "read meta.yaml failed for cluster %s, err = %s", clusterMeta.Cluster.ID, err)
		framework.LogWithContext(context).Errorf(err.Error())
		return  err
	}

	node.Record(fmt.Sprintf("sync topology config for cluster %s ", clusterMeta.Cluster.ID))
	return models.GetClusterReaderWriter().UpdateTopologySnapshotConfig(context, clusterMeta.Cluster.ID, metaYaml)
}

func readTiUPFile(ctx context.Context, clusterHome string, file string) (string, error) {
	fileName := fmt.Sprintf("%s%s", clusterHome, file)
	fileData, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer fileData.Close()

	dataByte, err := ioutil.ReadAll(fileData)
	if err != nil {
		return "", err
	}

	return string(dataByte), nil
}

// stopCluster
// @Description: execute command, stop
func stopCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster

	framework.LogWithContext(context.Context).Infof(
		"stop cluster %s, version = %s", cluster.ID, cluster.Version)
	taskId, err := secondparty.Manager.ClusterStop(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ID, handler.DefaultTiupTimeOut, []string{}, node.ID,
	)

	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s stop error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"get stop cluster %s task id: %s", clusterMeta.Cluster.ID, taskId)

	node.Record(fmt.Sprintf("stop cluster: %s, version: %s ", clusterMeta.Cluster.ID, cluster.Version))
	return nil
}

// destroyCluster
// @Description: execute command, destroy
func destroyCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster

	_, err := models.GetClusterReaderWriter().GetCurrentClusterTopologySnapshot(context, cluster.ID)
	if err != nil {
		framework.LogWithContext(context).Warnf("cluster %s is not really existed", cluster.ID)
		node.Success("skip because cluster is not existed")
		return nil
	}

	taskId, err := secondparty.Manager.ClusterDestroy(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.ID, handler.DefaultTiupTimeOut, []string{}, node.ID,
	)

	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s destroy error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"get destroy cluster %s task id: %s", clusterMeta.Cluster.ID, taskId)

	node.Record(fmt.Sprintf("destroy cluster: %s, version: %s ", clusterMeta.Cluster.ID, cluster.Version))
	return nil
}

// deleteCluster
// @Description: delete cluster from database
func deleteCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	return clusterMeta.Delete(context)
}

// freedClusterResource
// @Description: freed all resource owned by cluster
func freedClusterResource(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	request := &resourceStructs.RecycleRequest{
		RecycleReqs: []resourceStructs.RecycleRequire{
			{
				RecycleType: resourceStructs.RecycleHolder,
				HolderID:    clusterMeta.Cluster.ID,
			},
		},
	}
	err := resourceManagement.GetManagement().GetAllocatorRecycler().RecycleResources(context, request)

	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s freed resource error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"cluster %s freed resource succeed", clusterMeta.Cluster.ID)

	node.Record(fmt.Sprintf("cluster %s freed resource ", clusterMeta.Cluster.ID))
	return nil
}

// initDatabaseAccount
// @Description: init database account for new cluster
func initDatabaseAccount(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster

	tidbServerHost := clusterMeta.GetClusterConnectAddresses()[0].IP
	tidbServerPort := clusterMeta.GetClusterConnectAddresses()[0].Port
	req := secondparty.ClusterSetDbPswReq{
		DbConnParameter: secondparty.DbConnParam{
			Username: cluster.DBUser,
			Password: cluster.DBPassword,
			IP:       tidbServerHost,
			Port:     strconv.Itoa(tidbServerPort),
		},
	}
	err := secondparty.Manager.SetClusterDbPassword(context, req, node.ID)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s init database account error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"cluster %s init database account successfully", clusterMeta.Cluster.ID)

	node.Record(fmt.Sprintf("cluster %s init database account ", clusterMeta.Cluster.ID))
	return nil
}

func fetchTopologyFile(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	req := context.GetData(ContextTakeoverRequest).(cluster.TakeoverClusterReq)

	sshClient, sftpClient, err := openSftpClient(context.Context, req)

	if err != nil {
		return err
	} else {
		defer sshClient.Close()
		defer sftpClient.Close()
	}

	clusterHome := fmt.Sprintf("%sstorage/cluster/clusters/%s/", req.TiUPPath, clusterMeta.Cluster.ID)

	privateKey, err := readRemoteFile(context, sftpClient, clusterHome, "/ssh/id_rsa")
	if err != nil {
		framework.LogWithContext(context).Errorf("read private key failed, error: %s", err.Error())
		return errors.WrapError(errors.TIEM_TAKEOVER_SFTP_ERROR, "read remote file error", err)
	}

	publicKey, err := readRemoteFile(context, sftpClient, clusterHome, "/ssh/id_rsa.pub")
	if err != nil {
		framework.LogWithContext(context).Errorf("read public key failed, error: %s", err.Error())
		return errors.WrapError(errors.TIEM_TAKEOVER_SFTP_ERROR, "read remote file error", err)
	}

	metaData, err := readRemoteFile(context, sftpClient, clusterHome, "meta.yaml")
	if err != nil {
		framework.LogWithContext(context).Errorf("fetch topology file meta.yaml failed, error: %s", err.Error())
		return errors.WrapError(errors.TIEM_TAKEOVER_SFTP_ERROR, "read remote file error", err)
	}

	context.SetData(ContextPrivateKey, privateKey)
	context.SetData(ContextPublicKey, privateKey)
	context.SetData(ContextTopologyConfig, metaData)

	err = models.GetClusterReaderWriter().CreateClusterTopologySnapshot(context, management.ClusterTopologySnapshot{
		ClusterID: clusterMeta.Cluster.ID,
		TenantID: clusterMeta.Cluster.TenantId,
		PrivateKey: string(privateKey),
		PublicKey: string(publicKey),
	})
	if err != nil {
		return err
	}
	err = models.GetClusterReaderWriter().UpdateTopologySnapshotConfig(context, clusterMeta.Cluster.ID, string(metaData))
	if err != nil {
		return err
	}
	node.Record(fmt.Sprintf("fetch topology of cluster %s ", clusterMeta.Cluster.ID))

	return err
}

func readRemoteFile(ctx context.Context, sftp *sftp.Client, clusterHome string, file string) ([]byte, error) {
	filePath := fmt.Sprintf("%s%s", clusterHome, file)
	fileData, err := sftp.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fileData.Close()

	dataByte, err := ioutil.ReadAll(fileData)
	if err != nil {
		return nil, err
	}

	return dataByte, nil
}

// rebuildTopologyFromConfig
// @Description:
// @Parameter node
// @Parameter context
// @return error
func rebuildTopologyFromConfig(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	dataByte := context.GetData(ContextTopologyConfig).([]byte)
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	metadata := &spec.ClusterMeta{}
	err := yaml.Unmarshal(dataByte, metadata)
	if err != nil {
		err = errors.WrapError(errors.TIEM_UNMARSHAL_ERROR, "rebuild topology config failed", err)
		return err
	}

	clusterMeta.Cluster.Type = string(constants.EMProductIDTiDB)
	clusterMeta.Cluster.Version = metadata.Version
	clusterSpec := metadata.GetTopology().(*spec.Specification)
	err = clusterMeta.ParseTopologyFromConfig(context, clusterSpec)
	if err != nil {
		framework.LogWithContext(context).Errorf(
			"add instances into cluster %s topology error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}
	node.Record(fmt.Sprintf("rebuild topology config, add instances into cluster %s topology ", clusterMeta.Cluster.ID))
	return nil
}

// rebuildTiupSpaceForCluster
// @Description:
// @Parameter node
// @Parameter context
// @return error
func rebuildTiupSpaceForCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	snapshot, err := models.GetClusterReaderWriter().GetCurrentClusterTopologySnapshot(context, clusterMeta.Cluster.ID)
	if err != nil {
		return err
	}

	home := getClusterSpaceInTiUP(context, clusterMeta.Cluster.ID)
	err = os.MkdirAll(home + "ssh", 0750)
	if err != nil {
		framework.LogWithContext(context).Errorf("mkdir for cluster %s failed, err = %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}

	metaFile, err := os.Create(home + "meta.yaml")
	if err != nil {
		framework.LogWithContext(context).Errorf("open meta.yaml failed")
		return err
	} else {
		defer metaFile.Close()
	}
	privateKeyFile, err := os.Create(home + "ssh/id_rsa")
	if err != nil {
		framework.LogWithContext(context).Errorf("open private key file failed")
		return err
	} else {
		defer metaFile.Close()
	}
	publicKeyFile, err := os.Create(home + "ssh/id_rsa.pub")
	if err != nil {
		framework.LogWithContext(context).Errorf("open public key file failed")
		return err
	} else {
		defer metaFile.Close()
	}
	metaFile.Write([]byte(snapshot.Config))
	node.Record(fmt.Sprintf("write topology config into meta.yaml for cluster %s ", clusterMeta.Cluster.ID))

	privateKeyFile.Write([]byte(snapshot.PrivateKey))
	publicKeyFile.Write([]byte(snapshot.PublicKey))
	return nil
}

func takeoverResource(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	requirements, instances := clusterMeta.GenerateTakeoverResourceRequirements(context)

	batchReq := &resourceStructs.BatchAllocRequest{
		BatchRequests: []resourceStructs.AllocReq{
			{
				Applicant: resourceStructs.Applicant{
					HolderId:          clusterMeta.Cluster.ID,
					RequestId:         uuidutil.GenerateID(),
					TakeoverOperation: true,
				},
				Requires: requirements,
			},
		},
	}
	allocResponse, err := resourceManagement.GetManagement().GetAllocatorRecycler().AllocResources(context, batchReq)

	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster %s alloc resource error: %s", clusterMeta.Cluster.ID, err.Error())
		return err
	}


	resourceResult := allocResponse.BatchResults[0]
	clusterMeta.Cluster.Region = resourceResult.Results[0].Location.Region

	for i, instance := range instances {
		instance.HostID = resourceResult.Results[i].HostId
		instance.Zone = resourceResult.Results[i].Location.Zone
		instance.Rack = resourceResult.Results[i].Location.Rack
		node.Record(fmt.Sprintf("type: %s, zone: %s, host IP: %s; ", instance.Type, instance.Zone, instance.HostIP[0]))
	}
	node.Record(fmt.Sprintf("cluster %s alloc resource ", clusterMeta.Cluster.ID))
	return nil
}

func testConnectivity(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	connectAddress := clusterMeta.GetClusterConnectAddresses()[0]

	var db *sql.DB
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	return errors.OfNullable(nil).
		BreakIf(func() error {
			sqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql", clusterMeta.Cluster.DBUser, clusterMeta.Cluster.DBPassword, connectAddress.IP, connectAddress.Port))
			db = sqlDB
			return err
		}).
		BreakIf(func() error {
			_, err := db.Query("select * from mysql.db")
			return err
		}).
		If(func(err error) {
			framework.LogWithContext(context).Errorf("test connectivity failed, err = %s", err.Error())
		}).
		Else(func() {
			node.Record(fmt.Sprintf("test TiDB server %s:%d connection ", connectAddress.IP, connectAddress.Port))
		}).
		Present()
}
