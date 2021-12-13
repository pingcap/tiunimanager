/*******************************************************************************
 * @File: executor
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/9
*******************************************************************************/

package management

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/common"
	resourceType "github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	resourceManagement "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

func prepareResource(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	// Alloc resource for new instances
	allocID, err := clusterMeta.AllocInstanceResource(context.Context)
	if err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"cluster[%s] alloc resource error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	context.SetData(ContextAllocResource, allocID)
	framework.LogWithContext(context.Context).Infof(
		"cluster[%s] alloc resource request id: %s", clusterMeta.Cluster.Name, allocID)
	return nil
}

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
	framework.LogWithContext(context.Context).Infof("get scale out cluster task id: %d", taskId)
	return nil
}

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
	framework.LogWithContext(context.Context).Infof("get scale in cluster task id: %d", taskId)
	return nil
}

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

func setClusterOnline(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	if clusterMeta.Cluster.Status == string(constants.ClusterInitializing) {
		if err := clusterMeta.UpdateClusterStatus(context.Context, constants.ClusterRunning); err != nil {
			framework.LogWithContext(context.Context).Errorf(
				"update cluster[%s] status into running error: %s", clusterMeta.Cluster.Name, err.Error())
			return err
		}
	}
	if err := clusterMeta.UpdateInstancesStatus(context.Context,
		constants.ClusterInitializing, constants.ClusterRunning); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"update cluster[%s] instances status into running error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	framework.LogWithContext(context.Context).Infof(
		"set cluster[%s] online successfully", clusterMeta.Cluster.Name)
	return nil
}

func clusterFail(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	allocID := context.GetData(ContextAllocResource)
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

func clusterEnd(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) error {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	if err := clusterMeta.UpdateClusterMaintenanceStatus(context.Context, constants.ClusterMaintenanceNone); err != nil {
		framework.LogWithContext(context.Context).Errorf(
			"set cluster[%s] maintenance status error: %s", clusterMeta.Cluster.Name, err.Error())
		return err
	}
	return nil
}
