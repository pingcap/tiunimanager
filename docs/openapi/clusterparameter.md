# Cluster Parameters

The current implementation mainly includes cluster parameter management.

## Query cluster parameter

### Request URI

```HTTP
GET /api/v1/clusters/{clusterId}/params
```

### Request Parameters

#### clusterId

Cluster id for query cluster.

Type：String

Category: path

Required: Yes

#### instanceType

Instance type for query cluster.

Type：String

Category: query

Required: No

#### paramName

Param name for query cluster.

Type：String

Category: query

Required: No

#### page

Paging parameters, current page number.

Type：Integer

Category: query

Required: No

#### pageSize

Paging parameters, number of entries per page.

Type：Integer

Category: query

Required: No

### Responses

200 OK
```json
{
  "code": 0,
  "data": {
    "paramGroupId": "string",
    "params": [
      {
        "category": "basic",
        "createTime": 1636698675,
        "defaultValue": "1",
        "description": "binlog cache size",
        "hasApply": 1,
        "hasReboot": 0,
        "instanceType": "tidb",
        "name": "binlog_size",
        "note": "binlog cache size",
        "paramId": "1",
        "range": [
          "1",
          " 1000"
        ],
        "rangeType": 1,
        "readOnly": 0,
        "realValue": {
          "clusterValue": "string",
          "instanceValue": [
            {
              "instanceId": "string",
              "value": "string"
            }
          ]
        },
        "systemVariable": "log.log_level",
        "type": 0,
        "unit": "MB",
        "unitOptions": [
          "KB",
          "MB",
          "GB"
        ],
        "updateSource": 0,
        "updateTime": 1636698675
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

401 Unauthorized
``` json
{
  "code": int,
  "data": {},
  "message": "string"
}
```

500 Internal Server Error
``` json
{
  "code": int,
  "data": {},
  "message": "string"
}
```

### Request Sample
``` shell
curl --location --request GET \
'http://$IP:$PORT/api/v1/clusters/CkylV1LUToGsPPUoIHUcmg/params?page=1&pageSize=10' \
--header 'Authorization: Bearer $TOKEN'
```

## Modify cluster parameter

### Request URI

```HTTP
PUT /api/v1/clusters/{clusterId}/params
```

### Request Parameters

#### clusterId

Cluster id for inspect cluster.

Type：String

Category: path

Required: Yes

#### params

Modify parameter for cluster.

Type：Array

Category: body

Required: Yes

#### reboot

Specifies whether to restart the cluster.

Type：String

Category: body

Required: Yes

### Responses

200 OK
```json
{
  "code": 0,
  "data": {
    "clusterId": "string",
    "workFlowId": "string"
  },
  "message": "string"
}
```

401 Unauthorized
``` json
{
  "code": int,
  "data": {},
  "message": "string"
}
```

500 Internal Server Error
``` json
{
  "code": int,
  "data": {},
  "message": "string"
}
```

### Request Sample
``` shell
curl --location --request PUT \
'http://$IP:$PORT/api/v1/clusters/$CLUSTER_ID/params' \
--header 'Authorization: Bearer $TOKEN' \
--header 'Content-Type: application/json' \
--data-raw '{
    "params": [
        {
            "paramId": "400",
            "realValue": {
                "clusterValue": "info"
            }
        }
    ],
    "reboot": false
}'
```


## Inspect cluster parameter

### Request URI

```HTTP
POST /api/v1/clusters/{clusterId}/params/inspect
```

### Request Parameters

#### clusterId

Cluster id for inspect cluster.

Type：String

Category: path

Required: Yes

#### instanceId

Instance id for inspect cluster.

Type：String

Category: body

Required: No

### Responses

200 OK
```json
{
  "code": 0,
  "data": {
    "params": [
      {
        "instanceId": "string",
        "instanceType": "string",
        "parameterInfos": [
          {
            "category": "log",
            "description": "binlog cache size",
            "hasReboot": 0,
            "inspectValue": {},
            "name": "binlog_cache",
            "note": "binlog cache size",
            "paramId": "1",
            "range": [
              "1",
              " 1000"
            ],
            "rangeType": 1,
            "readOnly": 0,
            "realValue": {
              "clusterValue": "string",
              "instanceValue": [
                {
                  "instanceId": "string",
                  "value": "string"
                }
              ]
            },
            "systemVariable": "log.log_level",
            "type": 0,
            "unit": "MB",
            "unitOptions": [
              "KB",
              "MB",
              "GB"
            ]
          }
        ]
      }
    ]
  },
  "message": "string"
}
```

401 Unauthorized
``` json
{
  "code": int,
  "data": {},
  "message": "string"
}
```

500 Internal Server Error
``` json
{
  "code": int,
  "data": {},
  "message": "string"
}
```
### Request Sample
``` shell
curl --location --request POST \
'http://$IP:$PORT/api/v1/clusters/$CLUSTER_ID/params/inspect' \
--header 'Authorization: Bearer $TOKEN' \
--header 'Content-Type: application/json' \
--data-raw '{
    "instanceId": ""
}'
```
