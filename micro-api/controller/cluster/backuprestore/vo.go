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

package backuprestore

import (
	"time"

	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type BackupRecord struct {
	ID           int64                 `json:"id"`
	ClusterId    string                `json:"clusterId"`
	StartTime    time.Time             `json:"startTime"`
	EndTime      time.Time             `json:"endTime"`
	BackupType   string                `json:"backupType"`
	BackupMethod string                `json:"backupMethod"`
	BackupMode   string                `json:"backupMode"`
	Operator     controller.Operator   `json:"operator"`
	Size         float32               `json:"size"`
	BackupTso    uint64                `json:"backupTso"`
	Status       controller.StatusInfo `json:"status"`
	FilePath     string                `json:"filePath"`
}
