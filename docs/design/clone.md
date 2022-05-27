# TiEM clone cluster design doc

- Author(s): [YaozhengWang](https://github.com/YaozhengWang)

## table of Contents

- [TiEM clone cluster design doc](#TiEM-clone-cluster-design-doc)
    - [table of Contents](#table-f-Contents)
    - [clone cluster API](#clone-cluster-API)
    - [design](#design)
        - [workflow](#workflow)
        - [clone cluster process](#clone-cluster-process)
        - [errors](#errors)

## clone cluster API
micro-api/route/route.go, "/clusters" group define clone cluster API,
```go
        cluster := apiV1.Group("/clusters")
        {
        // ...
		// Clone cluster
		cluster.POST("/clone", metrics.HandleMetrics(constants.MetricsClusterClone), clusterApi.Clone)
        // ...
        }
```
The `HandleFunc` of each route can be tracked as the entry of each API，
- `clusterApi.Clone` clone the specified TiDB cluster;

## design

clusterManager is manager of clone cluster (micro-cluster/cluster/management), main files function:
- `manger.go` implement of clone cluster；
- `excutor.go` implement of clone cluster workflow nodes；
- `meta/metabuilder.go` implement cluster meta operators related to builder;
- `meta/metaconfig.go` implement cluster meta operators related to config;
- `meta/metadisplay.go` implement cluster meta operators related to display;
- `meta/common.go` common functions and constants definition of clone cluster;

### workflow
workflow definition of clone cluster:
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
The workflow consists of multiple workflow nodes，
- `clone` this process mainly includes allocating resource, creating cluster and setting cluster status etc.;

### clone cluster process

The clone cluster process will support to create a new cluster based on the specified source cluster, It mainly includes the
following steps:
- `prepareResource` Allocating resources for cloning TiDB cluster
- `modifySourceClusterGCTime` Modifying source cluster gc time
- `backupSourceCluster` Backup the source cluster
- `buildConfig` Generating topology configuration for cloning cluster
- `deployCluster` Deploying the TiDB cluster
- `setClusterOnline` Setting TiDB cluster status into running
- `initAccount and initRootAccount` Initialling related accounts for the target TiDB cluster
- `syncParameters` Syncing the source cluster parameters into the target cluster
- `restoreCluster` Restoring the backup data into the target cluster
- `syncIncrData` Syncing incremental data into the target cluster

### errors

clone cluster API errors define in file common/errors/errorcode.go
```go
    // clone cluster
    TIEM_CHECK_CLUSTER_VERSION_ERROR: {"Cluster version not support", 500},
    TIEM_CDC_NOT_FOUND:               {"TiCDC not found when full clone", 404},
    TIEM_CLONE_TIKV_ERROR:            {"Failed to clone TiKV component", 500},
    TIEM_CLONE_SLAVE_ERROR:           {"Failed to clone when this cluster has been cloned", 500},
```
