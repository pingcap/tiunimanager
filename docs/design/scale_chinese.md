# TiUniManager 扩缩容设计文档

- Author(s): [YaozhengWang](https://github.com/YaozhengWang)

## 目录

- [TiUniManager扩缩容设计文档](#TiUniManager-扩缩容设计文档)
    - [目录](#目录)
    - [扩缩容API](#扩缩容API)
    - [设计](#设计)
        - [workflow](#workflow)
        - [扩容流程](#扩容流程)
        - [缩容流程](#缩容流程)
        - [异常错误](#异常错误)

## 扩缩容API
在micro-api/route/route.go中, "/clusters" group定义了扩缩容模块提供的API,
```go
        cluster := apiV1.Group("/clusters")
        {
        // ...
		// Scale cluster
		cluster.POST("/:clusterId/scale-out", metrics.HandleMetrics(constants.MetricsClusterScaleOut), clusterApi.ScaleOut)
		cluster.POST("/:clusterId/scale-in", metrics.HandleMetrics(constants.MetricsClusterScaleIn), clusterApi.ScaleIn)
        // ...
        }
```
可以跟踪每条route后的HandleFunc进入每个API的入口，其中，
- `clusterApi.ScaleOut` 执行扩容操作接口；
- `clusterApi.ScaleIn` 执行缩容操作接口；

## 设计

clusterManager是扩缩容操作具体流程的入口 (micro-cluster/cluster/management), 主要文件功能如下:
- `manger.go` 扩缩容接口的具体实现；
- `excutor.go` 扩缩容workflow的步骤实现；
- `meta/metabuilder.go` 和创建相关的集群元信息操作实现；
- `meta/metaconfig.go` 和配置相关的集群元信息操作实现；
- `meta/metadisplay.go` 和展示相关的集群元信息操作实现；
- `meta/common.go` 扩缩容相关的通用方法和常量定义；

### workflow
扩容和缩容任务流程步骤如下：
```go
// scale-out workflow definition
var scaleOutDefine = workflow.WorkFlowDefine{
	FlowName: constants.FlowScaleOutCluster,
	TaskNodes: map[string]*workflow.NodeDefine{
		"start":            {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
		"resourceDone":     {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
		"configDone":       {"scaleOutCluster", "scaleOutDone", "fail", workflow.PollingNode, scaleOutCluster},
		"scaleOutDone":     {"syncTopology", "syncTopologyDone", "fail", workflow.SyncFuncNode, syncTopology},
		"syncTopologyDone": {"getTypes", "getTypesDone", "fail", workflow.SyncFuncNode, getFirstScaleOutTypes},
		"getTypesDone":     {"setClusterOnline", "onlineDone", "fail", workflow.SyncFuncNode, setClusterOnline},
		"onlineDone":       {"updateClusterParameters", "updateDone", "failAfterScale", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, updateClusterParameters)},
		"updateDone":       {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance, asyncBuildLog)},
		"fail":             {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(revertResourceAfterFailure, endMaintenance)},
		"failAfterScale":   {"failAfterScale", "", "", workflow.SyncFuncNode, endMaintenance},
	},
}
workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleOutCluster, &scaleOutDefine)

// scale-in workflow definition
var scaleInDefine = workflow.WorkFlowDefine{
    FlowName: constants.FlowScaleInCluster,
    TaskNodes: map[string]*workflow.NodeDefine{
        "start":       {"scaleInCluster", "scaleInDone", "fail", workflow.PollingNode, scaleInCluster},
        "scaleInDone": {"checkInstanceStatus", "checkDone", "fail", workflow.SyncFuncNode, checkInstanceStatus},
        "checkDone":   {"freeInstanceResource", "freeDone", "fail", workflow.SyncFuncNode, freeInstanceResource},
        "freeDone":    {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, endMaintenance)},
        "fail":        {"fail", "", "", workflow.SyncFuncNode, endMaintenance},
    },
}
workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowScaleInCluster, &scaleInDefine)
```
扩缩容的任务流程有多个步骤组成：
- `scale-out` 扩容过程主要包括资源分配，扩容集群和设置集群状态等；
- `scale-in` 缩容过程主要包括缩容集群，检查集群状态和回收资源等；
### 扩容流程

扩容过程支持PD，TiDB，TiKV，TiFlash和TiCDC等组件的扩容，主要包括如下步骤:
- `prepareResource` 为扩容TiDB集群分配资源；
- `buildConfig` 为扩容TiDB集群产生拓扑配置文件；
- `scaleOutCluster` 扩容TiDB集群；
- `setClusterOnline` 将集群状态设为running；
- `updateClusterParameters` 通过指定的参数模板更新集群参数；

### 缩容流程

缩容过程支持PD，TiDB，TiKV，TiFlash和TiCDC等组件的缩容，主要包括如下步骤：
- `scaleInCluster` 缩容TiDB集群
- `checkInstanceStatus` 检查TiDB集群的状态
- `freeInstanceResource` 回收相关资源

### 异常错误

扩容和缩容API报错定义在common/errors/errorcode.go
```go
    // scale out & scale in
    TIUNIMANAGER_INSTANCE_NOT_FOUND:               {"Instance of cluster is not found", 404},
    TIUNIMANAGER_CONNECT_TIDB_ERROR:               {"Failed to connect TiDB instances", 500},
    TIUNIMANAGER_DELETE_INSTANCE_ERROR:            {"Failed to delete cluster instance", 500},
    TIUNIMANAGER_CHECK_PLACEMENT_RULES_ERROR:      {"Placement rule is not set when scale out TiFlash", 409},
    TIUNIMANAGER_CHECK_TIFLASH_MAX_REPLICAS_ERROR: {"The number of remaining TiFlash instances is less than the maximum replicas of data tables", 409},
    TIUNIMANAGER_SCAN_MAX_REPLICA_COUNT_ERROR:     {"Failed to scan max replicas of data tables of TiFlash", 500},    
```
