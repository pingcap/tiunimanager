/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package importexport

import (
	"github.com/pingcap/tiunimanager/models/common"
	"time"
)

// DataTransportRecord import and export record information
type DataTransportRecord struct {
	common.Entity
	ClusterID       string `gorm:"index;"`
	TransportType   string `gorm:"not null;"`
	FilePath        string `gorm:"not null;"`
	ConfigPath      string
	ZipName         string
	StorageType     string `gorm:"not null;"`
	Comment         string
	ReImportSupport bool
	StartTime       time.Time
	EndTime         time.Time
}
