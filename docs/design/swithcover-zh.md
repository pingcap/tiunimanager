# 背景

为满足用户的容灾需求，TiEM 需要备支持 TiDB 集群主备切换的功能，要求尽可能减少甚至在某些情况下能完全杜绝数据丢失。

# 概览

当前实现主要包含三个功能部分：主备集群状态检查、主备切换、主备切换异常回滚。

* 主备集群状态检查：主备切换前、后的主备集群状态检查，比如，检测当前主备集群状态是否满足能够进行主备切换的前提条件，或者检测当前主备集群状态是否符合主备切换完成之后的预期
* 主备切换：主要的主备切换逻辑实现（包含一定的自动回滚能力）
* 主备切换异常回滚：在异常情况下，如果主备切换的自动回滚失败，此时主备集群将可能会处于主备切换的中间异常状态，这时就需要手动触发主备切换异常回滚流程

# OpenAPI 示例

主备切换请求 URI: 

```
POST /api/v1/clusters/switchover
```

请求入参：

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

## 主备切换预检查

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

## 主备切换

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

## 主备切换异常回滚

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

关于接口的更多信息请查阅 OpenAPI 文档。

# 业务流程

## 主备切换

流程图如下：



![master-slave-switchover-workflow-diagram](./images/switchover/master-slave-switchover-workflow-diagram.png)



因为当前 TiEM 中的 WorkFlow 组件不支持分支执行，故将之拆分为三个基本流程，在收到主备切换指令时根据具体输入参数和当前各主备集群健康状态进行 Switch：

* WorkFlowMasterSlaveSwitchoverNormal

  普通切换：非强制

* WorkFlowMasterSlaveSwitchoverForce	

  强制切换：强制，主备集群均正常工作并且可达

* WorkFlowMasterSlaveSwitchoverForceWithMasterUnavailable	

  强制切换：强制，主集群不可达或者不能正常工作，备集群能够正常工作并且可达

### WorkFlowMasterSlaveSwitchoverNormal

假设当前主集群为 A，备集群分别为 B 和 C，切换后主集群为 B，备集群分别为 A 和 C，如下图所示：



![switchover-demo](./images/switchover/switchover-demo.png)



工作流节点顺序执行列表：

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

这一步会把整个主备集群的一些必要元信息加密和序列化之后保存在 workflow 的 Result 中，以便于后续可能的异常恢复。

#### checkHealthStatus

确认集群 AB 之间的 CDC 同步任务处于健康状态。

#### checkSyncChangeFeedTaskMaxLagTime

确认主备之间数据差距小于 30s，若经过多次等待和检查之后发现依然不满足条件，返回错误。

#### setOldMasterReadOnly

将集群 A 设置为只读。

#### waitOldMasterCDCsCaughtUp

确认集群 A 上的所有 CDC 同步任务全部将数据追平，若经过多次等待和检查之后发现依然不满足条件，返回错误。

#### pauseOldSyncChangeFeedTask

暂停从集群 A 向集群 B 的 CDC 同步任务。

#### createReverseSyncChangeFeedTask

创建从集群 B 向集群 A 的 CDC 同步任务。

#### checkNewSyncChangeFeedTaskHealth

确认上一个流程所创建的新 CDC 同步处于正常同步状态，若经过多次等待和检查之后发现依然不满足条件，返回错误。

#### migrateAllDownStreamSyncChangeFeedTasksToNewMaster

将集群 A 上除了已经被暂停的从集群 A 向集群 B 的 CDC 同步任务之外的所有 CDC 同步任务迁移到新主集群上，即集群 B。

#### setNewMasterReadWrite

将集群 B 设置为可写状态。

#### swapMasterSlaveRelationInDB

更新 TiEM 数据库中的集群主备关系。

#### end

做一些收尾工作，比如将从集群 A 向集群 B 的 CDC 同步任务删除，并集群 ABC 的“切换中”集群维护状态清除，最终结束整个切换流程。

#### 自动回滚

在主备切换过程中会包含很多步骤，如果某一个步骤在操作过程中遇到错误（比如可能因为网络不可达或者某些组件故障），主备切换会尽可能地进行自动回滚，努力回退到主备切换之前的状态。在少数情况下，如果在回滚过程中某一步遇到了错误，切换流程会直接结束，那么此时系统可能会处于主备切换的中间异常状态，这时就需要进行手工触发异常回滚流程。

![switchover-rollback-successfully](./images/switchover/switchover-rollback-successfully.png)

比如在上述情景中，主备切换在执行过程中遇到故障，执行自动回滚操作，错误信息中显示 ”Rollback Successfully“ 即为回滚成功，那么此时虽然主备切换失败，但是由于自动回滚过程中没有遇到错误，集群将会顺利回退到主备切换之前的状态，所以无需执行异常恢复相关的操作。但如果错误信息中显示 ”Rollback Failed“，即为回滚失败，就意味着集群很可能正处于异常状态，此时需要进行手工触发异常回滚流程。

### WorkFlowMasterSlaveSwitchoverForce

工作流节点顺序执行列表：

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

由于这一部分和 WorkFlowMasterSlaveSwitchoverNormal 部分的唯一区别就是跳过了等待同步的过程，所以不再一一赘述各个步骤的执行细节。

#### 自动回滚

此部分和 WorkFlowMasterSlaveSwitchoverNormal 内容相同，不再赘述。

### WorkFlowMasterSlaveSwitchoverForceWithMasterUnavailable

工作流节点顺序执行列表：

```go
marshalSwitchoverMasterSlavesState
setOldMasterReadOnly
setNewMasterReadWrite
swapMasterSlaveRelationInDB
end
```

这一步只针对强制切换并且主集群不可达的情况，所以不再一一赘述各个步骤的执行细节。

#### 自动回滚

此部分和 WorkFlowMasterSlaveSwitchoverNormal 内容相同，不再赘述。

## 主备切换异常回滚

工作流节点顺序执行列表：

```go
rollback
end
```

主备切换异常回滚的 API 形如下：

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

在 rollback 节点，会首先检测 rollbackWorkFlowID 是否已经自动回滚成功，如果是，将会直接返回成功，否则，将根据要回滚工作流 marshalSwitchoverMasterSlavesState 节点里存储备份的主备集群元数据进行回滚。

主备集群元数据形如下：

```go
type switchoverMasterSlavesState struct {
	OldMasterClusterID                string
	OldSlavesClusterIDMapToSyncTaskID map[string]string
	OldSlavesClusterIDMapToCDCIDs     map[string][]string
	CDCsOnMaster                      []*cluster.ChangeFeedTask
}
```

rollback 节点详细执行过程：

```
1. set cluster A B C readonly
2. remove newly created CDC tasks on B C
3. remove newly created CDC tasks on A
4. resume old CDC tasks on A
5. create previous deleted CDC tasks on A
6. set cluster A writeable
7. reset master-slave relations in TiEM DB
8. endMaintenance of cluster A B C
```

如果在任何步骤出现错误，异常回滚流程会直接错误退出，当错误得到修复之后可再次执行异常回滚流程。

end 节点目前为 NOP，没有包含任何逻辑操作。

## 主备切换预检查

API 形如下：

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

这一部分比较简单，没有涉及到工作流。在 onlyCheck 为真的情况下，在进行完预检查之后就会立即返回，并不会执行实际的主备切换逻辑。

除了可以进行主集群是否可写、备集群是否只读的预检查之外，默认还进行了 targetCluster 是否包含 CDC 组件和 TiEM DB 中主备关系条目是否合法的预检查。

# FT 设计

![switchover-ft](./images/switchover/switchover-ft.png)

