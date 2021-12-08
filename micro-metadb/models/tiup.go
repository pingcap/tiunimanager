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
	"time"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"

	"gorm.io/gorm"
)

type TiupTask struct {
	ID        uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	BizID     uint64         `gorm:"index"`
	Type      int
	Status    int
	ErrorStr  string `gorm:"size:8192"`
}

func (t TiupTask) TableName() string {
	return "tiup_task"
}

type TiupTaskStatus string
type TiupTaskType string

const (
	TaskStatusInit       TiupTaskStatus = "INIT"
	TaskStatusProcessing TiupTaskStatus = "PROCESSING"
	TaskStatusFinished   TiupTaskStatus = "FINISHED"
	TaskStatusError      TiupTaskStatus = "ERROR"
)

const (
	TiupTaskTypeDeploy  TiupTaskType = "DEPLOY"
	TiupTaskTypeDestroy TiupTaskType = "DESTROY"
)

func CreateTiupTask(db *gorm.DB, ctx context.Context, taskType dbpb.TiupTaskType, bizID uint64) (id uint64, err error) {
	t := TiupTask{
		Type:     int(taskType),
		Status:   int(dbpb.TiupTaskStatus_Init),
		ErrorStr: "",
		BizID:    bizID,
	}
	log := framework.LogForkFile(common.LogFileLibTiUP).WithField("models", "CreateTiupTask").WithField("TiupTask", t)
	log.Debug("entry")
	err = db.Select("Type", "Status", "ErrorStr", "BizID").Create(&t).Error
	id = t.ID
	if err != nil {
		getLogger().Error("err:", err, "t:", t)
	} else {
		getLogger().Info("err:", err, "t:", t)
	}
	return id, err
}

func UpdateTiupTaskStatus(db *gorm.DB, ctx context.Context, id uint64, taskStatus dbpb.TiupTaskStatus, errStr string) error {
	t := TiupTask{
		ID: id,
	}
	log := framework.LogForkFile(common.LogFileLibTiUP).WithField("models", "UpdateTiupTaskStatus").WithField("TiupTask", t)
	log.Debug("entry")
	err := db.Model(&t).Updates(map[string]interface{}{"Status": taskStatus, "ErrorStr": errStr}).Error
	if err != nil {
		getLogger().Error("err:", err, "t:", t)
	} else {
		getLogger().Info("err:", err, "t:", t)
	}
	return err
}

func FindTiupTaskByID(db *gorm.DB, ctx context.Context, id uint64) (task TiupTask, err error) {
	log := framework.LogForkFile(common.LogFileLibTiUP).WithField("models", "FindTiupTaskByID").WithField("id", id)
	log.Debug("entry")
	err = db.First(&task, id).Error
	if err != nil {
		getLogger().Error("err:", err, "t:", &task)
	} else {
		getLogger().Info("err:", err, "t:", &task)
	}
	return
}

func FindTiupTasksByBizID(db *gorm.DB, ctx context.Context, bizID uint64) (tasks []TiupTask, err error) {
	log := framework.LogForkFile(common.LogFileLibTiUP).WithField("models", "FindTiupTasksByBizID").WithField("bizID", bizID)
	log.Debug("entry")
	err = db.Where(&TiupTask{BizID: bizID}).Find(&tasks).Error

	if err != nil {
		getLogger().Error("err:", err, "t:", tasks)
	} else {
		getLogger().Info("err:", err, "t:", tasks)
	}
	return
}

type TiupOperator struct {
	ID        uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	BizID     string         `gorm:"index"`
	Type      int
	Status    int
	ErrorStr  string `gorm:"size:8192"`
}

func (t TiupOperator) TableName() string {
	return "tiup_operator"
}

func CreateTiupOperatorRecord(db *gorm.DB, ctx context.Context, taskType dbpb.TiupTaskType, bizID string) (id uint64, err error) {
	t := TiupOperator{
		Type:     int(taskType),
		Status:   int(dbpb.TiupTaskStatus_Init),
		ErrorStr: "",
		BizID:    bizID,
	}
	log := framework.LogForkFile(common.LogFileLibTiUP).WithField("models", "CreateTiupOperatorRecord").WithField("TiupOperator", t)
	log.Debug("entry")
	err = db.Select("Type", "Status", "ErrorStr", "BizID").Create(&t).Error
	id = t.ID
	if err != nil {
		getLogger().Error("err:", err, "t:", t)
	} else {
		getLogger().Info("err:", err, "t:", t)
	}
	return id, err
}

func UpdateTiupOperatorRecordStatus(db *gorm.DB, ctx context.Context, id uint64, taskStatus dbpb.TiupTaskStatus, errStr string) error {
	t := TiupOperator{
		ID: id,
	}
	log := framework.LogForkFile(common.LogFileLibTiUP).WithField("models", "UpdateTiupOperatorRecordStatus").WithField("TiupOperator", t)
	log.Debug("entry")
	err := db.Model(&t).Updates(map[string]interface{}{"Status": taskStatus, "ErrorStr": errStr}).Error
	if err != nil {
		getLogger().Error("err:", err, "t:", t)
	} else {
		getLogger().Info("err:", err, "t:", t)
	}
	return err
}

func FindTiupOperatorRecordByID(db *gorm.DB, ctx context.Context, id uint64) (task TiupTask, err error) {
	log := framework.LogForkFile(common.LogFileLibTiUP).WithField("models", "FindTiupOperatorRecordByID").WithField("id", id)
	log.Debug("entry")
	err = db.First(&task, id).Error
	if err != nil {
		getLogger().Error("err:", err, "t:", &task)
	} else {
		getLogger().Info("err:", err, "t:", &task)
	}
	return
}

func FindTiupOperatorRecordsByBizID(db *gorm.DB, ctx context.Context, bizID string) (tasks []TiupTask, err error) {
	log := framework.LogForkFile(common.LogFileLibTiUP).WithField("models", "FindTiupOperatorRecordsByBizID").WithField("bizID", bizID)
	log.Debug("entry")
	err = db.Where(&TiupOperator{BizID: bizID}).Find(&tasks).Error

	if err != nil {
		getLogger().Error("err:", err, "t:", tasks)
	} else {
		getLogger().Info("err:", err, "t:", tasks)
	}
	return
}
