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

package account

import (
	cryrand "crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/pingcap-inc/tiem/models/common"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        string    `gorm:"primarykey"`
	TenantID  string    `gorm:"primarykey"`
	Creator   string    `gorm:"default:null;not null;"`
	Name      string    `gorm:"default:null;not null;uniqueIndex"`
	Salt      string    `gorm:"default:null;not null;"` //password
	FinalHash string    `gorm:"default:null;not null"`
	Email     string    `gorm:"default:null"`
	Phone     string    `gorm:"default:null"`
	Status    string    `gorm:"not null;"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type Tenant struct {
	ID               string    `gorm:"primarykey"`
	Creator          string    `gorm:"default:null;not null"`
	Name             string    `gorm:"default:null;not null"`
	Status           string    `gorm:"default:null;not null"`
	OnBoardingStatus string    `gorm:"default:null;not null"`
	MaxCluster       int32     `gorm:"default:1024;"`
	MaxCPU           int32     `gorm:"default:102400"`
	MaxMemory        int32     `gorm:"default:1024000"`
	MaxStorage       int32     `gorm:"default:10240000"`
	CreatedAt        time.Time `gorm:"<-:create"`
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt
}

func (user *User) GenSaltAndHash(passwd string) error {
	b := make([]byte, 16)
	_, err := cryrand.Read(b)

	if err != nil {
		return err
	}

	salt := base64.URLEncoding.EncodeToString(b)

	finalSalt, err := common.FinalHash(salt, passwd)

	if err != nil {
		return err
	}

	user.Salt = salt
	user.FinalHash = string(finalSalt)

	return nil
}

func (user *User) CheckPassword(passwd string) (bool, error) {
	if passwd == "" {
		return false, errors.New("password cannot be empty")
	}
	if len(passwd) > 20 {
		return false, errors.New("password is too long")
	}
	s := user.Salt + passwd

	err := bcrypt.CompareHashAndPassword([]byte(user.FinalHash), []byte(s))

	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}
