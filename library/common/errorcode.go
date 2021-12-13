/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package common

type TIEM_ERROR_CODE int32

// all tiem error code
const (
	TIEM_SUCCESS TIEM_ERROR_CODE = 0

	// common, client
	TIEM_PARAMETER_INVALID TIEM_ERROR_CODE = 10001

	// common, system
	TIEM_MARSHAL_ERROR             TIEM_ERROR_CODE = 10101
	TIEM_UNMARSHAL_ERROR           TIEM_ERROR_CODE = 10102
	TIEM_UNRECOGNIZED_ERROR        TIEM_ERROR_CODE = 10103
	TIEM_CLUSTER_SERVER_CALL_ERROR TIEM_ERROR_CODE = 10104
	TIEM_METADB_SERVER_CALL_ERROR  TIEM_ERROR_CODE = 10105
	TIEM_TIMEOUT                   TIEM_ERROR_CODE = 10106

	// cluster, common
	TIEM_CLUSTER_NOT_FOUND            TIEM_ERROR_CODE = 20001
	TIEM_CLUSTER_META_BROKEN          TIEM_ERROR_CODE = 20002
	TIEM_CLUSTER_MAINTENANCE_CONFLICT TIEM_ERROR_CODE = 20003

	// cluster, create
	TIEM_CLUSTER_DUPLICATED_NAME           TIEM_ERROR_CODE = 20101
	TIEM_CLUSTER_INVALIE_TOPOLOGY          TIEM_ERROR_CODE = 20102
	TIEM_CLUSTER_UNSUPPORTED_PRODUCT       TIEM_ERROR_CODE = 20103
	TIEM_CLUSTER_RESOURCE_NOT_ENOUGH       TIEM_ERROR_CODE = 20104
	TIEM_CLUSTER_HOST_NOT_REACHABLE        TIEM_ERROR_CODE = 20105
	TIEM_CLUSTER_FAILED_TO_DEPLOY          TIEM_ERROR_CODE = 20106
	TIEM_CLUSTER_FAILED_TO_START           TIEM_ERROR_CODE = 20107
	TIEM_CLUSTER_NOT_ALIGNED_WITH_SETTINGS TIEM_ERROR_CODE = 20108
	TIEM_CLUSTER_FAILED_TO_INITIATE        TIEM_ERROR_CODE = 20109

	// cluster, takeover
	TIEM_TAKEOVER_NOT_REACHABLE          TIEM_ERROR_CODE = 20201
	TIEM_TAKEOVER_INCORRECT_ACCOUNT      TIEM_ERROR_CODE = 20202
	TIEM_TAKEOVER_INCORRECT_PATH         TIEM_ERROR_CODE = 20203
	TIEM_TAKEOVER_INCORRECT_CLUSTER_NAME TIEM_ERROR_CODE = 20204
	TIEM_TAKEOVER_ALREADY_MANAGED        TIEM_ERROR_CODE = 20205
	TIEM_TAKEOVER_HOST_NOT_MANAGED       TIEM_ERROR_CODE = 20206
	TIEM_TAKEOVER_HOST_NOT_REACHABLE     TIEM_ERROR_CODE = 20207
	TIEM_TAKEOVER_FAILED_TO_TAKEOVER     TIEM_ERROR_CODE = 20208

	// cluster, restart, 203xx
	// cluster, destroy, 204xx
	// cluster, parameter, 205xx
	// cluster, backup, 206xx
	// cluster, recover, 207xx
	// cluster, scale out, 208xx
	// cluster, scale in, 209xx
	// cluster, stand by, 210xx
	// cluster, upgrade, 211xx
	// cluster, change feed, 212xx

	TIEM_CHANGE_FEED_NOT_FOUND              TIEM_ERROR_CODE = 21201
	TIEM_CHANGE_FEED_DUPLICATE_TASK         TIEM_ERROR_CODE = 21202
	TIEM_CHANGE_FEED_STATUS_CONFLICT        TIEM_ERROR_CODE = 21203
	TIEM_CHANGE_FEED_LOCK_EXPIRED           TIEM_ERROR_CODE = 21204
	TIEM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM TIEM_ERROR_CODE = 21205
	TIEM_CHANGE_FEED_CREATE_ERROR           TIEM_ERROR_CODE = 21206
	TIEM_CHANGE_FEED_CONNECT_ERROR          TIEM_ERROR_CODE = 21207

	// resource, import, 301xx
	// resource, management, 302xx
	// resource, stock, 303xx

	// workflow, 4xxxx
	// user, 5xxxx

	// data transfer, import, 601xx
	// data transfer, export, 602xx
	// data transfer, migration, 603xx

	// cluster diagnose, monitor, 801xx
	// cluster diagnose, alert, 802xx
	// cluster diagnose, log, 803xx

	// platform, 9xxxx

	TIEM_TASK_TIMEOUT          TIEM_ERROR_CODE = 9997
	TIEM_FLOW_NOT_FOUND        TIEM_ERROR_CODE = 9996
	TIEM_TASK_FAILED           TIEM_ERROR_CODE = 9995
	TIEM_TASK_POLLING_TIME_OUT TIEM_ERROR_CODE = 9994
	TIEM_TASK_CONFLICT         TIEM_ERROR_CODE = 9993
	TIEM_TASK_CANCELED         TIEM_ERROR_CODE = 9992

	TIEM_ACCOUNT_NOT_FOUND       TIEM_ERROR_CODE = 100
	TIEM_TENANT_NOT_FOUND        TIEM_ERROR_CODE = 200
	TIEM_QUERY_PERMISSION_FAILED TIEM_ERROR_CODE = 300
	TIEM_ADD_TOKEN_FAILED        TIEM_ERROR_CODE = 400
	TIEM_TOKEN_NOT_FOUND         TIEM_ERROR_CODE = 401

	TIEM_RESOURCE_SQL_ERROR                        TIEM_ERROR_CODE = 500
	TIEM_RESOURCE_HOST_NOT_FOUND                   TIEM_ERROR_CODE = 501
	TIEM_RESOURCE_NO_ENOUGH_HOST                   TIEM_ERROR_CODE = 502
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED    TIEM_ERROR_CODE = 503
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER TIEM_ERROR_CODE = 504
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_HOST_FILTER TIEM_ERROR_CODE = 505
	TIEM_RESOURCE_NO_ENOUGH_PORT                   TIEM_ERROR_CODE = 506
	TIEM_RESOURCE_NOT_ALL_SUCCEED                  TIEM_ERROR_CODE = 507
	TIEM_RESOURCE_INVALID_STRATEGY                 TIEM_ERROR_CODE = 508
	TIEM_RESOURCE_INVAILD_RECYCLE_TYPE             TIEM_ERROR_CODE = 509
	TIEM_UPDATE_HOST_STATUS_FAIL                   TIEM_ERROR_CODE = 510
	TIEM_RESERVE_HOST_FAIL                         TIEM_ERROR_CODE = 511
	TIEM_RESOURCE_NO_STOCK                         TIEM_ERROR_CODE = 512
	TIEM_RESOURCE_GET_DISK_ID_FAIL                 TIEM_ERROR_CODE = 513
	TIEM_RESOURCE_TRAIT_NOT_FOUND                  TIEM_ERROR_CODE = 514
	TIEM_RESOURCE_INVALID_LOCATION                 TIEM_ERROR_CODE = 515
	TIEM_RESOURCE_INVALID_ARCH                     TIEM_ERROR_CODE = 516

	TIEM_DASHBOARD_NOT_FOUND           TIEM_ERROR_CODE = 600
	TIEM_EXPORT_PARAM_INVALID          TIEM_ERROR_CODE = 601
	TIEM_EXPORT_PROCESS_FAILED         TIEM_ERROR_CODE = 602
	TIEM_IMPORT_PARAM_INVALID          TIEM_ERROR_CODE = 603
	TIEM_IMPORT_PROCESS_FAILED         TIEM_ERROR_CODE = 604
	TIEM_TRANSPORT_RECORD_NOT_FOUND    TIEM_ERROR_CODE = 605
	TIEM_BACKUP_PROCESS_FAILED         TIEM_ERROR_CODE = 606
	TIEM_RECOVER_PARAM_INVALID         TIEM_ERROR_CODE = 607
	TIEM_RECOVER_PROCESS_FAILED        TIEM_ERROR_CODE = 608
	TIEM_BACKUP_RECORD_DELETE_FAILED   TIEM_ERROR_CODE = 609
	TIEM_BACKUP_RECORD_QUERY_FAILED    TIEM_ERROR_CODE = 610
	TIEM_BACKUP_STRATEGY_PARAM_INVALID TIEM_ERROR_CODE = 611
	TIEM_BACKUP_STRATEGY_SAVE_FAILED   TIEM_ERROR_CODE = 612
	TIEM_BACKUP_STRATEGY_QUERY_FAILED  TIEM_ERROR_CODE = 613
	TIEM_MONITOR_NOT_FOUND             TIEM_ERROR_CODE = 614
	TIEM_TRANSPORT_RECORD_DEL_FAILED   TIEM_ERROR_CODE = 615
	TIEM_BACKUP_RECORD_NOT_FOUND       TIEM_ERROR_CODE = 616
	TIEM_BACKUP_STRATEGY_NOT_FOUND     TIEM_ERROR_CODE = 617
	TIEM_LIST_WORKFLOW_FAILED          TIEM_ERROR_CODE = 618
	TIEM_DETAIL_WORKFLOW_FAILED        TIEM_ERROR_CODE = 619
	TIEM_GET_CONFIG_FAILED             TIEM_ERROR_CODE = 620
	TIEM_TRANSPORT_RECORD_QUERY_FAIL   TIEM_ERROR_CODE = 621

	TIEM_DEFAULT_PARAM_GROUP_NOT_DEL TIEM_ERROR_CODE = 20500
	TIEM_MODIFY_PARAM_FAILED         TIEM_ERROR_CODE = 20501
	TIEM_CONVERT_OBJ_FAILED          TIEM_ERROR_CODE = 20502
)

type ErrorCodeExplanation struct {
	code        TIEM_ERROR_CODE
	explanation string
	httpCode    int
}

func (t TIEM_ERROR_CODE) GetHttpCode() int {
	return explanationContainer[t].httpCode
}

func (t TIEM_ERROR_CODE) Equal(code int32) bool {
	return code == int32(t)
}

func (t TIEM_ERROR_CODE) Explain() string {
	return explanationContainer[t].explanation
}

var explanationContainer = map[TIEM_ERROR_CODE]ErrorCodeExplanation{
	TIEM_SUCCESS: {code: TIEM_SUCCESS, explanation: "succeed", httpCode: 200},

	// common, client error
	TIEM_PARAMETER_INVALID: {TIEM_PARAMETER_INVALID, "parameter invalid", 400},

	// common, system error
	TIEM_MARSHAL_ERROR:             {TIEM_MARSHAL_ERROR, "marshal error", 500},
	TIEM_UNMARSHAL_ERROR:           {TIEM_UNMARSHAL_ERROR, "Unmarshal error", 500},
	TIEM_UNRECOGNIZED_ERROR:        {TIEM_UNRECOGNIZED_ERROR, "unrecognized error", 500},
	TIEM_METADB_SERVER_CALL_ERROR:  {TIEM_METADB_SERVER_CALL_ERROR, "call metadb-Server failed", 500},
	TIEM_CLUSTER_SERVER_CALL_ERROR: {TIEM_CLUSTER_SERVER_CALL_ERROR, "call cluster-Server failed", 500},
	TIEM_TIMEOUT:                   {TIEM_TIMEOUT, "request timeout", 500},

	// cluster, common error
	TIEM_CLUSTER_NOT_FOUND:            {TIEM_CLUSTER_NOT_FOUND, "cluster not found", 404},
	TIEM_CLUSTER_META_BROKEN:          {TIEM_CLUSTER_META_BROKEN, "cluster meta broken", 400},
	TIEM_CLUSTER_MAINTENANCE_CONFLICT: {TIEM_CLUSTER_MAINTENANCE_CONFLICT, "cluster maintenance conflict", 400},

	// cluster, create
	TIEM_CLUSTER_DUPLICATED_NAME:           {TIEM_CLUSTER_DUPLICATED_NAME, "Duplicated cluster name", 400},
	TIEM_CLUSTER_INVALIE_TOPOLOGY:          {TIEM_CLUSTER_INVALIE_TOPOLOGY, "Invalid cluster topology", 400},
	TIEM_CLUSTER_UNSUPPORTED_PRODUCT:       {TIEM_CLUSTER_UNSUPPORTED_PRODUCT, "Unsupported cluster type or version", 400},
	TIEM_CLUSTER_RESOURCE_NOT_ENOUGH:       {TIEM_CLUSTER_RESOURCE_NOT_ENOUGH, "Host resource is not enough", 500},
	TIEM_CLUSTER_HOST_NOT_REACHABLE:        {TIEM_CLUSTER_HOST_NOT_REACHABLE, "Host is not reachable", 500},
	TIEM_CLUSTER_FAILED_TO_DEPLOY:          {TIEM_CLUSTER_FAILED_TO_DEPLOY, "Failed to deploy cluster ", 500},
	TIEM_CLUSTER_FAILED_TO_START:           {TIEM_CLUSTER_FAILED_TO_START, "Failed to start cluster", 500},
	TIEM_CLUSTER_NOT_ALIGNED_WITH_SETTINGS: {TIEM_CLUSTER_NOT_ALIGNED_WITH_SETTINGS, "Cluster topology is not aligned with settings", 500},
	TIEM_CLUSTER_FAILED_TO_INITIATE:        {TIEM_CLUSTER_FAILED_TO_INITIATE, "Failed to initiate database", 500},

	// cluster, takeover
	TIEM_TAKEOVER_NOT_REACHABLE:          {TIEM_TAKEOVER_NOT_REACHABLE, "Original control machine is not reachable", 400},
	TIEM_TAKEOVER_INCORRECT_ACCOUNT:      {TIEM_TAKEOVER_INCORRECT_ACCOUNT, "Incorrect username or password", 400},
	TIEM_TAKEOVER_INCORRECT_PATH:         {TIEM_TAKEOVER_INCORRECT_PATH, "Incorrect path of TiUP on control machine", 400},
	TIEM_TAKEOVER_INCORRECT_CLUSTER_NAME: {TIEM_TAKEOVER_INCORRECT_CLUSTER_NAME, "Cluster to take over does not exist", 400},
	TIEM_TAKEOVER_ALREADY_MANAGED:        {TIEM_TAKEOVER_ALREADY_MANAGED, "Cluster is already managed by TiEM", 409},
	TIEM_TAKEOVER_HOST_NOT_MANAGED:       {TIEM_TAKEOVER_HOST_NOT_MANAGED, "Host is not imported into resource pool yet", 409},
	TIEM_TAKEOVER_HOST_NOT_REACHABLE:     {TIEM_TAKEOVER_HOST_NOT_REACHABLE, "Host in cluster is not reachable", 500},
	TIEM_TAKEOVER_FAILED_TO_TAKEOVER:     {TIEM_TAKEOVER_FAILED_TO_TAKEOVER, "Failed to take over cluster", 500},

	// cluster, change feed
	TIEM_CHANGE_FEED_NOT_FOUND:              {TIEM_CHANGE_FEED_NOT_FOUND, "change feed task not found", 404},
	TIEM_CHANGE_FEED_DUPLICATE_TASK:         {TIEM_CHANGE_FEED_DUPLICATE_TASK, "duplicate task", 500},
	TIEM_CHANGE_FEED_STATUS_CONFLICT:        {TIEM_CHANGE_FEED_STATUS_CONFLICT, "task status conflict", 409},
	TIEM_CHANGE_FEED_LOCK_EXPIRED:           {TIEM_CHANGE_FEED_LOCK_EXPIRED, "task status lock expired", 409},
	TIEM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM: {TIEM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM, "task downstream type not supported", 400},
	TIEM_CHANGE_FEED_CREATE_ERROR:           {TIEM_CHANGE_FEED_CREATE_ERROR, "failed to create change feed task", 500},
	TIEM_CHANGE_FEED_CONNECT_ERROR:          {TIEM_CHANGE_FEED_CONNECT_ERROR, "failed to connect cdc server", 500},

	TIEM_TASK_TIMEOUT:          {TIEM_TASK_TIMEOUT, "task timeout", 500},
	TIEM_FLOW_NOT_FOUND:        {TIEM_FLOW_NOT_FOUND, "flow not found", 500},
	TIEM_TASK_FAILED:           {TIEM_TASK_FAILED, "task failed", 500},
	TIEM_TASK_CONFLICT:         {TIEM_TASK_CONFLICT, "task polling time out", 500},
	TIEM_TASK_CANCELED:         {TIEM_TASK_CONFLICT, "task canceled", 500},
	TIEM_TASK_POLLING_TIME_OUT: {TIEM_TASK_POLLING_TIME_OUT, "task polling time out", 500},

	TIEM_DASHBOARD_NOT_FOUND: {TIEM_DASHBOARD_NOT_FOUND, "dashboard is not found", 500},

	// cluster import export
	TIEM_EXPORT_PARAM_INVALID:        {TIEM_EXPORT_PARAM_INVALID, "export data param invalid", 500},
	TIEM_EXPORT_PROCESS_FAILED:       {TIEM_EXPORT_PROCESS_FAILED, "export process failed", 500},
	TIEM_IMPORT_PARAM_INVALID:        {TIEM_IMPORT_PARAM_INVALID, "import data param invalid", 500},
	TIEM_IMPORT_PROCESS_FAILED:       {TIEM_IMPORT_PROCESS_FAILED, "import process failed", 500},
	TIEM_TRANSPORT_RECORD_DEL_FAILED: {TIEM_TRANSPORT_RECORD_DEL_FAILED, "delete data transport failed", 500},

	// cluster backup
	TIEM_TRANSPORT_RECORD_NOT_FOUND:    {TIEM_TRANSPORT_RECORD_NOT_FOUND, "transport record is not found", 500},
	TIEM_BACKUP_PROCESS_FAILED:         {TIEM_BACKUP_PROCESS_FAILED, "backup process failed", 500},
	TIEM_RECOVER_PARAM_INVALID:         {TIEM_RECOVER_PARAM_INVALID, "recover param invalid", 500},
	TIEM_RECOVER_PROCESS_FAILED:        {TIEM_RECOVER_PROCESS_FAILED, "recover process failed", 500},
	TIEM_BACKUP_RECORD_DELETE_FAILED:   {TIEM_BACKUP_RECORD_DELETE_FAILED, "delete backup record failed", 500},
	TIEM_BACKUP_RECORD_QUERY_FAILED:    {TIEM_BACKUP_RECORD_QUERY_FAILED, "query backup record failed", 500},
	TIEM_BACKUP_STRATEGY_PARAM_INVALID: {TIEM_BACKUP_STRATEGY_PARAM_INVALID, "backup strategy param invalid", 500},
	TIEM_BACKUP_STRATEGY_SAVE_FAILED:   {TIEM_BACKUP_STRATEGY_SAVE_FAILED, "save backup strategy failed", 500},
	TIEM_BACKUP_STRATEGY_QUERY_FAILED:  {TIEM_BACKUP_STRATEGY_QUERY_FAILED, "query backup strategy failed", 500},

	// resource
	TIEM_RESOURCE_SQL_ERROR:                        {TIEM_RESOURCE_SQL_ERROR, "resource sql error", 500},
	TIEM_RESOURCE_HOST_NOT_FOUND:                   {TIEM_RESOURCE_HOST_NOT_FOUND, "host not found", 500},
	TIEM_RESOURCE_NO_ENOUGH_HOST:                   {TIEM_RESOURCE_NO_ENOUGH_HOST, "no enough host resource", 500},
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED:    {TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED, "no enough disk resource after excluded", 500},
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER: {TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER, "no enough disk after disk filter", 500},
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_HOST_FILTER: {TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_HOST_FILTER, "no enough disk after host filter", 500},
	TIEM_RESOURCE_NO_ENOUGH_PORT:                   {TIEM_RESOURCE_NO_ENOUGH_PORT, "no enough port resource", 500},
	TIEM_RESOURCE_NOT_ALL_SUCCEED:                  {TIEM_RESOURCE_NOT_ALL_SUCCEED, "not all request succeed", 500},
	TIEM_RESOURCE_INVALID_STRATEGY:                 {TIEM_RESOURCE_INVALID_STRATEGY, "invalid alloc strategy", 500},
	TIEM_RESOURCE_TRAIT_NOT_FOUND:                  {TIEM_RESOURCE_TRAIT_NOT_FOUND, "trait not found by label name", 500},
	TIEM_RESOURCE_INVALID_LOCATION:                 {TIEM_RESOURCE_INVALID_LOCATION, "invalid location", 500},
	TIEM_RESOURCE_INVALID_ARCH:                     {TIEM_RESOURCE_INVALID_ARCH, "invalid arch", 500},

	// tenant
	TIEM_ACCOUNT_NOT_FOUND:       {TIEM_ACCOUNT_NOT_FOUND, "account not found", 404},
	TIEM_TENANT_NOT_FOUND:        {TIEM_TENANT_NOT_FOUND, "tenant not found", 404},
	TIEM_QUERY_PERMISSION_FAILED: {TIEM_QUERY_PERMISSION_FAILED, "query permission failed", 500},
	TIEM_ADD_TOKEN_FAILED:        {TIEM_ADD_TOKEN_FAILED, "add token failed", 500},
	TIEM_TOKEN_NOT_FOUND:         {TIEM_TOKEN_NOT_FOUND, "token not found", 404},

	// param group & cluster param
	TIEM_DEFAULT_PARAM_GROUP_NOT_DEL: {TIEM_DEFAULT_PARAM_GROUP_NOT_DEL, "The default param group cannot be deleted", 500},
	TIEM_MODIFY_PARAM_FAILED:         {TIEM_MODIFY_PARAM_FAILED, "apply or modify parameters failed!", 500},
	TIEM_CONVERT_OBJ_FAILED:          {TIEM_CONVERT_OBJ_FAILED, "convert obj failed!", 500},

}
