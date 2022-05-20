# Upgrade Related API

## Query Upgrade Path

Query upgrade path for given cluster ID

### Request URI

```HTTP
GET /clusters/{clusterId}/upgrade/path
```

### Request Parameters

None

### Responses

200 OK

```JSON
{
  "code": 0,
  "data": {
    "paths": [
      {
        "upgradeType": "in-place",
        "upgradeWays": [
          "offline",
          "online"
        ],
        "versions": [
          "v5.3.0",
          "v5.4.0"
        ]
      }
    ]
  },
  "message": "string"
}
```

401 Unauthorized

```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

500 Internal Server Error

```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

### Request Sample

```Shell
curl -X 'GET' \
  'http://$IP:$PORT/api/v1/clusters/$CLUSTER_ID/upgrade/path' \
  -H 'accept: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json'
```

## Query Config Difference

Query config difference between the current running cluster's and default one's of target upgrade version

### Request URI

```HTTP
GET /clusters/{clusterId}/upgrade/diff
```

### Request Parameters

#### targetVersion

The targetVersion is used for defining the TiDB version to be upgraded

Type: String

Required: Yes

### Responses

200 OK

```JSON
{
  "code": 0,
  "data": {
    "configDiffInfos": [
      {
        "category": "basic",
        "currentValue": "20",
        "description": "desc for max-merge-region-size",
        "instanceType": "pd-server",
        "name": "max-merge-region-size",
        "paramId": "1",
        "range": [
          "1",
          " 1000"
        ],
        "rangeType": 1,
        "suggestValue": "30",
        "type": 0,
        "unit": "MB",
        "unitOptions": [
          "KB",
          "MB",
          "GB"
        ]
      }
    ]
  },
  "message": "string"
}
```

401 Unauthorized

```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

500 Internal Server Error

```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

### Request Sample

```Shell
curl -X 'GET' \
  'http://$IP:$PORT/api/v1/clusters/$CLUSTER_ID/upgrade/diff?targetVersion=$UPGRADE_VERSION' \
  -H 'accept: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json'
```

## Request upgrade TiDB cluster

request for upgrading TiDB cluster

### Request URI

```HTTP
POST /clusters/{clusterId}/upgrade
```

### Request Parameters

#### configs

The configs are arrays of user's choice of parameter difference between running cluster's and the target version TiDB's default template's

Type: Array

Required: No

#### instanceType

The instanceType is used for defining the instance type of TiDB parameter

Type: String

Required: No

#### name

The name is used for defining the name of TiDB parameter

Type: String

Required: No

#### paramId

The paramId is used for defining the id of TiDB parameter specified in EM

Type: String

Required: No

#### value

The value is used for defining the value of TiDB parameter chosen by user

Type: String

Required: No

#### targetVersion

The targetVersion is used for defining the TiDB version to be upgraded

Type: String

Required: No

#### upgradeType

The upgradeType is used for defining the TiDB upgrade type: in-place or migration

Type: String

Required: No

#### upgradeWay

The upgradeWay is used for defining the TiDB upgrade way: online or offline

Type: String

Required: No

### Responses

200 OK

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

401 Unauthorized

```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

500 Internal Server Error

```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

### Request Sample

```Shell
curl -X 'POST' \
 'http://$IP:$PORT/api/v1/clusters/$CLUSTER_ID/upgrade' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
  "configs": [
    {
      "instanceType": "pd-server",
      "name": "max-merge-region-size",
      "paramId": "1",
      "value": "20"
    }
  ],
  "targetVersion": "v5.4.0",
  "upgradeType": "in-place",
  "upgradeWay": "offline"
}'
```