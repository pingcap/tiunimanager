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
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: tiup
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/30
*******************************************************************************/

package tiup

import (
	"github.com/pingcap/tiunimanager/util/uuidutil"
	"time"

	"gorm.io/gorm"
)

// TiupConfig Record config, i.e. TIUP_HOME for different component of TiUP
type TiupConfig struct {
	ID            string    `gorm:"primaryKey;"`
	ComponentType string    `gorm:"not null;comment:'TiUP component type, eg: cluster, tiunimanager, dm, ctl;'"`
	TiupHome      string    `gorm:"not null;comment:'TiUP_HOME path'"`
	CreatedAt     time.Time `gorm:"<-:create"`
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (s *TiupConfig) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuidutil.GenerateID()
	return nil
}
