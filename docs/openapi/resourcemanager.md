# ResourceManager API 
- Author(s): [Jarivs Zheng](http://github.com/jiayang-zheng)

## Table of Contents
- [ResourceManager API](#resourcemanager-api)
  - [Table of Contents](#table-of-contents)
  - [Download Host Template File](#download-host-template-file)
    - [Request URI](#request-uri)
    - [Request Parameters](#request-parameters)
    - [Responses](#responses)
    - [Request Sample](#request-sample)
  - [Import Hosts](#import-hosts)
    - [Request URI](#request-uri-1)
    - [Request Parameters](#request-parameters-1)
      - [hostReserved](#hostreserved)
      - [skipHostInit](#skiphostinit)
      - [ignorewarns](#ignorewarns)
      - [file](#file)
    - [Responses](#responses-1)
    - [Request Sample](#request-sample-1)
  - [Query Hosts](#query-hosts)
    - [Request URI](#request-uri-2)
    - [Request Parameters](#request-parameters-2)
      - [arch](#arch)
      - [region](#region)
      - [zone](#zone)
      - [rack](#rack)
      - [hostIp](#hostip)
      - [hostName](#hostname)
      - [hostId](#hostid)
      - [status](#status)
      - [loadStat](#loadstat)
      - [purpose](#purpose)
      - [clusterType](#clustertype)
      - [hostDiskType](#hostdisktype)
      - [pageSize](#pagesize)
      - [page](#page)
    - [Responses](#responses-2)
    - [Request Sample](#request-sample-2)
  - [Update Host Info](#update-host-info)
    - [Request URI](#request-uri-3)
    - [Request Parameters](#request-parameters-3)
      - [updateReq](#updatereq)
    - [Responses](#responses-3)
    - [Request Sample](#request-sample-3)
  - [Update Host Status](#update-host-status)
    - [Request URI](#request-uri-4)
    - [Request Parameters](#request-parameters-4)
      - [updateReq](#updatereq-1)
    - [Responses](#responses-4)
    - [Request Sample](#request-sample-4)
  - [Update Host Reserved](#update-host-reserved)
    - [Request URI](#request-uri-5)
    - [Request Parameters](#request-parameters-5)
      - [updateReq](#updatereq-2)
    - [Responses](#responses-5)
    - [Request Sample](#request-sample-5)
  - [Delete Hosts](#delete-hosts)
    - [Request URI](#request-uri-6)
    - [Request Parameters](#request-parameters-6)
      - [hostIds](#hostids)
    - [Responses](#responses-6)
    - [Request Sample](#request-sample-6)
  - [Add Disks](#add-disks)
    - [Request URI](#request-uri-7)
    - [Request Parameters](#request-parameters-7)
      - [createDisksReq](#createdisksreq)
    - [Responses](#responses-7)
    - [Request Sample](#request-sample-7)
  - [Update Disk Info](#update-disk-info)
    - [Request URI](#request-uri-8)
    - [Request Parameters](#request-parameters-8)
      - [updateReq](#updatereq-3)
    - [Responses](#responses-8)
    - [Request Sample](#request-sample-8)
  - [Delete Disks](#delete-disks)
    - [Request URI](#request-uri-9)
    - [Request Parameters](#request-parameters-9)
      - [diskIds](#diskids)
    - [Responses](#responses-9)
    - [Request Sample](#request-sample-9)
  - [Get Resource Hierarchy](#get-resource-hierarchy)
    - [Request URI](#request-uri-10)
    - [Request Parameters](#request-parameters-10)
      - [arch](#arch-1)
      - [level](#level)
      - [depth](#depth)
      - [status](#status-1)
      - [loadStat](#loadstat-1)
      - [purpose](#purpose-1)
      - [clusterType](#clustertype-1)
      - [hostDiskType](#hostdisktype-1)
    - [Responses](#responses-10)
    - [Request Sample](#request-sample-10)
  - [Get Resource Stocks](#get-resource-stocks)
    - [Request URI](#request-uri-11)
    - [Request Parameters](#request-parameters-11)
      - [arch](#arch-2)
      - [region](#region-1)
      - [zone](#zone-1)
      - [rack](#rack-1)
      - [hostIp](#hostip-1)
      - [hostName](#hostname-1)
      - [hostId](#hostid-1)
      - [status](#status-2)
      - [loadStat](#loadstat-2)
      - [purpose](#purpose-2)
      - [clusterType](#clustertype-2)
      - [hostDiskType](#hostdisktype-2)
      - [diskStatus](#diskstatus)
      - [capacity](#capacity)
    - [Responses](#responses-11)
    - [Request Sample](#request-sample-11)

## Download Host Template File
Download the host template file for importing.

### Request URI
```
GET resources/hosts-template
```

### Request Parameters 
No parameters. 

### Responses
200 OK
Return the template file(*.xlsx) in the response body.

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
  'http://172.16.6.252:4100/api/v1/resources/hosts-template' \
  -H 'accept: application/octet-stream' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29'
```

## Import Hosts 
Import hosts to the TiUniManager platform with a hosts xlsx file. 

### Request URI
```
POST resources/hosts
```

### Request Parameters 
#### hostReserved 
The `hostReserved` is used to indicate whether the batch of hosts is reserved (only set to true while importing hosts for a take-over cluster).   
Type：string   
Required: No  
#### skipHostInit 
The `skipHostInit` is used for skipping the host initialization while importing.  
Type：string  
Required: No  
#### ignorewarns 
The `ignorewarns` indicates whether to ignore warnings while importing.  
Type：string  
Required: No    
#### file
The `file` is the xlsx file containing the hosts to be imported.  
Type：file  
Required: Yes   

### Responses
200 OK
``` json
{
  "code": 0,
  "data": {
    "flowInfo": [
      {
        "workFlowId": "string"
      }
    ],
    "hostIds": [
      "string"
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
curl -X 'POST' \
  'http://172.16.6.252:4100/api/v1/resources/hosts' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29' \
  -H 'Content-Type: multipart/form-data' \
  -F 'hostReserved=false' \
  -F 'skipHostInit=false' \
  -F 'ignorewarns=false' \
  -F 'file=@hostInfo.xlsx;type=application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
```

## Query Hosts 
Query host(s) by filter conditions. 

### Request URI 
```
Get resources/hosts
```

### Request Parameters 
#### arch 
Filter hosts by arch("X86" / "X86_64" / "ARM" / "ARM64").   
Type：string   
Required: No  
#### region 
Filter hosts by region.  
Type：string  
Required: No   
#### zone
Filter hosts by zone. If `region` not specified, `zone` does not work.  
Type：string  
Required: No   
#### rack
Filter hosts by rack. If `region` and `zone` are not specified,  `rack` does not work.   
Type：string  
Required: No   
#### hostIp
Filter hosts by host IP.  
Type：string  
Required: No   
#### hostName
Filter hosts by host name.  
Type：string  
Required: No   
#### hostId
Filter hosts by host ID.  
Type：string  
Required: No   
#### status
Filter hosts by status ("Init" / "Online" / "Offline"/ "Failed" / "Deleting" / "Deleted").   
Type：string  
Required: No   
#### loadStat
Filter hosts by host load stat ("LoadLess" / "Inused" / "Exclusive").  
Type：string  
Required: No   
#### purpose
Filter hosts by host purpose ("Compute" / "Schedule" / "Storage").  
Type：string  
Required: No   
#### clusterType
Filter hosts by cluster type of host ("TiDB" / "DataMigration" / "EnterpriseManager").  
Type：string  
Required: No   
#### hostDiskType
Filter hosts by host disk type ("NVMeSSD" / "SSD" / "SATA").    
Type：string  
Required: No   
#### pageSize
Limit the number of this request.  
Type：int  
Required: No   
#### page
Current page location.  
Type：string  
Required: No   

### Responses
200 OK  
``` json
{
  "code": 0,
  "data": {
    "hosts": [
      {
        "arch": "string",
        "availableDiskCount": 0,
        "az": "string",
        "clusterType": "string",
        "cpuCores": 0,
        "createTime": 0,
        "diskType": "string",
        "disks": [
          {
            "capacity": 0,
            "diskId": "string",
            "hostId": "string",
            "name": "string",
            "path": "string",
            "status": "string",
            "type": "string"
          }
        ],
        "hostId": "string",
        "hostName": "string",
        "instances": {
          "additionalProp1": [
            "string"
          ],
          "additionalProp2": [
            "string"
          ],
          "additionalProp3": [
            "string"
          ]
        },
        "ip": "string",
        "kernel": "string",
        "loadStat": "string",
        "memory": 0,
        "nic": "string",
        "os": "string",
        "passwd": "string",
        "purpose": "string",
        "rack": "string",
        "region": "string",
        "reserved": true,
        "spec": "string",
        "sshPort": 0,
        "status": "string",
        "sysLabels": [
          "string"
        ],
        "traits": 0,
        "updateTime": 0,
        "usedCpuCores": 0,
        "usedMemory": 0,
        "userName": "string",
        "vendor": "string"
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
curl -X 'GET' \
  'http://172.16.6.252:4100/api/v1/resources/hosts?arch=X86_64&clusterType=TiDB&region=Region1&zone=Zone1' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29'
```

## Update Host Info 
Update host information. 

### Request URI 
```
PUT resources/host
```

### Request Parameters 
#### updateReq
The `updateReq` is the new host info to be updated.  
Type：json in body  
Required: Yes   
``` json
{
  "newHostInfo": {
    "hostId": "string",
    "hostName": "string",
    "userName": "string",
    "passwd": "string",
    "cpuCores": 0,
    "memory": 0,
    "kernel": "string",
    "os": "string", 
    "nic": "string",
    "purpose": "string"
  }
}
```

### Responses
200 OK  
``` json
{
  "code": 0,
  "message": "OK",
  "data": {}
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
curl -X 'PUT' \
  'http://172.16.6.252:4100/api/v1/resources/host' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29' \
  -H 'Content-Type: application/json' \
  -d '{
  "newHostInfo": {
    "hostId": "ykw1WxMHT0WF63GA3JOWpw",
    "hostName": "NewHostName",
    "cpuCores": 128,
    "memory": 256,
    "os": "Ubuntu",
    "purpose": "Schedule"
  }
}'
```

## Update Host Status 
Update Host Status by hostIDs. 

### Request URI 
```
PUT resources/host-status
```

### Request Parameters 
#### updateReq
The `updateReq` is the new status to be updated.  
Type：json in body  
Required: Yes   
``` json
{
  "hostIds": [
    "string"
  ],
  "status": "string"
}
```

### Responses
200 OK  
``` json
{
  "code": 0,
  "message": "OK",
  "data": {}
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
curl -X 'PUT' \
  'http://172.16.6.252:4100/api/v1/resources/host-status' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29' \
  -H 'Content-Type: application/json' \
  -d '{
  "hostIds": [
    "ykw1WxMHT0WF63GA3JOWpw",
    "5dN7bs4AQbGC3AFd2lAW0w"
  ],
  "status": "Offline"
}'
```

## Update Host Reserved
Update Host's Reserved field by hostIDs. 

### Request URI 
```
PUT resources/host-reserved
```

### Request Parameters 
#### updateReq
The `updateReq` is the new reserved status to be updated.  
Type：json in body  
Required: Yes   
``` json
{
  "hostIds": [
    "string"
  ],
  "reserved": true
}
```

### Responses
200 OK  
``` json
{
  "code": 0,
  "message": "OK",
  "data": {}
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
curl -X 'PUT' \
  'http://172.16.6.252:4100/api/v1/resources/host-reserved' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29' \
  -H 'Content-Type: application/json' \
  -d '{
  "hostIds": [
    "ykw1WxMHT0WF63GA3JOWpw",
    "5dN7bs4AQbGC3AFd2lAW0w"
  ],
  "reserved": false
}'
```

## Delete Hosts 
Delete a batch of hosts by hostIDs. 

### Request URI 
```
DELETE resources/hosts
```

### Request Parameters 
#### hostIds
The `hostIds` contains the hosts' ID to be deleted and whether to delete the hosts by force.   
If force field is true, only move out the hosts' information in meta database. Otherwise, the deletion workflow will do some clean work before deleting hosts' info from meta database.   
Type：json in body  
Required: Yes   
``` json
{
  "force": true,
  "hostIds": [
    "string"
  ]
}
```

### Responses
200 OK
``` json
{
  "code": 0,
  "data": {
    "flowInfo": [
      {
        "workFlowId": "string"
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
curl -X 'DELETE' \
  'http://172.16.6.252:4100/api/v1/resources/hosts' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29' \
  -H 'Content-Type: application/json' \
  -d '{
  "force": false,
  "hostIds": [
    "5dN7bs4AQbGC3AFd2lAW0w"
  ]
}'
```

## Add Disks 
Add disks for a specified host. 

### Request URI 
```
POST resources/disks
```

### Request Parameters 
#### createDisksReq
The `createDisksReq` is disk info to be added.  
Type：json in body  
Required: Yes   
``` json
{
  "disks": [
    {
      "capacity": int,
      "name": "string",
      "path": "string"
    }
  ],
  "hostId": "string"
}
```

### Responses
200 OK
``` json
{
  "code": 0,
  "data": {
    "diskIds": [
      "string"
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
curl -X 'POST' \
  'http://172.16.6.252:4100/api/v1/resources/disks' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29' \
  -H 'Content-Type: application/json' \
  -d '{
  "disks": [
    {
      "capacity": 1024,
      "name": "sdc",
      "path": "/mnt/sdc"
    }
  ],
  "hostId": "ykw1WxMHT0WF63GA3JOWpw"
}'
```

## Update Disk Info 
Update disk info by DiskID. 

### Request URI 
```
PUT resources/disk
```

### Request Parameters 
#### updateReq
The `updateReq` is the new disk info to be updated.  
Type：json in body  
Required: Yes   
``` json
{
  "newDiskInfo": {
    "diskId": "string",
    "capacity": int,
    "name": "string",
    "status": "string"
  }
}
```

### Responses
200 OK
``` json
{
  "code": 0,
  "message": "OK",
  "data": {}
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
curl -X 'PUT' \
  'http://172.16.6.252:4100/api/v1/resources/disk' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29' \
  -H 'Content-Type: application/json' \
  -d '{
  "newDiskInfo": {
    "diskId": "dTo8Gg4PQdS4kwI11Rlwtg",
    "capacity": 256,
    "name": "sdk",
    "status": "Error"
  }
}'
```

## Delete Disks 
Delete disks by DiskIDs.

### Request URI 
```
DELETE resources/disks
```

### Request Parameters 
#### diskIds
The `diskIds` is a batch of disks to be deleted.  
Type：json in body  
Required: Yes  
``` json 
{
  "diskIds": [
    "string"
  ]
}
```

### Responses
200 OK
``` json
{
  "code": 0,
  "message": "OK",
  "data": {}
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
curl -X 'DELETE' \
  'http://172.16.6.252:4100/api/v1/resources/disks' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29' \
  -H 'Content-Type: application/json' \
  -d '{
  "diskIds": [
    "dTo8Gg4PQdS4kwI11Rlwtg"
  ]
}'
```

## Get Resource Hierarchy
Get the hierarchy tree of  resources. 

### Request URI 
```
GET resources/hierarchy
```

### Request Parameters 
#### arch 
Filter hosts by arch("X86" / "X86_64" / "ARM" / "ARM64").   
Type：string   
Required: No  
#### level
The first hierarchy level (1:Region, 2:Zone, 3:Rack, 4:Host).  
Type：interger   
Required: No  
#### depth
The depth of the hierarchy tree ( level+depth <= 4:Host ).  
Type：interger   
Required: No  
#### status
Filter hosts by status ("Init" / "Online" / "Offline"/ "Failed" / "Deleting" / "Deleted").  
Type：string  
Required: No   
#### loadStat
Filter hosts by host load stat ("LoadLess" / "Inused" / "Exclusive").  
Type：string  
Required: No   
#### purpose
Filter hosts by host purpose ("Compute" / "Schedule" / "Storage").  
Type：string  
Required: No   
#### clusterType
Filter hosts by cluster type of host ("TiDB" / "DataMigration" / "EnterpriseManager").  
Type：string  
Required: No   
#### hostDiskType
Filter hosts by host disk type ("NVMeSSD" / "SSD" / "SATA").  
Type：string  
Required: No   

### Responses
200 OK
``` json
{
  "code": 0,
  "data": {
    "root": {
      "code": "string",
      "name": "string",
      "prefix": "string",
      "subNodes": [
        "string"
      ]
    }
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
curl -X 'GET' \
  'http://172.16.6.252:4100/api/v1/resources/hierarchy?Depth=1&Level=1&arch=X86_64' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29'
```

## Get Resource Stocks 
Get the stocks of resources after filtering. 

### Request URI 
```
GET resources/stocks
```

### Request Parameters 
#### arch 
Filter hosts by arch("X86" / "X86_64" / "ARM" / "ARM64").   
Type：string   
Required: No  
#### region 
Filter hosts by region.  
Type：string  
Required: No   
#### zone
Filter hosts by zone. If `region` not specified, `zone` does not work.  
Type：string  
Required: No   
#### rack
Filter hosts by rack. If `region` and `zone` are not specified,  `rack` does not work.  
Type：string  
Required: No   
#### hostIp
Filter hosts by host IP.  
Type：string  
Required: No   
#### hostName
Filter hosts by host name.  
Type：string  
Required: No   
#### hostId
Filter hosts by host ID.  
Type：string  
Required: No   
#### status
Filter hosts by status ("Init" / "Online" / "Offline"/ "Failed" / "Deleting" / "Deleted").  
Type：string  
Required: No   
#### loadStat
Filter hosts by host load stat ("LoadLess" / "Inused" / "Exclusive").  
Type：string  
Required: No   
#### purpose
Filter hosts by host purpose ("Compute" / "Schedule" / "Storage").  
Type：string  
Required: No   
#### clusterType
Filter hosts by cluster type of host ("TiDB" / "DataMigration" / "EnterpriseManager").  
Type：string  
Required: No   
#### hostDiskType
Filter hosts by host disk type ("NVMeSSD" / "SSD" / "SATA").  
Type：string  
Required: No  
#### diskStatus
Filter disks by disk status ("Available" / "Reserved" / "Exhaust" / "Error").  
Type：string  
Required: No  
#### capacity
Filter disks' capacity which  no less than `capacity`.  
Type：interger  
Required: No  

### Responses
200 OK
``` json
{
  "code": 0,
  "data": {
    "stocks": {
      "additionalProp1": {
        "freeCpuCores": 0,
        "freeDiskCapacity": 0,
        "freeDiskCount": 0,
        "freeHostCount": 0,
        "freeMemory": 0,
        "zone": "string"
      },
      "additionalProp2": {
        "freeCpuCores": 0,
        "freeDiskCapacity": 0,
        "freeDiskCount": 0,
        "freeHostCount": 0,
        "freeMemory": 0,
        "zone": "string"
      },
      "additionalProp3": {
        "freeCpuCores": 0,
        "freeDiskCapacity": 0,
        "freeDiskCount": 0,
        "freeHostCount": 0,
        "freeMemory": 0,
        "zone": "string"
      }
    }
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
curl -X 'GET' \
  'http://172.16.6.252:4100/api/v1/resources/stocks?arch=X86_64&loadStat=LoadLess&region=Region1&zone=Zone1_1' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer a1cd2093-a11b-4c6f-bde0-72a0e852ef29'
```
