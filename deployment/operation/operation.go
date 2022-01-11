/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: operation
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/11
*******************************************************************************/

package operation

import (
	"time"

	"github.com/pingcap-inc/tiem/util/uuidutil"
	"gorm.io/gorm"
)

// Operation Record information about each TiUP operation
type Operation struct {
	ID         string    `gorm:"primaryKey;"`
	Type       string    `gorm:"not null;comment:'second party operation of type, eg: deploy, start, stop...'"`
	WorkFlowID string    `gorm:"not null;index;comment:'workflow ID of operation'"`
	Status     Status    `gorm:"default:null"`
	Result     string    `gorm:"default:null"`
	ErrorStr   string    `gorm:"size:8192;comment:'operation error msg'"`
	CreatedAt  time.Time `gorm:"<-:create"`
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (s *Operation) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuidutil.GenerateID()
	return nil
}

type Status string

const (
	Init       Status = "init"
	Processing Status = "processing"
	Finished   Status = "finished"
	Error      Status = "error"
)
