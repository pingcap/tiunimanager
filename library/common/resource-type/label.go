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

package resource

import (
	"time"

	"gorm.io/gorm"
)

type LabelCategory int8

const (
	UserSpecify LabelCategory = 0
	Cluster     LabelCategory = 1
	Componennt  LabelCategory = 2
	DiskPerf    LabelCategory = 3
)

type Label struct {
	Name      string `gorm:"PrimaryKey"`
	Category  int8
	Trait     int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

var DefaultLabelTypes = map[string]Label{
	string(Database):      {Name: string(Database), Category: int8(Cluster), Trait: 0x0000000000000001},
	string(DataMigration): {Name: string(DataMigration), Category: int8(Cluster), Trait: 0x0000000000000002},
	string(Compute):       {Name: string(Compute), Category: int8(Componennt), Trait: 0x0000000000000004},
	string(Storage):       {Name: string(Storage), Category: int8(Componennt), Trait: 0x0000000000000008},
	string(Dispatch):      {Name: string(Dispatch), Category: int8(Componennt), Trait: 0x0000000000000010},
	string(NvmeSSD):       {Name: string(NvmeSSD), Category: int8(DiskPerf), Trait: 0x0000000000000020},
	string(SSD):           {Name: string(SSD), Category: int8(DiskPerf), Trait: 0x0000000000000040},
	string(Sata):          {Name: string(Sata), Category: int8(DiskPerf), Trait: 0x0000000000000080},
}
