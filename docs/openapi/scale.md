# Scale

The current implementation mainly includes scale-in cluster and scale-out cluster.

## Scale out

### Request URI

```HTTP
POST /api/v1/clusters/{clusterId}/scale-out
```

### Request Parameters

#### clusterId

Cluster id for scaling out.

Type：String

Required: Yes

#### instanceResource

Instance resources for scaling out.

Type：Array

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

scale-out：

```Shell
curl -vX 'POST' \
  'http://$IP:$PORT/api/v1/clusters/$clusterId/scale-out' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
    "instanceResource":[
        {
            "componentType":"TiDB",
            "resource":[
                {
                    "count":1,
                    "diskCapacity":0,
                    "diskType":"SATA",
                    "specCode":"4C8G",
                    "zoneCode":"Region1,Zone1_1"
                }
            ],
            "totalNodeCount":1
        }
    ]
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

## Scale in

### Request URI

```HTTP
POST /api/v1/clusters/{clusterId}/scale-in
```

### Request Parameters

#### clusterId

Cluster id for scaling in.

Type：String

Required: Yes

#### instanceId

Instance id for scaling in.

Type：String

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

scale-in：

```Shell
curl -vX 'POST' \
  'http://$IP:$PORT/api/v1/clusters/$clusterId/scale-in' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{ "instanceId": "$instanceId" }'
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