
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

package service

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"time"

	"github.com/pingcap-inc/tiem/micro-metadb/models"
)

func (handler *DBServiceHandler) CreateTiupTask(ctx context.Context, req *dbpb.CreateTiupTaskRequest, rsp *dbpb.CreateTiupTaskResponse) error {
	db := handler.Dao().Db()
	id, e := models.CreateTiupTask(db, ctx, req.Type, req.BizID)
	if e == nil {
		rsp.Id = id
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}

func (handler *DBServiceHandler) UpdateTiupTask(ctx context.Context, req *dbpb.UpdateTiupTaskRequest, rsp *dbpb.UpdateTiupTaskResponse) error {
	db := handler.Dao().Db()
	e := models.UpdateTiupTaskStatus(db, ctx, req.Id, req.Status, req.ErrStr)
	if e == nil {
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}

func (handler *DBServiceHandler) FindTiupTaskByID(ctx context.Context, req *dbpb.FindTiupTaskByIDRequest, rsp *dbpb.FindTiupTaskByIDResponse) error {
	db := handler.Dao().Db()
	task, e := models.FindTiupTaskByID(db, ctx, req.Id)
	if e == nil {
		var deleteAt string
		if task.DeletedAt.Valid {
			deleteAt = task.DeletedAt.Time.String()
		} else {
			var zeroTime time.Time
			deleteAt = zeroTime.String()
		}
		rsp.TiupTask = &dbpb.TiupTask{
			ID:        task.ID,
			CreatedAt: task.CreatedAt.String(),
			UpdatedAt: task.UpdatedAt.String(),
			DeletedAt: deleteAt,
			Type:      dbpb.TiupTaskType(task.Type),
			Status:    dbpb.TiupTaskStatus(task.Status),
			ErrorStr:  task.ErrorStr,
		}
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}

func (handler *DBServiceHandler) GetTiupTaskStatusByBizID(ctx context.Context, req *dbpb.GetTiupTaskStatusByBizIDRequest,
	rsp *dbpb.GetTiupTaskStatusByBizIDResponse) error {

	db := handler.Dao().Db()
	tasks, e := models.FindTiupTasksByBizID(db, ctx, req.BizID)
	if e == nil {
		errCt := 0
		errStatStr := ""
		processingCt := 0
		for _, task := range tasks {
			if task.Status == int(dbpb.TiupTaskStatus_Finished) {
				rsp.Stat = dbpb.TiupTaskStatus_Finished
				return nil
			}
			if task.Status == int(dbpb.TiupTaskStatus_Error) {
				errStatStr = task.ErrorStr
				errCt++
				continue
			}
			if task.Status == int(dbpb.TiupTaskStatus_Processing) {
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
			rsp.Stat = dbpb.TiupTaskStatus_Error
			rsp.StatErrStr = errStatStr
			return nil
		} else {
			if processingCt > 0 {
				rsp.Stat = dbpb.TiupTaskStatus_Processing
			} else {
				rsp.Stat = dbpb.TiupTaskStatus_Init
			}
			return nil
		}
	} else {
		rsp.ErrCode = 1
		rsp.ErrStr = e.Error()
	}
	return nil
}
