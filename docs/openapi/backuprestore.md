## Backup && Restore
### GetBackupStrategy
User get auto backup strategy of cluster
#### Request URI
```HTTP
GET /clusters/[cluster-id]/strategy
```

#### Request Parameters
None

#### Responses
200 OK
```JSON
{
  "code":0,
  "message":"OK",
  "data":{
    "strategy":{
      "clusterId":"oEcox8lMQ7KMswDebRwIZg",
      "backupDate":"",
      "period":"0:00-1:00"
    }
  }
}
```

#### Request Sample
```HTTP
curl -v -X GET "http://172.16.5.136:4116/api/v1/clusters/test-tidb/strategy" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 6b43c6e2-8e96-48f8-acbc-eddfff41d40a"
```

### SaveBackupStrategy
User save auto backup strategy of cluster
#### Request URI
```HTTP
POST /clusters/[cluster-id]/strategy
```

#### Request Parameters
##### strategy
Struct of backup strategy
Type：struct
Required: Yes

```JSON
{
  "strategy":{
    "backupDate":"Monday,Tuesday",
    "period":"00:00-01:00"
  }
}
```

#### Responses
200 OK
```JSON
{
    "code": 0,
    "message": "OK",
    "data": {}
}
```

#### Request Sample
```HTTP
curl -v -X PUT "http://172.16.5.136:4116/api/v1/clusters/test-tidb/strategy" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 6b43c6e2-8e96-48f8-acbc-eddfff41d40a" -d "{\"strategy\":{\"clusterId\":\"test-tidb\",\"backupDate\":\"Tuesday,Wednesday\",\"period\":\"7:00-9:00\"}}"
```

### Backup
User do backup
#### equest URI
POST /backups

#### Request Parameters
##### clusterId
clusterId to be backup
Type：struct
Required: Yes
##### backupMode
Manual Auto backup mode
Type：struct
Required: No

```JSON
{
    "clusterId": "oEcox8lMQ7KMswDebRwIZg",
    "backupMode": "manual"
}
```

#### Responses
200 OK
```JSON
{
"code": 0,
"message": "OK",
"data": {}
}
```

#### Request Sample
```HTTP
curl -v -X POST "http://172.16.4.147:4116/api/v1/backups/" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer ca9f709d-0c1c-458d-80ca-209a4ec1f9ee" -d "{ \"clusterId\": \"AM0kcZ1qSkyE_PKal6jQmA\", \"backupMode\": \"manual\"}"
```

### CancelBackup
User cancel running backup
#### Request URI
```HTTP
POST /backups/cancel
```

#### Request Parameters
##### clusterId
clusterId to be backup
Type：string
Required: Yes
##### backupId
Running backupId of cluster
Type：string
Required: Yes

```JSON
{
    "clusterId": "oEcox8lMQ7KMswDebRwIZg",
    "backupId": "xxxxxxxx"
}
```

#### Responses
200 OK
```JSON
{
    "code": 0,
    "message": "OK",
    "data": {}
}
```
#### Request Sample
```HTTP
curl -v -X POST "http://172.16.4.147:4116/api/v1/backups/cancel" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer ca9f709d-0c1c-458d-80ca-209a4ec1f9ee" -d "{ \"clusterId\": \"AM0kcZ1qSkyE_PKal6jQmA\", \"backupId\": \"xxxxxxx\"}"
```

### QueryBackupRecords
Query backup records of one cluster
#### Request URI
```HTTP
GET /backups
```

#### Request Parameters
##### clusterId
clusterId to be backup
Type：string
Required: Yes
##### backupId
Running backupId of cluster
Type：string
Required: Yes
```JSON
{
    "clusterId": "oEcox8lMQ7KMswDebRwIZg",
    "backupId": "xxxxxxxx"
}
```

#### Responses
200 OK
```JSON
{
  "code":0,
  "message":"OK",
  "data":{
    "backupRecords":[
      {
        "id":"h6IhUQh-Q32MDOfcVieYgA",
        "clusterId":"oEcox8lMQ7KMswDebRwIZg",
        "backupType":"full",
        "backupMethod":"physical",
        "backupMode":"manual",
        "filePath":"nfs/em/backup/oEcox8lMQ7KMswDebRwIZg/2022-04-28-16-00-41-full",
        "size":0,
        "backupTso":"432834567783317505",
        "status":"Finished",
        "startTime":"2022-04-28T16:00:41.697491505+08:00",
        "endTime":"2022-04-28T16:00:42.263135237+08:00",
        "createTime":"2022-04-28T16:00:41.697748779+08:00",
        "updateTime":"2022-04-28T16:00:42.267644671+08:00",
        "deleteTime":"0001-01-01T00:00:00Z"
      }
    ]
  },
  "page":{
    "page":1,
    "pageSize":10,
    "total":1
  }
}
```
#### Request Sample
curl -v -X GET "http://172.16.5.136:4116/api/v1/backups/?clusterId=test-tidb&page=1&pageSize=30" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 9771156d-b56a-4e32-83e1-db3203d1d682"

### DeleteBackup
Delete backup record with backup files
#### Request URI
```HTTP
DELETE /backups/[backupId]
```

#### Request Parameters
None
Responses
200 OK
```JSON
{
    "code": 0,
    "message": "OK",
    "data": {}
}
```
#### Request Sample
```HTTP
curl -v -X DELETE "http://172.16.5.136:4116/api/v1/backups/xxxxxx" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 9771156d-b56a-4e32-83e1-db3203d1d682"
```
