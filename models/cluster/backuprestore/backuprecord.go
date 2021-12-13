/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package backuprestore

import (
	"github.com/pingcap-inc/tiem/models/common"
	"time"
)

// BackupRecord backup record information
type BackupRecord struct {
	common.Entity
	StorageType  string `gorm:"not null"`
	ClusterID    string `gorm:"not null;type:varchar(22);default:null"`
	BackupType   string
	BackupMethod string
	BackupMode   string
	FilePath     string
	Size         uint64
	BackupTso    uint64
	StartTime    time.Time
	EndTime      time.Time
}
