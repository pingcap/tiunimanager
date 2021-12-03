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

package errors

type EM_ERROR_CODE int32

// all error code
// todo standardize
const (
	EM_SUCCESS EM_ERROR_CODE = 0

	EM_METADB_SERVER_CALL_ERROR  EM_ERROR_CODE = 9998
	EM_CLUSTER_SERVER_CALL_ERROR EM_ERROR_CODE = 9999
	EM_TASK_TIMEOUT              EM_ERROR_CODE = 9997
	EM_FLOW_NOT_FOUND            EM_ERROR_CODE = 9996
	EM_TASK_FAILED               EM_ERROR_CODE = 9995
	EM_TASK_POLLING_TIME_OUT     EM_ERROR_CODE = 9994
	EM_TASK_CONFLICT             EM_ERROR_CODE = 9993
	EM_TASK_CANCELED             EM_ERROR_CODE = 9992

	EM_PARAMETER_INVALID  EM_ERROR_CODE = 1
	EM_UNRECOGNIZED_ERROR EM_ERROR_CODE = 2

	EM_TAKEOVER_SSH_CONNECT_ERROR     EM_ERROR_CODE = 20109
	EM_TAKEOVER_SFTP_ERROR            EM_ERROR_CODE = 20110
	EM_CLUSTER_NOT_FOUND              EM_ERROR_CODE = 20111
	EM_CLUSTER_RESOURCE_NOT_ENOUGH    EM_ERROR_CODE = 20112
	EM_CLUSTER_GET_CLUSTER_PORT_ERROR EM_ERROR_CODE = 20113

	EM_ACCOUNT_NOT_FOUND       EM_ERROR_CODE = 100
	EM_TENANT_NOT_FOUND        EM_ERROR_CODE = 200
	EM_QUERY_PERMISSION_FAILED EM_ERROR_CODE = 300
	EM_ADD_TOKEN_FAILED        EM_ERROR_CODE = 400
	EM_TOKEN_NOT_FOUND         EM_ERROR_CODE = 401

	EM_RESOURCE_SQL_ERROR                        EM_ERROR_CODE = 500
	EM_RESOURCE_HOST_NOT_FOUND                   EM_ERROR_CODE = 501
	EM_RESOURCE_NO_ENOUGH_HOST                   EM_ERROR_CODE = 502
	EM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED    EM_ERROR_CODE = 503
	EM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER EM_ERROR_CODE = 504
	EM_RESOURCE_NO_ENOUGH_DISK_AFTER_HOST_FILTER EM_ERROR_CODE = 505
	EM_RESOURCE_NO_ENOUGH_PORT                   EM_ERROR_CODE = 506
	EM_RESOURCE_NOT_ALL_SUCCEED                  EM_ERROR_CODE = 507
	EM_RESOURCE_INVALID_STRATEGY                 EM_ERROR_CODE = 508
	EM_RESOURCE_INVAILD_RECYCLE_TYPE             EM_ERROR_CODE = 509
	EM_UPDATE_HOST_STATUS_FAIL                   EM_ERROR_CODE = 510
	EM_RESERVE_HOST_FAIL                         EM_ERROR_CODE = 511
	EM_RESOURCE_NO_STOCK                         EM_ERROR_CODE = 512
	EM_RESOURCE_GET_DISK_ID_FAIL                 EM_ERROR_CODE = 513
	EM_RESOURCE_TRAIT_NOT_FOUND                  EM_ERROR_CODE = 514
	EM_RESOURCE_INVALID_LOCATION                 EM_ERROR_CODE = 515
	EM_RESOURCE_INVALID_ARCH                     EM_ERROR_CODE = 516
	EM_RESOURCE_ADD_TABLE_ERROR                  EM_ERROR_CODE = 517
	EM_RESOURCE_INIT_LABELS_ERROR                EM_ERROR_CODE = 518
	EM_RESOURCE_CREATE_HOST_ERROR                EM_ERROR_CODE = 519
	EM_RESOURCE_LOCK_TABLE_ERROR                 EM_ERROR_CODE = 520
	EM_RESOURCE_DELETE_HOST_ERROR                EM_ERROR_CODE = 521

	EM_DASHBOARD_NOT_FOUND           EM_ERROR_CODE = 600
	EM_EXPORT_PARAM_INVALID          EM_ERROR_CODE = 601
	EM_EXPORT_PROCESS_FAILED         EM_ERROR_CODE = 602
	EM_IMPORT_PARAM_INVALID          EM_ERROR_CODE = 603
	EM_IMPORT_PROCESS_FAILED         EM_ERROR_CODE = 604
	EM_TRANSPORT_RECORD_NOT_FOUND    EM_ERROR_CODE = 605
	EM_BACKUP_PROCESS_FAILED         EM_ERROR_CODE = 606
	EM_RECOVER_PARAM_INVALID         EM_ERROR_CODE = 607
	EM_RECOVER_PROCESS_FAILED        EM_ERROR_CODE = 608
	EM_BACKUP_RECORD_DELETE_FAILED   EM_ERROR_CODE = 609
	EM_BACKUP_RECORD_QUERY_FAILED    EM_ERROR_CODE = 610
	EM_BACKUP_STRATEGY_PARAM_INVALID EM_ERROR_CODE = 611
	EM_BACKUP_STRATEGY_SAVE_FAILED   EM_ERROR_CODE = 612
	EM_BACKUP_STRATEGY_QUERY_FAILED  EM_ERROR_CODE = 613
	EM_MONITOR_NOT_FOUND             EM_ERROR_CODE = 614
	EM_TRANSPORT_RECORD_DEL_FAIL     EM_ERROR_CODE = 615

	EM_DEFAULT_PARAM_GROUP_NOT_DEL EM_ERROR_CODE = 20500

	EM_CHANGE_FEED_NOT_FOUND EM_ERROR_CODE = 701

	EM_CHANGE_FEED_DUPLICATE_ID           EM_ERROR_CODE = 702
	EM_CHANGE_FEED_STATUS_CONFLICT        EM_ERROR_CODE = 703
	EM_CHANGE_FEED_LOCK_EXPIRED           EM_ERROR_CODE = 704
	EM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM EM_ERROR_CODE = 705
)

type ErrorCodeExplanation struct {
	code        EM_ERROR_CODE
	explanation string
	httpCode    int
}

func (t EM_ERROR_CODE) GetHttpCode() int {
	return explanationContainer[t].httpCode
}

func (t EM_ERROR_CODE) Equal(code int32) bool {
	return code == int32(t)
}

func (t EM_ERROR_CODE) Explain() string {
	return explanationContainer[t].explanation
}

var explanationContainer = map[EM_ERROR_CODE]ErrorCodeExplanation{
	EM_SUCCESS: {code: EM_SUCCESS, explanation: "succeed", httpCode: 200},

	// system error
	EM_METADB_SERVER_CALL_ERROR:  {EM_METADB_SERVER_CALL_ERROR, "call metadb-Server failed", 500},
	EM_CLUSTER_SERVER_CALL_ERROR: {EM_CLUSTER_SERVER_CALL_ERROR, "call cluster-Server failed", 500},
	EM_PARAMETER_INVALID:         {EM_PARAMETER_INVALID, "parameter is invalid", 500},
	EM_UNRECOGNIZED_ERROR:        {EM_UNRECOGNIZED_ERROR, "unrecognized error", 500},

	EM_TASK_TIMEOUT:          {EM_TASK_TIMEOUT, "task timeout", 500},
	EM_FLOW_NOT_FOUND:        {EM_FLOW_NOT_FOUND, "flow not found", 500},
	EM_TASK_FAILED:           {EM_TASK_FAILED, "task failed", 500},
	EM_TASK_CONFLICT:         {EM_TASK_CONFLICT, "task polling time out", 500},
	EM_TASK_CANCELED:         {EM_TASK_CONFLICT, "task canceled", 500},
	EM_TASK_POLLING_TIME_OUT: {EM_TASK_POLLING_TIME_OUT, "task polling time out", 500},

	// cluster management
	EM_CLUSTER_NOT_FOUND:           {EM_CLUSTER_NOT_FOUND, "cluster not found", 500},
	EM_TAKEOVER_SSH_CONNECT_ERROR:  {EM_TAKEOVER_SSH_CONNECT_ERROR, "ssh connect failed", 500},
	EM_TAKEOVER_SFTP_ERROR:         {EM_TAKEOVER_SFTP_ERROR, "sftp failed", 500},
	EM_CLUSTER_RESOURCE_NOT_ENOUGH: {EM_CLUSTER_RESOURCE_NOT_ENOUGH, "no enough resource for cluster", 500},

	EM_DASHBOARD_NOT_FOUND: {EM_DASHBOARD_NOT_FOUND, "dashboard is not found", 500},

	// cluster import export
	EM_EXPORT_PARAM_INVALID:      {EM_EXPORT_PARAM_INVALID, "export data param invalid", 500},
	EM_EXPORT_PROCESS_FAILED:     {EM_EXPORT_PROCESS_FAILED, "export process failed", 500},
	EM_IMPORT_PARAM_INVALID:      {EM_IMPORT_PARAM_INVALID, "import data param invalid", 500},
	EM_IMPORT_PROCESS_FAILED:     {EM_IMPORT_PROCESS_FAILED, "import process failed", 500},
	EM_TRANSPORT_RECORD_DEL_FAIL: {EM_TRANSPORT_RECORD_DEL_FAIL, "delete data transport failed", 500},

	// cluster backup
	EM_TRANSPORT_RECORD_NOT_FOUND:    {EM_TRANSPORT_RECORD_NOT_FOUND, "transport record is not found", 500},
	EM_BACKUP_PROCESS_FAILED:         {EM_BACKUP_PROCESS_FAILED, "backup process failed", 500},
	EM_RECOVER_PARAM_INVALID:         {EM_RECOVER_PARAM_INVALID, "recover param invalid", 500},
	EM_RECOVER_PROCESS_FAILED:        {EM_RECOVER_PROCESS_FAILED, "recover process failed", 500},
	EM_BACKUP_RECORD_DELETE_FAILED:   {EM_BACKUP_RECORD_DELETE_FAILED, "delete backup record failed", 500},
	EM_BACKUP_RECORD_QUERY_FAILED:    {EM_BACKUP_RECORD_QUERY_FAILED, "query backup record failed", 500},
	EM_BACKUP_STRATEGY_PARAM_INVALID: {EM_BACKUP_STRATEGY_PARAM_INVALID, "backup strategy param invalid", 500},
	EM_BACKUP_STRATEGY_SAVE_FAILED:   {EM_BACKUP_STRATEGY_SAVE_FAILED, "save backup strategy failed", 500},
	EM_BACKUP_STRATEGY_QUERY_FAILED:  {EM_BACKUP_STRATEGY_QUERY_FAILED, "query backup strategy failed", 500},

	// resource
	EM_RESOURCE_SQL_ERROR:                        {EM_RESOURCE_SQL_ERROR, "resource sql error", 500},
	EM_RESOURCE_HOST_NOT_FOUND:                   {EM_RESOURCE_HOST_NOT_FOUND, "host not found", 500},
	EM_RESOURCE_NO_ENOUGH_HOST:                   {EM_RESOURCE_NO_ENOUGH_HOST, "no enough host resource", 500},
	EM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED:    {EM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED, "no enough disk resource after excluded", 500},
	EM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER: {EM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER, "no enough disk after disk filter", 500},
	EM_RESOURCE_NO_ENOUGH_DISK_AFTER_HOST_FILTER: {EM_RESOURCE_NO_ENOUGH_DISK_AFTER_HOST_FILTER, "no enough disk after host filter", 500},
	EM_RESOURCE_NO_ENOUGH_PORT:                   {EM_RESOURCE_NO_ENOUGH_PORT, "no enough port resource", 500},
	EM_RESOURCE_NOT_ALL_SUCCEED:                  {EM_RESOURCE_NOT_ALL_SUCCEED, "not all request succeed", 500},
	EM_RESOURCE_INVALID_STRATEGY:                 {EM_RESOURCE_INVALID_STRATEGY, "invalid alloc strategy", 500},
	EM_RESOURCE_TRAIT_NOT_FOUND:                  {EM_RESOURCE_TRAIT_NOT_FOUND, "trait not found by label name", 500},
	EM_RESOURCE_INVALID_LOCATION:                 {EM_RESOURCE_INVALID_LOCATION, "invalid location", 500},
	EM_RESOURCE_INVALID_ARCH:                     {EM_RESOURCE_INVALID_ARCH, "invalid arch", 500},

	// tenant
	EM_ACCOUNT_NOT_FOUND:       {EM_ACCOUNT_NOT_FOUND, "account not found", 500},
	EM_TENANT_NOT_FOUND:        {EM_TENANT_NOT_FOUND, "tenant not found", 500},
	EM_QUERY_PERMISSION_FAILED: {EM_QUERY_PERMISSION_FAILED, "query permission failed", 500},
	EM_ADD_TOKEN_FAILED:        {EM_ADD_TOKEN_FAILED, "add token failed", 500},
	EM_TOKEN_NOT_FOUND:         {EM_TOKEN_NOT_FOUND, "token not found", 500},

	// param group & cluster param
	EM_DEFAULT_PARAM_GROUP_NOT_DEL: {EM_DEFAULT_PARAM_GROUP_NOT_DEL, "The default param group cannot be deleted", 500},

	// change feed
	EM_CHANGE_FEED_NOT_FOUND:              {EM_CHANGE_FEED_NOT_FOUND, "change feed task not found", 404},
	EM_CHANGE_FEED_DUPLICATE_ID:           {EM_CHANGE_FEED_DUPLICATE_ID, "duplicate id", 500},
	EM_CHANGE_FEED_STATUS_CONFLICT:        {EM_CHANGE_FEED_STATUS_CONFLICT, "task status conflict", 409},
	EM_CHANGE_FEED_LOCK_EXPIRED:           {EM_CHANGE_FEED_LOCK_EXPIRED, "task status lock expired", 409},
	EM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM: {EM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM, "task downstream type not supported", 500},
}
