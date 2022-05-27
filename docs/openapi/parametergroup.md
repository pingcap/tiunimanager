# Parameter Groups

The current implementation mainly includes parameter group management.

## Query parameter group

### Request URI

```HTTP
GET /api/v1/param-groups/
```

### Request Parameters

#### name

Name for parameter group.

Type：String

Category: query

Required: No

#### clusterSpec

Cluster spec for parameter group.

Type：String

Category: query

Required: No

#### clusterVersion

Cluster version for parameter group.

Type：String

Category: query

Required: No

#### dbType

DB type for parameter group.

Type：Integer

Category: query

Required: No

#### hasDefault

Has default for parameter group.

Type：Integer

Category: query

Required: No

#### hasDetail

Show parameter group details.

Type：Integer

Category: query

Required: No

#### page

Paging parameters, current page number.

Type：integer

Category: query

Required: No

#### pageSize

Paging parameters, number of entries per page.

Type：integer

Category: query

Required: No

### Responses

200 OK
```json
{
  "code": 0,
  "data": [
    {
      "clusterSpec": "8C16G",
      "clusterVersion": "v5.0",
      "createTime": 1636698675,
      "dbType": 1,
      "groupType": 0,
      "hasDefault": 1,
      "name": "default",
      "note": "default param group",
      "paramGroupId": "1",
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
      ],
      "updateTime": 1636698675
    }
  ],
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
'http://$IP:$PORT/api/v1/param-groups/?page=1&pageSize=10' \
--header 'Authorization: Bearer $TOKEN'
```

## Create parameter group

### Request URI

```HTTP
POST /api/v1/param-groups/
```

### Request Parameters

#### name

Name for parameter group.

Type：String

Category: body

Required: Yes

#### clusterSpec

Cluster spec for parameter group.

Type：String

Category: body

Required: Yes

#### clusterVersion

Cluster version for parameter group.

Type：String

Category: body

Required: Yes

#### dbType

DB type for parameter group.

Type：Integer

Category: body

Required: Yes

#### groupType

Group type for parameter group.

Type：Integer

Category: body

Required: Yes

#### params

Parameter list managed by parameter group.

Type：Array

Category: body

Required: Yes

#### note

Note for parameter group.

Type：String

Category: body

Required: No

### Responses

200 OK
```json
{
  "code": 0,
  "data": {
    "paramGroupId": "1"
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
'http://$IP:$PORT/api/v1/param-groups/' \
--header 'Authorization: Bearer $TOKEN' \
--header 'Content-Type: application/json' \
--data-raw '{
  "dbType": 1,
  "groupType": 1,
  "hasDefault": 1,
  "name": "8C16GV4_default",
  "note": "default param group",
  "params": [
    {
      "defaultValue": "1",
      "description": "binlog cache size",
      "paramId": "1"
    }
  ],
  "clusterSpec": "8C16G",
  "clusterVersion": "v5.0"
}'
```

## Detail parameter group

### Request URI

```HTTP
GET /api/v1/param-groups/{paramGroupId}
```

### Request Parameters

#### paramGroupId

ID for parameter group.

Type：String

Category: path

Required: Yes

#### instanceType

Name for parameter group.

Type：String

Category: query

Required: No

#### instanceType

Instance type for parameter group.

Type：String

Category: query

Required: No

### Responses

200 OK
```json
{
  "code": 0,
  "data": {
    "clusterSpec": "8C16G",
    "clusterVersion": "v5.0",
    "createTime": 1636698675,
    "dbType": 1,
    "groupType": 0,
    "hasDefault": 1,
    "name": "default",
    "note": "default param group",
    "paramGroupId": "1",
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
    ],
    "updateTime": 1636698675
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
curl --location --request GET \
'http://$IP:$PORT/api/v1/param-groups/$PARAM_GROUP_ID' \
--header 'Authorization: Bearer $TOKEN'
```

## Update parameter group

### Request URI

```HTTP
PUT /api/v1/param-groups/{paramGroupId}
```

### Request Parameters

#### paramGroupId

ID for parameter group.

Type：String

Category: path

Required: Yes

#### name

Name for parameter group.

Type：String

Category: body

Required: Yes

#### clusterSpec

Cluster spec for parameter group.

Type：String

Category: body

Required: Yes

#### clusterVersion

Cluster version for parameter group.

Type：String

Category: body

Required: Yes

#### params

Parameter list managed by parameter group.

Type：Array

Category: body

Required: Yes

#### note

Note for parameter group.

Type：String

Category: body

Required: No

### Responses

200 OK
```json
{
  "code": 0,
  "data": {
    "paramGroupId": "1"
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
'http://$IP:$PORT/api/v1/param-groups/iBy7ix8VRo2ehWv5z90u2A' \
--header 'Authorization: Bearer $TOKEN' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "8C16GV5_update",
    "note": "update param group",
    "params": [
        {
            "defaultValue": "2",
            "description": "binlog cache size update",
            "paramId": "159"
        }
    ],
    "addParams": [
        {
            "category": "log",
            "defaultValue": "1024",
            "description": "binlog size",
            "hasApply": 1,
            "hasReboot": 0,
            "instanceType": "TiDB",
            "name": "binlog_size",
            "note": "binlog cache",
            "range": [
                "1",
                "1024"
            ],
            "readOnly": 0,
            "systemVariable": "log.binlog_size",
            "type": 0,
            "unit": "mb",
            "updateSource": 0
        }
    ],
    "delParams": [],
    "clusterSpec": "18C26G",
    "clusterVersion": "v5.1"
}'
```

## Delete parameter group

### Request URI

```HTTP
DELETE /api/v1/param-groups/{paramGroupId}
```

### Request Parameters

#### paramGroupId

ID for parameter group.

Type：String

Category: path

Required: Yes

### Responses

200 OK
```json
{
  "code": 0,
  "data": {
    "paramGroupId": "1"
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
curl --location --request DELETE \
'http://$IP:$PORT/api/v1/param-groups/$PARAM_GROUP_ID' \
--header 'Authorization: Bearer $TOKEN'
```

## Apply parameter group

### Request URI

```HTTP
POST /api/v1/param-groups/{paramGroupId}/apply
```

### Request Parameters

#### paramGroupId

ID for parameter group.

Type：String

Category: path

Required: Yes

#### clusterId

Cluster applying parameter groups.

Type：String

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
    "clusterId": "123",
    "paramGroupId": "123",
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
curl --location --request POST \
'http://$IP:$PORT/api/v1/param-groups/1/apply' \
--header 'Authorization: Bearer $TOKEN' \
--header 'Content-Type: application/json' \
--data-raw '{
    "clusterId": "ahQ-OO1cQS2DALlZqkCfLg",
    "reboot": false
}'
```

## Copy parameter group

### Request URI

```HTTP
POST /api/v1/param-groups/{paramGroupId}/copy
```

### Request Parameters

#### paramGroupId

ID for parameter group.

Type：String

Category: path

Required: Yes

#### name

Name for copy parameter group.

Type：String

Category: body

Required: Yes

#### note

Note for copy parameter group.

Type：String

Category: body

Required: No

### Responses

200 OK
```json
{
  "code": 0,
  "data": {
    "paramGroupId": "1"
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
'http://$IP:$PORT/api/v1/param-groups/$PARAM_GROUP_ID/copy' \
--header 'Authorization: Bearer $TOKEN' \
--header 'Content-Type: application/json' \
--data-raw '{
  "name": "8C16GV5_default",
  "note": "copy param group"
}'
```
