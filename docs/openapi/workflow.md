## Workflow
### QueryWorkflows
Query workflow list
#### Request URI
```HTTP
GET /workflow
```
#### Request Parameters
None
Responses
200 OK
```HTTP
{
    "code":0,
    "message":"OK",
    "data":{
        "workFlows":[
            {
                "id":"fXgXFrNVRv2GLE0vYJvsYw",
                "name":"BackupCluster",
                "bizId":"oEcox8lMQ7KMswDebRwIZg",
                "bizType":"cluster",
                "status":"Finished",
                "createTime":"2022-04-28T16:00:41.700877107+08:00",
                "updateTime":"2022-04-28T16:00:42.278458882+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            },
            {
                "id":"afzkygyoQlSGKwZBqbxqTA",
                "name":"CheckPlatform",
                "bizId":"8N78RjBvSZ6KVofkdw_AIg",
                "bizType":"checkReport",
                "status":"Finished",
                "createTime":"2022-04-27T23:00:00.014602876+08:00",
                "updateTime":"2022-04-27T23:00:02.459227573+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            },
            {
                "id":"6tE95RwiQs-LNKMr7H0uyQ",
                "name":"CheckPlatform",
                "bizId":"rXot_sb3SnCyqQPZETKkMA",
                "bizType":"checkReport",
                "status":"Finished",
                "createTime":"2022-04-26T23:00:00.023465673+08:00",
                "updateTime":"2022-04-26T23:00:02.707831158+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            },
            {
                "id":"3ExEgLkzTkOFZMPKVc0C1w",
                "name":"CheckPlatform",
                "bizId":"8levKkQCRoSjhTxN_53k1g",
                "bizType":"checkReport",
                "status":"Finished",
                "createTime":"2022-04-25T23:00:00.023215066+08:00",
                "updateTime":"2022-04-25T23:00:02.744281982+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            },
            {
                "id":"24iPOazYTeqDFrquRazujg",
                "name":"CheckPlatform",
                "bizId":"URWhdLUURNyzO_R5MVas-Q",
                "bizType":"checkReport",
                "status":"Finished",
                "createTime":"2022-04-24T23:00:00.02769702+08:00",
                "updateTime":"2022-04-24T23:00:02.921806068+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            },
            {
                "id":"R1fJDiy6TyyEgAFMMl8ILA",
                "name":"CheckPlatform",
                "bizId":"hqtcwBy-R0qqaWQr6NHpJQ",
                "bizType":"checkReport",
                "status":"Finished",
                "createTime":"2022-04-23T23:00:00.01499896+08:00",
                "updateTime":"2022-04-23T23:00:03.978269044+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            },
            {
                "id":"1h-GokaKQf2QDL_SLmv52Q",
                "name":"CheckPlatform",
                "bizId":"HTxHYkbwQHig30FlmsP7PA",
                "bizType":"checkReport",
                "status":"Finished",
                "createTime":"2022-04-22T23:00:00.039406449+08:00",
                "updateTime":"2022-04-22T23:00:03.277873906+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            },
            {
                "id":"Qz-8TRmSTQivHec0xrfBPw",
                "name":"CheckPlatform",
                "bizId":"NGXpqUQARfW8wMKJdwheMQ",
                "bizType":"checkReport",
                "status":"Finished",
                "createTime":"2022-04-21T23:00:00.029002578+08:00",
                "updateTime":"2022-04-21T23:00:03.11685909+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            },
            {
                "id":"2s0GtcmXQgCmj6cEPuVH2w",
                "name":"CheckPlatform",
                "bizId":"ngWaM31_Qdmsf6Aeje_ENw",
                "bizType":"checkReport",
                "status":"Finished",
                "createTime":"2022-04-20T23:00:00.025229372+08:00",
                "updateTime":"2022-04-20T23:00:03.984810883+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            },
            {
                "id":"t5Q_bvRrTpGPL5VMkzCEIQ",
                "name":"CheckPlatform",
                "bizId":"ucN4uRUDSpOiNxkSLDViEQ",
                "bizType":"checkReport",
                "status":"Finished",
                "createTime":"2022-04-19T23:00:00.010003889+08:00",
                "updateTime":"2022-04-19T23:00:02.18091762+08:00",
                "deleteTime":"0001-01-01T00:00:00Z"
            }
        ]
    },
    "page":{
        "page":1,
        "pageSize":10,
        "total":162
    }
}
```

#### Request Sample
```HTTP
curl -v -X GET "http://172.16.4.147:5000/api/v1/workflow/?page=1&pageSize=10" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 9771156d-b56a-4e32-83e1-db3203d1d682"
```
### DetailWorkflow
Query workflow detail
#### Request URI
```HTTP
GET /workflow/[workflowId]
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
    "info":{
      "id":"fXgXFrNVRv2GLE0vYJvsYw",
      "name":"BackupCluster",
      "bizId":"oEcox8lMQ7KMswDebRwIZg",
      "bizType":"cluster",
      "status":"Finished",
      "createTime":"2022-04-28T16:00:41.700877107+08:00",
      "updateTime":"2022-04-28T16:00:42.278458882+08:00",
      "deleteTime":"0001-01-01T00:00:00Z"
    },
    "nodes":[
      {
        "id":"ZoSWpolYR9OugKVjUetDag",
        "name":"backup",
        "parameters":"",
        "result":"get cluster oEcox8lMQ7KMswDebRwIZg tidb address: 172.16.5.148:10002 \nconvert storage type: s3 \nbackup cluster oEcox8lMQ7KMswDebRwIZg \nCompleted.\n",
        "status":"Finished",
        "startTime":"2022-04-28T16:00:41.732208493+08:00",
        "endTime":"2022-04-28T16:00:42.256659781+08:00"
      },
      {
        "id":"07LWNs-WSQGADeXpgu2zfQ",
        "name":"updateBackupRecord",
        "parameters":"",
        "result":"update backup record h6IhUQh-Q32MDOfcVieYgA of cluster oEcox8lMQ7KMswDebRwIZg \nCompleted.\n",
        "status":"Finished",
        "startTime":"2022-04-28T16:00:42.256696813+08:00",
        "endTime":"2022-04-28T16:00:42.269754982+08:00"
      },
      {
        "id":"Y1Nk8KdpTiujsMoG6H3MsA",
        "name":"end",
        "parameters":"",
        "result":"Completed.\n",
        "status":"Finished",
        "startTime":"2022-04-28T16:00:42.26977709+08:00",
        "endTime":"2022-04-28T16:00:42.277849774+08:00"
      }
    ],
    "nodeNames":[
      "backup",
      "updateBackupRecord",
      "end"
    ]
  }
}
```
#### Request Sample
```HTTP
curl -v -X GET "http://172.16.4.147:5000/api/v1/workflow/xxxxxxx" -H "accept: application/json" -H "Content-Type: application/json" -H "Authorization: Bearer 9771156d-b56a-4e32-83e1-db3203d1d682" 
```
