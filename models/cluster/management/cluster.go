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
	"encoding/json"

	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/common/errors"
	"github.com/pingcap-inc/tiunimanager/models/common"
	"gorm.io/gorm"
	"time"
)

type Cluster struct {
	common.Entity
	Name              string                             `gorm:"not null;size:64;uniqueIndex:uniqueName;comment:'user name of the cluster''"`
	Type              string                             `gorm:"not null;size:16;comment:'type of the cluster, eg. TiDB、TiDB Migration';"`
	Version           string                             `gorm:"not null;size:64;comment:'version of the cluster'"`
	TLS               bool                               `gorm:"default:false;comment:'whether to enable TLS, value: true or false'"`
	Tags              []string                           `gorm:"-"`
	TagInfo           string                             `gorm:"comment:'cluster tag information'"`
	OwnerId           string                             `gorm:"not null;size:32;<-:create;->"`
	ParameterGroupID  string                             `gorm:"comment: parameter group id"`
	Copies            int                                `gorm:"comment: copies"`
	Exclusive         bool                               `gorm:"comment: exclusive"`
	Vendor            string                             `gorm:"comment: vendorID"`
	Region            string                             `gorm:"comment: region location"`
	CpuArchitecture   constants.ArchType                 `gorm:"not null;type:varchar(64);comment:'user name of the cluster''"`
	MaintenanceStatus constants.ClusterMaintenanceStatus `gorm:"not null;type:varchar(64);comment:'user name of the cluster''"`
	MaintainWindow    string                             `gorm:"not null;type:varchar(64);comment:'maintain window''"`
	// only for database
	DeleteTime int64 `gorm:"uniqueIndex:uniqueName"`
}

func (t *Cluster) BeforeSave(tx *gorm.DB) (err error) {
	if t.Status == "" {
		t.Status = string(constants.ClusterInitializing)
	}

	if t.Tags == nil {
		t.Tags = make([]string, 0)
	}
	b, jsonErr := json.Marshal(t.Tags)
	if jsonErr == nil {
		t.TagInfo = string(b)
	} else {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, jsonErr.Error())
	}

	if len(t.ID) == 0 {
		return t.Entity.BeforeCreate(tx)
	}
	return nil
}

func (t *Cluster) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Model(t).Update("delete_time", time.Now().Unix())
	return nil
}

func (t *Cluster) AfterFind(tx *gorm.DB) (err error) {
	if len(t.TagInfo) > 0 {
		t.Tags = make([]string, 0)
		json.Unmarshal([]byte(t.TagInfo), &t.Tags)
	}
	return nil
}
