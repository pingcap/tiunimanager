package service

import (
	"context"
	"github.com/pingcap/ticp/micro-metadb/models"
	"github.com/pingcap/ticp/micro-metadb/proto"
	"time"
)

func (d *DBServiceHandler)CreateTransportRecord(ctx context.Context, in *db.DBCreateTransportRecordRequest, out *db.DBCreateTransportRecordResponse) error {
	record := &models.TransportRecord{
		ID: in.GetRecord().GetID(),
		ClusterId: in.GetRecord().GetClusterId(),
		TransportType: in.GetRecord().GetTransportType(),
		FilePath: in.GetRecord().GetFilePath(),
		TenantId: in.GetRecord().GetTenantId(),
		Status: in.GetRecord().GetStatus(),
		StratTime: time.Unix(in.GetRecord().GetStratTime(), 0),
	}
	id, err := models.CreateTransportRecord(record)
	if err != nil {
		return err
	}
	out.Id = id
	return nil
}

func (d *DBServiceHandler)UpdateTransportRecord(ctx context.Context, in *db.DBUpdateTransportRecordRequest, out *db.DBUpdateTransportRecordResponse) error {
	err := models.UpdateTransportRecord(in.GetRecord().GetID(), in.GetRecord().GetClusterId(), in.GetRecord().GetStatus(), time.Unix(in.GetRecord().GetEndTime(), 0))
	if err != nil {
		return err
	}
	return nil
}

func (d *DBServiceHandler)FindTrasnportRecordByID(ctx context.Context, in *db.DBFindTransportRecordByIDRequest, out *db.DBFindTransportRecordByIDResponse) error {
	record, err := models.FindTransportRecordById(in.GetRecord().GetID())
	if err != nil {
		return err
	}
	recordDTO := &db.TransportRecordDTO{
		ID: record.ID,
		ClusterId: record.ClusterId,
		TransportType: record.TransportType,
		TenantId: record.TenantId,
		FilePath: record.FilePath,
		Status: record.Status,
		StratTime: record.StratTime.Unix(),
		EndTime: record.EndTime.Unix(),
	}
	out.Record = recordDTO
	return nil
}