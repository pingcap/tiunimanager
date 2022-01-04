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

	ConfigKeyImportShareStoragePath string = "ImportShareStoragePath"
	ConfigKeyExportShareStoragePath string = "ExportShareStoragePath"

	ConfigTelemetrySwitch   string = "config_telemetry_switch"
	ConfigPrometheusAddress string = "config_prometheus_address"
)
