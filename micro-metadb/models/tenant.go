
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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

package models

import (
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"github.com/pingcap/errors"
	"gorm.io/gorm"
	"time"
)

const (
	TENANT_STATUS_NORMAL int8 = iota
	TENANT_STATUS_DEACTIVATE
)

type Tenant struct {
	ID        string `gorm:"PrimaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Name   string `gorm:"default:null;not null"`
	Type   int8   `gorm:"default:0"`
	Status int8   `gorm:"default:0"`
}

func (e *Tenant) BeforeCreate(*gorm.DB) (err error) {
	e.ID = uuidutil.GenerateID()
	e.Status = TENANT_STATUS_NORMAL
	return nil
}

func (m *DAOAccountManager) AddTenant(name string, tenantType, status int8) (t *Tenant, err error) {
	if name == " " || status < TENANT_STATUS_NORMAL || status > TENANT_STATUS_NORMAL {
		return nil, errors.Errorf("add tenant has invalid parameter,name: %s, type: %d, status: %d", name, tenantType, status)
	}
	t = &Tenant{Type: tenantType, Name: name}
	return t, m.Db().Create(t).Error
}

func (m *DAOAccountManager) FindTenantById(tenantId string) (t *Tenant, err error) {
	if tenantId == "" {
		return nil, errors.Errorf("FindTenantByDd has invalid parameter, tenantId: %s", tenantId)
	}
	t = &Tenant{}
	return t, m.Db().Where("id = ?", tenantId).First(t).Error
}

func (m *DAOAccountManager) FindTenantByName(name string) (t *Tenant, err error) {
	if name == "" {
		return nil, errors.Errorf("FindTenantByName has invalid parameter, name: %s", name)
	}
	t = &Tenant{}
	return t, m.Db().Where("name = ?", name).First(t).Error
}
