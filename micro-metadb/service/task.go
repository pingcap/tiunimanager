package service

import (
	"context"
	"github.com/pingcap/ticp/micro-metadb/models"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"

)

var TaskSuccessResponseStatus = &dbPb.DBTaskResponseStatus{Code: 0}

func (d *DBServiceHandler) CreateFlow(ctx context.Context, req *dbPb.DBCreateFlowRequest, rsp *dbPb.DBCreateFlowResponse) error {
	flow, err := models.CreateFlow(req.Flow.FlowName, req.Flow.GetFlowName(), req.Flow.GetBizId())
	if err != nil {
		// todo

	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Flow = convertFlowToDTO(flow)
	}

	return nil
}

func (d *DBServiceHandler) CreateTask(ctx context.Context, req *dbPb.DBCreateTaskRequest, rsp *dbPb.DBCreateTaskResponse) error {
	task, err := models.CreateTask(
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
func (d *DBServiceHandler) UpdateFlow(ctx context.Context, req *dbPb.DBUpdateFlowRequest, rsp *dbPb.DBUpdateFlowResponse) error {
	flow, err := models.UpdateFlow(*parseFlowDTO(req.FlowWithTasks.Flow))
	if err != nil {
		// todo
	} else {
		tasks, err := models.BatchSaveTasks(batchParseTaskDTO(req.FlowWithTasks.Tasks))
		if err != nil {
			// todo

		} else {
			rsp.Status = TaskSuccessResponseStatus
			rsp.FlowWithTasks = &dbPb.DBFlowWithTaskDTO{
				Flow:  convertFlowToDTO(&flow),
				Tasks: batchConvertTaskToDTO(tasks),
			}
		}
	}
	return nil
}
func (d *DBServiceHandler) UpdateTask(ctx context.Context, req *dbPb.DBUpdateTaskRequest, rsp *dbPb.DBUpdateTaskResponse) error {
	task, err := models.UpdateTask(*parseTaskDTO(req.Task))
	if err != nil {
		// todo

	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Task = convertTaskToDTO(&task)
	}

	return nil
}
func (d *DBServiceHandler) LoadFlow(ctx context.Context, req *dbPb.DBLoadFlowRequest, rsp *dbPb.DBLoadFlowResponse) error {
	flow , tasks, err  := models.FetchFlowDetail(uint(req.Id))
	if err != nil {
		// todo
	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.FlowWithTasks = &dbPb.DBFlowWithTaskDTO{
			Flow:  convertFlowToDTO(flow),
			Tasks: batchConvertTaskToDTO(tasks),
		}
	}

	return nil
}
func (d *DBServiceHandler) LoadTask(ctx context.Context, req *dbPb.DBLoadTaskRequest, rsp *dbPb.DBLoadTaskResponse) error {
	task, err := models.FetchTask(uint(req.Id))
	if err != nil {
		// todo

	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Task = convertTaskToDTO(&task)
	}

	return nil
}

func convertFlowToDTO(do *models.FlowDO) (dto *dbPb.DBFlowDTO) {
	dto = &dbPb.DBFlowDTO{}
	dto.Id = int64(do.ID)
	dto.FlowName = do.FlowName
	dto.StatusAlias = do.StatusAlias
	dto.BizId = do.BizId
	dto.Status = int32(do.Status)
	dto.CreateTime = do.CreatedAt.Unix()
	dto.UpdateTime = do.UpdatedAt.Unix()

	dto.DeleteTime = DeletedAtUnix(do.DeletedAt)
	return
}

func parseFlowDTO(dto *dbPb.DBFlowDTO) (do *models.FlowDO) {
	do = &models.FlowDO{}
	do.ID = uint(dto.Id)
	do.FlowName = dto.FlowName
	do.StatusAlias = dto.StatusAlias
	do.BizId = dto.BizId
	do.Status = int8(dto.Status)
	return
}

func convertTaskToDTO(do *models.TaskDO) (dto *dbPb.DBTaskDTO) {
	dto = &dbPb.DBTaskDTO{}
	dto.Id = int64(do.ID)

	dto.BizId = do.BizId
	dto.Status = int32(do.Status)
	dto.CreateTime = do.CreatedAt.Unix()
	dto.UpdateTime = do.UpdatedAt.Unix()

	dto.DeleteTime = DeletedAtUnix(do.DeletedAt)

	dto.ParentId = do.ParentId
	dto.ParentType = int32(do.ParentType)

	dto.Parameters = do.Parameters
	dto.Result = do.Result
	dto.TaskName = do.TaskName
	dto.TaskReturnType = do.TaskReturnType

	return
}

func batchConvertTaskToDTO(dos []*models.TaskDO) (dtos []*dbPb.DBTaskDTO) {
	dtos = make([]*dbPb.DBTaskDTO, len(dos), len(dos))

	for i,v := range dos {
		dtos[i] = convertTaskToDTO(v)
	}
	return
}

func batchParseTaskDTO(dtos []*dbPb.DBTaskDTO) (dos []*models.TaskDO) {
	dos = make([]*models.TaskDO, len(dtos), len(dtos))

	for i,v := range dtos {
		dos[i] = parseTaskDTO(v)
	}
	return
}

func parseTaskDTO(dto *dbPb.DBTaskDTO) (do *models.TaskDO) {
	do = &models.TaskDO{}
	do.ID = uint(dto.Id)
	do.BizId = dto.BizId
	do.Status = int8(dto.Status)

	do.Parameters = dto.Parameters
	do.Result = dto.Result
	do.TaskName = dto.TaskName
	do.TaskReturnType = dto.TaskReturnType

	do.ParentId = dto.ParentId
	do.ParentType = int8(dto.ParentType)
	return
}
