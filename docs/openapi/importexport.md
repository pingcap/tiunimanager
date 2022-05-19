## Import && Export
### ImportData
User import  data to cluster
#### Request URI
```HTTP
GET /clusters/import
```

#### Request Parameters
##### clusterId
clusterId to be backup
Type：string
Required: Yes
##### userName
DB account username
Type：string
Required: Yes
##### password
DB account password
Type：string
Required: Yes
##### recordId
History of import and export record Id
Type：string
Required: No
##### storageType
Storage type of data source（nfs/s3）
Type：string
Required: No
##### endpointUrl
s3 endpoint url
Type：string
Required: No（when choose s3 storageType）
##### bucketUrl
s3 bucket url
Type：string
Required: No（when choose s3 storageType）
##### accessKey
s3 access key
Type：string
Required: No（when choose s3 accesskey）
##### secretAccessKey
s3 SecretAccess Key
Type：string
Required: No（when choose s3 storageType）
##### comment
Comment of this import
Type：string
Required: No
#### Responses
200 OK
```HTTP
{
    "code":0,
    "message":"OK",
    "data":{
        "RecordId":"xxxxxxx"
    }
}
```
#### Request Sample
```HTTP
curl -v -X POST "http://172.16.4.147:4116/api/v1/clusters/import" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 2b0febff-8863-46e0-9dcf-67815e61b54a" -d "{ \"clusterId\": \"AM0kcZ1qSkyE_PKal6jQmA\", \"userName\": \"root\", \"password\": \"mypassword\", , \"storageType\": \"nfs\"}"
```

### ExportData
User export  data from cluster
#### Request URI
```HTTP
GET /clusters/export
```

#### Request Parameters
##### clusterId
clusterId to be backup
Type：string
Required: Yes
##### userName
DB account username
Type：string
Required: Yes
##### password
DB account password
Type：string
Required: Yes
##### fileType
Export file type(csv/sql)
Type：string
Required: Yes
##### storageType
Storage type of data source（nfs/s3）
Type：string
Required: No
##### endpointUrl
s3 endpoint url
Type：string
Required: No（when choose s3 storageType）
##### bucketUrl
s3 bucket url
Type：string
Required: No（when choose s3 storageType）
##### accessKey
s3 access key
Type：string
Required: No（when choose s3 accesskey）
##### secretAccessKey
s3 SecretAccess Key
Type：string
Required: No（when choose s3 storageType）
##### comment
Comment of this import
Type：string
Required: No
##### filter
Export filter option
Type：string
Required: No
##### sql
Export fileter sql option
Type：string
Required: No
##### zipName
zipName of export zip
Type：string
Required: No
#### Responses
200 OK
```JSON
{
    "code": 0,
    "message": "OK",
    "data": {
      "RecordId": "xxxxxxx"
    }
}
```
#### Request Sample
```HTTP
curl -v -X POST "http://172.16.4.147:4116/api/v1/clusters/export" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 2b0febff-8863-46e0-9dcf-67815e61b54a" -d "{ \"clusterId\": \"AM0kcZ1qSkyE_PKal6jQmA\", \"userName\": \"root\", \"password\": \"mypassword\", \"fileType\": \"csv\", \"storageType\":\"nfs\"}"
```

### QueryDataTransport
Query transport records
#### Request URI
```HTTP
GET /clusters/transport/
```

#### Request Parameters
None
Responses
200 OK
```JSON
{
  "code":0,
  "message":"OK",
  "data":{
    "transportRecords":[
      {
        "recordId":"nyKmqQ0jSvyJ-AhCKyjNmQ",
        "clusterId":"Qbb5wzT5SDKt8wioSO9C7w",
        "transportType":"import",
        "filePath":"s3://nfs/em/export5G?\u0026endpoint=http://minio.pingcap.net:9000",
        "zipName":"data.zip",
        "storageType":"s3",
        "comment":"",
        "status":"Finished",
        "startTime":"2022-04-06T20:46:21.930278284+08:00",
        "endTime":"2022-04-06T20:55:58.075322535+08:00",
        "createTime":"2022-04-06T20:46:21.930398162+08:00",
        "updateTime":"2022-04-06T20:55:58.078725906+08:00",
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
```HTTP
curl -v -X GET "http://172.16.5.136:4116/api/v1/backups/?clusterId=test-tidb&page=1&pageSize=30" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 9771156d-b56a-4e32-83e1-db3203d1d682"
```

### DeleteDataTransport
Query transport records
#### Request URI
```HTTP
DELETE /clusters/transport/[recordId]
```

#### Request Parameters
None
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
curl -v -X DELETE "http://172.16.4.147:4116/api/v1/clusters/transport/_anzJTgmTbGk7cO_Pm_mAA" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 2b0febff-8863-46e0-9dcf-67815e61b54a"
```
