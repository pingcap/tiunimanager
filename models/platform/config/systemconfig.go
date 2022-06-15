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

package config

import (
	"github.com/pingcap/tiunimanager/util/uuidutil"
	"gorm.io/gorm"
	"time"
)

type SystemConfig struct {
	ID          string    `gorm:"primarykey"`
	CreatedAt   time.Time `gorm:"<-:create"`
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	ConfigKey   string         `gorm:"not null;uniqueIndex"`
	ConfigValue string         `gorm:"not null;"`
}

func (e *SystemConfig) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuidutil.GenerateID()
	return nil
}
