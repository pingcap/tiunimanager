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

	"strconv"
	"time"
)

func (handler *DBServiceHandler) CreateTransportRecord(ctx context.Context, in *dbpb.DBCreateTransportRecordRequest, out *dbpb.DBCreateTransportRecordResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "QueryBackupStrategyByTime", int(out.GetStatus().GetCode()))
	uintId, err := strconv.ParseInt(in.GetRecord().GetID(), 10, 64)
	log := framework.LogWithContext(ctx)
	record := &models.TransportRecord{
		Record: models.Record{
			ID: uint(uintId),
		},
		ClusterId:     in.GetRecord().GetClusterId(),
		TransportType: in.GetRecord().GetTransportType(),
		FilePath:      in.GetRecord().GetFilePath(),
		TenantId:      in.GetRecord().GetTenantId(),
		Status:        in.GetRecord().GetStatus(),
		StartTime:     time.Unix(in.GetRecord().GetStartTime(), 0),
	}
	id, err := handler.Dao().ClusterManager().CreateTransportRecord(ctx, record)
	if err != nil {
		out.Status = BizErrResponseStatus
		out.Status.Message = err.Error()
		log.Errorf("CreateTransportRecord failed, %s", err.Error())
	} else {
		out.Status = ClusterSuccessResponseStatus
		out.Id = id
		log.Infof("CreateTransportRecord success")
	}

	return nil
}

func (handler *DBServiceHandler) UpdateTransportRecord(ctx context.Context, in *dbpb.DBUpdateTransportRecordRequest, out *dbpb.DBUpdateTransportRecordResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "UpdateTransportRecord", int(out.GetStatus().GetCode()))
	log := framework.LogWithContext(ctx)
	err := handler.Dao().ClusterManager().UpdateTransportRecord(ctx, in.GetRecord().GetID(), in.GetRecord().GetClusterId(), in.GetRecord().GetStatus(), time.Unix(in.GetRecord().GetEndTime(), 0))
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
	record, err := handler.Dao().ClusterManager().FindTransportRecordById(ctx, in.GetRecordId())
	if err != nil {
		out.Status = BizErrResponseStatus
		out.Status.Message = err.Error()
		log.Errorf("FindTransportRecordById failed, %s", err.Error())
	} else {
		out.Status = ClusterSuccessResponseStatus
		out.Record = convertRecordDTO(record)
		log.Infof("FindTrasnportRecordByID success, %v", out)
	}

	return nil
}

func (handler *DBServiceHandler) ListTrasnportRecord(ctx context.Context, in *dbpb.DBListTransportRecordRequest, out *dbpb.DBListTransportRecordResponse) error {
	start := time.Now()
	defer handler.HandleMetrics(start, "ListTrasnportRecord", int(out.GetStatus().GetCode()))
	log := framework.LogWithContext(ctx)
	records, total, err := handler.Dao().ClusterManager().ListTransportRecord(ctx, in.GetClusterId(), in.GetRecordId(), (in.GetPage().GetPage()-1)*in.GetPage().GetPageSize(), in.GetPage().GetPageSize())
	if err != nil {
		out.Status = BizErrResponseStatus
		out.Status.Message = err.Error()
		log.Errorf("ListTrasnportRecord failed, %s", err.Error())
	} else {
		out.Status = ClusterSuccessResponseStatus
		out.Records = make([]*dbpb.TransportRecordDTO, len(records))
		for index := 0; index < len(records); index++ {
			out.Records[index] = convertRecordDTO(records[index])
		}
		out.Page = &dbpb.DBPageDTO{
			Page:     in.GetPage().GetPage(),
			PageSize: in.GetPage().GetPageSize(),
			Total:    int32(total),
		}
		log.Infof("ListTrasnportRecord success, %v", out)
	}

	return nil
}

func convertRecordDTO(record *models.TransportRecord) *dbpb.TransportRecordDTO {
	recordDTO := &dbpb.TransportRecordDTO{
		ID:            strconv.Itoa(int(record.ID)),
		ClusterId:     record.ClusterId,
		TransportType: record.TransportType,
		TenantId:      record.TenantId,
		FilePath:      record.FilePath,
		Status:        record.Status,
		StartTime:     record.StartTime.Unix(),
		EndTime:       record.EndTime.Unix(),
	}
	return recordDTO
}
