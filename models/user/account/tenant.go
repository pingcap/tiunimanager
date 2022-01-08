/*
 * Copyright (c)  2022 PingCAP, Inc.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*******************************************************************************
 * @File: tenant_api.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/4
*******************************************************************************/

package account

import (
	"gorm.io/gorm"
	"time"
)

type Tenant struct {
	ID               string    `gorm:"default:null;not null"`
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
