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
 * @File: platform.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package constants

// System config key
const (
	ConfigKeyBackupStorageType       string = "BackupStorageType"
	ConfigKeyBackupStoragePath       string = "BackupStoragePath"
	ConfigKeyBackupS3Endpoint        string = "BackupS3Endpoint"
	ConfigKeyBackupS3AccessKey       string = "BackupS3AccessKey"
	ConfigKeyBackupS3SecretAccessKey string = "BackupS3SecretAccessKey"
	ConfigKeyBackupRateLimit         string = "BackupRateLimit"
	ConfigKeyRestoreRateLimit        string = "RestoreRateLimit"
	ConfigKeyBackupConcurrency       string = "BackupConcurrency"
	ConfigKeyRestoreConcurrency      string = "RestoreConcurrency"

	ConfigKeyImportShareStoragePath string = "ImportShareStoragePath"
	ConfigKeyExportShareStoragePath string = "ExportShareStoragePath"
	ConfigKeyDumplingThreadNum      string = "DumplingThreadNum"

	ConfigTelemetrySwitch   string = "config_telemetry_switch"
	ConfigPrometheusAddress string = "config_prometheus_address"

	ConfigKeyRetainedPortRange string = "config_retained_port_range"
)

type SystemStatus string

const (
	SystemInitialing    SystemStatus = "Initialing"
	SystemServiceReady  SystemStatus = "ServiceReady"
	SystemDataReady     SystemStatus = "DataReady"
	SystemUpgrading     SystemStatus = "Upgrading"
	SystemUnserviceable SystemStatus = "Unserviceable"
	SystemRunning       SystemStatus = "Running"
	SystemFailure       SystemStatus = "Failure"
)

type SystemEvent string

const (
	SystemStarted         SystemEvent = "Started"
	SystemDataInitialized SystemEvent = "DataInitialized"
	SystemRestart         SystemEvent = "Restart"
	SystemUpgrade         SystemEvent = "Upgrade"
	SystemServe           SystemEvent = "Serve"
	SystemStop            SystemEvent = "Stop"
	SystemFailureDetected SystemEvent = "FailureDetected"
)
