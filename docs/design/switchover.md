# Background

In order to meet the disaster recovery needs of users, TiUniManager needs to support the function of master-slave switchover of tidb clusters. It is required to reduce or even completely eliminate data loss in some cases.

# Overview

The current implementation mainly includes three functional parts: master-slave clusters' status check, master-slave switchover, and rollback on error.

* Master-slave clusters' status check：Check the status of master-slave clusters before or after switchover. For example, check whether the current status of master-slave clusters meet the preconditions for switchover, or whether the current status of master-slave clusters meet the expectations after a successful switchover
* Master-slave switchover：The actual function of master-slave switchover
* Rollback on failed workflow：Under some abnormal circumstances, if the auto-rollback of the switchover fails, the master/slave clusters may be in the intermediate abnormal state of the switchover. In this case, it is necessary to manually trigger the rollback on failed workflow process

# OpenAPI Example

Master-slave switchover request URI: 

```
POST /api/v1/clusters/switchover
```

Parameters of the request：

```go
type MasterSlaveClusterSwitchoverReq struct {
    // old master/new slave
    SourceClusterID string `json:"sourceClusterID" validate:"required,min=4,max=64"`
    // new master/old slave
    TargetClusterID string `json:"targetClusterID" validate:"required,min=4,max=64"`
    Force           bool   `json:"force"`
    // if this field is not empty, that means this is a rollback request
    RollbackWorkFlowID string `json:"rollbackWorkFlowID"`
    // clear previous un-cleared "Switching" maintenance state
    RollbackClearPreviousMaintenanceFlag bool `json:"rollbackClearPreviousMaintenanceFlag"`
    // only check if this flag is true
    OnlyCheck               bool `json:"onlyCheck"`
    CheckSlaveReadOnlyFlag  bool `json:"checkSlaveReadOnlyFlag"`
    CheckMasterWritableFlag bool `json:"checkMasterWritableFlag"`
    // check if cluster specified in `SourceClusterID` is standalone, i.e. no cluster relation and no cdc
    // if this flag is true, always only check
    CheckStandaloneClusterFlag bool `json:"checkStandaloneClusterFlag"`
}
```

## Master-slave clusters' status check

```bash
curl -vX 'POST' \
 'http://$IP:$PORT/api/v1/clusters/switchover' \
 -H 'accept: application/json' \
 -H 'Authorization: Bearer 931f170a-7943-485b-a072-817038f7a5d8' \
 -H 'Content-Type: application/json' \
 -d '{
  "sourceClusterId" : "NK66Ks_lRAaHg8r1FRBpQg",
  "targetClusterID" : "A_B1lr0BQCCg9y5cg4Jedg",
  "onlyCheck": true,
  "checkSlaveReadOnlyFlag": true,
  "checkMasterWritableFlag": true
}'
```

## Master-slave switchover

```bash
curl -vX 'POST' \
 'http://$IP:$PORT/api/v1/clusters/switchover' \
 -H 'accept: application/json' \
 -H 'Authorization: Bearer 931f170a-7943-485b-a072-817038f7a5d8' \
 -H 'Content-Type: application/json' \
 -d '{
  "sourceClusterId" : "NK66Ks_lRAaHg8r1FRBpQg",
  "targetClusterID" : "A_B1lr0BQCCg9y5cg4Jedg"
}'
```

## Rollback on error workflow

```bash
curl -vX 'POST' \
 'http://$IP:$PORT/api/v1/clusters/switchover' \
 -H 'accept: application/json' \
 -H 'Authorization: Bearer 931f170a-7943-485b-a072-817038f7a5d8' \
 -H 'Content-Type: application/json' \
 -d '{
  "sourceClusterId" : "NK66Ks_lRAaHg8r1FRBpQg",
  "targetClusterID" : "A_B1lr0BQCCg9y5cg4Jedg",
  "rollbackWorkFlowID":"CHx6yvG4SPyHuS_h4xAzWA"
}'
```

For more information about this interface, please refer to the detailed OpenAPI documentation.

# Detailed Design

## Master-slave switchover

The flow chart:

![master-slave-switchover-workflow-diagram](./images/switchover/master-slave-switchover-workflow-diagram-en.png)



Because the current implementation of workflow component in TiUniManager does not support branches, the workflow of master-slave switchover is divided into three straightforward workflows. When receiving the master-slave switchover instruction, select one of these workflows to execute according to the specific input parameters and the current health status of master-slave clusters:

* WorkFlowMasterSlaveSwitchoverNormal

  Normal switchover: non-force

* WorkFlowMasterSlaveSwitchoverForce	

  Forced switchover: the master-slave clusters work normally and could reach

* WorkFlowMasterSlaveSwitchoverForceWithMasterUnavailable	

  Forced switchover: the master cluster cannot reach or cannot work normally, while the slave cluster can work normally and is reachable

### WorkFlowMasterSlaveSwitchoverNormal

Suppose that the current master cluster is A and the slave cluster is B and C respectively. After a successful switchover A to B, the master cluster becomes B and the slave cluster is A and C respectively, as shown in the following figures:

![switchover-demo](./images/switchover/switchover-demo.png)



List of workflow nodes sorted by the execution sequence:

```go
marshalSwitchoverMasterSlavesState
checkHealthStatus
checkSyncChangeFeedTaskMaxLagTime
setOldMasterReadOnly
waitOldMasterCDCsCaughtUp
pauseOldSyncChangeFeedTask
createReverseSyncChangeFeedTask
checkNewSyncChangeFeedTaskHealth
migrateAllDownStreamSyncChangeFeedTasksToNewMaster
setNewMasterReadWrite
swapMasterSlaveRelationInDB
end
```

#### marshalSwitchoverMasterSlavesState

In this step, some necessary meta information of the whole master-slave clusters will be encrypted, serialized and saved in the Result of the current workflow node to make subsequent rollback-on-error possible. 

#### checkHealthStatus

Confirm that the CDC sync task between cluster A & B is in a healthy state.

#### checkSyncChangeFeedTaskMaxLagTime

Confirm that the data gap between master and slave clusters is less than 30s. If this condition are still not met after some waiting and retries, an error will be returned.

#### setOldMasterReadOnly

Set cluster A readonly.

#### waitOldMasterCDCsCaughtUp

Confirm that all CDC sync tasks on cluster A is finished. If these conditions are still not met after some waiting and retries, an error is returned.

#### pauseOldSyncChangeFeedTask

Suspend the CDC sync task from cluster A to cluster B.

#### createReverseSyncChangeFeedTask

Create a CDC sync task from cluster B to cluster A.

#### checkNewSyncChangeFeedTaskHealth

Confirm that the new CDC sync task created in the previous node is in a normal sync state. If the condition are still not met after some waiting and retries, an error will be returned.

#### migrateAllDownStreamSyncChangeFeedTasksToNewMaster

Migrate all CDC sync tasks on cluster A (but except the previous suspended CDC sync task) from cluster A to cluster B, i.e. to the new master cluster.

#### setNewMasterReadWrite

Set cluster B writeable.

#### swapMasterSlaveRelationInDB

Update the master-slave cluster relations in the TiUniManager DB.

#### end

Do some finalizers, such as deleting the CDC sync task from cluster A to cluster B, clearing the "switching" cluster maintenance status of cluster A, B and C, and finally finish the whole switchover workflow.

#### Auto-Rollback

There are many steps in the process of master-slave switchover. If a step encounters an error during operation (for example, it may be due to inaccessibility of the network or failure of some components), the master-slave switchover workflow will rollback as much as possible and try to return to the state before the master-slave switchover. However, in a few cases, if a step in the rollback process of a failed switchover workflow encounters an error, the switchover workflow will end directly. At such time, the system will be in an intermediate abnormal state of master-slave switchover, and it is necessary to manually trigger this rollback on error wokflow.

For example, if the switchover encounters a failure during execution, it will execute the auto-rollback operation. If "Rollback Successfully" is displayed in the workflow node error message, means the auto-rollback is successful. At such time, although the switchover fails, these clusters will automatically return to the previous state before the switchover workflow is triggered, because no error is encountered in the auto-rollback process, and so there is no need to perform operations like rollback on error workflow. However, if something like "Rollback failed" is displayed in the error message, that means the rollback fails, which is indicating that these clusters are probably in an abnormal state. At such time, the rollback on error workflow should to be triggered manually.

### WorkFlowMasterSlaveSwitchoverForce

List of workflow nodes sorted by the execution sequence:

```go
marshalSwitchoverMasterSlavesState
setOldMasterReadOnly
pauseOldSyncChangeFeedTask
createReverseSyncChangeFeedTask
checkNewSyncChangeFeedTaskHealth
migrateAllDownStreamSyncChangeFeedTasksToNewMaster
setNewMasterReadWrite
swapMasterSlaveRelationInDB
end
```

The only difference between this part and the WorkflowMasterSlaveSwitchoverNormal part is that it skips the process of waiting for CDC synchronization, so such redundant description is skipped.

#### Auto-Rollback

Same as the WorkflowMasterSlaveSwitchoverNormal part.

### WorkFlowMasterSlaveSwitchoverForceWithMasterUnavailable

List of workflow nodes sorted by the execution sequence:

```go
marshalSwitchoverMasterSlavesState
setOldMasterReadOnly
setNewMasterReadWrite
swapMasterSlaveRelationInDB
end
```

This step is only for the case of forced switchover with an unreachable master cluster. And such redundant description is skipped.

#### Auto-Rollback

Same as the WorkflowMasterSlaveSwitchoverNormal part.

## Rollback on failed workflow

List of workflow nodes sorted by the execution sequence:

```go
rollback
end
```

API example:

```bash
curl -vX 'POST' \
 'http://$IP:$PORT/api/v1/clusters/switchover' \
 -H 'accept: application/json' \
 -H 'Authorization: Bearer 931f170a-7943-485b-a072-817038f7a5d8' \
 -H 'Content-Type: application/json' \
 -d '{
  "sourceClusterId" : "NK66Ks_lRAaHg8r1FRBpQg",
  "targetClusterID" : "A_B1lr0BQCCg9y5cg4Jedg",
  "rollbackWorkFlowID":"CHx6yvG4SPyHuS_h4xAzWA"
}'
```

At first in the rollback node, whether the workflow provided in the `rollbackWorkFlowID` has already been automatically rollbacked successfully is checked. If true, success will be returned directly. Otherwise, it will be rollbacked according to the state metadata which has already been stored in the `marshalSwitchoverMasterSlavesState` node of this workflow.

The state metadata of the master-slave clusters:

```go
type switchoverMasterSlavesState struct {
	OldMasterClusterID                string
	OldSlavesClusterIDMapToSyncTaskID map[string]string
	OldSlavesClusterIDMapToCDCIDs     map[string][]string
	CDCsOnMaster                      []*cluster.ChangeFeedTask
}
```

Detailed execution steps of this `rollback` node:

```
1. set cluster A B C readonly
2. remove newly created CDC tasks on B C
3. remove newly created CDC tasks on A
4. resume old CDC tasks on A
5. create previous deleted CDC tasks on A
6. set cluster A writeable
7. reset master-slave relations in TiUniManager DB
8. endMaintenance of cluster A B C
```

If an error occurs at any step, the rollback workflow will exit with error directly. After such error is repaired, the rollback workflow could be triggered again.

The `end` node is currently a NOP and does not contain any logic.

## Master-slave clusters' status check

API example:

```bash
curl -vX 'POST' \
 'http://$IP:$PORT/api/v1/clusters/switchover' \
 -H 'accept: application/json' \
 -H 'Authorization: Bearer 931f170a-7943-485b-a072-817038f7a5d8' \
 -H 'Content-Type: application/json' \
 -d '{
  "sourceClusterId" : "NK66Ks_lRAaHg8r1FRBpQg",
  "targetClusterID" : "A_B1lr0BQCCg9y5cg4Jedg",
  "onlyCheck": true,
  "checkSlaveReadOnlyFlag": true,
  "checkMasterWritableFlag": true
}'
```

This part is relatively simple and does not involve workflow. That is, if `onlyCheck` is true, it will return immediately after the pre-check, and the actual master/slave switchover will not be executed.

In addition to pre-check whether the master cluster is writable and the slave cluster is read-only, pre-checks like whether the target cluster contains CDC component and whether the master/slave relations entry in TiUniManager DB is legal is also implemented by default.

# FT Design

![switchover-ft](./images/switchover/switchover-ft-en.png)

