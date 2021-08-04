package service

import (
	"fmt"
	"github.com/pingcap/errors"
	"github.com/pingcap/ticp/micro-metadb/models"
	db "github.com/pingcap/ticp/micro-metadb/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
)

func testInit() {
	var err error
	models.MetaDB, err = gorm.Open(sqlite.Open("../ticp.sqlite.db"), &gorm.Config{})
	if err != nil {
		fmt.Printf("init open database failed, %s", err.Error())
		return
	}
	err = models.MetaDB.Migrator().CreateTable(
		&models.TransportRecord{},
	)
	if err != nil && !errors.IsAlreadyExists(err) {
		fmt.Printf(err.Error())
		return
	}
}

func TestDBServiceHandler_CreateTransportRecord(t *testing.T) {
	testInit()
	record := &db.TransportRecordDTO{
		ID: "uuid-abcd",
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
	db := new(DBServiceHandler)
	var err error
	err = db.CreateTransportRecord(nil, in, out)
	if err != nil {
		t.Errorf("TestDBServiceHandler_CreateTransportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_CreateTransportRecord success")
}

func TestDBServiceHandler_UpdateTransportRecord(t *testing.T) {
	testInit()
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
	db := new(DBServiceHandler)
	var err error
	err = db.UpdateTransportRecord(nil, in, out)
	if err != nil{
		t.Errorf("TestDBServiceHandler_UpdateTransportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_CreateTransportRecord success")
}

func TestDBServiceHandler_FindTrasnportRecordByID(t *testing.T) {
	testInit()
	record := &db.TransportRecordDTO{
		ID: "uuid-abcd",
	}
	in := &db.DBFindTransportRecordByIDRequest{
		Record: record,
	}
	out :=  &db.DBFindTransportRecordByIDResponse{}
	db := new(DBServiceHandler)
	var err error
	err = db.FindTrasnportRecordByID(nil, in, out)
	if err != nil{
		t.Errorf("TestDBServiceHandler_FindTrasnportRecordByID failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_FindTrasnportRecordByID success, record: %v", out)
}