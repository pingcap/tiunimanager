# TiEM克隆集群设计文档

- Author(s): [YaozhengWang](https://github.com/YaozhengWang)

## 目录

- [TiEM克隆集群设计文档](#TiEM-克隆集群设计文档)
    - [目录](#目录)
    - [克隆集群API](#克隆集群API)
    - [设计](#设计)
        - [workflow](#workflow)
        - [克隆集群流程](#克隆集群流程)
        - [异常错误](#异常错误)

## 克隆集群API
在micro-api/route/route.go中， "/clusters" group定义了克隆模块提供的API,
```go
        cluster := apiV1.Group("/clusters")
        {
        // ...
		// Clone cluster
		cluster.POST("/clone", metrics.HandleMetrics(constants.MetricsClusterClone), clusterApi.Clone)
        // ...
        }
```
可以跟踪每条route后的HandleFunc进入每个API的入口，其中，
- `clusterApi.Clone` 执行克隆操作的接口;

## 设计

clusterManager是扩缩容操作具体流程的入口 (micro-cluster/cluster/management), 主要文件功能如下:
- `manger.go` 克隆集群接口的具体实现；
- `excutor.go` 克隆集群workflow的步骤实现；
- `meta/metabuilder.go` 和创建相关的集群元信息操作实现；
- `meta/metaconfig.go` 和配置相关的集群元信息操作实现；
- `meta/metadisplay.go` 和展示相关的集群元信息操作实现；
- `meta/common.go` 克隆集群相关的通用方法和常量定义；

### workflow
克隆集群的workflow定义:
```go
// clone cluster workflow definition
var cloneDefine = workflow.WorkFlowDefine{
    FlowName: constants.FlowCloneCluster,
    TaskNodes: map[string]*workflow.NodeDefine{
        "start":                   {"prepareResource", "resourceDone", "fail", workflow.SyncFuncNode, prepareResource},
        "resourceDone":            {"modifySourceClusterGCTime", "modifyGCTimeDone", "fail", workflow.SyncFuncNode, modifySourceClusterGCTime},
        "modifyGCTimeDone":        {"backupSourceCluster", "backupDone", "fail", workflow.SyncFuncNode, backupSourceCluster},
        "backupDone":              {"waitBackup", "waitBackupDone", "fail", workflow.SyncFuncNode, waitWorkFlow},
        "waitBackupDone":          {"buildConfig", "configDone", "fail", workflow.SyncFuncNode, buildConfig},
        "configDone":              {"deployCluster", "deployDone", "fail", workflow.PollingNode, deployCluster},
        "deployDone":              {"syncConnectionKey", "syncConnectionKeyDone", "failAfterDeploy", workflow.SyncFuncNode, syncConnectionKey},
        "syncConnectionKeyDone":   {"syncTopology", "syncTopologyDone", "failAfterDeploy", workflow.SyncFuncNode, syncTopology},
        "syncTopologyDone":        {"startCluster", "startDone", "fail", workflow.PollingNode, startCluster},
        "startDone":               {"setClusterOnline", "onlineDone", "failAfterDeploy", workflow.SyncFuncNode, setClusterOnline},
        "onlineDone":              {"initRootAccount", "initRootAccountDone", "failAfterDeploy", workflow.SyncFuncNode, initRootAccount},
        "initRootAccountDone":     {"initAccount", "initAccountDone", "failAfterDeploy", workflow.SyncFuncNode, initDatabaseAccount},
        "initAccountDone":         {"applyParameterGroup", "applyParameterGroupDone", "failAfterDeploy", workflow.SyncFuncNode, workflow.CompositeExecutor(persistCluster, applyParameterGroup)},
        "applyParameterGroupDone": {"syncBackupStrategy", "syncBackupStrategyDone", "failAfterDeploy", workflow.SyncFuncNode, syncBackupStrategy},
        "syncBackupStrategyDone":  {"syncParameters", "syncParametersDone", "failAfterDeploy", workflow.SyncFuncNode, syncParameters},
        "syncParametersDone":      {"waitSyncParam", "waitSyncParamDone", "failAfterDeploy", workflow.SyncFuncNode, waitWorkFlow},
        "waitSyncParamDone":       {"adjustParameters", "initParametersDone", "failAfterDeploy", workflow.SyncFuncNode, adjustParameters},
        "initParametersDone":      {"restoreCluster", "restoreClusterDone", "failAfterDeploy", workflow.SyncFuncNode, restoreCluster},
        "restoreClusterDone":      {"waitRestore", "waitRestoreDone", "failAfterDeploy", workflow.SyncFuncNode, waitWorkFlow},
        "waitRestoreDone":         {"syncIncrData", "syncIncrDataDone", "failAfterDeploy", workflow.SyncFuncNode, syncIncrData},
        "syncIncrDataDone":        {"end", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(recoverSourceClusterGCTime, persistCluster, endMaintenance, asyncBuildLog)},
        "fail":                    {"fail", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(recoverSourceClusterGCTime, setClusterFailure, revertResourceAfterFailure, endMaintenance)},
        "failAfterDeploy":         {"failAfterDeploy", "", "", workflow.SyncFuncNode, workflow.CompositeExecutor(recoverSourceClusterGCTime, setClusterFailure, endMaintenance)},
    },
}
workflowManager.RegisterWorkFlow(context.TODO(), constants.FlowCloneCluster, &cloneDefine)
```
克隆集群的任务流程有多个步骤组成：
- `clone` 克隆集群过程主要包括分配资源，创建目标集群和同步集群参数等;

### 克隆集群流程

克隆集群流程支持基于源集群克隆创建一个新集群，其主要包括如下步骤:
- `prepareResource` 为克隆的目标集群分配资源；
- `modifySourceClusterGCTime` 修改源集群的gc设置；
- `backupSourceCluster` 对源集群进行备份；
- `buildConfig` 为目标集群参数拓扑配置；
- `deployCluster` 部署目标集群；
- `setClusterOnline` 将目标TiDB集群的状态设为running；
- `initAccount and initRootAccount` 为目标TiDB集群初始化相关账户配置；
- `syncParameters` 如果目标集群和源集群的版本一样，同步参数配置到目标集群；
- `restoreCluster` 在目标集群恢复源集群的备份数据；
- `syncIncrData` 将源集群的增量数据同步到目标集群；

### 异常错误

克隆集群API报错定义在common/errors/errorcode.go
```go
    // clone cluster
    TIEM_CHECK_CLUSTER_VERSION_ERROR: {"Cluster version not support", 500},
    TIEM_CDC_NOT_FOUND:               {"TiCDC not found when full clone", 404},
    TIEM_CLONE_TIKV_ERROR:            {"Failed to clone TiKV component", 500},
    TIEM_CLONE_SLAVE_ERROR:           {"Failed to clone when this cluster has been cloned", 500},
```
