## /changefeeds/

### GET
#### Summary

query change feed tasks of a cluster

#### Description

query change feed tasks of a cluster

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | query |  cluster id | Yes | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & [cluster.QueryChangeFeedTaskResp](#querychangefeedtaskresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### cluster.QueryChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string | _Example:_ `"CLUSTER_ID_IN_TIEM__22"` | No |
| createTime | string |  | No |
| downstream | object |  | No |
| downstreamFetchTs | string | _Example:_ `"415241823337054209"` | No |
| downstreamFetchUnix | integer | _Example:_ `1642402879000` | No |
| downstreamSyncTs | string | _Example:_ `"415241823337054209"` | No |
| downstreamSyncUnix | integer | _Example:_ `1642402879000` | No |
| downstreamType | string | _Enum:_ `"tidb"`, `"kafka"`, `"mysql"`<br>_Example:_ `"tidb"` | No |
| id | string | _Example:_ `"CLUSTER_ID_IN_TIEM__22"` | No |
| name | string | _Example:_ `"my_sync_name"` | No |
| rules | [ string ] | _Example:_ `["*.*"]` | No |
| startTS | string | _Example:_ `"415241823337054209"` | No |
| startUnix | integer | _Example:_ `1642402879000` | No |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |
| unsteady | boolean | _Example:_ `false` | No |
| updateTime | string |  | No |
| upstreamUpdateTs | string | _Example:_ `"415241823337054209"` | No |
| upstreamUpdateUnix | integer | _Example:_ `1642402879000` | No |

### POST
#### Summary

create a change feed task

#### Description

create a change feed task

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTask | body | change feed task request | Yes | [cluster.CreateChangeFeedTaskReq](#clustercreatechangefeedtaskreq) |

#### cluster.CreateChangeFeedTaskReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string | _Example:_ `"CLUSTER_ID_IN_TIEM__22"` | Yes |
| downstream | object |  | No |
| downstreamType | string | _Enum:_ `"tidb"`, `"kafka"`, `"mysql"`<br>_Example:_ `"tidb"` | Yes |
| name | string | _Example:_ `"my_sync_name"` | Yes |
| rules | [ string ] | _Example:_ `["*.*"]` | No |
| startTS | string | _Example:_ `"415241823337054209"` | No |

#### cluster.TiDBDownstream

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| concurrentThreads | integer | _Example:_ `5` | No |
| ip | string | _Example:_ `"127.0.0.1"` | No |
| maxTxnRow | integer | _Example:_ `4` | No |
| password | string | _Example:_ `"my_password"` | No |
| port | integer | _Example:_ `4534` | No |
| targetClusterId | string | _Example:_ `"CLUSTER_ID_IN_TIEM__22"` | No |
| tls | boolean | _Example:_ `false` | No |
| username | string | _Example:_ `"tidb"` | No |
| workerCount | integer | _Example:_ `2` | No |

#### cluster.KafkaDownstream

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clientId | string | _Example:_ `"213"` | No |
| dispatchers | [ [cluster.Dispatcher](#clusterdispatcher) ] |  | No |
| ip | string | _Example:_ `"127.0.0.1"` | No |
| maxBatchSize | integer | _Example:_ `5` | No |
| maxMessageBytes | integer | _Example:_ `16` | No |
| partitions | integer | _Example:_ `1` | No |
| port | integer | _Example:_ `9001` | No |
| protocol | string | _Enum:_ `"default"`, `"canal"`, `"avro"`, `"maxwell"`<br>_Example:_ `"default"` | No |
| replicationFactor | integer | _Example:_ `1` | No |
| tls | boolean | _Example:_ `false` | No |
| topicName | string | _Example:_ `"my_topic"` | No |
| version | string | _Example:_ `"2.4.0"` | No |

#### cluster.TiDBDownstream

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| concurrentThreads | integer | _Example:_ `5` | No |
| ip | string | _Example:_ `"127.0.0.1"` | No |
| maxTxnRow | integer | _Example:_ `4` | No |
| password | string | _Example:_ `"my_password"` | No |
| port | integer | _Example:_ `4534` | No |
| targetClusterId | string | _Example:_ `"CLUSTER_ID_IN_TIEM__22"` | No |
| tls | boolean | _Example:_ `false` | No |
| username | string | _Example:_ `"tidb"` | No |
| workerCount | integer | _Example:_ `2` | No |

#### cluster.Dispatcher

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| dispatcher | string | _Example:_ `"ts"` | No |
| matcher | string | _Example:_ `"test1.*"` | No |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.CreateChangeFeedTaskResp](#clustercreatechangefeedtaskresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### cluster.CreateChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | _Example:_ `"TASK_ID_IN_TIEM____22"` | No |

## /changefeeds/{changeFeedTaskId}

### DELETE
#### Summary

delete a change feed task

#### Description

delete a change feed task

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.DeleteChangeFeedTaskResp](#clusterdeletechangefeedtaskresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### cluster.DeleteChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | _Example:_ `"TASK_ID_IN_TIEM____22"` | No |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |

## /changefeeds/{changeFeedTaskId}/

### GET
#### Summary

get change feed detail

#### Description

get change feed detail

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.DetailChangeFeedTaskResp](#clusterdetailchangefeedtaskresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### cluster.DetailChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string | _Example:_ `"CLUSTER_ID_IN_TIEM__22"` | No |
| createTime | string |  | No |
| downstream | [cluster.TiDBDownstream](#clustertidbdownstream) OR [cluster.MysqlDownstream](#clustermysqldownstream) OR [cluster.KafkaDownstream](#clusterkafkadownstream) |  | No |
| downstreamFetchTs | string | _Example:_ `"415241823337054209"` | No |
| downstreamFetchUnix | integer | _Example:_ `1642402879000` | No |
| downstreamSyncTs | string | _Example:_ `"415241823337054209"` | No |
| downstreamSyncUnix | integer | _Example:_ `1642402879000` | No |
| downstreamType | string | _Enum:_ `"tidb"`, `"kafka"`, `"mysql"`<br>_Example:_ `"tidb"` | No |
| id | string | _Example:_ `"CLUSTER_ID_IN_TIEM__22"` | No |
| name | string | _Example:_ `"my_sync_name"` | No |
| rules | [ string ] | _Example:_ `["*.*"]` | No |
| startTS | string | _Example:_ `"415241823337054209"` | No |
| startUnix | integer | _Example:_ `1642402879000` | No |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |
| unsteady | boolean | _Example:_ `false` | No |
| updateTime | string |  | No |
| upstreamUpdateTs | string | _Example:_ `"415241823337054209"` | No |
| upstreamUpdateUnix | integer | _Example:_ `1642402879000` | No |

## /changefeeds/{changeFeedTaskId}/pause

### POST
#### Summary

pause a change feed task

#### Description

pause a change feed task

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.PauseChangeFeedTaskResp](#clusterpausechangefeedtaskresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

### cluster.PauseChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |

## /changefeeds/{changeFeedTaskId}/resume

### POST
#### Summary

resume a change feed task

#### Description

resume a change feed task

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.ResumeChangeFeedTaskResp](#clusterresumechangefeedtaskresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### cluster.ResumeChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |

## /changefeeds/{changeFeedTaskId}/update

### POST
#### Summary

resume a change feed

#### Description

resume a change feed

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |
| task | body | change feed task | Yes | [cluster.UpdateChangeFeedTaskReq](#clusterupdatechangefeedtaskreq) |

#### cluster.UpdateChangeFeedTaskReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| downstream | [cluster.TiDBDownstream](#clustertidbdownstream) OR [cluster.MysqlDownstream](#clustermysqldownstream) OR [cluster.KafkaDownstream](#clusterkafkadownstream) |  | No |
| downstreamType | string | _Enum:_ `"tidb"`, `"kafka"`, `"mysql"`<br>_Example:_ `"tidb"` | No |
| name | string | _Example:_ `"my_sync_name"` | Yes |
| rules | [ string ] | _Example:_ `["*.*"]` | No |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.UpdateChangeFeedTaskResp](#clusterupdatechangefeedtaskresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### cluster.UpdateChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |

## commonModel

### controller.CommonResult

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| data | object |  | No |
| message | string |  | No |

### controller.Page

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| page | integer |  | No |
| pageSize | integer |  | No |
| total | integer |  | No |

### controller.ResultWithPage

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| data | object |  | No |
| message | string |  | No |
| page | [controller.Page](#controllerpage) |  | No |
