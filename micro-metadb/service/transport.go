
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
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

func (d *DBServiceHandler) CreateTransportRecord(ctx context.Context, in *dbpb.DBCreateTransportRecordRequest, out *dbpb.DBCreateTransportRecordResponse) error {
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
	id, err := d.Dao().ClusterManager().CreateTransportRecord(record)
	if err != nil {
		log.Errorf("CreateTransportRecord failed, %s", err.Error())
		return err
	}
	out.Id = id
	log.Infof("CreateTransportRecord success")
	return nil
}

func (d *DBServiceHandler) UpdateTransportRecord(ctx context.Context, in *dbpb.DBUpdateTransportRecordRequest, out *dbpb.DBUpdateTransportRecordResponse) error {
	log := framework.LogWithContext(ctx)
	err := d.Dao().ClusterManager().UpdateTransportRecord(in.GetRecord().GetID(), in.GetRecord().GetClusterId(), in.GetRecord().GetStatus(), time.Unix(in.GetRecord().GetEndTime(), 0))
	if err != nil {
		log.Errorf("UpdateTransportRecord failed, %s", err.Error())
		return err
	}
	log.Infof("UpdateTransportRecord success")
	return nil
}

func (d *DBServiceHandler) FindTrasnportRecordByID(ctx context.Context, in *dbpb.DBFindTransportRecordByIDRequest, out *dbpb.DBFindTransportRecordByIDResponse) error {
	log := framework.LogWithContext(ctx)
	record, err := d.Dao().ClusterManager().FindTransportRecordById(in.GetRecordId())
	if err != nil {
		log.Errorf("FindTransportRecordById failed, %s", err.Error())
		return err
	}
	out.Record = convertRecordDTO(record)
	log.Infof("FindTrasnportRecordByID success, %v", out)
	return nil
}

func (d *DBServiceHandler) ListTrasnportRecord(ctx context.Context, in *dbpb.DBListTransportRecordRequest, out *dbpb.DBListTransportRecordResponse) error {
	log := framework.LogWithContext(ctx)
	records, total, err := d.Dao().ClusterManager().ListTransportRecord(in.GetClusterId(), in.GetRecordId(), (in.GetPage().GetPage()-1)*in.GetPage().GetPageSize(), in.GetPage().GetPageSize())
	if err != nil {
		log.Errorf("ListTrasnportRecord failed, %s", err.Error())
		return err
	}
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
