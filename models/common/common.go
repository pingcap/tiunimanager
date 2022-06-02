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
	"database/sql/driver"
	"encoding/json"
	"github.com/pingcap-inc/tiem/common/constants"
	crypto "github.com/pingcap-inc/tiem/util/encrypt"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"golang.org/x/crypto/bcrypt"
	"time"

	"github.com/pingcap-inc/tiem/common/errors"
	"gorm.io/gorm"
)

type Entity struct {
	ID        string    `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	TenantId string `gorm:"default:null;not null;<-:create"`
	Status   string `gorm:"not null;"`
}

type Password string

// Scan implements the Scanner interface.
func (p *Password) Scan(value interface{}) error {
	enc, err := crypto.AesDecryptCFB(value.(string))
	if err != nil {
		return err
	}
	*p = Password(enc)
	return nil
}

// Value implements the driver Valuer interface.
func (p Password) Value() (driver.Value, error) {
	return crypto.AesEncryptCFB(string(p))
}

type PasswordInExpired struct {
	Val        string    // Value is the value of the password
	UpdateTime time.Time // UpdateTime is the last update time
}

// Scan implements the Scanner interface.
func (p *PasswordInExpired) Scan(value interface{}) error {
	if value == nil {
		p.Val, p.UpdateTime = "", time.Time{}
		return nil
	}
	val := []byte(value.(string))
	err := json.Unmarshal(val, p)
	if err != nil {
		return err
	}
	dec, err := crypto.AesDecryptCFB(p.Val)

	if err != nil {
		return err
	}
	p.Val = dec
	return nil
}

// Value implements the driver Valuer interface.
func (p PasswordInExpired) Value() (driver.Value, error) {
	enc, err := crypto.AesEncryptCFB(p.Val)
	if err != nil {
		return nil, err
	}
	p.Val = enc
	res, err := json.Marshal(p)
	return string(res), err
}

// CheckUpdateTimeExpired
// @Description: check if the update time of the password is expired
// @Parameter
// @return bool
func (p PasswordInExpired) CheckUpdateTimeExpired() (bool, error) {
	duration := time.Since(p.UpdateTime)
	return duration > constants.ExpirationTime, nil
}

func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	if len(e.ID) == 0 {
		e.ID = uuidutil.GenerateID()
	}

	return nil
}

type GormDB struct {
	db *gorm.DB
}

func WrapDB(db *gorm.DB) GormDB {
	return GormDB{db: db}
}

type ctxTransactionKeyStruct struct{}

var ctxTransactionKey = ctxTransactionKeyStruct{}

func CtxWithTransaction(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxTransactionKey, db)
}

func (m *GormDB) DB(ctx context.Context) *gorm.DB {
	iface := ctx.Value(ctxTransactionKey)

	if iface != nil {
		tx, ok := iface.(*gorm.DB)
		if !ok {
			return nil
		}

		return tx
	}

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
	case errors.EMError:
		return err
	default:
		return errors.NewError(errors.TIEM_SQL_ERROR, err.Error())
	}
}

func FinalHash(salt string, password string) ([]byte, error) {
	if password == "" {
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "password cannot be empty")
	}
	s := salt + password
	finalSalt, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)

	return finalSalt, err
}
