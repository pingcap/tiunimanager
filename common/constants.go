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

/*******************************************************************************
 * @File: constants
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/3
*******************************************************************************/

package common

const SlowSqlThreshold = 100
const (
	TiEM         string = "EnterpriseManager" // ??
	LocalAddress string = "0.0.0.0"
)

const (
	RegistryMicroServicePrefix = "/micro/registry/"
	HttpProtocol               = "http://"
)

const (
	Sunday    string = "Sunday"
	Monday    string = "Monday"
	Tuesday   string = "Tuesday"
	Wednesday string = "Wednesday"
	Thursday  string = "Thursday"
	Friday    string = "Friday"
	Saturday  string = "Saturday"
)

var WeekDayMap = map[string]int{
	Sunday:    0,
	Monday:    1,
	Tuesday:   2,
	Wednesday: 3,
	Thursday:  4,
	Friday:    5,
	Saturday:  6}

//System log-related constants
const (
	LogFileSystem      string = "system"
	LogFileSecondParty string = "2rd"
	LogFileLibTiUP     string = "libTiUP"
	LogFileAccess      string = "access"
	LogFileAudit       string = "audit"
	LogDirPrefix       string = "/logs/"
)

// Enterprise Manager database constants
const (
	DBDirPrefix    string = "/"
	SqliteFileName string = "em.db"
)

// Enterprise Manager Certificates constants
const (
	CertDirPrefix string = "/cert/"
	CrtFileName   string = "server.crt"
	KeyFileName   string = "server.key"
)

// micro service default port
const (
	DefaultMicroMetaDBPort  int = 4100
	DefaultMicroClusterPort int = 4110
	DefaultMicroApiPort     int = 4116
	DefaultMicroFilePort    int = 4118
	DefaultMetricsPort      int = 4121
)

// TiDB component default port
const (
	DefaultTiDBPort       int = 4000
	DefaultTiDBStatusPort int = 10080
	DefaultPDClientPort   int = 2379
	DefaultAlertPort      int = 9093
	DefaultGrafanaPort    int = 3000
)

type EMProductNameType string

//Definition of product names provided by Enterprise manager
const (
	EMProductNameTiDB              EMProductNameType = "TiDB"
	EMProductNameDataMigration     EMProductNameType = "DataMigration"
	EMProductNameEnterpriseManager EMProductNameType = "EnterpriseManager"
)

//cluster

type ClusterStatus string

//Definition of cluster running status information
const (
	ClusterInitializing ClusterStatus = "Initializing"
	ClusterStopped      ClusterStatus = "Stopped"
	ClusterRunning      ClusterStatus = "Running"
	ClusterRecovering   ClusterStatus = "Recovering"
	ClusterMaintenance  ClusterStatus = "Maintenance"
)

type ClusterMaintenanceStatus string

// Definition cluster maintenance status information
const (
	ClusterMaintenanceCreating                     ClusterMaintenanceStatus = "Creating"
	ClusterMaintenanceCloning                      ClusterMaintenanceStatus = "Cloning"
	ClusterMaintenanceDeleting                     ClusterMaintenanceStatus = "Deleting"
	ClusterMaintenanceStopping                     ClusterMaintenanceStatus = "Stopping"
	ClusterMaintenanceRestarting                   ClusterMaintenanceStatus = "Restarting"
	ClusterMaintenanceBackingUp                    ClusterMaintenanceStatus = "BackingUp"
	ClusterMaintenanceRecovering                   ClusterMaintenanceStatus = "Recovering"
	ClusterMaintenanceScaleIn                      ClusterMaintenanceStatus = "ScaleIn"
	ClusterMaintenanceScaleOut                     ClusterMaintenanceStatus = "ScaleOut"
	ClusterMaintenanceUpgrading                    ClusterMaintenanceStatus = "Upgrading"
	ClusterMaintenanceSwitching                    ClusterMaintenanceStatus = "Switching"
	ClusterMaintenanceModifyParameterAndRestarting ClusterMaintenanceStatus = "ModifyParameterRestarting"
)

type ClusterBackupStatus string

//Definition of cluster backup status information
const (
	ClusterBackupInitializing ClusterBackupStatus = "Initializing"
	ClusterBackupProcessing   ClusterBackupStatus = "Processing"
	ClusterBackupFinished     ClusterBackupStatus = "Finished"
	ClusterBackupFailed       ClusterBackupStatus = "Failed"
)

type ClusterRelationType string

//Constants for the relationships between clusters
const (
	ClusterRelationSlaveTo     ClusterRelationType = "SlaveTo"
	ClusterRelationStandBy     ClusterRelationType = "StandBy"
	ClusterRelationCloneFrom   ClusterRelationType = "CloneFrom"
	ClusterRelationRecoverFrom ClusterRelationType = "RecoverFrom"
)

type BackupType string
type BackupMethod string

//Definition backup data method and type
const (
	BackupTypeFull      BackupType   = "full"
	BackupTypeIncrement BackupType   = "incr"
	BackupMethodLogic   BackupMethod = "logical"
	BackupMethodPhysics BackupMethod = "physical"
)

type BackupMode string

//Definition backup data mode
const (
	BackupModeAuto   BackupMode = "auto"
	BackupModeManual BackupMode = "manual"
)

type StorageType string

//Definition backup data storage type
const (
	StorageTypeLocal StorageType = "local"
	StorageTypeS3    StorageType = "s3"
	StorageTypeNFS   StorageType = "nfs"
)

//Definition export & import data constants
const (
	DefaultImportDir    string = "/tmp/em/import"
	DefaultExportDir    string = "/tmp/em/export"
	DefaultZipName      string = "data.zip"
	TransportTypeExport string = "export"
	TransportTypeImport string = "import"
)

type DataExportStatus string

//Definition data export status information
const (
	DataExportInitializing DataExportStatus = "Initializing"
	DataExportProcessing   DataExportStatus = "Processing"
	DataExportFinished     DataExportStatus = "Finished"
	DataExportFailed       DataExportStatus = "Failed"
)

type DataImportStatus string

//Definition data import status information
const (
	DataImportInitializing DataImportStatus = "Initializing"
	DataImportProcessing   DataImportStatus = "Processing"
	DataImportFinished     DataImportStatus = "Finished"
	DataImportFailed       DataImportStatus = "Failed"
)

//user status

type TenantStatus string

//Definition tenant status information
const (
	TenantStatusNormal     TenantStatus = "Normal"
	TenantStatusDeactivate TenantStatus = "Deactivate"
)

type UserStatus string

//Definition user status information
const (
	UserStatusNormal     UserStatus = "Normal"
	UserStatusDeactivate UserStatus = "Deactivate"
)

type TokenStatus string

//Definition token status information
const (
	TokenStatusNormal     UserStatus = "Normal"
	TokenStatusDeactivate UserStatus = "Deactivate"
)

// product

type ProductStatus string

//Definition product status information
const (
	ProductStatusOnline    ProductStatus = "Online"
	ProductStatusOffline   ProductStatus = "Offline"
	ProductStatusException ProductStatus = "Exception" // only TiDB Enterprise Manager
)

type ProductSpecStatus string

//Definition product spec status information
const (
	ProductSpecStatusOnline  ProductSpecStatus = "Online"
	ProductSpecStatusOffline ProductSpecStatus = "Offline"
)

type ProductUpgradePathStatus string

//Definition product spec status information
const (
	ProductUpgradePathAvailable   ProductUpgradePathStatus = "Available"
	ProductUpgradePathUnAvailable ProductUpgradePathStatus = "UnAvailable"
)

//WorkFlow

type WorkFlowStatus string

//Definition workflow status information
const (
	WorkFlowStatusInitializing = "Initializing"
	WorkFlowStatusProcessing   = "Processing"
	WorkFlowStatusFinished     = "Finished"
	WorkFlowStatusError        = "Error"
	WorkFlowStatusCanceled     = "Canceled"
)

type WorkFlowReturnType string

//Definition workflow return type information
const (
	WorkFlowReturnTypeUserTask     WorkFlowReturnType = "UserTask"
	WorkFlowReturnTypeSyncFuncTask WorkFlowReturnType = "SyncFuncTask"
	WorkFlowReturnTypeCallbackTask WorkFlowReturnType = "CallbackTask"
	WorkFlowReturnTypePollingTasK  WorkFlowReturnType = "PollingTasK"
)

// Definition workflow name
const (
	WorkFlowCreateCluster    = "CreateCluster"
	WorkFlowDeleteCluster    = "DeleteCluster"
	WorkFlowBackupCluster    = "BackupCluster"
	WorkFlowRecoverCluster   = "RecoverCluster"
	WorkFlowModifyParameters = "ModifyParameters"
	WorkFlowExportData       = "ExportData"
	WorkFlowImportData       = "ImportData"
	WorkFlowRestartCluster   = "RestartCluster"
	WorkFlowStopCluster      = "StopCluster"
	WorkFlowTakeoverCluster  = "TakeoverCluster"
	WorkFlowBuildLogConfig   = "BuildLogConfig"
	WorkFlowScaleOutCluster  = "ScaleOutCluster"
	WorkFlowScaleInCluster   = "ScaleInCluster"
)
