/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
 * @File: parametergroup.go
 * @Description: parameter group table orm
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 14:43
*******************************************************************************/

package parametergroup

import "time"

// ParameterGroup
// @Description: parameter_group table orm
type ParameterGroup struct {
	ID             string    `gorm:"primaryKey;comment:'parameter group id'"`
	Name           string    `gorm:"uniqueIndex;not null;comment:'parameter group name'"`
	ParentID       string    `gorm:"comment:'copy parameter group id'"`
	ClusterSpec    string    `gorm:"not null;comment:'cluster specifications'"`
	HasDefault     int       `gorm:"default:1;comment:'whether it is the default parameter group. optional values: [1: default, 2: customize]'"`
	DBType         int       `gorm:"default:1;comment:'database type. optional values: [1: tidb, 2: dm]'"`
	GroupType      int       `gorm:"default:1;comment:'parameter group type. optional values: [1: cluster, 2: instance]'"`
	ClusterVersion string    `gorm:"comment:'cluster version'"`
	Note           string    `gorm:"comment:'parameter group remarks information'"`
	CreatedAt      time.Time `gorm:"<-:create"`
	UpdatedAt      time.Time
}
