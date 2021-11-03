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
	TIEM_SUCCESS           = 0
	TIEM_PARAMETER_INVALID = 1

	TIEM_ACCOUNT_NOT_FOUND = 100
	TIME_ACCOUNT_EXIST     = 101

	TIEM_TENANT_NOT_FOUND = 200
	TIEM_TENANT_EXIST     = 201

	TIEM_QUERY_PERMISSION_FAILED = 300

	TIEM_ADD_TOKEN_FAILED = 400
	TIEM_TOKEN_NOT_FOUND  = 401

	TIEM_RESOURCE_SQL_ERROR                        = 500
	TIEM_RESOURCE_HOST_NOT_FOUND                   = 501
	TIEM_RESOURCE_NO_ENOUGH_HOST                   = 502
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED    = 503
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER = 504
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_HOST_FILTER = 505
	TIEM_RESOURCE_NO_ENOUGH_PORT                   = 506
	TIEM_RESOURCE_NOT_ALL_SUCCEED                  = 507
	TIEM_RESOURCE_INVALID_STRATEGY                 = 508
	TIEM_RESOURCE_INVAILD_RECYCLE_TYPE             = 509
	TIEM_UPDATE_HOST_STATUS_FAIL                   = 510
	TIEM_RESERVE_HOST_FAIL                         = 511
	TIEM_RESOURCE_NO_STOCK                         = 512

	TIEM_DASHBOARD_NOT_FOUND           = 600
	TIEM_EXPORT_PARAM_INVALID          = 601
	TIEM_EXPORT_PROCESS_FAILED         = 602
	TIEM_IMPORT_PARAM_INVALID          = 603
	TIEM_IMPORT_PROCESS_FAILED         = 604
	TIEM_TRANSPORT_RECORD_NOT_FOUND    = 605
	TIEM_BACKUP_PROCESS_FAILED         = 606
	TIEM_RECOVER_PARAM_INVALID         = 607
	TIEM_RECOVER_PROCESS_FAILED        = 608
	TIEM_BACKUP_RECORD_DELETE_FAILED   = 609
	TIEM_BACKUP_RECORD_QUERY_FAILED    = 610
	TIEM_BACKUP_STRATEGY_PARAM_INVALID = 611
	TIEM_BACKUP_STRATEGY_SAVE_FAILED   = 612
	TIEM_BACKUP_STRATEGY_QUERY_FAILED  = 613
	TIEM_MONITOR_NOT_FOUND             = 614
)

var TiEMErrMsg = map[TIEM_ERROR_CODE]string{
	TIEM_SUCCESS:           "successful",
	TIEM_PARAMETER_INVALID: "parameter is invalid",
	TIEM_ACCOUNT_NOT_FOUND: "account is not found",
	TIME_ACCOUNT_EXIST:     "account is exist",

	TIEM_TENANT_NOT_FOUND: "tenant is not found",
	TIEM_TENANT_EXIST:     "tenant is exist",

	TIEM_QUERY_PERMISSION_FAILED: "query permission failed",

	TIEM_ADD_TOKEN_FAILED: "add token failed",

	TIEM_TOKEN_NOT_FOUND: "token not found",

	TIEM_RESOURCE_SQL_ERROR:                        "resource sql error",
	TIEM_RESOURCE_HOST_NOT_FOUND:                   "host is not found",
	TIEM_RESOURCE_NO_ENOUGH_HOST:                   "no enough host resource",
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_EXCLUDED:    "no enough disk resource after excluded",
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_DISK_FILTER: "no enough disk after disk filter",
	TIEM_RESOURCE_NO_ENOUGH_DISK_AFTER_HOST_FILTER: "no enouth disk after host filter",
	TIEM_RESOURCE_NO_ENOUGH_PORT:                   "no enough port resource",
	TIEM_RESOURCE_NOT_ALL_SUCCEED:                  "not all request succeed",
	TIEM_RESOURCE_INVALID_STRATEGY:                 "invalid alloc strategy",

	TIEM_DASHBOARD_NOT_FOUND:           "dashboard is not found",
	TIEM_EXPORT_PARAM_INVALID:          "export data param invalid",
	TIEM_EXPORT_PROCESS_FAILED:         "export process failed",
	TIEM_IMPORT_PARAM_INVALID:          "import data param invalid",
	TIEM_IMPORT_PROCESS_FAILED:         "import process failed",
	TIEM_TRANSPORT_RECORD_NOT_FOUND:    "transport record is not found",
	TIEM_BACKUP_PROCESS_FAILED:         "backup process failed",
	TIEM_RECOVER_PARAM_INVALID:         "recover param invalid",
	TIEM_RECOVER_PROCESS_FAILED:        "recover process failed",
	TIEM_BACKUP_RECORD_DELETE_FAILED:   "delete backup record failed",
	TIEM_BACKUP_RECORD_QUERY_FAILED:    "query backup record failed",
	TIEM_BACKUP_STRATEGY_PARAM_INVALID: "backup strategy param invalid",
	TIEM_BACKUP_STRATEGY_SAVE_FAILED:   "save backup strategy failed",
	TIEM_BACKUP_STRATEGY_QUERY_FAILED:  "query backup strategy failed",
}
