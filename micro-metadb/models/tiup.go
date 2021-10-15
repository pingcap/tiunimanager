package models

import (
	"context"
	"time"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"

	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"

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

func CreateTiupTask(db *gorm.DB, ctx context.Context, taskType dbPb.TiupTaskType, bizID uint64) (id uint64, err error) {
	t := TiupTask{
		Type:     int(taskType),
		Status:   int(dbPb.TiupTaskStatus_Init),
		ErrorStr: "",
		BizID:    bizID,
	}
	log := framework.LogForkFile(common.LogFileLibTiup).WithField("models", "CreateTiupTask").WithField("TiupTask", t)
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

func UpdateTiupTaskStatus(db *gorm.DB, ctx context.Context, id uint64, taskStatus dbPb.TiupTaskStatus, errStr string) error {
	t := TiupTask{
		ID: id,
	}
	log := framework.LogForkFile(common.LogFileLibTiup).WithField("models", "UpdateTiupTaskStatus").WithField("TiupTask", t)
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
	log := framework.LogForkFile(common.LogFileLibTiup).WithField("models", "FindTiupTaskByID").WithField("id", id)
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
	log := framework.LogForkFile(common.LogFileLibTiup).WithField("models", "FindTiupTasksByBizID").WithField("bizID", bizID)
	log.Debug("entry")
	err = db.Where(&TiupTask{BizID: bizID}).Find(&tasks).Error

	if err != nil {
		getLogger().Error("err:", err, "t:", tasks)
	} else {
		getLogger().Info("err:", err, "t:", tasks)
	}
	return
}
