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
	"github.com/pingcap/tiunimanager/common/constants"
	"gorm.io/gorm"
)

// ClusterRelation Cluster relationship, the system will establish a master-slave relationship
type ClusterRelation struct {
	gorm.Model
	RelationType         constants.ClusterRelationType `gorm:"not null;size:32"`
	SubjectClusterID     string                        `gorm:"not null;size:32"`
	ObjectClusterID      string                        `gorm:"not null;size:32"`
	SyncChangeFeedTaskID string                        `gorm:"not null;size:32;default:''"`
}
