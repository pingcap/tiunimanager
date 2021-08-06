package service

import (
	"context"
	"github.com/pingcap/tiem/micro-metadb/models"
	"github.com/pingcap/tiem/micro-metadb/proto"
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
		StartTime: time.Unix(in.GetRecord().GetStartTime(), 0),
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
	record, err := models.FindTransportRecordById(in.GetRecordId())
	if err != nil {
		return err
	}
	out.Record = convertRecordDTO(record)
	return nil
}

func (d *DBServiceHandler)ListTrasnportRecord(ctx context.Context, in *db.DBListTransportRecordRequest, out *db.DBListTransportRecordResponse) error {
	records, total, err := models.ListTransportRecord(in.GetClusterId(), in.GetRecordId(), in.GetPage().GetPage(), in.GetPage().GetPageSize())
	if err != nil {
		return err
	}
	out.Records = make([]*db.TransportRecordDTO, len(records))
	for index := 0; index < len(records); index ++ {
		out.Records[index] = convertRecordDTO(records[index])
	}
	out.Page = &db.DBPageDTO{
		Page: in.GetPage().GetPage(),
		PageSize: in.GetPage().GetPageSize(),
		Total: int32(total),
	}

	return nil
}

func convertRecordDTO(record *models.TransportRecord) *db.TransportRecordDTO {
	recordDTO := &db.TransportRecordDTO{
		ID: record.ID,
		ClusterId: record.ClusterId,
		TransportType: record.TransportType,
		TenantId: record.TenantId,
		FilePath: record.FilePath,
		Status: record.Status,
		StartTime: record.StartTime.Unix(),
		EndTime: record.EndTime.Unix(),
	}
	return recordDTO
}