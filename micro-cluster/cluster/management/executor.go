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
	resourceType "github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	resourceManagement "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
)

func PrepareResource(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) bool {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	// Alloc resource for new instances
	allocID, err := clusterMeta.AllocInstanceResource(context.Context)
	if err != nil {
		node.Fail(err)
		return false
	}
	context.SetData(ContextAllocResource, allocID)
	framework.LogWithContext(context.Context).Infof(
		"Scale out cluster[%s], alloc resource request id: %s", clusterMeta.Cluster.Name, allocID)
	node.Success("success")
	return true
}

func BuildConfig(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) bool {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)

	topology, err := clusterMeta.GenerateTopologyConfig(context.Context)
	if err != nil {
		node.Fail(err)
		return false
	}

	context.SetData(ContextTopology, topology)
	node.Success("success")
	return true
}

func ScaleOutCluster(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) bool {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	cluster := clusterMeta.Cluster
	yamlConfig := context.GetData(ContextTopology).(string)

	framework.LogWithContext(context.Context).Infof("Scale out cluster[%s], version = %s, yamlConfig = %s", cluster.Name, cluster.Version, yamlConfig)
	taskId, err := secondparty.Manager.ClusterScaleOut(
		context.Context, secondparty.ClusterComponentTypeStr, cluster.Name,
		yamlConfig, 0, []string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa"}, node.ID)
	if err != nil {
		node.Fail(err)
		return false
	}
	framework.LogWithContext(context.Context).Infof("Get scale out cluster task id: %d", taskId)
	node.Success("success")
	return true
}

func SetClusterOnline(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) bool {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	if err := clusterMeta.SetClusterOnline(context.Context); err != nil {
		node.Fail(err)
		return false
	}
	framework.LogWithContext(context.Context).Infof("Set cluster[%s] status into running", clusterMeta.Cluster.Name)
	node.Success("success")
	return true
}

func ClusterFail(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) bool {
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
			framework.LogWithContext(context).Errorf("RecycleResources error, %s", err.Error())
			node.Fail(err)
			return false
		}
	}
	node.Success("success")
	return true
}

func ClusterEnd(node *workflowModel.WorkFlowNode, context *workflow.FlowContext) bool {
	clusterMeta := context.GetData(ContextClusterMeta).(*handler.ClusterMeta)
	var status constants.ClusterMaintenanceStatus = ""
	if err := clusterMeta.SetClusterMaintenanceStatus(context.Context, status); err != nil {
		node.Fail(err)
		return false
	}
	node.Success("success")
	return true
}
