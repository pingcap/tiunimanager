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

package management

import (
	"gorm.io/gorm"
	"time"
)

type DBUser struct {
	gorm.Model
	ClusterID                string `gorm:"not null;type:varchar(22);default:null"`
	Name                     string `gorm:"default:null;not null;comment:'name of the user'"`
	Password                 string `gorm:"not null;size:64;comment:'password of the user'"`
	RoleType                 string `gorm:"not null;size:64;comment:'role type of the user'"`
	LastPasswordGenerateTime time.Time
}
