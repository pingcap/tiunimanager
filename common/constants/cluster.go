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
 ******************************************************************************/

/*******************************************************************************
 * @File: cluster.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package constants

type ClusterRunningStatus string

//Definition of cluster running status information
const (
	ClusterInitializing ClusterRunningStatus = "Initializing"
	ClusterStopped      ClusterRunningStatus = "Stopped"
	ClusterRunning      ClusterRunningStatus = "Running"
	ClusterRecovering   ClusterRunningStatus = "Recovering"
	ClusterFailure      ClusterRunningStatus = "Failure"
)

type ClusterMaintenanceStatus string

// Definition cluster maintenance status information
const (
	ClusterMaintenanceCreating                     ClusterMaintenanceStatus = "Creating"
	ClusterMaintenanceCloning                      ClusterMaintenanceStatus = "Cloning"
	ClusterMaintenanceDeleting                     ClusterMaintenanceStatus = "Deleting"
	ClusterMaintenanceStopping                     ClusterMaintenanceStatus = "Stopping"
	ClusterMaintenanceRestarting                   ClusterMaintenanceStatus = "Restarting"
	ClusterMaintenanceBackUp                       ClusterMaintenanceStatus = "BackUp"
	ClusterMaintenanceRestore                      ClusterMaintenanceStatus = "Restore"
	ClusterMaintenanceScaleIn                      ClusterMaintenanceStatus = "ScaleIn"
	ClusterMaintenanceScaleOut                     ClusterMaintenanceStatus = "ScaleOut"
	ClusterMaintenanceUpgrading                    ClusterMaintenanceStatus = "Upgrading"
	ClusterMaintenanceSwitching                    ClusterMaintenanceStatus = "Switching"
	ClusterMaintenanceModifyParameterAndRestarting ClusterMaintenanceStatus = "ModifyParameterRestarting"
	ClusterMaintenanceTakeover                     ClusterMaintenanceStatus = "Takeover"
	ClusterMaintenanceNone                         ClusterMaintenanceStatus = ""
)

const (
	FlowCreateCluster       = "CreateCluster"
	FlowDeleteCluster       = "DeleteCluster"
	FlowBackupCluster       = "BackupCluster"
	FlowRestoreNewCluster   = "RestoreNewCluster"
	FlowRestoreExistCluster = "RestoreExistCluster"
	FlowModifyParameters    = "ModifyParameters"
	FlowExportData          = "ExportData"
	FlowImportData          = "ImportData"
	FlowRestartCluster      = "RestartCluster"
	FlowStopCluster         = "StopCluster"
	FlowTakeoverCluster     = "BuildForTakeover"
	FlowBuildLogConfig      = "BuildLogConfig"
	FlowScaleOutCluster     = "ScaleOutCluster"
	FlowScaleInCluster      = "ScaleInCluster"
	FlowCloneCluster        = "CloneCluster"
)

type ClusterInstanceRunningStatus string

//Definition of cluster instance running status information
const (
	ClusterInstanceInitializing ClusterInstanceRunningStatus = "Initializing"
	ClusterInstanceStopped      ClusterInstanceRunningStatus = "Stopped"
	ClusterInstanceRunning      ClusterInstanceRunningStatus = "Running"
	ClusterInstanceRecovering   ClusterInstanceRunningStatus = "Recovering"
	ClusterInstanceFailure      ClusterInstanceRunningStatus = "Failure"
)

type ClusterInstanceMaintenanceStatus string

// Definition cluster instance maintenance status information
const (
	ClusterInstanceMaintenanceCreating                     ClusterInstanceMaintenanceStatus = "Creating"
	ClusterInstanceMaintenanceDeleting                     ClusterInstanceMaintenanceStatus = "Deleting"
	ClusterInstanceMaintenanceStopping                     ClusterInstanceMaintenanceStatus = "Stopping"
	ClusterInstanceMaintenanceRestarting                   ClusterInstanceMaintenanceStatus = "Restarting"
	ClusterInstanceMaintenanceUpgrading                    ClusterInstanceMaintenanceStatus = "Upgrading"
	ClusterInstanceMaintenanceModifyParameterAndRestarting ClusterInstanceMaintenanceStatus = "ModifyParameterRestarting"
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

type ClusterCloneStrategy string

// Definition cluster clone strategy
const (
	ClusterTopologyClone ClusterCloneStrategy = "TopologyClone"
	SnapShotClone        ClusterCloneStrategy = "Snapshot"
	CDCSyncClone         ClusterCloneStrategy = "CDCSync"
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

const (
	DefaultBackupStoragePath       string = "nfs/em/backup"
	DefaultBackupS3AccessKey       string = "minioadmin"
	DefaultBackupS3SecretAccessKey string = "minioadmin"
	DefaultBackupS3Endpoint        string = "http://minio.pingcap.net:9000"
)
