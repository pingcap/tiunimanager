package service

import (
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"testing"
	"time"
)

func TestDBServiceHandler_CreateTransportRecord(t *testing.T) {
	record := &db.TransportRecordDTO{
		ID: "uuid-abc",
		ClusterId: "tc-123",
		TransportType: "import",
		FilePath: "/tmp/tiem/datatransport/tc-123/import",
		TenantId: "admin",
		Status: "Running",
		StartTime: time.Now().Unix(),
	}
	in := &db.DBCreateTransportRecordRequest{
		Record: record,
	}
	out :=  &db.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(nil, in, out)
	if err != nil {
		t.Errorf("TestDBServiceHandler_CreateTransportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_CreateTransportRecord success")
}

func TestDBServiceHandler_UpdateTransportRecord(t *testing.T) {
	record := &db.TransportRecordDTO{
		ID: "uuid-abcd",
		ClusterId: "tc-123",
		Status: "Finish",
		EndTime: time.Now().Unix(),
	}
	in := &db.DBUpdateTransportRecordRequest{
		Record: record,
	}
	out :=  &db.DBUpdateTransportRecordResponse{}
	err := handler.UpdateTransportRecord(nil, in, out)
	if err != nil{
		t.Errorf("TestDBServiceHandler_UpdateTransportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_CreateTransportRecord success")
}

func TestDBServiceHandler_FindTrasnportRecordByID(t *testing.T) {
	in := &db.DBFindTransportRecordByIDRequest{
		RecordId: "uuid-abcd",
	}
	out :=  &db.DBFindTransportRecordByIDResponse{}
	err := handler.FindTrasnportRecordByID(nil, in, out)
	if err != nil{
		t.Errorf("TestDBServiceHandler_FindTrasnportRecordByID failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_FindTrasnportRecordByID success, record: %v", out)
}

func TestDBServiceHandler_ListTrasnportRecord(t *testing.T) {
	in := &db.DBListTransportRecordRequest{
		ClusterId: "tc-123",
	}
	out :=  &db.DBListTransportRecordResponse{}
	err := handler.ListTrasnportRecord(nil, in, out)
	if err != nil{
		t.Errorf("TestDBServiceHandler_ListTrasnportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_ListTrasnportRecord success, record: %v", out)
}