# Cluster Log

The current implementation mainly includes cluster log management.

## Query cluster log

### Request URI

```HTTP
GET /api/v1/clusters/{clusterId}/log
```

### Request Parameters

#### clusterId

Cluster id for cluster log.

Type：string

Category: path

Required: Yes

#### startTime

query log start time.

Type：Integer

Category: query

Required: No

#### endTime

query log end time.

Type：Integer

Category: query

Required: No

#### ip

filtering parameters, filtered by ip.

Type：string

Category: query

Required: No

#### level

filtering parameters, filtered by log level.

Type：string

Category: query

Required: No

#### message

filtering parameters, filtered by log content.

Type：string

Category: query

Required: No

#### module

filtering parameters, filtered by log module.

Type：string

Category: query

Required: No

#### page

paging parameters, current page number.

Type：Integer

Category: query

Required: No

#### pageSize

paging parameters, number of entries per page.

Type：Integer

Category: query

Required: No

### Responses

200 OK
```json
{
  "code": 0,
  "data": {
    "results": [
      {
        "clusterId": "string",
        "ext": {},
        "id": "string",
        "index": "string",
        "ip": "string",
        "level": "string",
        "message": "string",
        "module": "string",
        "sourceLine": "string",
        "timestamp": "string"
      }
    ],
    "took": 10
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
curl -X 'GET' \
curl --location --request GET \
    'http://$IP:$PORT/api/v1/clusters/$CLUSTER_ID/log?page=0&pageSize=10&module=tidb' \
    --header 'Authorization: Bearer $TOKEN'
```
