# EM UI API
EM UI API

## Version: 1.0

**Contact information:**  
zhangpeijin  
zhangpeijin@pingcap.com  

**License:** [Apache 2.0](http://www.apache.org/licenses/LICENSE-2.0.html)

### Security
**ApiKeyAuth**  

|apiKey|*API Key*|
|---|---|
|In|header|
|Name|Authorization|

### /backups/

#### GET
##### Summary

query backup records of a cluster

##### Description

query backup records of a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| backupId | query |  | No | string |
| clusterId | query |  | No | string |
| endTime | query |  | No | integer |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |
| startTime | query |  | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

backup

##### Description

backup

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| backupReq | body | backup request | Yes | [cluster.BackupClusterDataReq](#clusterbackupclusterdatareq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /backups/{backupId}

#### DELETE
##### Summary

delete backup record

##### Description

delete backup record

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| backupId | path | backup record id | Yes | integer |
| backupDeleteReq | body | backup delete request | Yes | [cluster.DeleteBackupDataReq](#clusterdeletebackupdatareq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /backups/cancel

#### POST
##### Summary

cancel backup

##### Description

cancel backup

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| backupReq | body | cancel backup request | Yes | [cluster.CancelBackupReq](#clustercancelbackupreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /changefeeds/

#### GET
##### Summary

query change feed tasks of a cluster

##### Description

query change feed tasks of a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | query |  | Yes | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

create a change feed task

##### Description

create a change feed task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTask | body | change feed task request | Yes | [cluster.CreateChangeFeedTaskReq](#clustercreatechangefeedtaskreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /changefeeds/{changeFeedTaskId}

#### DELETE
##### Summary

delete a change feed task

##### Description

delete a change feed task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /changefeeds/{changeFeedTaskId}/

#### GET
##### Summary

get change feed detail

##### Description

get change feed detail

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /changefeeds/{changeFeedTaskId}/pause

#### POST
##### Summary

pause a change feed task

##### Description

pause a change feed task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /changefeeds/{changeFeedTaskId}/resume

#### POST
##### Summary

resume a change feed task

##### Description

resume a change feed task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /changefeeds/{changeFeedTaskId}/update

#### POST
##### Summary

resume a change feed

##### Description

resume a change feed

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| changeFeedTaskId | path | changeFeedTaskId | Yes | string |
| task | body | change feed task | Yes | [cluster.UpdateChangeFeedTaskReq](#clusterupdatechangefeedtaskreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/

#### GET
##### Summary

query clusters

##### Description

query clusters

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | query |  | No | string |
| clusterName | query |  | No | string |
| clusterStatus | query |  | No | string |
| clusterTag | query |  | No | string |
| clusterType | query |  | No | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

create a cluster

##### Description

create a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| createReq | body | create request | Yes | [cluster.CreateClusterReq](#clustercreateclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}

#### DELETE
##### Summary

delete cluster

##### Description

delete cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |
| deleteReq | body | delete request | No | [cluster.DeleteClusterReq](#clusterdeleteclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### GET
##### Summary

show details of a cluster

##### Description

show details of a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/dashboard

#### GET
##### Summary

dashboard

##### Description

dashboard

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/log

#### GET
##### Summary

query cluster log

##### Description

query cluster log

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | clusterId | Yes | string |
| endTime | query |  | No | integer |
| ip | query |  | No | string |
| level | query |  | No | string |
| message | query |  | No | string |
| module | query |  | No | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |
| startTime | query |  | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/monitor

#### GET
##### Summary

describe monitoring link

##### Description

describe monitoring link

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/params

#### GET
##### Summary

query parameters of a cluster

##### Description

query parameters of a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| instanceType | query |  | No | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |
| paramName | query |  | No | string |
| clusterId | path | clusterId | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### PUT
##### Summary

submit parameters

##### Description

submit parameters

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| updateReq | body | update params request | Yes | [cluster.UpdateClusterParametersReq](#clusterupdateclusterparametersreq) |
| clusterId | path | clusterId | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/params/inspect

#### POST
##### Summary

inspect parameters

##### Description

inspect parameters

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | clusterId | Yes | string |
| inspectReq | body | inspect params request | Yes | [cluster.InspectParametersReq](#clusterinspectparametersreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/preview-scale-out

#### POST
##### Summary

preview cluster topology and capability

##### Description

preview cluster topology and capability

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |
| scaleOutReq | body | scale out request | Yes | [cluster.ScaleOutClusterReq](#clusterscaleoutclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/restart

#### POST
##### Summary

restart a cluster

##### Description

restart a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/scale-in

#### POST
##### Summary

scale in a cluster

##### Description

scale in a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |
| scaleInReq | body | scale in request | Yes | [cluster.ScaleInClusterReq](#clusterscaleinclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/scale-out

#### POST
##### Summary

scale out a cluster

##### Description

scale out a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |
| scaleOutReq | body | scale out request | Yes | [cluster.ScaleOutClusterReq](#clusterscaleoutclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/stop

#### POST
##### Summary

stop a cluster

##### Description

stop a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/strategy

#### GET
##### Summary

show the backup strategy of a cluster

##### Description

show the backup strategy of a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | clusterId | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### PUT
##### Summary

save the backup strategy of a cluster

##### Description

save the backup strategy of a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | clusterId | Yes | string |
| updateReq | body | backup strategy request | Yes | [cluster.SaveBackupStrategyReq](#clustersavebackupstrategyreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/upgrade

#### POST
##### Summary

request for upgrade TiDB cluster

##### Description

request for upgrade TiDB cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | clusterId | Yes | string |
| upgradeReq | body | upgrade request | Yes | [cluster.UpgradeClusterReq](#clusterupgradeclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/upgrade/diff

#### GET
##### Summary

query config diff between current cluster and target upgrade version

##### Description

query config diff between current cluster and target upgrade version

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | clusterId | Yes | string |
| targetVersion | query |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/{clusterId}/upgrade/path

#### GET
##### Summary

query upgrade path for given cluster id

##### Description

query upgrade path for given cluster id

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | clusterId | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/clone

#### POST
##### Summary

clone a cluster

##### Description

clone a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| cloneClusterReq | body | clone cluster request | Yes | [cluster.CloneClusterReq](#clustercloneclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/export

#### POST
##### Summary

export data from tidb cluster

##### Description

export

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| dataExport | body | cluster info for data export | Yes | [message.DataExportReq](#messagedataexportreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/import

#### POST
##### Summary

import data to tidb cluster

##### Description

import

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| dataImport | body | cluster info for import data | Yes | [message.DataImportReq](#messagedataimportreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/preview

#### POST
##### Summary

preview cluster topology and capability

##### Description

preview cluster topology and capability

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| createReq | body | preview request | Yes | [cluster.CreateClusterReq](#clustercreateclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/restore

#### POST
##### Summary

restore a new cluster by backup record

##### Description

restore a new cluster by backup record

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | restore request | Yes | [cluster.RestoreNewClusterReq](#clusterrestorenewclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/switchover

#### POST
##### Summary

master/slave switchover

##### Description

master/slave switchover

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| masterSlaveClusterSwitchoverReq | body | switchover request | Yes | [cluster.MasterSlaveClusterSwitchoverReq](#clustermasterslaveclusterswitchoverreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/takeover

#### POST
##### Summary

takeover a cluster

##### Description

takeover a cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| takeoverReq | body | takeover request | Yes | [cluster.TakeoverClusterReq](#clustertakeoverclusterreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/transport

#### GET
##### Summary

query records of import and export

##### Description

query records of import and export

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | query |  | No | string |
| endTime | query |  | No | integer |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |
| reImport | query |  | No | boolean |
| recordId | query |  | No | string |
| startTime | query |  | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /clusters/transport/{recordId}

#### DELETE
##### Summary

delete data transport record

##### Description

delete data transport record

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| recordId | path | data transport recordId | Yes | integer |
| DataTransportDeleteReq | body | data transport record delete request | Yes | [message.DeleteImportExportRecordReq](#messagedeleteimportexportrecordreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /config/

#### GET
##### Summary

get system config

##### Description

get system config

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| configKey | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /config/update

#### POST
##### Summary

update system config

##### Description

UpdateSystemConfig

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| systemConfig | body | system config for update | Yes | [message.UpdateSystemConfigReq](#messageupdatesystemconfigreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /downstream/

#### DELETE
##### Summary

unused, just display downstream config

##### Description

show display config

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| tidb | body | tidb | Yes | [cluster.TiDBDownstream](#clustertidbdownstream) |
| mysql | body | mysql | Yes | [cluster.MysqlDownstream](#clustermysqldownstream) |
| kafka | body | kafka | Yes | [cluster.KafkaDownstream](#clusterkafkadownstream) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /metadata/{clusterId}/

#### DELETE
##### Summary

delete cluster metadata in this system physically, but keep the real cluster alive

##### Description

for handling exceptions only

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |
| deleteReq | body | delete request | No | [cluster.DeleteMetadataPhysicallyReq](#clusterdeletemetadataphysicallyreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /param-groups/

#### GET
##### Summary

query parameter group

##### Description

query parameter group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterSpec | query |  | No | string |
| clusterVersion | query |  | No | string |
| dbType | query |  | No | integer |
| hasDefault | query |  | No | integer |
| hasDetail | query |  | No | boolean |
| name | query |  | No | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

create a parameter group

##### Description

create a parameter group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| createReq | body | create request | Yes | [message.CreateParameterGroupReq](#messagecreateparametergroupreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /param-groups/{paramGroupId}

#### DELETE
##### Summary

delete a parameter group

##### Description

delete a parameter group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| paramGroupId | path | parameter group id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### GET
##### Summary

show details of a parameter group

##### Description

show details of a parameter group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| paramGroupId | path | parameter group id | Yes | string |
| instanceType | query |  | No | string |
| paramName | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### PUT
##### Summary

update a parameter group

##### Description

update a parameter group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| paramGroupId | path | parameter group id | Yes | string |
| updateReq | body | update parameter group request | Yes | [message.UpdateParameterGroupReq](#messageupdateparametergroupreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /param-groups/{paramGroupId}/apply

#### POST
##### Summary

apply a parameter group

##### Description

apply a parameter group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| paramGroupId | path | parameter group id | Yes | string |
| applyReq | body | apply parameter group request | Yes | [message.ApplyParameterGroupReq](#messageapplyparametergroupreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /param-groups/{paramGroupId}/copy

#### POST
##### Summary

copy a parameter group

##### Description

copy a parameter group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| paramGroupId | path | parameter group id | Yes | string |
| copyReq | body | copy parameter group request | Yes | [message.CopyParameterGroupReq](#messagecopyparametergroupreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /platform/check

#### POST
##### Summary

platform check

##### Description

platform check

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| checkPlatformReq | body | check platform | No | [message.CheckPlatformReq](#messagecheckplatformreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /platform/check/{clusterId}

#### POST
##### Summary

platform check cluster

##### Description

platform check cluster

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /platform/log

#### GET
##### Summary

query platform log

##### Description

query platform log

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| endTime | query |  | No | integer |
| level | query |  | No | string |
| message | query |  | No | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |
| startTime | query |  | No | integer |
| traceId | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /platform/report/{checkId}

#### GET
##### Summary

get check report based on checkId

##### Description

get check report based on checkId

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| checkId | path | check id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /platform/reports

#### GET
##### Summary

query check reports

##### Description

get query check reports

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /products/

#### GET
##### Summary

query products

##### Description

query products

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| productIDs | query | product id collection | No | [ string ] |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

update products

##### Description

update products

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| UpdateProductsInfoReq | body | update products info request parameter | Yes | [message.UpdateProductsInfoReq](#messageupdateproductsinforeq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /products/available

#### GET
##### Summary

queries all products' information

##### Description

queries all products' information

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| internalProduct | query |  | No | integer |
| status | query |  | No | string |
| vendorId | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /products/detail

#### GET
##### Summary

query all product detail

##### Description

query all product detail

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| internalProduct | query |  | No | integer |
| productId | query |  | No | string |
| regionId | query |  | No | string |
| status | query |  | No | string |
| vendorId | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /rbac/permission/{userId}

#### GET
##### Summary

query permissions of user

##### Description

QueryPermissionsForUser

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path | rbac userId | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /rbac/permission/add

#### POST
##### Summary

add permissions for role

##### Description

AddPermissionsForRole

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| addPermissionsForRoleReq | body | AddPermissionsForRole request | Yes | [message.AddPermissionsForRoleReq](#messageaddpermissionsforrolereq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /rbac/permission/check

#### POST
##### Summary

check permissions of user

##### Description

CheckPermissionForUser

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| checkPermissionForUserReq | body | CheckPermissionForUser request | Yes | [message.CheckPermissionForUserReq](#messagecheckpermissionforuserreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /rbac/permission/delete

#### DELETE
##### Summary

delete permissions for role

##### Description

DeletePermissionsForRole

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| deletePermissionsForRoleReq | body | DeleteRoleForUser request | Yes | [message.DeletePermissionsForRoleReq](#messagedeletepermissionsforrolereq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /rbac/role/

#### GET
##### Summary

query rbac roles

##### Description

QueryRoles

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| role | path | rbac role | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

create rbac role

##### Description

CreateRbacRole

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| createReq | body | CreateRole request | Yes | [message.CreateRoleReq](#messagecreaterolereq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /rbac/role/{role}

#### DELETE
##### Summary

delete rbac role

##### Description

DeleteRbacRole

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| role | path | rbac role | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /rbac/role/bind

#### POST
##### Summary

bind user with roles

##### Description

BindRolesForUser

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| bindReq | body | BindRolesForUser request | Yes | [message.BindRolesForUserReq](#messagebindrolesforuserreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /rbac/user_role/delete

#### DELETE
##### Summary

unbind rbac role from user

##### Description

UnbindRoleForUser

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| deleteRoleForUserReq | body | UnbindRoleForUser request | Yes | [message.UnbindRoleForUserReq](#messageunbindroleforuserreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /resources/disk

#### PUT
##### Summary

Update disk info

##### Description

update disk information

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| updateReq | body | update disk information | Yes | [message.UpdateDiskReq](#messageupdatediskreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /resources/disks

#### DELETE
##### Summary

Remove a batch of disks

##### Description

remove disks by a list

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| diskIds | body | list of disk IDs | Yes | [message.DeleteDisksReq](#messagedeletedisksreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

Add disks to the specified host

##### Description

add disks to the specified host

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| createDisksReq | body | specify the hostId and disks | Yes | [message.CreateDisksReq](#messagecreatedisksreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /resources/hierarchy

#### GET
##### Summary

Show the resources hierarchy

##### Description

get resource hierarchy-tree

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| Depth | query |  | No | integer |
| Level | query | [1:Region, 2:Zone, 3:Rack, 4:Host] | No | integer |
| arch | query |  | No | string |
| clusterType | query |  | No | string |
| hostDiskType | query |  | No | string |
| hostId | query |  | No | string |
| hostName | query |  | No | string |
| loadStat | query |  | No | string |
| purpose | query |  | No | string |
| status | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /resources/host

#### PUT
##### Summary

Update host info

##### Description

update host information

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| updateReq | body | update host information | Yes | [message.UpdateHostInfoReq](#messageupdatehostinforeq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /resources/host-reserved

#### PUT
##### Summary

Update host reserved

##### Description

update host reserved by a list

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| updateReq | body | do update in host list | Yes | [message.UpdateHostReservedReq](#messageupdatehostreservedreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /resources/host-status

#### PUT
##### Summary

Update host status

##### Description

update host status by a list

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| updateReq | body | do update in host list | Yes | [message.UpdateHostStatusReq](#messageupdatehoststatusreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /resources/hosts

#### DELETE
##### Summary

Remove a batch of hosts

##### Description

remove hosts by a list

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| hostIds | body | list of host IDs | Yes | [message.DeleteHostsReq](#messagedeletehostsreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### GET
##### Summary

Show all hosts list in EM

##### Description

get hosts list

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| arch | query |  | No | string |
| clusterType | query |  | No | string |
| hostDiskType | query |  | No | string |
| hostId | query |  | No | string |
| hostIp | query |  | No | string |
| hostName | query |  | No | string |
| loadStat | query |  | No | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |
| purpose | query |  | No | string |
| rack | query |  | No | string |
| region | query |  | No | string |
| status | query |  | No | string |
| zone | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

Import a batch of hosts to EM

##### Description

import hosts by xlsx file

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| hostReserved | formData | whether hosts are reserved(won't be allocated) after import | No | string |
| skipHostInit | formData | whether to skip host init steps | No | string |
| ignorewarns | formData | whether to ignore warings in init steps | No | string |
| file | formData | hosts information in a xlsx file | Yes | file |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /resources/hosts-template

#### GET
##### Summary

Download the host information template file for importing

##### Description

get host template xlsx file

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | file |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /resources/stocks

#### GET
##### Summary

Show the resources stocks

##### Description

get resource stocks in specified conditions

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| arch | query |  | No | string |
| capacity | query |  | No | integer |
| clusterType | query |  | No | string |
| diskStatus | query |  | No | string |
| diskType | query |  | No | string |
| hostDiskType | query |  | No | string |
| hostId | query |  | No | string |
| hostIp | query |  | No | string |
| hostName | query |  | No | string |
| loadStat | query |  | No | string |
| purpose | query |  | No | string |
| rack | query |  | No | string |
| region | query |  | No | string |
| status | query |  | No | string |
| zone | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /system/info

#### GET
##### Summary

get system info

##### Description

get system info

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| withVersionDetail | query |  | No | boolean |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

### /tenant

#### GET
##### Summary

queries all tenant profile

##### Description

query all tenant profile

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| QueryTenantReq | body | query tenant profile request parameter | Yes | [message.QueryTenantReq](#messagequerytenantreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

update tenant onboarding status

##### Description

update tenant onboarding status

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| tenantId | path | tenant id | Yes | string |
| UpdateTenantOnBoardingStatusReq | body | query tenant profile request parameter | Yes | [message.UpdateTenantOnBoardingStatusReq](#messageupdatetenantonboardingstatusreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /tenant/{tenantId}

#### GET
##### Summary

get tenant profile

##### Description

get tenant profile

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| tenantId | path | tenant id | Yes | string |
| GetTenantReq | body | get tenant profile request parameter | Yes | [message.GetTenantReq](#messagegettenantreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /tenants/

#### POST
##### Summary

created  tenant

##### Description

created tenant

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| CreateTenantReq | body | create tenant request parameter | Yes | [message.CreateTenantReq](#messagecreatetenantreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /tenants/{tenantId}

#### DELETE
##### Summary

delete tenant

##### Description

delete tenant

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| tenantId | path | tenant id | Yes | string |
| DeleteTenantReq | body | delete tenant request parameter | Yes | [message.DeleteTenantReq](#messagedeletetenantreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /tenants/{tenantId}/update_profile

#### POST
##### Summary

update tenant profile

##### Description

update tenant profile

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| tenantId | path | tenant id | Yes | string |
| UpdateTenantProfileReq | body | query tenant profile request parameter | Yes | [message.UpdateTenantProfileReq](#messageupdatetenantprofilereq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /user/login

#### POST
##### Summary

login

##### Description

login

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| loginInfo | body | login info | Yes | [message.LoginReq](#messageloginreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

### /user/logout

#### POST
##### Summary

logout

##### Description

logout

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /users/

#### GET
##### Summary

queries all user profile

##### Description

query all user profile

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| queryUserRequest | body | query user profile request parameter | Yes | [message.QueryUserReq](#messagequeryuserreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

created  user

##### Description

created user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| createUserReq | body | create user request parameter | Yes | [message.CreateUserReq](#messagecreateuserreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /users/{userId}

#### DELETE
##### Summary

delete user

##### Description

delete user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path | user id | Yes | string |
| deleteUserReq | body | delete user request parameter | Yes | [message.DeleteUserReq](#messagedeleteuserreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### GET
##### Summary

get user profile

##### Description

get user profile

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path | user id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /users/{userId}/password

#### POST
##### Summary

update user password

##### Description

update user password

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path | user id | Yes | string |
| UpdateUserPasswordRequest | body | query user password request parameter | Yes | [message.UpdateUserPasswordReq](#messageupdateuserpasswordreq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /users/{userId}/update_profile

#### POST
##### Summary

update user profile

##### Description

update user profile

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path | user id | Yes | string |
| updateUserProfileRequest | body | query user profile request parameter | Yes | [message.UpdateUserProfileReq](#messageupdateuserprofilereq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /vendors/

#### GET
##### Summary

query vendors

##### Description

query vendors

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| vendorIDs | query | vendor id collection | No | [ string ] |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

update vendors

##### Description

update vendors

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| UpdateVendorInfoReq | body | update vendor info request parameter | Yes | [message.UpdateVendorInfoReq](#messageupdatevendorinforeq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /vendors/available

#### GET
##### Summary

query available vendors and regions

##### Description

query available vendors and regions

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /workflow/

#### GET
##### Summary

query flow works

##### Description

query flow works

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| bizId | query |  | No | string |
| bizType | query |  | No | string |
| flowName | query |  | No | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |
| status | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /workflow/{workFlowId}

#### GET
##### Summary

show details of a flow work

##### Description

show details of a flow work

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| workFlowId | path | flow work id | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### Models

#### cluster.BackupClusterDataReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| backupMode | string | auto,manual | No |
| backupType | string | full,incr | No |
| clusterId | string |  | No |

#### cluster.BackupClusterDataResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| backupId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.CancelBackupReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| backupId | string |  | No |
| clusterId | string |  | No |

#### cluster.CancelBackupResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| cluster.CancelBackupResp | object |  |  |

#### cluster.CloneClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| cloneStrategy | string | specify clone strategy, include empty, snapshot and sync, default empty(option) | Yes |
| clusterName | string |  | Yes |
| clusterType | string |  | Yes |
| clusterVersion | string |  | Yes |
| copies | integer | The number of copies of the newly created cluster data, consistent with the number of copies set in PD | No |
| cpuArchitecture | string | X86/X86_64/ARM | Yes |
| dbPassword | string |  | Yes |
| dbUser | string | todo delete? | No |
| exclusive | boolean | Whether the newly created cluster is exclusive to physical resources, when exclusive, a host will only deploy instances of the same cluster, which may result in poor resource utilization | No |
| parameterGroupID | string |  | No |
| region | string | The Region where the cluster is located | Yes |
| resourceParameters | [structs.ClusterResourceInfo](#structsclusterresourceinfo) |  | No |
| sourceClusterId | string | specify source cluster id(require) | Yes |
| tags | [ string ] |  | No |
| tls | boolean |  | No |
| vendor | string |  | No |

#### cluster.CloneClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.CreateChangeFeedTaskReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string | _Example:_ `"CLUSTER_ID_IN_TIEM__22"` | Yes |
| downstream | object |  | No |
| downstreamType | string | _Enum:_ `"tidb"`, `"kafka"`, `"mysql"`<br>_Example:_ `"tidb"` | Yes |
| name | string | _Example:_ `"my_sync_name"` | Yes |
| rules | [ string ] | _Example:_ `["*.*"]` | No |
| startTS | string | _Example:_ `"415241823337054209"` | No |

#### cluster.CreateChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | _Example:_ `"TASK_ID_IN_TIEM____22"` | No |

#### cluster.CreateClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterName | string |  | Yes |
| clusterType | string |  | Yes |
| clusterVersion | string |  | Yes |
| copies | integer | The number of copies of the newly created cluster data, consistent with the number of copies set in PD | No |
| cpuArchitecture | string | X86/X86_64/ARM | Yes |
| dbPassword | string |  | Yes |
| dbUser | string | todo delete? | No |
| exclusive | boolean | Whether the newly created cluster is exclusive to physical resources, when exclusive, a host will only deploy instances of the same cluster, which may result in poor resource utilization | No |
| parameterGroupID | string |  | No |
| region | string | The Region where the cluster is located | Yes |
| resourceParameters | [structs.ClusterResourceInfo](#structsclusterresourceinfo) |  | No |
| tags | [ string ] |  | No |
| tls | boolean |  | No |
| vendor | string |  | No |

#### cluster.CreateClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.DeleteBackupDataReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| backupMode | string |  | No |
| clusterId | string |  | No |

#### cluster.DeleteBackupDataResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| cluster.DeleteBackupDataResp | object |  |  |

#### cluster.DeleteChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | _Example:_ `"TASK_ID_IN_TIEM____22"` | No |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |

#### cluster.DeleteClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| autoBackup | boolean |  | No |
| force | boolean |  | No |
| keepHistoryBackupRecords | boolean |  | No |

#### cluster.DeleteClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterID | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.DeleteMetadataPhysicallyReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| reason | string |  | Yes |

#### cluster.DeleteMetadataPhysicallyResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| cluster.DeleteMetadataPhysicallyResp | object |  |  |

#### cluster.DetailChangeFeedTaskResp

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

#### cluster.Dispatcher

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| dispatcher | string | _Example:_ `"ts"` | No |
| matcher | string | _Example:_ `"test1.*"` | No |

#### cluster.GetBackupStrategyResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| strategy | [structs.BackupStrategy](#structsbackupstrategy) |  | No |

#### cluster.GetDashboardInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string | _Example:_ `"abc"` | No |
| token | string |  | No |
| url | string | _Example:_ `"http://127.0.0.1:9093"` | No |

#### cluster.InspectParameterInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| category | string | _Example:_ `"log"` | No |
| description | string | _Example:_ `"binlog cache size"` | No |
| hasReboot | integer | _Enum:_ `0`, `1`<br>_Example:_ `0` | No |
| inspectValue | object |  | No |
| name | string | _Example:_ `"binlog_cache"` | No |
| note | string | _Example:_ `"binlog cache size"` | No |
| paramId | string | _Example:_ `"1"` | No |
| range | [ string ] | _Example:_ `["1"," 1000"]` | No |
| rangeType | integer | _Enum:_ `0`, `1`, `2`<br>_Example:_ `1` | No |
| readOnly | integer | _Enum:_ `0`, `1`<br>_Example:_ `0` | No |
| realValue | [structs.ParameterRealValue](#structsparameterrealvalue) |  | No |
| systemVariable | string | _Example:_ `"log.log_level"` | No |
| type | integer | _Enum:_ `0`, `1`, `2`, `3`, `4`<br>_Example:_ `0` | No |
| unit | string | _Example:_ `"MB"` | No |
| unitOptions | [ string ] | _Example:_ `["KB","MB","GB"]` | No |

#### cluster.InspectParameters

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| instanceId | string |  | No |
| instanceType | string |  | No |
| parameterInfos | [ [cluster.InspectParameterInfo](#clusterinspectparameterinfo) ] |  | No |

#### cluster.InspectParametersReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| instanceId | string |  | No |

#### cluster.InspectParametersResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| params | [ [cluster.InspectParameters](#clusterinspectparameters) ] |  | No |

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

#### cluster.MasterSlaveClusterSwitchoverReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| checkMasterWritableFlag | boolean |  | No |
| checkSlaveReadOnlyFlag | boolean |  | No |
| checkStandaloneClusterFlag | boolean | check if cluster specified in `SourceClusterID` is standalone, i.e. no cluster relation and no cdc if this flag is true, always only check | No |
| force | boolean |  | No |
| onlyCheck | boolean | only check if this flag is true | No |
| sourceClusterID | string | old master/new slave | Yes |
| targetClusterID | string | new master/old slave | Yes |

#### cluster.MasterSlaveClusterSwitchoverResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.MysqlDownstream

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| concurrentThreads | integer | _Example:_ `7` | No |
| ip | string | _Example:_ `"127.0.0.1"` | No |
| maxTxnRow | integer | _Example:_ `5` | No |
| password | string | _Example:_ `"my_password"` | No |
| port | integer | _Example:_ `8001` | No |
| tls | boolean | _Example:_ `false` | No |
| username | string | _Example:_ `"root"` | No |
| workerCount | integer | _Example:_ `2` | No |

#### cluster.PauseChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |

#### cluster.PreviewClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| capabilityIndexes | [ [structs.Index](#structsindex) ] |  | No |
| clusterName | string |  | No |
| clusterType | string |  | No |
| clusterVersion | string |  | No |
| cpuArchitecture | string |  | No |
| region | string |  | No |
| stockCheckResult | [ [structs.ResourceStockCheckResult](#structsresourcestockcheckresult) ] |  | No |

#### cluster.QueryBackupRecordsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| backupRecords | [ [structs.BackupRecord](#structsbackuprecord) ] |  | No |

#### cluster.QueryChangeFeedTaskResp

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

#### cluster.QueryClusterDetailResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| info | [structs.ClusterInfo](#structsclusterinfo) |  | No |
| instanceResource | [ [structs.ClusterResourceParameterCompute](#structsclusterresourceparametercompute) ] |  | No |
| requestResourceMode | string | _Enum:_ `"SpecificZone"`, `"SpecificHost"` | No |
| topology | [ [structs.ClusterInstanceInfo](#structsclusterinstanceinfo) ] |  | No |

#### cluster.QueryClusterLogResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| results | [ [structs.ClusterLogItem](#structsclusterlogitem) ] |  | No |
| took | integer | _Example:_ `10` | No |

#### cluster.QueryClusterParametersResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| paramGroupId | string |  | No |
| params | [ [structs.ClusterParameterInfo](#structsclusterparameterinfo) ] |  | No |

#### cluster.QueryClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusters | [ [structs.ClusterInfo](#structsclusterinfo) ] |  | No |

#### cluster.QueryMonitorInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| alertUrl | string | _Example:_ `"http://127.0.0.1:9093"` | No |
| clusterId | string | _Example:_ `"abc"` | No |
| grafanaUrl | string | _Example:_ `"http://127.0.0.1:3000"` | No |

#### cluster.QueryUpgradePathRsp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| paths | [ [structs.ProductUpgradePathItem](#structsproductupgradepathitem) ] |  | No |

#### cluster.QueryUpgradeVersionDiffInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| configDiffInfos | [ [structs.ProductUpgradeVersionConfigDiffItem](#structsproductupgradeversionconfigdiffitem) ] |  | No |

#### cluster.RestartClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.RestoreNewClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| backupId | string |  | Yes |
| clusterName | string |  | Yes |
| clusterType | string |  | Yes |
| clusterVersion | string |  | Yes |
| copies | integer | The number of copies of the newly created cluster data, consistent with the number of copies set in PD | No |
| cpuArchitecture | string | X86/X86_64/ARM | Yes |
| dbPassword | string |  | Yes |
| dbUser | string | todo delete? | No |
| exclusive | boolean | Whether the newly created cluster is exclusive to physical resources, when exclusive, a host will only deploy instances of the same cluster, which may result in poor resource utilization | No |
| parameterGroupID | string |  | No |
| region | string | The Region where the cluster is located | Yes |
| resourceParameters | [structs.ClusterResourceInfo](#structsclusterresourceinfo) |  | No |
| tags | [ string ] |  | No |
| tls | boolean |  | No |
| vendor | string |  | No |

#### cluster.RestoreNewClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterID | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.ResumeChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |

#### cluster.SaveBackupStrategyReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| strategy | [structs.BackupStrategy](#structsbackupstrategy) |  | No |

#### cluster.SaveBackupStrategyResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| cluster.SaveBackupStrategyResp | object |  |  |

#### cluster.ScaleInClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| instanceId | string |  | No |

#### cluster.ScaleInClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.ScaleOutClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| instanceResource | [ [structs.ClusterResourceParameterCompute](#structsclusterresourceparametercompute) ] |  | No |
| requestResourceMode | string | _Enum:_ `"SpecificZone"`, `"SpecificHost"` | No |

#### cluster.ScaleOutClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.StopClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.TakeoverClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| TiUPIp | string | _Example:_ `"172.16.4.147"` | Yes |
| TiUPPath | string | _Example:_ `".tiup/"` | Yes |
| TiUPPort | integer | _Example:_ `22` | Yes |
| TiUPUserName | string | _Example:_ `"root"` | Yes |
| TiUPUserPassword | string | _Example:_ `"password"` | Yes |
| clusterName | string | _Example:_ `"myClusterName"` | Yes |
| dbPassword | string | _Example:_ `"myPassword"` | Yes |

#### cluster.TakeoverClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

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

#### cluster.UpdateChangeFeedTaskReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| downstream | object |  | No |
| downstreamType | string | _Enum:_ `"tidb"`, `"kafka"`, `"mysql"`<br>_Example:_ `"tidb"` | No |
| name | string | _Example:_ `"my_sync_name"` | Yes |
| rules | [ string ] | _Example:_ `["*.*"]` | No |

#### cluster.UpdateChangeFeedTaskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | string | _Enum:_ `"Initial"`, `"Normal"`, `"Stopped"`, `"Finished"`, `"Error"`, `"Failed"`<br>_Example:_ `"Normal"` | No |

#### cluster.UpdateClusterParametersReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| params | [ [structs.ClusterParameterSampleInfo](#structsclusterparametersampleinfo) ] |  | Yes |
| reboot | boolean |  | No |

#### cluster.UpdateClusterParametersResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string | _Example:_ `"1"` | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### cluster.UpgradeClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| configs | [ [structs.ClusterUpgradeVersionConfigItem](#structsclusterupgradeversionconfigitem) ] |  | No |
| targetVersion | string | _Example:_ `"v5.0.0"` | Yes |
| upgradeType | string | _Enum:_ `"in-place"`, `"migration"` | Yes |
| upgradeWay | string | _Enum:_ `"offline"`, `"online"` | No |

#### cluster.UpgradeClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### controller.CommonResult

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| data | object |  | No |
| message | string |  | No |

#### controller.Page

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| page | integer |  | No |
| pageSize | integer |  | No |
| total | integer |  | No |

#### controller.ResultWithPage

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| data | object |  | No |
| message | string |  | No |
| page | [controller.Page](#controllerpage) |  | No |

#### message.AddPermissionsForRoleReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permissions | [ [structs.RbacPermission](#structsrbacpermission) ] |  | No |
| role | string |  | No |

#### message.AddPermissionsForRoleResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.AddPermissionsForRoleResp | object |  |  |

#### message.ApplyParameterGroupReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string | _Example:_ `"123"` | Yes |
| reboot | boolean |  | No |

#### message.ApplyParameterGroupResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string | _Example:_ `"123"` | No |
| paramGroupId | string | _Example:_ `"123"` | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### message.BindRolesForUserReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| roles | [ string ] |  | No |
| userId | string |  | No |

#### message.BindRolesForUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.BindRolesForUserResp | object |  |  |

#### message.CheckClusterRsp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| checkId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### message.CheckPermissionForUserReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permissions | [ [structs.RbacPermission](#structsrbacpermission) ] |  | No |
| userId | string |  | No |

#### message.CheckPermissionForUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| result | boolean |  | No |

#### message.CheckPlatformReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| displayMode | string |  | No |

#### message.CheckPlatformRsp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| checkId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### message.CopyParameterGroupReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string | _Example:_ `"8C16GV4_copy"` | Yes |
| note | string | _Example:_ `"copy param group"` | No |

#### message.CopyParameterGroupResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| paramGroupId | string | _Example:_ `"1"` | No |

#### message.CreateDisksReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| disks | [ [structs.DiskInfo](#structsdiskinfo) ] |  | No |
| hostId | string |  | No |

#### message.CreateDisksResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| diskIds | [ string ] |  | No |

#### message.CreateParameterGroupReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| addParams | [ [message.ParameterInfo](#messageparameterinfo) ] |  | No |
| clusterSpec | string | _Example:_ `"8C16G"` | Yes |
| clusterVersion | string | _Example:_ `"v5.0"` | Yes |
| dbType | integer | _Enum:_ `1`, `2`<br>_Example:_ `1` | Yes |
| groupType | integer | _Enum:_ `1`, `2`<br>_Example:_ `1` | Yes |
| name | string | _Example:_ `"8C16GV4_default"` | Yes |
| note | string | _Example:_ `"default param group"` | No |
| params | [ [structs.ParameterGroupParameterSampleInfo](#structsparametergroupparametersampleinfo) ] |  | Yes |

#### message.CreateParameterGroupResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| paramGroupId | string | _Example:_ `"1"` | No |

#### message.CreateRoleReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | string |  | No |

#### message.CreateRoleResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.CreateRoleResp | object |  |  |

#### message.CreateTenantReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | Yes |
| maxCluster | integer |  | No |
| maxCpu | integer |  | No |
| maxMemory | integer |  | No |
| maxStorage | integer |  | No |
| name | string |  | No |
| onBoardingStatus | string |  | No |
| status | string |  | No |

#### message.CreateTenantResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.CreateTenantResp | object |  |  |

#### message.CreateUserReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| email | string |  | Yes |
| name | string |  | Yes |
| nickname | string |  | No |
| password | string |  | Yes |
| phone | string |  | No |
| tenantId | string |  | No |

#### message.CreateUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.CreateUserResp | object |  |  |

#### message.DataExportReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| accessKey | string |  | No |
| bucketUrl | string |  | No |
| clusterId | string |  | No |
| comment | string |  | No |
| endpointUrl | string |  | No |
| fileType | string |  | No |
| filter | string |  | No |
| password | string |  | No |
| secretAccessKey | string |  | No |
| sql | string |  | No |
| storageType | string |  | No |
| userName | string |  | No |
| zipName | string |  | No |

#### message.DataExportResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| recordId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### message.DataImportReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| accessKey | string |  | No |
| bucketUrl | string |  | No |
| clusterId | string |  | No |
| comment | string |  | No |
| endpointUrl | string |  | No |
| password | string |  | No |
| recordId | string |  | No |
| secretAccessKey | string |  | No |
| storageType | string |  | No |
| userName | string |  | No |

#### message.DataImportResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| recordId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### message.DeleteDisksReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| diskIds | [ string ] |  | No |

#### message.DeleteDisksResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.DeleteDisksResp | object |  |  |

#### message.DeleteHostsReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| force | boolean |  | No |
| hostIds | [ string ] |  | No |

#### message.DeleteHostsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| flowInfo | [ [structs.AsyncTaskWorkFlowInfo](#structsasynctaskworkflowinfo) ] |  | No |

#### message.DeleteImportExportRecordReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.DeleteImportExportRecordReq | object |  |  |

#### message.DeleteImportExportRecordResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| recordId | string |  | No |

#### message.DeleteParameterGroupResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| paramGroupId | string | _Example:_ `"1"` | No |

#### message.DeletePermissionsForRoleReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permissions | [ [structs.RbacPermission](#structsrbacpermission) ] |  | No |
| role | string |  | No |

#### message.DeletePermissionsForRoleResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.DeletePermissionsForRoleResp | object |  |  |

#### message.DeleteRoleResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.DeleteRoleResp | object |  |  |

#### message.DeleteTenantReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |

#### message.DeleteTenantResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.DeleteTenantResp | object |  |  |

#### message.DeleteUserReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.DeleteUserReq | object |  |  |

#### message.DeleteUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.DeleteUserResp | object |  |  |

#### message.DetailParameterGroupResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterSpec | string | _Example:_ `"8C16G"` | No |
| clusterVersion | string | _Example:_ `"v5.0"` | No |
| createTime | integer | _Example:_ `1636698675` | No |
| dbType | integer | _Enum:_ `1`, `2`<br>_Example:_ `1` | No |
| groupType | integer | _Enum:_ `1`, `2`<br>_Example:_ `0` | No |
| hasDefault | integer | _Enum:_ `1`, `2`<br>_Example:_ `1` | No |
| name | string | _Example:_ `"default"` | No |
| note | string | _Example:_ `"default param group"` | No |
| paramGroupId | string | _Example:_ `"1"` | No |
| params | [ [structs.ParameterGroupParameterInfo](#structsparametergroupparameterinfo) ] |  | No |
| updateTime | integer | _Example:_ `1636698675` | No |

#### message.GetCheckReportRsp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| reportInfo | object |  | No |

#### message.GetHierarchyResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| root | [structs.HierarchyTreeNode](#structshierarchytreenode) |  | No |

#### message.GetStocksResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| stocks | object | map[zone] -> stocks | No |

#### message.GetSystemConfigResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| configKey | string |  | No |
| configValue | string |  | No |

#### message.GetSystemInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| currentVersion | [structs.SystemVersionInfo](#structssystemversioninfo) |  | No |
| info | [structs.SystemInfo](#structssysteminfo) |  | No |
| lastVersion | [structs.SystemVersionInfo](#structssystemversioninfo) |  | No |

#### message.GetTenantReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |

#### message.GetTenantResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| info | [structs.TenantInfo](#structstenantinfo) |  | No |

#### message.GetUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [structs.UserInfo](#structsuserinfo) |  | No |

#### message.ImportHostsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| flowInfo | [ [structs.AsyncTaskWorkFlowInfo](#structsasynctaskworkflowinfo) ] |  | No |
| hostIds | [ string ] |  | No |

#### message.LoginReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userName | string |  | Yes |
| userPassword | string |  | Yes |

#### message.LoginResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| passwordExpired | boolean |  | No |
| tenantId | string |  | No |
| token | string |  | No |
| userId | string |  | No |

#### message.LogoutResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userId | string |  | No |

#### message.ParameterInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| category | string | _Example:_ `"log"` | No |
| defaultValue | string | _Example:_ `"1024"` | No |
| description | string | _Example:_ `"binlog size"` | No |
| hasApply | integer | _Enum:_ `0`, `1`<br>_Example:_ `1` | No |
| hasReboot | integer | _Enum:_ `0`, `1`<br>_Example:_ `0` | No |
| instanceType | string | _Example:_ `"TiDB"` | No |
| name | string | _Example:_ `"binlog_size"` | No |
| note | string |  | No |
| range | [ string ] |  | No |
| rangeType | integer | _Enum:_ `0`, `1`, `2`<br>_Example:_ `1` | No |
| readOnly | integer | _Enum:_ `0`, `1`<br>_Example:_ `0` | No |
| systemVariable | string | _Example:_ `"log.binlog_size"` | No |
| type | integer | _Enum:_ `0`, `1`, `2`, `3`, `4`<br>_Example:_ `0` | No |
| unit | string | _Example:_ `"MB"` | No |
| unitOptions | [ string ] | _Example:_ `["KB","MB","GB"]` | No |
| updateSource | integer | _Enum:_ `0`, `1`, `2`, `3`, `4`<br>_Example:_ `0` | No |

#### message.PlatformLogItem

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | _Example:_ `"zvadfwf"` | No |
| index | string | _Example:_ `"em-system-logs-2021.12.30"` | No |
| level | string | _Example:_ `"warn"` | No |
| message | string | _Example:_ `"some do something"` | No |
| microMethod | string | _Example:_ `"em.cluster.ClusterService.GetSystemInfo"` | No |
| timestamp | string | _Example:_ `"2021-09-23 14:23:10"` | No |
| traceId | string | _Example:_ `"UNe7K1uERa-2fwSxGJ6CFQ"` | No |

#### message.QueryAvailableProductsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | object | arch version | No |

#### message.QueryAvailableVendorsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| vendors | object |  | No |

#### message.QueryCheckReportsRsp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| reportMetas | object |  | No |

#### message.QueryDataImportExportRecordsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| transportRecords | [ [structs.DataImportExportRecordInfo](#structsdataimportexportrecordinfo) ] |  | No |

#### message.QueryHostsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| hosts | [ [structs.HostInfo](#structshostinfo) ] |  | No |

#### message.QueryParameterGroupResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterSpec | string | _Example:_ `"8C16G"` | No |
| clusterVersion | string | _Example:_ `"v5.0"` | No |
| createTime | integer | _Example:_ `1636698675` | No |
| dbType | integer | _Enum:_ `1`, `2`<br>_Example:_ `1` | No |
| groupType | integer | _Enum:_ `1`, `2`<br>_Example:_ `0` | No |
| hasDefault | integer | _Enum:_ `1`, `2`<br>_Example:_ `1` | No |
| name | string | _Example:_ `"default"` | No |
| note | string | _Example:_ `"default param group"` | No |
| paramGroupId | string | _Example:_ `"1"` | No |
| params | [ [structs.ParameterGroupParameterInfo](#structsparametergroupparameterinfo) ] |  | No |
| updateTime | integer | _Example:_ `1636698675` | No |

#### message.QueryPermissionsForUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permissions | [ [structs.RbacPermission](#structsrbacpermission) ] |  | No |
| userId | string |  | No |

#### message.QueryPlatformLogResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| results | [ [message.PlatformLogItem](#messageplatformlogitem) ] |  | No |
| took | integer | _Example:_ `10` | No |

#### message.QueryProductDetailResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | object |  | No |

#### message.QueryProductsInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | [ [structs.ProductConfigInfo](#structsproductconfiginfo) ] |  | No |

#### message.QueryRolesResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| roles | [ string ] |  | No |

#### message.QueryTenantReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| page | integer | Current page location | No |
| pageSize | integer | Number of this request | No |

#### message.QueryTenantResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| tenants | object |  | No |

#### message.QueryUserReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| page | integer | Current page location | No |
| pageSize | integer | Number of this request | No |

#### message.QueryUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | object |  | No |

#### message.QueryVendorInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| vendors | [ [structs.VendorConfigInfo](#structsvendorconfiginfo) ] |  | No |

#### message.QueryWorkFlowDetailResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| info | [structs.WorkFlowInfo](#structsworkflowinfo) |  | No |
| nodeNames | [ string ] |  | No |
| nodes | [ [structs.WorkFlowNodeInfo](#structsworkflownodeinfo) ] |  | No |

#### message.QueryWorkFlowsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| workFlows | [ [structs.WorkFlowInfo](#structsworkflowinfo) ] |  | No |

#### message.UnbindRoleForUserReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | string |  | No |
| userId | string |  | No |

#### message.UnbindRoleForUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UnbindRoleForUserResp | object |  |  |

#### message.UpdateDiskReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| newDiskInfo | [structs.DiskInfo](#structsdiskinfo) |  | No |

#### message.UpdateDiskResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateDiskResp | object |  |  |

#### message.UpdateHostInfoReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| newHostInfo | [structs.HostInfo](#structshostinfo) |  | No |

#### message.UpdateHostInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateHostInfoResp | object |  |  |

#### message.UpdateHostReservedReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| hostIds | [ string ] |  | No |
| reserved | boolean |  | No |

#### message.UpdateHostReservedResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateHostReservedResp | object |  |  |

#### message.UpdateHostStatusReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| hostIds | [ string ] |  | No |
| status | string |  | No |

#### message.UpdateHostStatusResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateHostStatusResp | object |  |  |

#### message.UpdateParameterGroupReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| addParams | [ [message.ParameterInfo](#messageparameterinfo) ] |  | No |
| clusterSpec | string | _Example:_ `"8C16G"` | No |
| clusterVersion | string | _Example:_ `"v5.0"` | No |
| delParams | [ string ] | _Example:_ `["1"]` | No |
| name | string | _Example:_ `"8C16GV4_new"` | No |
| note | string | _Example:_ `"update param group"` | No |
| params | [ [structs.ParameterGroupParameterSampleInfo](#structsparametergroupparametersampleinfo) ] |  | Yes |

#### message.UpdateParameterGroupResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| paramGroupId | string | _Example:_ `"1"` | No |

#### message.UpdateProductsInfoReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | [ [structs.ProductConfigInfo](#structsproductconfiginfo) ] |  | Yes |

#### message.UpdateProductsInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateProductsInfoResp | object |  |  |

#### message.UpdateSystemConfigReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| configKey | string |  | No |
| configValue | string |  | No |

#### message.UpdateSystemConfigResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateSystemConfigResp | object |  |  |

#### message.UpdateTenantOnBoardingStatusReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| onBoardingStatus | string |  | No |

#### message.UpdateTenantOnBoardingStatusResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateTenantOnBoardingStatusResp | object |  |  |

#### message.UpdateTenantProfileReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| maxCluster | integer |  | No |
| maxCpu | integer |  | No |
| maxMemory | integer |  | No |
| maxStorage | integer |  | No |
| name | string |  | No |

#### message.UpdateTenantProfileResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateTenantProfileResp | object |  |  |

#### message.UpdateUserPasswordReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| password | string |  | Yes |

#### message.UpdateUserPasswordResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateUserPasswordResp | object |  |  |

#### message.UpdateUserProfileReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| email | string |  | No |
| nickname | string |  | No |
| phone | string |  | No |

#### message.UpdateUserProfileResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateUserProfileResp | object |  |  |

#### message.UpdateVendorInfoReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| vendors | [ [structs.VendorConfigInfo](#structsvendorconfiginfo) ] |  | No |

#### structs.AsyncTaskWorkFlowInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| workFlowId | string | Asynchronous task workflow ID | No |

#### structs.BackupRecord

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| backupMethod | string |  | No |
| backupMode | string |  | No |
| backupTso | string |  | No |
| backupType | string |  | No |
| clusterId | string |  | No |
| createTime | string |  | No |
| deleteTime | string |  | No |
| endTime | string |  | No |
| filePath | string |  | No |
| id | string |  | No |
| size | number |  | No |
| startTime | string |  | No |
| status | string |  | No |
| updateTime | string |  | No |

#### structs.BackupStrategy

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| backupDate | string |  | No |
| clusterId | string |  | No |
| period | string |  | No |

#### structs.CheckReportMeta

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| checkID | string |  | No |
| createAt | string |  | No |
| creator | string |  | No |
| status | string |  | No |
| type | string |  | No |

#### structs.ClusterInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| alertUrl | string | _Example:_ `"http://127.0.0.1:9093"` | No |
| backupFileUsage | [structs.Usage](#structsusage) |  | No |
| clusterId | string |  | No |
| clusterName | string |  | No |
| clusterType | string |  | No |
| clusterVersion | string |  | No |
| copies | integer | The number of copies of the newly created cluster data, consistent with the number of copies set in PD | No |
| cpuArchitecture | string | X86/X86_64/ARM | No |
| cpuUsage | [structs.Usage](#structsusage) |  | No |
| createTime | string |  | No |
| deleteTime | string |  | No |
| exclusive | boolean | Whether the newly created cluster is exclusive to physical resources, when exclusive, a host will only deploy instances of the same cluster, which may result in poor resource utilization | No |
| extranetConnectAddresses | [ string ] |  | No |
| grafanaUrl | string | _Example:_ `"http://127.0.0.1:3000"` | No |
| intranetConnectAddresses | [ string ] |  | No |
| maintainStatus | string |  | No |
| maintainWindow | string |  | No |
| memoryUsage | [structs.Usage](#structsusage) |  | No |
| region | string |  | No |
| relations | [structs.ClusterRelations](#structsclusterrelations) |  | No |
| role | string |  | No |
| status | string |  | No |
| storageUsage | [structs.Usage](#structsusage) |  | No |
| tags | [ string ] |  | No |
| tls | boolean |  | No |
| updateTime | string |  | No |
| userId | string |  | No |
| vendor | string | DBUser                   string    `json:"dbUser"` //The username and password for the newly created database cluster, default is the root user, which is not valid for Data Migration clusters | No |
| whitelist | [ string ] |  | No |

#### structs.ClusterInstanceInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| addresses | [ string ] |  | No |
| cpuUsage | [structs.Usage](#structsusage) |  | No |
| diskId | string |  | No |
| hostID | string |  | No |
| id | string |  | No |
| ioUtil | number |  | No |
| iops | [ number ] |  | No |
| memoryUsage | [structs.Usage](#structsusage) |  | No |
| ports | [ integer ] |  | No |
| role | string |  | No |
| spec | [structs.ProductSpecInfo](#structsproductspecinfo) | ?? | No |
| status | string |  | No |
| storageUsage | [structs.Usage](#structsusage) |  | No |
| type | string |  | No |
| version | string |  | No |
| zone | [structs.ZoneFullInfo](#structszonefullinfo) | ?? | No |

#### structs.ClusterInstanceParameterValue

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| instanceId | string |  | No |
| value | string |  | No |

#### structs.ClusterLogItem

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string | _Example:_ `"abc"` | No |
| ext | object |  | No |
| id | string | _Example:_ `"zvadfwf"` | No |
| index | string | _Example:_ `"em-tidb-cluster-2021.09.23"` | No |
| ip | string | _Example:_ `"127.0.0.1"` | No |
| level | string | _Example:_ `"warn"` | No |
| message | string | _Example:_ `"tidb log"` | No |
| module | string | _Example:_ `"tidb"` | No |
| sourceLine | string | _Example:_ `"main.go:210"` | No |
| timestamp | string | _Example:_ `"2021-09-23 14:23:10"` | No |

#### structs.ClusterParameterInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| category | string | _Example:_ `"basic"` | No |
| createTime | integer | _Example:_ `1636698675` | No |
| defaultValue | string | _Example:_ `"1"` | No |
| description | string | _Example:_ `"binlog cache size"` | No |
| hasApply | integer | _Enum:_ `0`, `1`<br>_Example:_ `1` | No |
| hasReboot | integer | _Enum:_ `0`, `1`<br>_Example:_ `0` | No |
| instanceType | string | _Example:_ `"tidb"` | No |
| name | string | _Example:_ `"binlog_size"` | No |
| note | string | _Example:_ `"binlog cache size"` | No |
| paramId | string | _Example:_ `"1"` | No |
| range | [ string ] | _Example:_ `["1"," 1000"]` | No |
| rangeType | integer | _Enum:_ `0`, `1`, `2`<br>_Example:_ `1` | No |
| readOnly | integer | _Enum:_ `0`, `1`<br>_Example:_ `0` | No |
| realValue | [structs.ParameterRealValue](#structsparameterrealvalue) |  | No |
| systemVariable | string | _Example:_ `"log.log_level"` | No |
| type | integer | _Enum:_ `0`, `1`, `2`, `3`, `4`<br>_Example:_ `0` | No |
| unit | string | _Example:_ `"MB"` | No |
| unitOptions | [ string ] | _Example:_ `["KB","MB","GB"]` | No |
| updateSource | integer | _Enum:_ `0`, `1`, `2`, `3`, `4`<br>_Example:_ `0` | No |
| updateTime | integer | _Example:_ `1636698675` | No |

#### structs.ClusterParameterSampleInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| paramId | string | _Example:_ `"1"` | Yes |
| realValue | [structs.ParameterRealValue](#structsparameterrealvalue) |  | Yes |

#### structs.ClusterRelations

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| masters | [ string ] |  | No |
| slaves | [ string ] |  | No |

#### structs.ClusterResourceInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| instanceResource | [ [structs.ClusterResourceParameterCompute](#structsclusterresourceparametercompute) ] |  | No |
| requestResourceMode | string | _Enum:_ `"SpecificZone"`, `"SpecificHost"` | No |

#### structs.ClusterResourceParameterCompute

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| componentType | string | TiDB/TiKV/PD/TiFlash/CDC/DM-Master/DM-Worker | No |
| resource | [ [structs.ClusterResourceParameterComputeResource](#structsclusterresourceparametercomputeresource) ] |  | No |
| totalNodeCount | integer |  | No |

#### structs.ClusterResourceParameterComputeResource

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| count | integer |  | No |
| diskCapacity | integer |  | No |
| diskId | string |  | No |
| diskType | string | NVMeSSD/SSD/SATA | No |
| hostIp | string |  | No |
| specCode | string | 4C8G/8C16G ? | No |
| zoneCode | string |  | No |

#### structs.ClusterUpgradeVersionConfigItem

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| instanceType | string | _Example:_ `"pd-server"` | Yes |
| name | string | _Example:_ `"max-merge-region-size"` | Yes |
| paramId | string | _Example:_ `"1"` | Yes |
| value | string | _Example:_ `"20"` | Yes |

#### structs.ComponentInstanceResourceSpec

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| cpu | integer |  | No |
| diskType | string | eg: NVMeSSD/SSD/SATA | No |
| id | string | ID of the instance resource specification | No |
| memory | integer | The amount of memory occupied by the instance, in GiB | No |
| name | string | Name of the instance resource specification,eg: TiDB.c1.large | No |
| zoneId | string |  | No |
| zoneName | string |  | No |

#### structs.ComponentInstanceZoneWithSpecs

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| specs | [ [structs.ComponentInstanceResourceSpec](#structscomponentinstanceresourcespec) ] |  | No |
| zoneId | string |  | No |
| zoneName | string |  | No |

#### structs.DataImportExportRecordInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| comment | string |  | No |
| createTime | string |  | No |
| deleteTime | string |  | No |
| endTime | string |  | No |
| filePath | string |  | No |
| recordId | string |  | No |
| startTime | string |  | No |
| status | string |  | No |
| storageType | string |  | No |
| transportType | string |  | No |
| updateTime | string |  | No |
| zipName | string |  | No |

#### structs.DiskInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| capacity | integer | Disk size, Unit: GB | No |
| diskId | string |  | No |
| hostId | string |  | No |
| name | string | [sda/sdb/nvmep0...] | No |
| path | string | Disk mount path: [/data1] | No |
| status | string | Disk Status, 0 for available, 1 for inused | No |
| type | string | Disk type: [nvme-ssd/ssd/sata] | No |

#### structs.HierarchyTreeNode

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | string |  | No |
| name | string |  | No |
| prefix | string |  | No |
| subNodes | [ [structs.HierarchyTreeNode](#structshierarchytreenode) ] |  | No |

#### structs.HostInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| arch | string | x86 or arm64 | No |
| availableDiskCount | integer | available disk count which could be used for allocation | No |
| az | string |  | No |
| clusterType | string | What cluster is the host used for? [database/data migration] | No |
| cpuCores | integer | Host cpu cores spec, init while importing | No |
| createTime | integer |  | No |
| diskType | string | Disk type of this host [SATA/SSD/NVMeSSD] | No |
| disks | [ [structs.DiskInfo](#structsdiskinfo) ] |  | No |
| hostId | string |  | No |
| hostName | string |  | No |
| instances | object |  | No |
| ip | string |  | No |
| kernel | string |  | No |
| loadStat | string | Host load stat, Loadless, Inused, Exhaust, etc | No |
| memory | integer | Host memory, init while importing | No |
| nic | string | Host network type: 1GE or 10GE | No |
| os | string |  | No |
| passwd | string |  | No |
| purpose | string | What Purpose is the host used for? [compute/storage/schedule] | No |
| rack | string |  | No |
| region | string |  | No |
| reserved | boolean | Whether this host is reserved - will not be allocated | No |
| spec | string | Host Spec, init while importing | No |
| sshPort | integer |  | No |
| status | string | Host status, Online, Offline, Failed, Deleted, etc | No |
| sysLabels | [ string ] |  | No |
| traits | integer | Traits of labels | No |
| updateTime | integer |  | No |
| usedCpuCores | integer | Unused CpuCore, used for allocation | No |
| usedMemory | integer | Unused memory size, Unit:GiB, used for allocation | No |
| userName | string |  | No |
| vendor | string |  | No |

#### structs.Index

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| description | string |  | No |
| name | string |  | No |
| unit | string |  | No |
| value | object |  | No |

#### structs.ParameterGroupParameterInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| category | string | _Example:_ `"basic"` | No |
| createTime | integer | _Example:_ `1636698675` | No |
| defaultValue | string | _Example:_ `"1"` | No |
| description | string | _Example:_ `"binlog cache size"` | No |
| hasApply | integer | _Enum:_ `0`, `1`<br>_Example:_ `1` | No |
| hasReboot | integer | _Enum:_ `0`, `1`<br>_Example:_ `0` | No |
| instanceType | string | _Example:_ `"tidb"` | No |
| name | string | _Example:_ `"binlog_size"` | No |
| note | string | _Example:_ `"binlog cache size"` | No |
| paramId | string | _Example:_ `"1"` | No |
| range | [ string ] | _Example:_ `["1"," 1000"]` | No |
| rangeType | integer | _Enum:_ `0`, `1`, `2`<br>_Example:_ `1` | No |
| readOnly | integer | _Enum:_ `0`, `1`<br>_Example:_ `0` | No |
| systemVariable | string | _Example:_ `"log.log_level"` | No |
| type | integer | _Enum:_ `0`, `1`, `2`, `3`, `4`<br>_Example:_ `0` | No |
| unit | string | _Example:_ `"MB"` | No |
| unitOptions | [ string ] | _Example:_ `["KB","MB","GB"]` | No |
| updateSource | integer | _Enum:_ `0`, `1`, `2`, `3`, `4`<br>_Example:_ `0` | No |
| updateTime | integer | _Example:_ `1636698675` | No |

#### structs.ParameterGroupParameterSampleInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| defaultValue | string | _Example:_ `"1"` | Yes |
| note | string | _Example:_ `"binlog cache size"` | No |
| paramId | string | _Example:_ `"123"` | Yes |

#### structs.ParameterRealValue

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterValue | string |  | No |
| instanceValue | [ [structs.ClusterInstanceParameterValue](#structsclusterinstanceparametervalue) ] |  | No |

#### structs.Product

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| arch | string |  | No |
| id | string | The ID of the product | No |
| internal | integer |  | No |
| name | string | the Name of the product | No |
| regionId | string |  | No |
| regionName | string |  | No |
| status | string |  | No |
| vendorId | string | the vendor ID of the vendor, e.go AWS | No |
| vendorName | string | the Vendor name of the vendor, e.g AWS/Aliyun | No |
| version | string |  | No |

#### structs.ProductComponentPropertyWithZones

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| availableZones | [ [structs.ComponentInstanceZoneWithSpecs](#structscomponentinstancezonewithspecs) ] | Information on the specifications of the resources online for the running of product components,organized by different Zone | No |
| endPort | integer |  | No |
| id | string | ID of the product component, globally unique | No |
| maxInstance | integer | Maximum number of instances when the product component is running, e.g. PD can run up to 7 instances, other components have no upper limit | No |
| maxPort | integer |  | No |
| minInstance | integer | Minimum number of instances of product components at runtime, e.g. at least 1 instance of PD, at least 3 instances of TiKV | No |
| name | string | Name of the product component, globally unique | No |
| purposeType | string | The type of resources required by the product component at runtime, e.g. storage class | No |
| startPort | integer |  | No |
| suggestedInstancesCount | [ integer ] |  | No |

#### structs.ProductConfigInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| components | [ [structs.ProductComponentPropertyWithZones](#structsproductcomponentpropertywithzones) ] |  | No |
| productId | string |  | No |
| productName | string |  | No |
| versions | [ [structs.SpecificVersionProduct](#structsspecificversionproduct) ] |  | No |

#### structs.ProductDetail

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | The ID of the product consists of the product ID | No |
| name | string | The name of the product consists of the product name and the version | No |
| versions | object | Organize product information by version | No |

#### structs.ProductSpecInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |

#### structs.ProductUpgradePathItem

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| upgradeType | string | _Enum:_ `"in-place"`, `"migration"` | Yes |
| upgradeWays | [ string ] | _Example:_ `["offline","online"]` | No |
| versions | [ string ] | _Example:_ `["v5.3.0","v5.4.0"]` | Yes |

#### structs.ProductUpgradeVersionConfigDiffItem

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| category | string | _Example:_ `"basic"` | Yes |
| currentValue | string | _Example:_ `"20"` | Yes |
| description | string | _Example:_ `"desc for max-merge-region-size"` | No |
| instanceType | string | _Example:_ `"pd-server"` | Yes |
| name | string | _Example:_ `"max-merge-region-size"` | Yes |
| paramId | string | _Example:_ `"1"` | Yes |
| range | [ string ] | _Example:_ `["1"," 1000"]` | Yes |
| rangeType | integer | _Enum:_ `0`, `1`, `2`<br>_Example:_ `1` | Yes |
| suggestValue | string | _Example:_ `"30"` | Yes |
| type | integer | _Enum:_ `0`, `1`, `2`, `3`, `4`<br>_Example:_ `0` | Yes |
| unit | string | _Example:_ `"MB"` | Yes |
| unitOptions | [ string ] | _Example:_ `["KB","MB","GB"]` | Yes |

#### structs.ProductVersion

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| arch | object | Arch information of the product, e.g. X86/X86_64 | No |
| version | string | Version information of the product, e.g. v5.0.0 | No |

#### structs.ProductWithVersions

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| productId | string |  | No |
| productName | string |  | No |
| versions | [ [structs.SpecificVersionProduct](#structsspecificversionproduct) ] |  | No |

#### structs.RbacPermission

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| action | string |  | No |
| resource | string |  | No |

#### structs.RegionConfigInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| zones | [ [structs.ZoneInfo](#structszoneinfo) ] |  | No |

#### structs.RegionInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |

#### structs.ResourceStockCheckResult

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| componentName | string |  | No |
| componentType | string |  | No |
| count | integer |  | No |
| diskCapacity | integer |  | No |
| diskId | string |  | No |
| diskType | string | NVMeSSD/SSD/SATA | No |
| enough | boolean |  | No |
| hostIp | string |  | No |
| specCode | string | 4C8G/8C16G ? | No |
| zoneCode | string |  | No |

#### structs.SpecInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| cpu | integer |  | No |
| diskType | string | eg: NVMeSSD/SSD/SATA | No |
| id | string | ID of the resource specification | No |
| memory | integer | The amount of memory occupied by the instance, in GiB | No |
| name | string | Name of the resource specification,eg: TiDB.c1.large | No |
| purposeType | string | eg:Compute/Storage/Schedule | No |

#### structs.SpecificVersionProduct

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| arch | string |  | No |
| productId | string |  | No |
| version | string |  | No |

#### structs.Stocks

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| freeCpuCores | integer |  | No |
| freeDiskCapacity | integer |  | No |
| freeDiskCount | integer |  | No |
| freeHostCount | integer |  | No |
| freeMemory | integer |  | No |
| zone | string |  | No |

#### structs.SystemInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| currentVersionId | string |  | No |
| lastVersionId | string |  | No |
| productComponentsInitialized | boolean |  | No |
| productVersionsInitialized | boolean |  | No |
| state | string |  | No |
| supportedProducts | [ [structs.ProductWithVersions](#structsproductwithversions) ] |  | No |
| supportedVendors | [ [structs.VendorInfo](#structsvendorinfo) ] |  | No |
| systemLogo | string |  | No |
| systemName | string |  | No |
| vendorSpecsInitialized | boolean |  | No |
| vendorZonesInitialized | boolean |  | No |

#### structs.SystemVersionInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| desc | string |  | No |
| releaseNote | string |  | No |
| versionId | string |  | No |

#### structs.TenantInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| createAt | string |  | No |
| creator | string |  | No |
| id | string |  | No |
| maxCluster | integer |  | No |
| maxCpu | integer |  | No |
| maxMemory | integer |  | No |
| maxStorage | integer |  | No |
| name | string |  | No |
| onBoardingStatus | string |  | No |
| status | string |  | No |
| updateAt | string |  | No |

#### structs.Usage

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| total | number |  | No |
| usageRate | number |  | No |
| used | number |  | No |

#### structs.UserInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| createAt | string |  | No |
| creator | string |  | No |
| defaultTenantId | string |  | No |
| email | string |  | No |
| id | string |  | No |
| names | [ string ] |  | No |
| nickname | string |  | No |
| phone | string |  | No |
| status | string |  | No |
| tenantIds | [ string ] |  | No |
| updateAt | string |  | No |

#### structs.VendorConfigInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | The value of the VendorID is similar to AWS | No |
| name | string | The value of the Name is similar to AWS | No |
| regions | [ [structs.RegionConfigInfo](#structsregionconfiginfo) ] |  | No |
| specs | [ [structs.SpecInfo](#structsspecinfo) ] |  | No |

#### structs.VendorInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | The value of the VendorID is similar to AWS | No |
| name | string | The value of the Name is similar to AWS | No |

#### structs.VendorWithRegion

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | The value of the VendorID is similar to AWS | No |
| name | string | The value of the Name is similar to AWS | No |
| regions | object |  | No |

#### structs.WorkFlowInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| bizId | string |  | No |
| bizType | string |  | No |
| createTime | string |  | No |
| deleteTime | string |  | No |
| id | string |  | No |
| name | string |  | No |
| status | string | _Enum:_ `"Initializing"`, `"Processing"`, `"Finished"`, `"Error"`, `"Canceled"` | No |
| updateTime | string |  | No |

#### structs.WorkFlowNodeInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| endTime | string |  | No |
| id | string |  | No |
| name | string |  | No |
| parameters | string |  | No |
| result | string |  | No |
| startTime | string |  | No |
| status | string | _Enum:_ `"Initializing"`, `"Processing"`, `"Finished"`, `"Error"`, `"Canceled"` | No |

#### structs.ZoneFullInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| comment | string |  | No |
| regionId | string | The value of the RegionID is similar to CN-HANGZHOU | No |
| regionName | string | The value of the Name is similar to East China(Hangzhou) | No |
| vendorId | string | The value of the VendorID is similar to AWS | No |
| vendorName | string | The value of the Name is similar to AWS | No |
| zoneId | string | The value of the ZoneID is similar to CN-HANGZHOU-H | No |
| zoneName | string | The value of the Name is similar to Hangzhou(H) | No |

#### structs.ZoneInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| comment | string |  | No |
| zoneId | string | The value of the ZoneID is similar to CN-HANGZHOU-H | No |
| zoneName | string | The value of the Name is similar to Hangzhou(H) | No |
