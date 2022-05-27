## Clone

### Request URI

```HTTP
POST /api/v1/clusters/clone
```

### Major Request Parameters

#### sourceClusterId

Source cluster id for cloning.

Type：String

Required: Yes

#### cloneStrategy

Include 'CDCSync' and 'Snapshot'

Type：String

Required: Yes

#### resourceParameters

Resource parameters for the target cluster

Type：Object

Required: Yes

### Responses

200 OK

```JSON
{
  "code": 0,
  "data": {
    "workflowID": "CHx6yvG4SPyHuS_h4xAzWA",
  },
  "message": "OK"
}
```

### Request Sample

clone cluster：

```Shell
curl -vX 'POST' \
  'http://$IP:$PORT/api/v1/clusters/clone' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
   "clusterName" : "$clusterName",
   "clusterType" : "TiDB",
   "clusterVersion" : "v5.2.2",
   "copies" : 1,
   "cpuArchitecture" : "X86_64",
   "dbPassword" : "mypassword",
   "dbUser" : "",
   "exclusive" : false,
   "parameterGroupID" : "",
   "region" : "TEST_Region1",
   "tags" : [
      "tag2"
   ],
   "resourceParameters": {
    "instanceResource": [
      {
        "componentType": "TiDB",
        "resource": [
          {
            "count": 1,
            "diskCapacity": 0,
            "diskType": "SATA",
            "specCode": "4C8G",
            "zoneCode": "TEST_Zone1"
          }
        ],
        "totalNodeCount": 1
      },
      {
        "componentType": "TiKV",
        "resource": [
          {
            "count": 1,
            "diskCapacity": 0,
            "diskType": "SATA",
            "specCode": "4C8G",
            "zoneCode": "TEST_Zone1"
          }
        ],
        "totalNodeCount": 1
      },
      {
        "componentType": "PD",
        "resource": [
          {
            "count": 1,
            "diskCapacity": 0,
            "diskType": "SATA",
            "specCode": "4C8G",
            "zoneCode": "TEST_Zone1"
          }
        ],
        "totalNodeCount": 1
      }
    ]
    },
   "tls" : false,
   "vendor": "Local",
   "cloneStrategy" : "CDCSync",
   "sourceClusterId" : "$sourceClusterId"
}'
```

Successful response:

```bash
{
  "code": 0,
  "data": {
    "workflowID": "$NewWorkFlowID",
  },
  "message": "OK"
}
```