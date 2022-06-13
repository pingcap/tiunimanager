## /clusters/

### GET
#### Summary

query clusters

#### Description

query clusters

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | query |  | No | string |
| clusterName | query |  | No | string |
| clusterStatus | query |  | No | string |
| clusterTag | query |  | No | string |
| clusterType | query |  | No | string |
| page | query | Current page location | No | integer |
| pageSize | query | Number of this request | No | integer |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & [cluster.QueryClusterResp](#clusterqueryclusterresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### cluster.QueryClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusters | [ [structs.ClusterInfo](#structsclusterinfo) ] |  | No |
#### Example
request
```
curl -X 'GET' \
  'http://localhost:4116/api/v1/clusters/?page=1&pageSize=10' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken'
```
response
```JSON
{
  "code": 0,
  "data": {
    "clusters": [
      {
        "alertUrl": "http://127.0.0.1:9093",
        "backupFileUsage": {
          "total": 0,
          "usageRate": 0,
          "used": 0
        },
        "clusterId": "string",
        "clusterName": "string",
        "clusterType": "string",
        "clusterVersion": "string",
        "copies": 0,
        "cpuArchitecture": "string",
        "cpuUsage": {
          "total": 0,
          "usageRate": 0,
          "used": 0
        },
        "createTime": "string",
        "deleteTime": "string",
        "exclusive": true,
        "extranetConnectAddresses": [
          "string"
        ],
        "grafanaUrl": "http://127.0.0.1:3000",
        "intranetConnectAddresses": [
          "string"
        ],
        "maintainStatus": "string",
        "maintainWindow": "string",
        "memoryUsage": {
          "total": 0,
          "usageRate": 0,
          "used": 0
        },
        "region": "string",
        "relations": {
          "masters": [
            "string"
          ],
          "slaves": [
            "string"
          ]
        },
        "role": "string",
        "status": "string",
        "storageUsage": {
          "total": 0,
          "usageRate": 0,
          "used": 0
        },
        "tags": [
          "string"
        ],
        "tls": true,
        "updateTime": "string",
        "userId": "string",
        "vendor": "string",
        "whitelist": [
          "string"
        ]
      }
    ]
  },
  "message": "string",
  "page": {
    "page": 0,
    "pageSize": 0,
    "total": 0
  }
}
```

### POST
#### Summary

create a cluster

#### Description

create a cluster

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| createReq | body | create request | Yes | [cluster.CreateClusterReq](#clustercreateclusterreq) |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.CreateClusterResp](#clustercreateclusterresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### cluster.CreateClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |
#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/clusters/' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{
  "clusterName": "string",
  "clusterType": "string",
  "clusterVersion": "string",
  "copies": 0,
  "cpuArchitecture": "string",
  "dbPassword": "string",
  "dbUser": "string",
  "exclusive": true,
  "parameterGroupID": "string",
  "region": "string",
  "resourceParameters": {
    "instanceResource": [
      {
        "componentType": "string",
        "resource": [
          {
            "count": 0,
            "diskCapacity": 0,
            "diskId": "string",
            "diskType": "string",
            "hostIp": "string",
            "specCode": "string",
            "zoneCode": "string"
          }
        ],
        "totalNodeCount": 0
      }
    ],
    "requestResourceMode": "SpecificZone"
  },
  "tags": [
    "string"
  ],
  "tls": true,
  "vendor": "string"
}'
```
response
```JSON
{
  "code": 0,
  "data": {
    "clusterId": "string",
    "workFlowId": "string"
  },
  "message": "string"
}
```

### /clusters/{clusterId}

### DELETE
#### Summary

delete cluster

#### Description

delete cluster

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |
| deleteReq | body | delete request | No | [cluster.DeleteClusterReq](#clusterdeleteclusterreq) |

#### cluster.DeleteClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| autoBackup | boolean |  | No |
| force | boolean |  | No |
| keepHistoryBackupRecords | boolean |  | No |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.DeleteClusterResp](#clusterdeleteclusterresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### cluster.DeleteClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterID | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### Example
request
```
curl -X 'DELETE' \
  'http://localhost:4116/api/v1/clusters/121212' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{
  "autoBackup": true,
  "force": true,
  "keepHistoryBackupRecords": true
}'
```
response
```JSON
{
  "code": 0,
  "data": {
    "clusterID": "string",
    "workFlowId": "string"
  },
  "message": "string"
}
```

### GET
#### Summary

show details of a cluster

#### Description

show details of a cluster

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.QueryClusterDetailResp](#clusterqueryclusterdetailresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### cluster.QueryClusterDetailResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| info | [structs.ClusterInfo](#structsclusterinfo) |  | No |
| instanceResource | [ [structs.ClusterResourceParameterCompute](#structsclusterresourceparametercompute) ] |  | No |
| requestResourceMode | string | _Enum:_ `"SpecificZone"`, `"SpecificHost"` | No |
| topology | [ [structs.ClusterInstanceInfo](#structsclusterinstanceinfo) ] |  | No |

##### structs.ClusterInstanceInfo

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

##### structs.ClusterResourceParameterCompute

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| componentType | string | TiDB/TiKV/PD/TiFlash/CDC/DM-Master/DM-Worker | No |
| resource | [ [structs.ClusterResourceParameterComputeResource](#structsclusterresourceparametercomputeresource) ] |  | No |
| totalNodeCount | integer |  | No |

##### structs.ClusterResourceParameterComputeResource

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| count | integer |  | No |
| diskCapacity | integer |  | No |
| diskId | string |  | No |
| diskType | string | NVMeSSD/SSD/SATA | No |
| hostIp | string |  | No |
| specCode | string | 4C8G/8C16G ? | No |
| zoneCode | string |  | No |
##### structs.ClusterInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| alertUrl | string | _Example:_ `"http://127.0.0.1:9093"` | No |
| clusterId | string |  | No |
| clusterName | string |  | No |
| clusterType | string |  | No |
| clusterVersion | string |  | No |
| copies | integer | The number of copies of the newly created cluster data, consistent with the number of copies set in PD | No |
| cpuArchitecture | string | X86/X86_64/ARM | No |
| createTime | string |  | No |
| deleteTime | string |  | No |
| exclusive | boolean | Whether the newly created cluster is exclusive to physical resources, when exclusive, a host will only deploy instances of the same cluster, which may result in poor resource utilization | No |
| extranetConnectAddresses | [ string ] |  | No |
| grafanaUrl | string | _Example:_ `"http://127.0.0.1:3000"` | No |
| intranetConnectAddresses | [ string ] |  | No |
| maintainStatus | string |  | No |
| maintainWindow | string |  | No |
| region | string |  | No |
| relations | [structs.ClusterRelations](#structsclusterrelations) |  | No |
| role | string |  | No |
| status | string |  | No |
| tags | [ string ] |  | No |
| tls | boolean |  | No |
| updateTime | string |  | No |
| userId | string |  | No |
| vendor | string | DBUser                   string    `json:"dbUser"` //The username and password for the newly created database cluster, default is the root user, which is not valid for Data Migration clusters | No |
| whitelist | [ string ] |  | No |

##### structs.ClusterRelations

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| masters | [ string ] |  | No |
| slaves | [ string ] |  | No |
#### Example
request
```
curl -X 'GET' \
  'http://localhost:4116/api/v1/clusters/121212' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken'
```
response
```JSON
{
  "code": 0,
  "data": {
    "info": {
      "alertUrl": "http://127.0.0.1:9093",
      "backupFileUsage": {
        "total": 0,
        "usageRate": 0,
        "used": 0
      },
      "clusterId": "string",
      "clusterName": "string",
      "clusterType": "string",
      "clusterVersion": "string",
      "copies": 0,
      "cpuArchitecture": "string",
      "createTime": "string",
      "deleteTime": "string",
      "exclusive": true,
      "extranetConnectAddresses": [
        "string"
      ],
      "grafanaUrl": "http://127.0.0.1:3000",
      "intranetConnectAddresses": [
        "string"
      ],
      "maintainStatus": "string",
      "maintainWindow": "string",
      "region": "string",
      "relations": {
        "masters": [
          "string"
        ],
        "slaves": [
          "string"
        ]
      },
      "role": "string",
      "status": "string",
      "tags": [
        "string"
      ],
      "tls": true,
      "updateTime": "string",
      "userId": "string",
      "vendor": "string",
      "whitelist": [
        "string"
      ]
    },
    "instanceResource": [
      {
        "componentType": "string",
        "resource": [
          {
            "count": 0,
            "diskCapacity": 0,
            "diskId": "string",
            "diskType": "string",
            "hostIp": "string",
            "specCode": "string",
            "zoneCode": "string"
          }
        ],
        "totalNodeCount": 0
      }
    ],
    "requestResourceMode": "SpecificZone",
    "topology": [
      {
        "addresses": [
          "string"
        ],
        "diskId": "string",
        "hostID": "string",
        "id": "string",
        "ioUtil": 0,
        "iops": [
          0
        ],
        "ports": [
          0
        ],
        "role": "string",
        "spec": {
          "id": "string",
          "name": "string"
        },
        "status": "string",
        "type": "string",
        "version": "string",
        "zone": {
          "comment": "string",
          "regionId": "string",
          "regionName": "string",
          "vendorId": "string",
          "vendorName": "string",
          "zoneId": "string",
          "zoneName": "string"
        }
      }
    ]
  },
  "message": "string"
}
```

## /clusters/{clusterId}/stop

### POST
#### Summary

stop a cluster

#### Description

stop a cluster

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.StopClusterResp](#clusterstopclusterresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### cluster.StopClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/clusters/121212/stop' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -d ''
```
response
```JSON
{
  "code": 0,
  "data": {
    "clusterId": "string",
    "workFlowId": "string"
  },
  "message": "string"
}
```

## /clusters/{clusterId}/restart

### POST
#### Summary

restart a cluster

#### Description

restart a cluster

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| clusterId | path | cluster id | Yes | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.RestartClusterResp](#clusterrestartclusterresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### cluster.RestartClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |

#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/clusters/121212/restart' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -d ''
```
response
```JSON
{
  "code": 0,
  "data": {
    "clusterId": "string",
    "workFlowId": "string"
  },
  "message": "string"
}
```

## /clusters/takeover

### POST
#### Summary

takeover a cluster

#### Description

takeover a cluster

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| takeoverReq | body | takeover request | Yes | [cluster.TakeoverClusterReq](#clustertakeoverclusterreq) |

##### cluster.TakeoverClusterReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| TiUPIp | string | _Example:_ `"172.16.4.147"` | Yes |
| TiUPPath | string | _Example:_ `".tiup/"` | Yes |
| TiUPPort | integer | _Example:_ `22` | Yes |
| TiUPUserName | string | _Example:_ `"root"` | Yes |
| TiUPUserPassword | string | _Example:_ `"password"` | Yes |
| clusterName | string | _Example:_ `"myClusterName"` | Yes |
| dbPassword | string | _Example:_ `"myPassword"` | Yes |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [cluster.TakeoverClusterResp](#clustertakeoverclusterresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### cluster.TakeoverClusterResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| clusterId | string |  | No |
| workFlowId | string | Asynchronous task workflow ID | No |
#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/clusters/takeover' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{
  "TiUPIp": "172.16.4.147",
  "TiUPPath": ".tiup/",
  "TiUPPort": 22,
  "TiUPUserName": "root",
  "TiUPUserPassword": "password",
  "clusterName": "myClusterName",
  "dbPassword": "myPassword"
}'
```
response
```JSON
{
  "code": 0,
  "data": {
    "clusterId": "string",
    "workFlowId": "string"
  },
  "message": "string"
}
```

## CommonModel

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
