package service

import (
	"context"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
)

var TaskSuccessResponseStatus = &dbPb.DBTaskResponseStatus{Code: 0}

func (handler *DBServiceHandler) CreateFlow(ctx context.Context, req *dbPb.DBCreateFlowRequest, rsp *dbPb.DBCreateFlowResponse) error {
	db := handler.Dao().Db()
	flow, err := models.CreateFlow(db, req.Flow.FlowName, req.Flow.StatusAlias, req.Flow.GetBizId(), req.Flow.GetOperator())
	if err != nil {
		// todo

	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Flow = convertFlowToDTO(flow)
	}

	return nil
}

func (handler *DBServiceHandler) CreateTask(ctx context.Context, req *dbPb.DBCreateTaskRequest, rsp *dbPb.DBCreateTaskResponse) error {
	db := handler.Dao().Db()
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

func (handler *DBServiceHandler) UpdateFlow(ctx context.Context, req *dbPb.DBUpdateFlowRequest, rsp *dbPb.DBUpdateFlowResponse) error {
	db := handler.Dao().Db()
	flow, err := models.UpdateFlowStatus(db, *parseFlowDTO(req.FlowWithTasks.Flow))
	if err != nil {
		// todo
	} else {
		tasks, err := models.BatchSaveTasks(db, batchParseTaskDTO(req.FlowWithTasks.Tasks))
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

func (handler *DBServiceHandler) UpdateTask(ctx context.Context, req *dbPb.DBUpdateTaskRequest, rsp *dbPb.DBUpdateTaskResponse) error {
	db := handler.Dao().Db()
	task, err := models.UpdateTask(db, *parseTaskDTO(req.Task))
	if err != nil {
		// todo

	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Task = convertTaskToDTO(&task)
	}

	return nil
}

func (handler *DBServiceHandler) LoadFlow(ctx context.Context, req *dbPb.DBLoadFlowRequest, rsp *dbPb.DBLoadFlowResponse) error {
	db := handler.Dao().Db()
	flow, tasks, err := models.FetchFlowDetail(db, uint(req.Id))
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

func (handler *DBServiceHandler) LoadTask(ctx context.Context, req *dbPb.DBLoadTaskRequest, rsp *dbPb.DBLoadTaskResponse) error {
	db := handler.Dao().Db()
	task, err := models.FetchTask(db, uint(req.Id))
	if err != nil {
		// todo

	} else {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Task = convertTaskToDTO(&task)
	}

	return nil
}

func (handler *DBServiceHandler) ListFlows(ctx context.Context, req *dbPb.DBListFlowsRequest, rsp *dbPb.DBListFlowsResponse) error {
	db := handler.Dao().Db()
	flows, total, err := models.ListFlows(db, req.BizId, req.Keyword, int(req.Status), int(req.Page.Page - 1) * int(req.Page.PageSize), int(req.Page.PageSize))
	if nil == err {
		rsp.Status = TaskSuccessResponseStatus
		rsp.Page = &dbPb.DBTaskPageDTO{
			Page:     req.Page.Page,
			PageSize: req.Page.PageSize,
			Total:    int32(total),
		}
		flowDTOs := make([]*dbPb.DBFlowDTO, len(flows), len(flows))
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

func convertFlowToDTO(do *models.FlowDO) (dto *dbPb.DBFlowDTO) {
	if do == nil {
		return
	}
	dto = &dbPb.DBFlowDTO{}
	dto.Id = int64(do.ID)
	dto.FlowName = do.Name
	dto.StatusAlias = do.StatusAlias
	dto.BizId = do.BizId
	dto.Status = int32(do.Status)
	dto.CreateTime = do.CreatedAt.Unix()
	dto.UpdateTime = do.UpdatedAt.Unix()

	dto.DeleteTime = deletedAtUnix(do.DeletedAt)
	return
}

func parseFlowDTO(dto *dbPb.DBFlowDTO) (do *models.FlowDO) {
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

func convertTaskToDTO(do *models.TaskDO) (dto *dbPb.DBTaskDTO) {
	if do == nil {
		return
	}
	dto = &dbPb.DBTaskDTO{}
	dto.Id = int64(do.ID)

	dto.BizId = do.BizId
	dto.Status = int32(do.Status)
	dto.CreateTime = do.CreatedAt.Unix()
	dto.UpdateTime = do.UpdatedAt.Unix()

	dto.DeleteTime = deletedAtUnix(do.DeletedAt)

	dto.ParentId = do.ParentId
	dto.ParentType = int32(do.ParentType)

	dto.Parameters = do.Parameters
	dto.Result = do.Result
	dto.TaskName = do.Name
	dto.TaskReturnType = do.ReturnType

	return
}

func batchConvertTaskToDTO(dos []*models.TaskDO) (dtos []*dbPb.DBTaskDTO) {
	dtos = make([]*dbPb.DBTaskDTO, len(dos), len(dos))

	for i, v := range dos {
		dtos[i] = convertTaskToDTO(v)
	}
	return
}

func batchParseTaskDTO(dtos []*dbPb.DBTaskDTO) (dos []*models.TaskDO) {
	dos = make([]*models.TaskDO, len(dtos), len(dtos))

	for i, v := range dtos {
		dos[i] = parseTaskDTO(v)
	}
	return
}

func parseTaskDTO(dto *dbPb.DBTaskDTO) (do *models.TaskDO) {
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
