/******************************************************************************
 * Copyright (c)  2023 PingCAP, Inc.                                          *
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

package sqleditorfile

import (
	"time"

	"github.com/pingcap/tiunimanager/util/uuidutil"
	"gorm.io/gorm"
)

type SqlEditorFile struct {
	ID        string    `gorm:"primarykey"`
	Name      string    `gorm:"default:null;not null;"`
	ClusterID string    `gorm:"default:null;not null;"`
	Database  string    `gorm:"default:null;not null;"`
	Content   string    `gorm:"default:null;not null;"`
	IsDeleted int       `gorm:"default:0;not null;"`
	CreatedBy string    `gorm:"default:null;not null;"`
	UpdatedBy string    `gorm:"default:null;not null;"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}

func (f *SqlEditorFile) BeforeCreate(tx *gorm.DB) (err error) {
	if len(f.ID) == 0 {
		f.ID = uuidutil.GenerateID()
	}

	return nil
}
