# TiUniManager scale-out and scale-in design doc

- Author(s): [YaozhengWang](https://github.com/YaozhengWang)

## table of Contents

- [TiUniManager scale-out and scale-in design doc](#TiUniManager-scale-out-and-scale-in-design-doc)
    - [table of Contents](#table-f-Contents)
    - [scale-out and scale-in API](#scale-out-and-scale-in-API)
    - [design](#design)
        - [workflow](#workflow)
        - [scale-out process](#scale-out-process)
        - [scale-in process](#scale-in-process)
        - [errors](#errors)

## scale-out and scale-in API
micro-api/route/route.go, "/clusters" group define scale-out and scale-in API,
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
The `HandleFunc` of each route can be tracked as the entry of each API，
- `clusterApi.ScaleOut` scale out the specified TiDB cluster;
- `clusterApi.ScaleIn` scale in the specified TiDB cluster;

## design

clusterManager is manager of scale-out and scale-in (micro-cluster/cluster/management), main files function:
- `manger.go` implement of scale-in and scale-out；
- `excutor.go` implement of scale-out and scale-in workflow nodes；
- `meta/metabuilder.go` implement cluster meta operators related to builder;
- `meta/metaconfig.go` implement cluster meta operators related to config;
- `meta/metadisplay.go` implement cluster meta operators related to display;
- `meta/common.go` common functions and constants definition of scale-out and scale-in;

### workflow
workflow definition of scale-out and scale-in:
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
The workflow consists of multiple workflow nodes，
- `scale-out` this process mainly includes allocating resource, scaling out cluster and setting cluster status etc.;
- `scale-in` this process mainly includes scaling in cluster, checking cluster status and recycling resources etc.;
### scale-out process

The scale-out process will support to scale out PD, TiDB, TiKV, TiFlash and TiCDC components, It mainly includes the 
following steps:
- `prepareResource` Allocating resources for scaling TiDB cluster
- `buildConfig` Generating topology configuration for scaling out 
- `scaleOutCluster` Scaling out TiDB cluster
- `setClusterOnline` Setting TiDB cluster status into running
- `updateClusterParameters` Updating cluster parameters by the specified parameter template

### scale-in process

The scale-in process will support to scale in PD, TiDB, TiKV, TiFlash and TiCDC components, It mainly includes the 
following steps:
- `scaleInCluster` Scaling in TiDB cluster
- `checkInstanceStatus` Checking status of instances in the TiDB cluster
- `freeInstanceResource` Recycling resources

### errors

scale-out and scale-in API errors define in file common/errors/errorcode.go
```go
    // scale out & scale in
    TIUNIMANAGER_INSTANCE_NOT_FOUND:               {"Instance of cluster is not found", 404},
    TIUNIMANAGER_CONNECT_TIDB_ERROR:               {"Failed to connect TiDB instances", 500},
    TIUNIMANAGER_DELETE_INSTANCE_ERROR:            {"Failed to delete cluster instance", 500},
    TIUNIMANAGER_CHECK_PLACEMENT_RULES_ERROR:      {"Placement rule is not set when scale out TiFlash", 409},
    TIUNIMANAGER_CHECK_TIFLASH_MAX_REPLICAS_ERROR: {"The number of remaining TiFlash instances is less than the maximum replicas of data tables", 409},
    TIUNIMANAGER_SCAN_MAX_REPLICA_COUNT_ERROR:     {"Failed to scan max replicas of data tables of TiFlash", 500},    
```
