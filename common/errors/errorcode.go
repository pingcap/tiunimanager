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

const (
	TIEM_PANIC   EM_ERROR_CODE = -1
	TIEM_SUCCESS EM_ERROR_CODE = 0

	TIEM_CLUSTER_SERVER_CALL_ERROR EM_ERROR_CODE = 9999
	TIEM_TASK_TIMEOUT              EM_ERROR_CODE = 9997
	TIEM_FLOW_NOT_FOUND            EM_ERROR_CODE = 9996
	TIEM_TASK_FAILED               EM_ERROR_CODE = 9995
	TIEM_TASK_POLLING_TIME_OUT     EM_ERROR_CODE = 9994
	TIEM_TASK_CONFLICT             EM_ERROR_CODE = 9993
	TIEM_TASK_CANCELED             EM_ERROR_CODE = 9992

	TIEM_SYSTEM_MISSING_DATA    EM_ERROR_CODE = 9899
	TIEM_SYSTEM_MISSING_CONFIG  EM_ERROR_CODE = 9898
	TIEM_SYSTEM_STATE_CONFLICT  EM_ERROR_CODE = 9897
	TIEM_SYSTEM_INVALID_VERSION EM_ERROR_CODE = 9896

	TIEM_UNRECOGNIZED_ERROR EM_ERROR_CODE = 10000
	TIEM_PARAMETER_INVALID  EM_ERROR_CODE = 10001
	TIEM_SQL_ERROR          EM_ERROR_CODE = 10002
	TIEM_CLUSTER_NOT_FOUND  EM_ERROR_CODE = 10003
	TIEM_MARSHAL_ERROR      EM_ERROR_CODE = 10004
	TIEM_UNMARSHAL_ERROR    EM_ERROR_CODE = 10005
	TIEM_INSTANCE_NOT_FOUND EM_ERROR_CODE = 10006
	TIEM_CONNECT_TIDB_ERROR EM_ERROR_CODE = 10007

	// cluster
	TIEM_DUPLICATED_NAME             EM_ERROR_CODE = 20101
	TIEM_INVALID_TOPOLOGY            EM_ERROR_CODE = 20102
	TIEM_UNSUPPORT_PRODUCT           EM_ERROR_CODE = 20103
	TIEM_CLUSTER_RESOURCE_NOT_ENOUGH EM_ERROR_CODE = 20104
	TIEM_CLUSTER_METADATA_BROKEN     EM_ERROR_CODE = 20105

	TIEM_TAKEOVER_SSH_CONNECT_ERROR     EM_ERROR_CODE = 20201
	TIEM_TAKEOVER_SFTP_ERROR            EM_ERROR_CODE = 20110
	TIEM_CLUSTER_GET_CLUSTER_PORT_ERROR EM_ERROR_CODE = 20113
	TIEM_CLUSTER_MAINTENANCE_CONFLICT   EM_ERROR_CODE = 20114

	// backup && restore
	TIEM_BACKUP_SYSTEM_CONFIG_NOT_FOUND EM_ERROR_CODE = 20600
	TIEM_BACKUP_SYSTEM_CONFIG_INVAILD   EM_ERROR_CODE = 20601
	TIEM_BACKUP_RECORD_CREATE_FAILED    EM_ERROR_CODE = 20602
	TIEM_BACKUP_RECORD_DELETE_FAILED    EM_ERROR_CODE = 20603
	TIEM_BACKUP_RECORD_QUERY_FAILED     EM_ERROR_CODE = 20604
	TIEM_BACKUP_STRATEGY_SAVE_FAILED    EM_ERROR_CODE = 20605
	TIEM_BACKUP_STRATEGY_QUERY_FAILED   EM_ERROR_CODE = 20606
	TIEM_BACKUP_STRATEGY_DELETE_FAILED  EM_ERROR_CODE = 20607
	TIEM_BACKUP_FILE_DELETE_FAILED      EM_ERROR_CODE = 20608
	TIEM_BACKUP_PATH_CREATE_FAILED      EM_ERROR_CODE = 20609
	TIEM_BACKUP_RECORD_INVALID          EM_ERROR_CODE = 20610
	TIEM_BACKUP_RECORD_CANCEL_FAILED    EM_ERROR_CODE = 20611

	// upgrade
	TIEM_UPGRADE_QUERY_PATH_FAILED EM_ERROR_CODE = 21100
	TIEM_UPGRADE_REGION_UNHEALTHY  EM_ERROR_CODE = 21104
	TIEM_UPGRADE_VERSION_INCORRECT EM_ERROR_CODE = 21105

	// switchover
	TIEM_MASTER_SLAVE_SWITCHOVER_NOT_FOUND               EM_ERROR_CODE = 21000
	TIEM_MASTER_SLAVE_SWITCHOVER_FAILED                  EM_ERROR_CODE = 21001
	TIEM_MASTER_SLAVE_SWITCHOVER_CDC_SYNC_TASK_NOT_FOUND EM_ERROR_CODE = 21002
	TIEM_MASTER_SLAVE_SWITCHOVER_SLAVE_NO_CDC_COMPONENT  EM_ERROR_CODE = 21003

	// workflow
	TIEM_WORKFLOW_CREATE_FAILED         EM_ERROR_CODE = 40100
	TIEM_WORKFLOW_QUERY_FAILED          EM_ERROR_CODE = 40101
	TIEM_WORKFLOW_DETAIL_FAILED         EM_ERROR_CODE = 40102
	TIEM_WORKFLOW_START_FAILED          EM_ERROR_CODE = 40103
	TIEM_WORKFLOW_DEFINE_NOT_FOUND      EM_ERROR_CODE = 40104
	TIEM_WORKFLOW_NODE_POLLING_TIME_OUT EM_ERROR_CODE = 40105
	TIEM_WORKFLOW_STOP_FAILED           EM_ERROR_CODE = 40106
	TIEM_WORKFLOW_CANCEL_FAILED         EM_ERROR_CODE = 40107

	// import && export
	TIEM_TRANSPORT_SYSTEM_CONFIG_NOT_FOUND EM_ERROR_CODE = 60100
	TIEM_TRANSPORT_SYSTEM_CONFIG_INVALID   EM_ERROR_CODE = 60101
	TIEM_TRANSPORT_RECORD_NOT_FOUND        EM_ERROR_CODE = 60102
	TIEM_TRANSPORT_RECORD_CREATE_FAILED    EM_ERROR_CODE = 60103
	TIEM_TRANSPORT_RECORD_DELETE_FAILED    EM_ERROR_CODE = 60104
	TIEM_TRANSPORT_RECORD_QUERY_FAILED     EM_ERROR_CODE = 60105
	TIEM_TRANSPORT_FILE_DELETE_FAILED      EM_ERROR_CODE = 60106
	TIEM_TRANSPORT_PATH_CREATE_FAILED      EM_ERROR_CODE = 60107
	TIEM_TRANSPORT_FILE_SIZE_INVALID       EM_ERROR_CODE = 60108
	TIEM_TRANSPORT_FILE_UPLOAD_FAILED      EM_ERROR_CODE = 60109
	TIEM_TRANSPORT_FILE_DOWNLOAD_FAILED    EM_ERROR_CODE = 60110
	TIEM_TRANSPORT_FILE_TRANSFER_LIMITED   EM_ERROR_CODE = 60111

	//user
	TIEM_UNAUTHORIZED_USER    EM_ERROR_CODE = 70600
	TIEM_USER_NOT_FOUND       EM_ERROR_CODE = 70601
	TIEM_ACCESS_TOKEN_EXPIRED EM_ERROR_CODE = 70602
	TIEM_LOGIN_FAILED         EM_ERROR_CODE = 70603

	TIEM_RBAC_PERMISSION_CHECK_FAILED  EM_ERROR_CODE = 70650
	TIEM_RBAC_PERMISSION_ADD_FAILED    EM_ERROR_CODE = 70651
	TIEM_RBAC_PERMISSION_DELETE_FAILED EM_ERROR_CODE = 70652
	TIEM_RBAC_PERMISSION_QUERY_FAILED  EM_ERROR_CODE = 70653
	TIEM_RBAC_ROLE_CREATE_FAILED       EM_ERROR_CODE = 70654
	TIEM_RBAC_ROLE_DELETE_FAILED       EM_ERROR_CODE = 70655
	TIEM_RBAC_ROLE_QUERY_FAILED        EM_ERROR_CODE = 70656
	TIEM_RBAC_ROLE_BIND_FAILED         EM_ERROR_CODE = 70657
	TIEM_RBAC_ROLE_UNBIND_FAILED       EM_ERROR_CODE = 70658

	// dashboard && monitor
	TIEM_DASHBOARD_NOT_FOUND EM_ERROR_CODE = 80100

	TIEM_RESOURCE_HOST_NOT_FOUND            EM_ERROR_CODE = 30101
	TIEM_RESOURCE_NO_ENOUGH_HOST            EM_ERROR_CODE = 30102
	TIEM_RESOURCE_NO_ENOUGH_DISK            EM_ERROR_CODE = 30103
	TIEM_RESOURCE_NO_ENOUGH_PORT            EM_ERROR_CODE = 30104
	TIEM_UPDATE_HOST_STATUS_FAIL            EM_ERROR_CODE = 30105
	TIEM_RESERVE_HOST_FAIL                  EM_ERROR_CODE = 30106
	TIEM_RESOURCE_NO_STOCK                  EM_ERROR_CODE = 30107
	TIEM_RESOURCE_TRAIT_NOT_FOUND           EM_ERROR_CODE = 30108
	TIEM_RESOURCE_INVALID_LOCATION          EM_ERROR_CODE = 30109
	TIEM_RESOURCE_INVALID_ARCH              EM_ERROR_CODE = 30110
	TIEM_RESOURCE_ADD_TABLE_ERROR           EM_ERROR_CODE = 30111
	TIEM_RESOURCE_INIT_LABELS_ERROR         EM_ERROR_CODE = 30112
	TIEM_RESOURCE_CREATE_HOST_ERROR         EM_ERROR_CODE = 30113
	TIEM_RESOURCE_DELETE_HOST_ERROR         EM_ERROR_CODE = 30114
	TIEM_RESOURCE_INVALID_PRODUCT_NAME      EM_ERROR_CODE = 30115
	TIEM_RESOURCE_INVALID_PURPOSE           EM_ERROR_CODE = 30116
	TIEM_RESOURCE_INVALID_DISKTYPE          EM_ERROR_CODE = 30117
	TIEM_RESOURCE_HOST_ALREADY_EXIST        EM_ERROR_CODE = 30118
	TIEM_RESOURCE_HOST_STILL_INUSED         EM_ERROR_CODE = 30119
	TIEM_RESOURCE_CREATE_DISK_ERROR         EM_ERROR_CODE = 30120
	TIEM_RESOURCE_TEMPLATE_FILE_NOT_FOUND   EM_ERROR_CODE = 30121
	TIEM_RESOURCE_PARSE_TEMPLATE_FILE_ERROR EM_ERROR_CODE = 30122
	TIEM_RESOURCE_DECRYPT_PASSWD_ERROR      EM_ERROR_CODE = 30123
	TIEM_RESOURCE_ALLOCATE_ERROR            EM_ERROR_CODE = 30124
	TIEM_RESOURCE_RECYCLE_ERROR             EM_ERROR_CODE = 30125
	TIEM_RESOURCE_CONNECT_TO_HOST_ERROR     EM_ERROR_CODE = 30126
	TIEM_RESOURCE_NEW_SESSION_ERROR         EM_ERROR_CODE = 30127
	TIEM_RESOURCE_RUN_COMMAND_ERROR         EM_ERROR_CODE = 30128
	TIEM_RESOURCE_HOST_NOT_EXPECTED         EM_ERROR_CODE = 30129
	TIEM_RESOURCE_INIT_FILEBEAT_ERROR       EM_ERROR_CODE = 30130
	TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR    EM_ERROR_CODE = 30131
	TIEM_RESOURCE_UNINSTALL_FILEBEAT_ERROR  EM_ERROR_CODE = 30132
	TIEM_RESOURCE_PREPARE_HOST_ERROR        EM_ERROR_CODE = 30133
	TIEM_RESOURCE_INVALID_VENDOR_NAME       EM_ERROR_CODE = 30134
	TIEM_RESOURCE_INVALID_ZONE_INFO         EM_ERROR_CODE = 30135
	TIEM_RESOURCE_CHECK_COMPUTES_ERROR      EM_ERROR_CODE = 30136
	TIEM_RESOURCE_CHECK_DISKS_ERROR         EM_ERROR_CODE = 30137
	TIEM_RESOURCE_INIT_DEPLOY_USER_ERROR    EM_ERROR_CODE = 30138
	TIEM_RESOURCE_INIT_HOST_AUTH_ERROR      EM_ERROR_CODE = 30139
	TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR     EM_ERROR_CODE = 30140
	TIEM_RESOURCE_VALIDATE_DISK_ERROR       EM_ERROR_CODE = 30141
	TIEM_RESOURCE_UPDATE_DISK_ERROR         EM_ERROR_CODE = 30142
	TIEM_RESOURCE_DELETE_DISK_ERROR         EM_ERROR_CODE = 30143
	TIEM_RESOURCE_DISK_STILL_INUSED         EM_ERROR_CODE = 30144
	TIEM_RESOURCE_DISK_ALREADY_EXIST        EM_ERROR_CODE = 30145

	TIEM_MONITOR_NOT_FOUND EM_ERROR_CODE = 614

	// param group & cluster param
	TIEM_DEFAULT_PARAM_GROUP_NOT_DEL                 EM_ERROR_CODE = 20500
	TIEM_DEFAULT_PARAM_GROUP_NOT_MODIFY              EM_ERROR_CODE = 20501
	TIEM_MODIFY_PARAM_FAILED                         EM_ERROR_CODE = 20502
	TIEM_CONVERT_OBJ_FAILED                          EM_ERROR_CODE = 20503
	TIEM_PARAMETER_GROUP_CREATE_ERROR                EM_ERROR_CODE = 20504
	TIEM_PARAMETER_GROUP_UPDATE_ERROR                EM_ERROR_CODE = 20505
	TIEM_PARAMETER_GROUP_DELETE_ERROR                EM_ERROR_CODE = 20506
	TIEM_PARAMETER_GROUP_QUERY_ERROR                 EM_ERROR_CODE = 20507
	TIEM_PARAMETER_GROUP_DETAIL_ERROR                EM_ERROR_CODE = 20508
	TIEM_PARAMETER_GROUP_COPY_ERROR                  EM_ERROR_CODE = 20509
	TIEM_PARAMETER_GROUP_APPLY_ERROR                 EM_ERROR_CODE = 20510
	TIEM_PARAMETER_GROUP_CREATE_RELATION_PARAM_ERROR EM_ERROR_CODE = 20511
	TIEM_PARAMETER_GROUP_DELETE_RELATION_PARAM_ERROR EM_ERROR_CODE = 20512
	TIEM_PARAMETER_GROUP_UPDATE_RELATION_PARAM_ERROR EM_ERROR_CODE = 20513
	TIEM_CLUSTER_PARAMETER_QUERY_ERROR               EM_ERROR_CODE = 20514
	TIEM_CLUSTER_PARAMETER_UPDATE_ERROR              EM_ERROR_CODE = 20515
	TIEM_PARAMETER_GROUP_NAME_ALREADY_EXISTS         EM_ERROR_CODE = 20516
	TIEM_PARAMETER_GROUP_RELATION_CLUSTER_NOT_DEL    EM_ERROR_CODE = 20517

	TIEM_PARAMETER_QUERY_ERROR    EM_ERROR_CODE = 20520
	TIEM_PARAMETER_CREATE_ERROR   EM_ERROR_CODE = 20521
	TIEM_PARAMETER_DELETE_ERROR   EM_ERROR_CODE = 20522
	TIEM_PARAMETER_DETAIL_ERROR   EM_ERROR_CODE = 20523
	TIEM_PARAMETER_UPDATE_ERROR   EM_ERROR_CODE = 20524
	TIEM_PARAMETER_ALREADY_EXISTS EM_ERROR_CODE = 20525

	TIEM_CHANGE_FEED_NOT_FOUND              EM_ERROR_CODE = 21201
	TIEM_CHANGE_FEED_DUPLICATE_ID           EM_ERROR_CODE = 21202
	TIEM_CHANGE_FEED_STATUS_CONFLICT        EM_ERROR_CODE = 21203
	TIEM_CHANGE_FEED_LOCK_EXPIRED           EM_ERROR_CODE = 21204
	TIEM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM EM_ERROR_CODE = 21205
	TIEM_CHANGE_FEED_EXECUTE_ERROR          EM_ERROR_CODE = 21206

	TIEM_DELETE_INSTANCE_ERROR            EM_ERROR_CODE = 20801
	TIEM_CHECK_PLACEMENT_RULES_ERROR      EM_ERROR_CODE = 20802
	TIEM_CHECK_TIFLASH_MAX_REPLICAS_ERROR EM_ERROR_CODE = 20803
	TIEM_SCAN_MAX_REPLICA_COUNT_ERROR     EM_ERROR_CODE = 20804
	TIEM_PD_NOT_FOUND_ERROR               EM_ERROR_CODE = 20806
	TIEM_CHECK_INSTANCE_TIEMOUT_ERROR     EM_ERROR_CODE = 20807
	TIEM_STORE_NOT_FOUND_ERROR            EM_ERROR_CODE = 20808

	TIEM_CHECK_CLUSTER_VERSION_ERROR EM_ERROR_CODE = 21301
	TIEM_CDC_NOT_FOUND               EM_ERROR_CODE = 21302
	TIEM_CLONE_TIKV_ERROR            EM_ERROR_CODE = 21303
	TIEM_CLONE_SLAVE_ERROR           EM_ERROR_CODE = 21304

	CreateZonesError              EM_ERROR_CODE = 70001
	DeleteZonesError              EM_ERROR_CODE = 70002
	QueryZoneScanRowError         EM_ERROR_CODE = 70003
	CreateProductError            EM_ERROR_CODE = 70004
	DeleteProductError            EM_ERROR_CODE = 70005
	QueryProductsScanRowError     EM_ERROR_CODE = 70006
	QueryProductComponentProperty EM_ERROR_CODE = 70007
	CreateSpecsError              EM_ERROR_CODE = 70008
	DeleteSpecsError              EM_ERROR_CODE = 70009
	QuerySpecScanRowError         EM_ERROR_CODE = 70010

	QueryUserScanRowError              EM_ERROR_CODE = 70604
	QueryTenantScanRowError            EM_ERROR_CODE = 70605
	TenantAlreadyExist                 EM_ERROR_CODE = 70606
	TenantNotExist                     EM_ERROR_CODE = 70607
	DeleteTenantFailed                 EM_ERROR_CODE = 70608
	UpdateTenantOnBoardingStatusFailed EM_ERROR_CODE = 70609

	UserAlreadyExist              EM_ERROR_CODE = 70610
	UserGenSaltAndHashValueFailed EM_ERROR_CODE = 70611
	UserNotExist                  EM_ERROR_CODE = 70612
	DeleteUserFailed              EM_ERROR_CODE = 70613
	UpdateUserProfileFailed       EM_ERROR_CODE = 70614

	TIEM_LOG_QUERY_FAILED EM_ERROR_CODE = 80300
	TIEM_LOG_TIME_AFTER   EM_ERROR_CODE = 80301

	QueryReportsScanRowError EM_ERROR_CODE = 90001
	CheckReportNotExist      EM_ERROR_CODE = 90002
)

type ErrorCodeExplanation struct {
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
	TIEM_SUCCESS: {"succeed", 200},
	TIEM_PANIC:   {"panic", 500},

	// system error
	TIEM_UNRECOGNIZED_ERROR: {"unrecognized error", 500},
	TIEM_PARAMETER_INVALID:  {"parameter is invalid", 400},
	TIEM_SQL_ERROR:          {"failed to execute SQL", 500},
	TIEM_CLUSTER_NOT_FOUND:  {"cluster not found", 404},
	TIEM_MARSHAL_ERROR:      {"marshal error", 500},
	TIEM_UNMARSHAL_ERROR:    {"marshal error", 500},

	// user
	TIEM_UNAUTHORIZED_USER:             {"unauthorized", 401},
	TIEM_USER_NOT_FOUND:                {"user not found", 404},
	TIEM_ACCESS_TOKEN_EXPIRED:          {"access token has been expired", 401},
	TIEM_LOGIN_FAILED:                  {"incorrect username or password", 400},
	TIEM_RBAC_PERMISSION_CHECK_FAILED:  {"rbac permission check failed", 403},
	TIEM_RBAC_PERMISSION_ADD_FAILED:    {"rbac permission add failed", 500},
	TIEM_RBAC_PERMISSION_DELETE_FAILED: {"rbac permission delete failed", 500},
	TIEM_RBAC_PERMISSION_QUERY_FAILED:  {"rbac permission query failed", 500},
	TIEM_RBAC_ROLE_CREATE_FAILED:       {"rbac role create failed", 500},
	TIEM_RBAC_ROLE_QUERY_FAILED:        {"rbac role query failed", 500},
	TIEM_RBAC_ROLE_DELETE_FAILED:       {"rbac role delete failed", 500},
	TIEM_RBAC_ROLE_BIND_FAILED:         {"rbac role bind user failed", 500},
	TIEM_RBAC_ROLE_UNBIND_FAILED:       {"rbac role unbind user failed", 500},

	TIEM_CLUSTER_SERVER_CALL_ERROR: {"call cluster-Server failed", 500},
	TIEM_SYSTEM_MISSING_DATA:       {"missing system data", 500},
	TIEM_SYSTEM_MISSING_CONFIG:     {"missing system config", 500},
	TIEM_SYSTEM_STATE_CONFLICT:     {"system state conflict", 500},
	TIEM_SYSTEM_INVALID_VERSION:    {"invalid system version", 500},

	TIEM_TASK_TIMEOUT:           {"task timeout", 500},
	TIEM_FLOW_NOT_FOUND:         {"flow not found", 500},
	TIEM_TASK_FAILED:            {"task failed", 500},
	TIEM_TASK_CONFLICT:          {"task conflict", 400},
	TIEM_TASK_CANCELED:          {"task canceled", 500},
	TIEM_TASK_POLLING_TIME_OUT:  {"task polling time out", 500},
	TIEM_WORKFLOW_STOP_FAILED:   {"workflow stop failed", 500},
	TIEM_WORKFLOW_CANCEL_FAILED: {"workflow cancel failed", 500},

	TIEM_DUPLICATED_NAME:              {"duplicated cluster name", 400},
	TIEM_INVALID_TOPOLOGY:             {"invalid cluster topology", 400},
	TIEM_UNSUPPORT_PRODUCT:            {"unsupported cluster product", 400},
	TIEM_CLUSTER_RESOURCE_NOT_ENOUGH:  {"host resource is not enough", 500},
	TIEM_CLUSTER_MAINTENANCE_CONFLICT: {"maintenance status conflict", 409},
	TIEM_CLUSTER_METADATA_BROKEN:      {"cluster meta is incomplete", 400},

	// cluster management
	TIEM_TAKEOVER_SSH_CONNECT_ERROR: {"ssh connect failed", 500},
	TIEM_TAKEOVER_SFTP_ERROR:        {"sftp failed", 500},

	// dashboard && monitor
	TIEM_DASHBOARD_NOT_FOUND: {"dashboard is not found", 500},

	// workflow
	TIEM_WORKFLOW_CREATE_FAILED:         {"workflow create failed", 500},
	TIEM_WORKFLOW_QUERY_FAILED:          {"workflow query failed", 500},
	TIEM_WORKFLOW_DETAIL_FAILED:         {"workflow detail failed", 500},
	TIEM_WORKFLOW_START_FAILED:          {"workflow start failed", 500},
	TIEM_WORKFLOW_DEFINE_NOT_FOUND:      {"workflow define not found", 404},
	TIEM_WORKFLOW_NODE_POLLING_TIME_OUT: {"workflow node polling time out", 500},

	// import && export
	TIEM_TRANSPORT_SYSTEM_CONFIG_NOT_FOUND: {"data transport system config not found", 404},
	TIEM_TRANSPORT_SYSTEM_CONFIG_INVALID:   {"data transport system config invalid", 400},
	TIEM_TRANSPORT_RECORD_NOT_FOUND:        {"data transport record is not found", 404},
	TIEM_TRANSPORT_RECORD_CREATE_FAILED:    {"create data transport record failed", 500},
	TIEM_TRANSPORT_RECORD_DELETE_FAILED:    {"delete data transport record failed", 500},
	TIEM_TRANSPORT_RECORD_QUERY_FAILED:     {"query data transport record failed", 500},
	TIEM_TRANSPORT_FILE_DELETE_FAILED:      {"remove transport file failed", 500},
	TIEM_TRANSPORT_PATH_CREATE_FAILED:      {"data transport filepath create failed", 500},
	TIEM_TRANSPORT_FILE_SIZE_INVALID:       {"data transport file size invalid", 400},
	TIEM_TRANSPORT_FILE_UPLOAD_FAILED:      {"data transport file upload failed", 500},
	TIEM_TRANSPORT_FILE_DOWNLOAD_FAILED:    {"data transport file download failed", 500},
	TIEM_TRANSPORT_FILE_TRANSFER_LIMITED:   {"exceed limit file transfer num", 400},

	// backup && restore
	TIEM_BACKUP_SYSTEM_CONFIG_NOT_FOUND: {"backup system config not found", 404},
	TIEM_BACKUP_SYSTEM_CONFIG_INVAILD:   {"backup system config invalid", 400},
	TIEM_BACKUP_RECORD_CREATE_FAILED:    {"create backup record failed", 500},
	TIEM_BACKUP_RECORD_DELETE_FAILED:    {"delete backup record failed", 500},
	TIEM_BACKUP_RECORD_QUERY_FAILED:     {"query backup record failed", 500},
	TIEM_BACKUP_STRATEGY_SAVE_FAILED:    {"save backup strategy failed", 500},
	TIEM_BACKUP_STRATEGY_QUERY_FAILED:   {"query backup strategy failed", 500},
	TIEM_BACKUP_STRATEGY_DELETE_FAILED:  {"delete backup strategy failed", 500},
	TIEM_BACKUP_FILE_DELETE_FAILED:      {"remove backup file failed", 500},
	TIEM_BACKUP_PATH_CREATE_FAILED:      {"backup filepath create failed", 500},
	TIEM_BACKUP_RECORD_INVALID:          {"backup record invalid", 400},
	TIEM_BACKUP_RECORD_CANCEL_FAILED:    {"cancel backup record failed", 500},

	// resource
	TIEM_RESOURCE_HOST_NOT_FOUND:            {"host not found", 500},
	TIEM_UPDATE_HOST_STATUS_FAIL:            {"update host status failed", 500},
	TIEM_RESERVE_HOST_FAIL:                  {"reserved host failed", 500},
	TIEM_RESOURCE_NO_STOCK:                  {"Insufficient resources for inventory inquiries", 400},
	TIEM_RESOURCE_TRAIT_NOT_FOUND:           {"trait not found by label name", 400},
	TIEM_RESOURCE_INVALID_LOCATION:          {"invalid location of host", 400},
	TIEM_RESOURCE_INVALID_ARCH:              {"invalid architecture of host", 400},
	TIEM_RESOURCE_ADD_TABLE_ERROR:           {"failed to create database for resources", 500},
	TIEM_RESOURCE_INIT_LABELS_ERROR:         {"failed to initiate label table ", 500},
	TIEM_RESOURCE_CREATE_HOST_ERROR:         {"failed to create hosts to db", 500},
	TIEM_RESOURCE_DELETE_HOST_ERROR:         {"failed to delete hosts", 500},
	TIEM_RESOURCE_INVALID_PRODUCT_NAME:      {"invalid product name", 400},
	TIEM_RESOURCE_INVALID_PURPOSE:           {"invalid purpose of host", 400},
	TIEM_RESOURCE_INVALID_DISKTYPE:          {"invalid disk type of host", 400},
	TIEM_RESOURCE_HOST_ALREADY_EXIST:        {"host already exists in the resource pool", 409},
	TIEM_RESOURCE_HOST_STILL_INUSED:         {"host is still in use", 409},
	TIEM_RESOURCE_CREATE_DISK_ERROR:         {"failed to create disk", 500},
	TIEM_RESOURCE_TEMPLATE_FILE_NOT_FOUND:   {"template file is not found", 500},
	TIEM_RESOURCE_PARSE_TEMPLATE_FILE_ERROR: {"parse template file failed", 400},
	TIEM_RESOURCE_CONNECT_TO_HOST_ERROR:     {"connect to host failed", 400},
	TIEM_RESOURCE_NEW_SESSION_ERROR:         {"new connect session to host failed", 500},
	TIEM_RESOURCE_RUN_COMMAND_ERROR:         {"run command on host failed", 500},
	TIEM_RESOURCE_HOST_NOT_EXPECTED:         {"host is not expected as import file", 400},
	TIEM_RESOURCE_INIT_FILEBEAT_ERROR:       {"install filebeat on host failed", 400},
	TIEM_RESOURCE_EXTRACT_FLOW_CTX_ERROR:    {"extract workflow context failed", 500},
	TIEM_RESOURCE_UNINSTALL_FILEBEAT_ERROR:  {"uninstall filebeat on host failed", 400},
	TIEM_RESOURCE_PREPARE_HOST_ERROR:        {"prepare host before verify failed", 500},
	TIEM_RESOURCE_INVALID_VENDOR_NAME:       {"invalid vendor", 400},
	TIEM_RESOURCE_INVALID_ZONE_INFO:         {"invalid zone info", 400},
	TIEM_RESOURCE_CHECK_COMPUTES_ERROR:      {"check compute resource mismatch", 500},
	TIEM_RESOURCE_CHECK_DISKS_ERROR:         {"check disk resource mismatch", 500},
	TIEM_RESOURCE_INIT_DEPLOY_USER_ERROR:    {"init deploy user failed", 500},
	TIEM_RESOURCE_INIT_HOST_AUTH_ERROR:      {"init host auth failed", 500},
	TIEM_RESOURCE_UPDATE_HOSTINFO_ERROR:     {"update host info failed", 400},
	TIEM_RESOURCE_VALIDATE_DISK_ERROR:       {"validate disk info failed", 400},
	TIEM_RESOURCE_UPDATE_DISK_ERROR:         {"update disk failed", 500},
	TIEM_RESOURCE_DELETE_DISK_ERROR:         {"delete disk failed", 500},
	TIEM_RESOURCE_DISK_STILL_INUSED:         {"disk is still in used", 409},
	TIEM_RESOURCE_DISK_ALREADY_EXIST:        {"disk is already existed", 409},

	// param group & cluster param
	TIEM_DEFAULT_PARAM_GROUP_NOT_DEL:                 {"Not allow to deleted the default parameter group", 409},
	TIEM_DEFAULT_PARAM_GROUP_NOT_MODIFY:              {"Not allow to modify the default parameter group", 409},
	TIEM_MODIFY_PARAM_FAILED:                         {"Failed to modify parameter", 500},
	TIEM_CONVERT_OBJ_FAILED:                          {"Failed to convert data type of parameter", 500},
	TIEM_PARAMETER_GROUP_CREATE_ERROR:                {"Failed to create parameter group", 500},
	TIEM_PARAMETER_GROUP_UPDATE_ERROR:                {"Failed to update parameter group", 500},
	TIEM_PARAMETER_GROUP_DELETE_ERROR:                {"Failed to delete parameter group", 500},
	TIEM_PARAMETER_GROUP_QUERY_ERROR:                 {"Failed to query parameter group", 500},
	TIEM_PARAMETER_GROUP_DETAIL_ERROR:                {"Failed to get parameter group details", 500},
	TIEM_PARAMETER_GROUP_COPY_ERROR:                  {"Failed to copy parameter group", 500},
	TIEM_PARAMETER_GROUP_APPLY_ERROR:                 {"Failed to apply parameter group", 500},
	TIEM_PARAMETER_GROUP_CREATE_RELATION_PARAM_ERROR: {"Failed to create relation parameter in parameter group", 500},
	TIEM_PARAMETER_GROUP_DELETE_RELATION_PARAM_ERROR: {"Failed to delete relation parameter in parameter group", 500},
	TIEM_PARAMETER_GROUP_UPDATE_RELATION_PARAM_ERROR: {"Failed to update relation parameter in parameter group", 500},
	TIEM_CLUSTER_PARAMETER_QUERY_ERROR:               {"Failed to query cluster parameters", 500},
	TIEM_CLUSTER_PARAMETER_UPDATE_ERROR:              {"Failed to update cluster parameters", 500},
	TIEM_PARAMETER_GROUP_NAME_ALREADY_EXISTS:         {"Parameter group name already exists", 500},
	TIEM_PARAMETER_GROUP_RELATION_CLUSTER_NOT_DEL:    {"Parameter group association clusters cannot be deleted", 500},

	TIEM_PARAMETER_QUERY_ERROR:    {"Failed to query parameter by parameter group id", 500},
	TIEM_PARAMETER_CREATE_ERROR:   {"Failed to create parameter", 500},
	TIEM_PARAMETER_DELETE_ERROR:   {"Failed to delete parameter", 500},
	TIEM_PARAMETER_DETAIL_ERROR:   {"Failed to detail parameter", 500},
	TIEM_PARAMETER_UPDATE_ERROR:   {"Failed to update parameter", 500},
	TIEM_PARAMETER_ALREADY_EXISTS: {"The parameter already exists", 500},

	// change feed
	TIEM_CHANGE_FEED_NOT_FOUND:              {"Change feed task is not found", 404},
	TIEM_CHANGE_FEED_DUPLICATE_ID:           {"Duplicate id", 500},
	TIEM_CHANGE_FEED_STATUS_CONFLICT:        {"Task status conflict", 409},
	TIEM_CHANGE_FEED_LOCK_EXPIRED:           {"Task status lock expired", 409},
	TIEM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM: {"Task downstream type not supported", 500},
	TIEM_CHANGE_FEED_EXECUTE_ERROR:          {"Failed to execute task command", 500},

	TIEM_MASTER_SLAVE_SWITCHOVER_NOT_FOUND:               {"master/slave relation not found", 404},
	TIEM_MASTER_SLAVE_SWITCHOVER_FAILED:                  {"master/slave switchover failed", 500},
	TIEM_MASTER_SLAVE_SWITCHOVER_CDC_SYNC_TASK_NOT_FOUND: {"master/slave CDC sync task not found", 400},
	TIEM_MASTER_SLAVE_SWITCHOVER_SLAVE_NO_CDC_COMPONENT:  {"slave has no CDC component", 400},

	TIEM_LOG_QUERY_FAILED: {"Failed to query cluster log", 500},
	TIEM_LOG_TIME_AFTER:   {"query log parameter startTime after endTime", 401},

	// scale out & scale in
	TIEM_INSTANCE_NOT_FOUND:               {"Instance of cluster is not found", 404},
	TIEM_CONNECT_TIDB_ERROR:               {"Failed to connect TiDB instances", 500},
	TIEM_DELETE_INSTANCE_ERROR:            {"Failed to delete cluster instance", 500},
	TIEM_CHECK_PLACEMENT_RULES_ERROR:      {"Placement rule is not set when scale out TiFlash", 409},
	TIEM_CHECK_TIFLASH_MAX_REPLICAS_ERROR: {"The number of remaining TiFlash instances is less than the maximum replicas of data tables", 409},
	TIEM_SCAN_MAX_REPLICA_COUNT_ERROR:     {"Failed to scan max replicas of data tables of TiFlash", 500},

	//product
	CreateZonesError:              {"create zone failed", 500},
	DeleteZonesError:              {"delete zone failed", 500},
	QueryZoneScanRowError:         {"query all zone failed", 401},
	CreateProductError:            {"create product failed", 500},
	DeleteProductError:            {"delete product failed", 500},
	QueryProductsScanRowError:     {"query all product failed", 401},
	QueryProductComponentProperty: {"query all component property failed", 401},
	CreateSpecsError:              {"create specs failed", 500},
	DeleteSpecsError:              {"delete specs failed", 500},
	QuerySpecScanRowError:         {"query all specs failed", 401},

	//user & tenant
	QueryUserScanRowError:              {"query all users failed", 401},
	QueryTenantScanRowError:            {"query all tenants failed", 401},
	TenantAlreadyExist:                 {"tenant already exist", 401},
	TenantNotExist:                     {"tenant not exist", 401},
	DeleteTenantFailed:                 {"delete tenant failed", 401},
	UpdateTenantOnBoardingStatusFailed: {"update tenant on boarding status failed", 401},
}
