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

package common

import (
	"context"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"time"

	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"gorm.io/gorm"
)

type Entity struct {
	ID        string    `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	TenantId string `gorm:"default:null;not null;<-:create"`
	Status   string `gorm:"not null;"`
}

func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuidutil.GenerateID()
	return nil
}

var split = []byte("_")

type GormDB struct {
	db *gorm.DB
}

func WrapDB(db *gorm.DB) GormDB {
	return GormDB{db: db}
}

func (m *GormDB) DB(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

// WrapDBError
// @Description:
// @Parameter err
// @return error is nil or TiEMError
func WrapDBError(err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case framework.TiEMError:
		return err
	default:
		return framework.NewTiEMErrorf(common.TIEM_UNRECOGNIZED_ERROR, err.Error())
	}
}