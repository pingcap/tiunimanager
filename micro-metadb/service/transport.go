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
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"

	"time"
)

func (handler *DBServiceHandler) CreateTransportRecord(ctx context.Context, in *dbpb.DBCreateTransportRecordRequest, out *dbpb.DBCreateTransportRecordResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "QueryBackupStrategyByTime", int(out.GetStatus().GetCode()))
	log := framework.LogWithContext(ctx)
	record := &models.TransportRecord{
		Record: models.Record{
			ID:       uint(in.GetRecord().GetRecordId()),
			TenantId: in.GetRecord().GetTenantId(),
		},
		ClusterId:       in.GetRecord().GetClusterId(),
		TransportType:   in.GetRecord().GetTransportType(),
		FilePath:        in.GetRecord().GetFilePath(),
		ZipName:         in.GetRecord().GetZipName(),
		Comment:         in.GetRecord().GetComment(),
		StorageType:     in.GetRecord().GetStorageType(),
		FlowId:          in.GetRecord().GetFlowId(),
		ReImportSupport: in.GetRecord().GetReImportSupport(),
		StartTime:       time.Unix(in.GetRecord().GetStartTime(), 0),
	}
	id, err := handler.Dao().ClusterManager().CreateTransportRecord(ctx, record)
	if err != nil {
		out.Status = BizErrResponseStatus
		out.Status.Message = err.Error()
		log.Errorf("CreateTransportRecord failed, %s", err.Error())
	} else {
		out.Status = ClusterSuccessResponseStatus
		out.RecordId = int64(id)
		log.Infof("CreateTransportRecord success")
	}

	return nil
}

func (handler *DBServiceHandler) UpdateTransportRecord(ctx context.Context, in *dbpb.DBUpdateTransportRecordRequest, out *dbpb.DBUpdateTransportRecordResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "UpdateTransportRecord", int(out.GetStatus().GetCode()))
	log := framework.LogWithContext(ctx)
	err := handler.Dao().ClusterManager().UpdateTransportRecord(ctx, int(in.GetRecord().GetRecordId()), in.GetRecord().GetClusterId(), time.Unix(in.GetRecord().GetEndTime(), 0))
	if err != nil {
		out.Status = BizErrResponseStatus
		out.Status.Message = err.Error()
		log.Errorf("UpdateTransportRecord failed, %s", err.Error())
	} else {
		out.Status = ClusterSuccessResponseStatus
		log.Infof("UpdateTransportRecord success")
	}

	return nil
}

func (handler *DBServiceHandler) FindTrasnportRecordByID(ctx context.Context, in *dbpb.DBFindTransportRecordByIDRequest, out *dbpb.DBFindTransportRecordByIDResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "FindTrasnportRecordByID", int(out.GetStatus().GetCode()))
	log := framework.LogWithContext(ctx)
	record, err := handler.Dao().ClusterManager().FindTransportRecordById(ctx, int(in.GetRecordId()))
	if err != nil {
		out.Status = BizErrResponseStatus
		out.Status.Message = err.Error()
		log.Errorf("FindTransportRecordById failed, %s", err.Error())
	} else {
		out.Status = ClusterSuccessResponseStatus
		out.Record = convertTransportRecordDTO(record)
		log.Infof("FindTrasnportRecordByID success, %v", out)
	}

	return nil
}

func (handler *DBServiceHandler) ListTrasnportRecord(ctx context.Context, in *dbpb.DBListTransportRecordRequest, out *dbpb.DBListTransportRecordResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "ListTrasnportRecord", int(out.GetStatus().GetCode()))
	log := framework.LogWithContext(ctx)
	result, total, err := handler.Dao().ClusterManager().ListTransportRecord(ctx, in.GetClusterId(), int(in.GetRecordId()), in.GetReImport(), (in.GetPage().GetPage()-1)*in.GetPage().GetPageSize(), in.GetPage().GetPageSize())
	if err != nil {
		out.Status = BizErrResponseStatus
		out.Status.Message = err.Error()
		log.Errorf("ListTrasnportRecord failed, %s", err.Error())
	} else {
		out.Page = &dbpb.DBPageDTO{
			Page:     in.Page.Page,
			PageSize: in.Page.PageSize,
			Total:    int32(total),
		}
		transportRecordDTOs := make([]*dbpb.DBTransportRecordDisplayDTO, len(result))
		for i, v := range result {
			transportRecordDTOs[i] = convertToTransportRecordDisplayDTO(v.TransportRecord, v.Flow)
		}
		out.Records = transportRecordDTOs
		out.Status = ClusterSuccessResponseStatus
		log.Infof("ListTrasnportRecord success, %v", out)
	}

	return nil
}

func (handler *DBServiceHandler) DeleteTransportRecord(ctx context.Context, in *dbpb.DBDeleteTransportRequest, out *dbpb.DBDeleteTransportResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "DeleteTransportRecord", int(out.GetStatus().GetCode()))
	log := framework.LogWithContext(ctx)
	record, err := handler.Dao().ClusterManager().DeleteTransportRecord(ctx, int(in.GetRecordId()))
	if err != nil {
		out.Status = BizErrResponseStatus
		out.Status.Message = err.Error()
		log.Errorf("DeleteTransportRecord failed, %s", err.Error())
	} else {
		out.Status = ClusterSuccessResponseStatus
		out.Record = convertTransportRecordDTO(record)
		log.Infof("DeleteTransportRecord success, %v", out)
	}

	return nil
}

func convertToTransportRecordDisplayDTO(do *models.TransportRecord, flow *models.FlowDO) (dto *dbpb.DBTransportRecordDisplayDTO) {
	if do == nil {
		return nil
	}

	dto = &dbpb.DBTransportRecordDisplayDTO{
		Record: convertTransportRecordDTO(do),
		Flow:   convertFlowToDTO(flow),
	}
	return
}

func convertTransportRecordDTO(record *models.TransportRecord) *dbpb.TransportRecordDTO {
	recordDTO := &dbpb.TransportRecordDTO{
		RecordId:        int64(record.ID),
		ClusterId:       record.ClusterId,
		TransportType:   record.TransportType,
		TenantId:        record.TenantId,
		FilePath:        record.FilePath,
		ZipName:         record.ZipName,
		StorageType:     record.StorageType,
		FlowId:          record.FlowId,
		Comment:         record.Comment,
		ReImportSupport: record.ReImportSupport,
		StartTime:       record.StartTime.Unix(),
		EndTime:         record.EndTime.Unix(),
	}
	return recordDTO
}
