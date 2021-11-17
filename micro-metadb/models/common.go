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
 *                                                                            *
 ******************************************************************************/

package models

import (
	"context"
	"errors"
	"time"

	gormLog "gorm.io/gorm/logger"

	"github.com/mozillazg/go-pinyin"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	TABLE_NAME_CLUSTER            = "clusters"
	TABLE_NAME_DEMAND_RECORD      = "demand_records"
	TABLE_NAME_ACCOUNT            = "accounts"
	TABLE_NAME_TENANT             = "tenants"
	TABLE_NAME_ROLE               = "roles"
	TABLE_NAME_ROLE_BINDING       = "role_bindings"
	TABLE_NAME_PERMISSION         = "permissions"
	TABLE_NAME_PERMISSION_BINDING = "permission_bindings"
	TABLE_NAME_TOKEN              = "tokens"
	TABLE_NAME_TASK               = "tasks"
	TABLE_NAME_HOST               = "hosts"
	TABLE_NAME_DISK               = "disks"
	TABLE_NAME_USED_COMPUTE       = "used_computes"
	TABLE_NAME_USED_PORT          = "used_ports"
	TABLE_NAME_USED_DISK          = "used_disks"
	TABLE_NAME_LABEL              = "labels"
	TABLE_NAME_TIUP_CONFIG        = "tiup_configs"
	TABLE_NAME_TIUP_TASK          = "tiup_tasks"
	TABLE_NAME_FLOW               = "flows"
	TABLE_NAME_PARAMETERS_RECORD  = "parameters_records"
	TABLE_NAME_BACKUP_RECORD      = "backup_records"
	TABLE_NAME_BACKUP_STRATEGY    = "backup_strategies"
	TABLE_NAME_TRANSPORT_RECORD   = "transport_records"
	TABLE_NAME_RECOVER_RECORD     = "recover_records"
	TABLE_NAME_COMPONENT_INSTANCE = "component_instances"
)

func getLogger() *log.Entry {
	return framework.Log()
}

type Entity struct {
	ID        string    `gorm:"primaryKey;"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Code     string `gorm:"uniqueIndex;default:null;not null;<-:create"`
	TenantId string `gorm:"default:null;not null;<-:create"`
	Status   int8   `gorm:"type:SMALLINT;default:0"`
}

func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuidutil.GenerateID()
	if e.Code == "" {
		e.Code = e.ID
	}
	e.Status = 0

	if len(e.Code) > 128 {
		return errors.New("entity code is too long, code = " + e.Code)
	}

	return nil
}

type Record struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	TenantId string `gorm:"default:null;not null;<-:create"`
}

type Data struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	BizId  string `gorm:"default:null;<-:create"`
	Status int8   `gorm:"type:SMALLINT;default:0"`
}

var split = []byte("_")

func generateEntityCode(name string) string {
	a := pinyin.NewArgs()

	a.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}

	code := pinyin.Pinyin(name, a)

	bytes := make([]byte, 0, len(name)*4)

	// split code with "_" before and after chinese word
	previousSplitFlag := false
	for _, v := range code {
		currentWord := v[0]

		if previousSplitFlag || len(currentWord) > 1 {
			bytes = append(bytes, split...)
		}

		bytes = append(bytes, []byte(currentWord)...)
		previousSplitFlag = len(currentWord) > 1
	}
	return string(bytes)
}

type DaoLogger struct {
	p             framework.Framework
	SlowThreshold time.Duration
}

// LogMode log mode
func (l *DaoLogger) LogMode(level gormLog.LogLevel) gormLog.Interface {
	return l
}

// Info print info
func (l DaoLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.p.LogWithContext(ctx).Infof(msg)
}

// Warn print warn messages
func (l DaoLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.p.LogWithContext(ctx).Warn(msg)
}

// Error print error messages
func (l DaoLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.p.LogWithContext(ctx).Error(msg)
}

// Trace print sql message
func (l DaoLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	logger := l.p.LogWithContext(ctx).WithField("rows", rows)

	elapsed := time.Since(begin).Milliseconds()

	if l.p.GetRootLogger().LogLevel == framework.LogDebug || l.p.GetRootLogger().LogLevel == framework.LogInfo {
		logger.Infof("execute sql : %s", sql)
	}

	if err != nil && (!errors.Is(err, gormLog.ErrRecordNotFound)) {
		logger.Errorf("sql error, sql : %s, err : %s", sql, err)
	}

	if elapsed > l.SlowThreshold.Milliseconds() {
		logger.Warnf("slow sql, cost %d milliseconds, sql : %s", elapsed, sql)
	}

}
