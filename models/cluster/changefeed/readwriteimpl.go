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

package changefeed

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
	"time"
)

type GormChangeFeedReadWrite struct {
	dbCommon.GormDB
}

func NewGormChangeFeedReadWrite(db *gorm.DB) *GormChangeFeedReadWrite {
	m := &GormChangeFeedReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (m *GormChangeFeedReadWrite) Create(ctx context.Context, task *ChangeFeedTask) (*ChangeFeedTask, error) {
	task.StatusLock = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	err := m.DB(ctx).Create(task).Error
	return task, dbCommon.WrapDBError(err)
}

func (m *GormChangeFeedReadWrite) Delete(ctx context.Context, taskId string) (err error) {
	task, err := m.Get(ctx, taskId)
	if err != nil {
		return err
	}

	err = m.DB(ctx).Delete(task).Error
	return dbCommon.WrapDBError(err)

}

func (m *GormChangeFeedReadWrite) LockStatus(ctx context.Context, taskId string) error {
	task, err := m.Get(ctx, taskId)
	if err != nil {
		return err
	}

	if task.Locked() {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_STATUS_CONFLICT)
	}

	err = m.DB(ctx).Model(task).
		Update("status_lock", sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}).Error
	return dbCommon.WrapDBError(err)
}


func (m *GormChangeFeedReadWrite) UnlockStatus(ctx context.Context, taskId string, targetStatus constants.ChangeFeedStatus) error {
	task, err := m.Get(ctx, taskId)
	if err != nil {
		return err
	}

	if !task.Locked() {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_LOCK_EXPIRED)
	}

	err = m.DB(ctx).Model(task).
		Update("status", targetStatus).
		Update("status_lock", sql.NullTime{
			Time:  time.Now(),
			Valid: false,
		}).Error
	return dbCommon.WrapDBError(err)
}

func (m *GormChangeFeedReadWrite) UpdateConfig(ctx context.Context, updateTemplate *ChangeFeedTask) error {
	_, err := m.Get(ctx, updateTemplate.ID)
	if err != nil {
		return err
	}

	err = m.DB(ctx).Omit("status_lock", "status", "cluster_id", "start_ts").
		Save(updateTemplate).Error
	return dbCommon.WrapDBError(err)
}

func (m *GormChangeFeedReadWrite) Get(ctx context.Context, taskId string) (*ChangeFeedTask, error) {
	if "" == taskId {
		return nil, framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "empty id")
	}

	task := &ChangeFeedTask{}
	err := m.DB(ctx).First(task, "id = ?", taskId).Error

	if err != nil {
		return nil, framework.NewTiEMError(common.TIEM_CHANGE_FEED_NOT_FOUND, fmt.Sprintf("task [%s]", taskId))
	} else {
		return task, nil
	}
}

func (m *GormChangeFeedReadWrite) QueryByClusterId(ctx context.Context, clusterId string, offset int, length int) (tasks []*ChangeFeedTask, total int64, err error) {
	if "" == clusterId {
		return nil, 0, framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "empty cluster id")
	}

	tasks = make([]*ChangeFeedTask, length)

	err = m.DB(ctx).Model(&ChangeFeedTask{}).
		Where("cluster_id = ?", clusterId).
		Order("created_at").Offset(offset).Limit(length).Find(&tasks).Count(&total).Error
	return tasks, total, dbCommon.WrapDBError(err)
}
