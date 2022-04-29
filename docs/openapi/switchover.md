## Switchover

The current implementation mainly includes three functional parts: master-slave clusters' status check, master-slave switchover, and rollback on failed workflow.

### Request URI

```HTTP
POST /api/v1/clusters/switchover
```

### Request Parameters

#### sourceClusterId

Source cluster ID.

Type：String

Required: Yes

#### targetClusterId

Target cluster ID.

Type：String

Required: Yes

#### force

Whether to perform a forced switchover.

Type：Bool

Required: No

#### onlyCheck

Whether to only perform the status pre-check of master/slave clusters. That is, if `onlyCheck` is true, it will return immediately after the pre-check, and the actual master/slave switchover will not be executed.

Type：Bool

Required: No

#### checkSlaveReadOnlyFlag

Whether to check the slave cluster is read-only.

Type：Bool

Required: No

#### checkMasterWritableFlag

Whether to check the master cluster is writeable.

Type：Bool

Required: No

#### checkStandaloneClusterFlag

Whether to check the cluster is standalone, i.e. neither a slave nor a master.

Type：Bool

Required: No

#### rollbackWorkFlowID

The ID of the workflow to perform rollback on. Empty indicates this request is not an rollback request.

Type：String

Required: No

#### rollbackClearPreviousMaintenanceFlag

Whether to clear the "Switching" cluster maintenance status which is possibly left by previous failed switchover workflow.

Type：Bool

Required: No

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

Switchover：

```Shell
curl -vX 'POST' \
  'http://$IP:$PORT/api/v1/clusters/switchover' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
   "sourceClusterId" : "NK66Ks_lRAaHg8r1FRBpQg",
   "targetClusterID" : "A_B1lr0BQCCg9y5cg4Jedg"
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

Master-slave clusters' status check：

```Shell
curl -vX 'POST' \
  'http://$IP:$PORT/api/v1/clusters/switchover' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
   "sourceClusterId" : "NK66Ks_lRAaHg8r1FRBpQg",
   "targetClusterID" : "A_B1lr0BQCCg9y5cg4Jedg",
   "onlyCheck": true,
   "checkSlaveReadOnlyFlag": true,
   "checkMasterWritableFlag": true
}'
```

Successful response (no new workflow is triggered):

```bash
{
  "code": 0,
  "data": {
    "workflowID": "",
  },
  "message": "OK"
}
```

Rollback on failed switchover workflow：

```Shell
curl -vX 'POST' \
  'http://$IP:$PORT/api/v1/clusters/switchover' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer $TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
   "sourceClusterId" : "NK66Ks_lRAaHg8r1FRBpQg",
   "targetClusterID" : "A_B1lr0BQCCg9y5cg4Jedg",
   "rollbackWorkFlowID":"$FailedWorkFlowID"
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

