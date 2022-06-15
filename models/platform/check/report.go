/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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
 * @File: report
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/14
*******************************************************************************/

package check

import (
	"github.com/pingcap/tiunimanager/util/uuidutil"
	"gorm.io/gorm"
	"time"
)

type CheckReport struct {
	ID        string    `gorm:"primarykey"`
	Report    string    `gorm:"default:null;not null;"`
	Creator   string    `gorm:"default:null;not null;"`
	Type      string    `gorm:"default:null;"`
	Status    string    `gorm:"default:null;not null;"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (report *CheckReport) BeforeCreate(tx *gorm.DB) (err error) {
	if len(report.ID) == 0 {
		report.ID = uuidutil.GenerateID()
	}

	return nil
}
