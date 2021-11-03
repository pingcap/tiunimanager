
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
	"database/sql"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
)

var TaskSuccessResponseStatus = &dbpb.DBTaskResponseStatus{Code: 0}

func (handler *DBServiceHandler) CreateFlow(ctx context.Context, req *dbpb.DBCreateFlowRequest, rsp *dbpb.DBCreateFlowResponse) error {
	db := handler.Dao().Db().WithContext(ctx)
	flow, err := models.CreateFlow(db, req.Flow.FlowName, req.Flow.StatusAlias, req.Flow.GetBizId(), req.Flow.GetOperator())
	if err != nil {
		// todo

	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Flow = convertFlowToDTO(flow)
	}

	return nil
}

func (handler *DBServiceHandler) CreateTask(ctx context.Context, req *dbpb.DBCreateTaskRequest, rsp *dbpb.DBCreateTaskResponse) error {
	db := handler.Dao().Db().WithContext(ctx)
	task, err := models.CreateTask(db,
		int8(req.Task.ParentType),
		req.Task.ParentId,
		req.Task.TaskName,
		req.Task.BizId,
		req.Task.TaskReturnType,
		req.Task.Parameters,
		req.Task.Result,
	)
	if err != nil {
		// todo
	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Task = convertTaskToDTO(task)
	}

	return nil
}

func (handler *DBServiceHandler) UpdateFlow(ctx context.Context, req *dbpb.DBUpdateFlowRequest, rsp *dbpb.DBUpdateFlowResponse) error {
	db := handler.Dao().Db().WithContext(ctx)
	flow, err := models.UpdateFlowStatus(db, *parseFlowDTO(req.FlowWithTasks.Flow))
	if err != nil {
		// todo
	} else {
		tasks, err := models.BatchSaveTasks(db, batchParseTaskDTO(req.FlowWithTasks.Tasks))
		if err != nil {
			// todo

		} else {
			rsp.Status = TaskSuccessResponseStatus
			rsp.FlowWithTasks = &dbpb.DBFlowWithTaskDTO{
				Flow:  convertFlowToDTO(&flow),
				Tasks: batchConvertTaskToDTO(tasks),
			}
		}
	}
	return nil
}

func (handler *DBServiceHandler) UpdateTask(ctx context.Context, req *dbpb.DBUpdateTaskRequest, rsp *dbpb.DBUpdateTaskResponse) error {
	db := handler.Dao().Db().WithContext(ctx)
	task, err := models.UpdateTask(db, *parseTaskDTO(req.Task))
	if err != nil {
		// todo

	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Task = convertTaskToDTO(&task)
	}

	return nil
}

func (handler *DBServiceHandler) LoadFlow(ctx context.Context, req *dbpb.DBLoadFlowRequest, rsp *dbpb.DBLoadFlowResponse) error {
	db := handler.Dao().Db().WithContext(ctx)
	flow, tasks, err := models.FetchFlowDetail(db, uint(req.Id))
	if err != nil {
		// todo
	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.FlowWithTasks = &dbpb.DBFlowWithTaskDTO{
			Flow:  convertFlowToDTO(flow),
			Tasks: batchConvertTaskToDTO(tasks),
		}
	}

	return nil
}

func (handler *DBServiceHandler) LoadTask(ctx context.Context, req *dbpb.DBLoadTaskRequest, rsp *dbpb.DBLoadTaskResponse) error {
	db := handler.Dao().Db().WithContext(ctx)
	task, err := models.FetchTask(db, uint(req.Id))
	if err != nil {
		// todo

	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Task = convertTaskToDTO(&task)
	}

	return nil
}

func (handler *DBServiceHandler) ListFlows(ctx context.Context, req *dbpb.DBListFlowsRequest, rsp *dbpb.DBListFlowsResponse) error {
	db := handler.Dao().Db().WithContext(ctx)
	flows, total, err := models.ListFlows(db, req.BizId, req.Keyword, int(req.Status), int(req.Page.Page-1)*int(req.Page.PageSize), int(req.Page.PageSize))
	if nil == err {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Page = &dbpb.DBTaskPageDTO{
			Page:     req.Page.Page,
			PageSize: req.Page.PageSize,
			Total:    int32(total),
		}
		flowDTOs := make([]*dbpb.DBFlowDTO, len(flows))
		for i, v := range flows {
			flowDTOs[i] = convertFlowToDTO(v)
		}
		rsp.Flows = flowDTOs
		framework.Log().Infof("ListFlows successful, total: %d", total)
	} else {
		framework.Log().Infof("ListFlows failed, error: %s", err.Error())
	}
	return err
}

func convertFlowToDTO(do *models.FlowDO) (dto *dbpb.DBFlowDTO) {
	if do == nil {
		return
	}
	dto = &dbpb.DBFlowDTO{}
	dto.Id = int64(do.ID)
	dto.FlowName = do.Name
	dto.StatusAlias = do.StatusAlias
	dto.BizId = do.BizId
	dto.Status = int32(do.Status)
	dto.CreateTime = do.CreatedAt.Unix()
	dto.UpdateTime = do.UpdatedAt.Unix()
	dto.Operator = do.Operator

	dto.DeleteTime = nullTimeUnix(sql.NullTime(do.DeletedAt))
	return
}

func parseFlowDTO(dto *dbpb.DBFlowDTO) (do *models.FlowDO) {
	if dto == nil {
		return
	}
	do = &models.FlowDO{}
	do.ID = uint(dto.Id)
	do.Name = dto.FlowName
	do.StatusAlias = dto.StatusAlias
	do.BizId = dto.BizId
	do.Status = int8(dto.Status)
	return
}

func convertTaskToDTO(do *models.TaskDO) (dto *dbpb.DBTaskDTO) {
	if do == nil {
		return
	}
	dto = &dbpb.DBTaskDTO{}
	dto.Id = int64(do.ID)

	dto.BizId = do.BizId
	dto.Status = int32(do.Status)
	dto.StartTime = do.CreatedAt.Unix()
	dto.EndTime = do.UpdatedAt.Unix()

	dto.ParentId = do.ParentId
	dto.ParentType = int32(do.ParentType)

	dto.StartTime = nullTimeUnix(do.StartTime)
	dto.EndTime = nullTimeUnix(do.EndTime)
	dto.Parameters = do.Parameters
	dto.Result = do.Result
	dto.TaskName = do.Name
	dto.TaskReturnType = do.ReturnType

	return
}

func batchConvertTaskToDTO(dos []*models.TaskDO) (dtos []*dbpb.DBTaskDTO) {
	dtos = make([]*dbpb.DBTaskDTO, len(dos))

	for i, v := range dos {
		dtos[i] = convertTaskToDTO(v)
	}
	return
}

func batchParseTaskDTO(dtos []*dbpb.DBTaskDTO) (dos []*models.TaskDO) {
	dos = make([]*models.TaskDO, len(dtos))

	for i, v := range dtos {
		dos[i] = parseTaskDTO(v)
	}
	return
}

func parseTaskDTO(dto *dbpb.DBTaskDTO) (do *models.TaskDO) {
	if dto == nil {
		return
	}
	do = &models.TaskDO{}
	do.ID = uint(dto.Id)
	do.BizId = dto.BizId
	do.Status = int8(dto.Status)

	do.Parameters = dto.Parameters
	do.Result = dto.Result
	do.Name = dto.TaskName
	do.ReturnType = dto.TaskReturnType

	do.ParentId = dto.ParentId
	do.ParentType = int8(dto.ParentType)
	return
}
