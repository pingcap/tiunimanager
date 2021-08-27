package service

import (
	"context"
	"time"

	"github.com/pingcap-inc/tiem/micro-metadb/models"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
)

func (handler *DBServiceHandler) CreateTiupTask(ctx context.Context, req *dbPb.CreateTiupTaskRequest, rsp *dbPb.CreateTiupTaskResponse) error {
	db := handler.Dao().Db()
	id, e := models.CreateTiupTask(db,ctx, req.Type, req.BizID)
	if e == nil {
		rsp.Id = id
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}

func (handler *DBServiceHandler) UpdateTiupTask(ctx context.Context, req *dbPb.UpdateTiupTaskRequest, rsp *dbPb.UpdateTiupTaskResponse) error {
	db := handler.Dao().Db()
	e := models.UpdateTiupTaskStatus(db,ctx, req.Id, req.Status, req.ErrStr)
	if e == nil {
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}

func (handler *DBServiceHandler) FindTiupTaskByID(ctx context.Context, req *dbPb.FindTiupTaskByIDRequest, rsp *dbPb.FindTiupTaskByIDResponse) error {
	db := handler.Dao().Db()
	task, e := models.FindTiupTaskByID(db,ctx, req.Id)
	if e == nil {
		var deleteAt string
		if task.DeletedAt.Valid {
			deleteAt = task.DeletedAt.Time.String()
		} else {
			var zeroTime time.Time
			deleteAt = zeroTime.String()
		}
		rsp.TiupTask = &dbPb.TiupTask{
			ID:        task.ID,
			CreatedAt: task.CreatedAt.String(),
			UpdatedAt: task.UpdatedAt.String(),
			DeletedAt: deleteAt,
			Type:      dbPb.TiupTaskType(task.Type),
			Status:    dbPb.TiupTaskStatus(task.Status),
			ErrorStr:  task.ErrorStr,
		}
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}

func (handler *DBServiceHandler) GetTiupTaskStatusByBizID(ctx context.Context, req *dbPb.GetTiupTaskStatusByBizIDRequest,
	rsp *dbPb.GetTiupTaskStatusByBizIDResponse) error {

	db := handler.Dao().Db()
	tasks, e := models.FindTiupTasksByBizID(db,ctx, req.BizID)
	if e == nil {
		errCt := 0
		errStatStr := ""
		processingCt := 0
		for _, task := range tasks {
			if task.Status == int(dbPb.TiupTaskStatus_Finished) {
				rsp.Stat = dbPb.TiupTaskStatus_Finished
				return nil
			}
			if task.Status == int(dbPb.TiupTaskStatus_Error) {
				errStatStr = task.ErrorStr
				errCt++
				continue
			}
			if task.Status == int(dbPb.TiupTaskStatus_Processing) {
				processingCt++
				continue
			}
		}
		if len(tasks) == 0 {
			rsp.ErrCode = 2
			rsp.ErrStr = "no match record was found"
			return nil
		}
		if errCt >= len(tasks) {
			rsp.Stat = dbPb.TiupTaskStatus_Error
			rsp.StatErrStr = errStatStr
			return nil
		} else {
			if processingCt > 0 {
				rsp.Stat = dbPb.TiupTaskStatus_Processing
			} else {
				rsp.Stat = dbPb.TiupTaskStatus_Init
			}
			return nil
		}
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}
