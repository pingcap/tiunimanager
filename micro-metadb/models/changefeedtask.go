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

package models

import (
	"context"
	"database/sql"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"gorm.io/gorm"
	"time"
)

type ChangeFeedTask struct {
	Entity
	Name              string       `gorm:"type:varchar(32)"`
	ClusterId         string       `gorm:"not null;type:varchar(22);index"`
	DownstreamType    string       `gorm:"not null;type:varchar(16)"`
	StartTS           int64        `gorm:"column:start_ts"`
	FilterRulesConfig string       `gorm:"type:text"`
	DownstreamConfig  string       `gorm:"type:text"`
	StatusLock        sql.NullTime `gorm:"column:status_lock"`
}

func (t ChangeFeedTask) Locked() bool {
	return t.StatusLock.Valid &&
		t.StatusLock.Time.Add(time.Minute).After(time.Now())
}

type DAOChangeFeedManager struct {
	db *gorm.DB
}

func NewDAOChangeFeedManager(d *gorm.DB) *DAOChangeFeedManager {
	m := new(DAOChangeFeedManager)
	m.SetDb(d)
	return m
}

func (m *DAOChangeFeedManager) SetDb(db *gorm.DB) {
	m.db = db
}

func (m *DAOChangeFeedManager) Db(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

func (m *DAOChangeFeedManager) Create(ctx context.Context, task *ChangeFeedTask) (*ChangeFeedTask, error) {
	task.StatusLock = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	return task, m.Db(ctx).Create(task).Error
}

func (m *DAOChangeFeedManager) Delete(ctx context.Context, taskId string) (err error) {
	if "" == taskId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}
	task := &ChangeFeedTask{}

	return m.Db(ctx).First(task, "id = ?", taskId).Delete(task).Error
}

func (m *DAOChangeFeedManager) LockStatus(ctx context.Context, taskId string) error {
	if "" == taskId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	task := &ChangeFeedTask{}
	err := m.Db(ctx).First(task, "id = ?", taskId).Error

	if err != nil {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_TASK_NOT_FOUND)
	}

	if task.Locked() {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_TASK_STATUS_CONFLICT)
	}

	return m.Db(ctx).Model(task).
		Update("status_lock", sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}).Error
}

func (m *DAOChangeFeedManager) UnlockStatus(ctx context.Context, taskId string, targetStatus int8) error {
	if "" == taskId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	task := &ChangeFeedTask{}
	err := m.Db(ctx).First(task, "id = ?", taskId).Error

	if err != nil {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_TASK_NOT_FOUND)
	}

	if !task.Locked() {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_TASK_LOCK_EXPIRED)
	}

	return m.Db(ctx).Model(task).
		Update("status", targetStatus).
		Update("status_lock", sql.NullTime{
			Time:  time.Now(),
			Valid: false,
		}).Error
}

func (m *DAOChangeFeedManager) UpdateConfig(ctx context.Context, updateTemplate ChangeFeedTask) error {
	if "" == updateTemplate.ID {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	targetTask := &ChangeFeedTask{}
	err := m.Db(ctx).First(&targetTask, "id = ?", updateTemplate.ID).Error
	if err != nil {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_TASK_NOT_FOUND)
	}

	return m.Db(ctx).Model(&targetTask).
		Omit("status_lock", "status", "cluster_id").
		Updates(updateTemplate).Error
}

func (m *DAOChangeFeedManager) Get(ctx context.Context, taskId string) (*ChangeFeedTask, error) {
	if "" == taskId {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	task := &ChangeFeedTask{}
	err := m.Db(ctx).First(task, "id = ?", taskId).Error

	if err != nil {
		return nil, framework.SimpleError(common.TIEM_CHANGE_FEED_TASK_NOT_FOUND)
	} else {
		return task, nil
	}
}

func (m *DAOChangeFeedManager) QueryByClusterId(ctx context.Context, clusterId string, offset int, length int) (tasks []*ChangeFeedTask, total int64, err error) {
	if "" == clusterId {
		return nil, 0, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	tasks = make([]*ChangeFeedTask, length)

	return tasks, total, m.Db(ctx).Table(TABLE_NAME_CHANGE_FEED_TASKS).
		Where("cluster_id = ?", clusterId).
		Order("created_at").Offset(offset).Limit(length).Find(&tasks).Count(&total).Error
}