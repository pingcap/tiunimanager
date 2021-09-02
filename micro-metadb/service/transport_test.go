package service

import (
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"testing"
	"time"
)


func TestDBServiceHandler_CreateTransportRecord(t *testing.T) {
	record := &db.TransportRecordDTO{
		ID: "1111",
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
		ID: "2222",
		ClusterId: "tc-123",
		TransportType: "import",
		FilePath: "/tmp/tiem/datatransport/tc-123/import",
		TenantId: "admin",
		Status: "Running",
		StartTime: time.Now().Unix(),
	}
	createIn := &db.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &db.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(nil, createIn, createOut)
	if err != nil {
		t.Errorf("TestDBServiceHandler_UpdateTransportRecord create record failed: %s", err.Error())
		return
	}

	record.Status = "Finish"
	record.EndTime = time.Now().Unix()

	updateIn := &db.DBUpdateTransportRecordRequest{
		Record: record,
	}
	updateOut :=  &db.DBUpdateTransportRecordResponse{}
	err = handler.UpdateTransportRecord(nil, updateIn, updateOut)
	if err != nil{
		t.Errorf("TestDBServiceHandler_UpdateTransportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_CreateTransportRecord success")
}

func TestDBServiceHandler_FindTrasnportRecordByID(t *testing.T) {
	record := &db.TransportRecordDTO{
		ID: "3333",
		ClusterId: "tc-123",
		TransportType: "import",
		FilePath: "/tmp/tiem/datatransport/tc-123/import",
		TenantId: "admin",
		Status: "Running",
		StartTime: time.Now().Unix(),
	}
	createIn := &db.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &db.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(nil, createIn, createOut)
	if err != nil {
		t.Errorf("TestDBServiceHandler_FindTrasnportRecordByID create record failed: %s", err.Error())
		return
	}

	in := &db.DBFindTransportRecordByIDRequest{
		RecordId: "3333",
	}
	out :=  &db.DBFindTransportRecordByIDResponse{}
	err = handler.FindTrasnportRecordByID(nil, in, out)
	if err != nil{
		t.Errorf("TestDBServiceHandler_FindTrasnportRecordByID failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_FindTrasnportRecordByID success, record: %v", out)
}

func TestDBServiceHandler_ListTrasnportRecord(t *testing.T) {
	record := &db.TransportRecordDTO{
		ID: "4444",
		ClusterId: "tc-123",
		TransportType: "import",
		FilePath: "/tmp/tiem/datatransport/tc-123/import",
		TenantId: "admin",
		Status: "Running",
		StartTime: time.Now().Unix(),
	}
	createIn := &db.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &db.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(nil, createIn, createOut)
	if err != nil {
		t.Errorf("TestDBServiceHandler_ListTrasnportRecord create record failed: %s", err.Error())
		return
	}

	in := &db.DBListTransportRecordRequest{
		ClusterId: "tc-123",
	}
	out :=  &db.DBListTransportRecordResponse{}
	err = handler.ListTrasnportRecord(nil, in, out)
	if err != nil{
		t.Errorf("TestDBServiceHandler_ListTrasnportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_ListTrasnportRecord success, record: %v", out)
}