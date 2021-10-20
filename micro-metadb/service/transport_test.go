
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
	"testing"
	"time"
)

func TestDBServiceHandler_CreateTransportRecord(t *testing.T) {
	record := &dbpb.TransportRecordDTO{
		ID:            "1111",
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "/tmp/tiem/datatransport/tc-123/import",
		TenantId:      "admin",
		Status:        "Running",
		StartTime:     time.Now().Unix(),
	}
	in := &dbpb.DBCreateTransportRecordRequest{
		Record: record,
	}
	out := &dbpb.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(context.TODO(), in, out)
	if err != nil {
		t.Errorf("TestDBServiceHandler_CreateTransportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_CreateTransportRecord success")
}

func TestDBServiceHandler_UpdateTransportRecord(t *testing.T) {
	record := &dbpb.TransportRecordDTO{
		ID:            "2222",
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "/tmp/tiem/datatransport/tc-123/import",
		TenantId:      "admin",
		Status:        "Running",
		StartTime:     time.Now().Unix(),
	}
	createIn := &dbpb.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &dbpb.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(context.TODO(), createIn, createOut)
	if err != nil {
		t.Errorf("TestDBServiceHandler_UpdateTransportRecord create record failed: %s", err.Error())
		return
	}

	record.Status = "Finish"
	record.EndTime = time.Now().Unix()

	updateIn := &dbpb.DBUpdateTransportRecordRequest{
		Record: record,
	}
	updateOut := &dbpb.DBUpdateTransportRecordResponse{}
	err = handler.UpdateTransportRecord(context.TODO(), updateIn, updateOut)
	if err != nil {
		t.Errorf("TestDBServiceHandler_UpdateTransportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_CreateTransportRecord success")
}

func TestDBServiceHandler_FindTrasnportRecordByID(t *testing.T) {
	record := &dbpb.TransportRecordDTO{
		ID:            "3333",
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "/tmp/tiem/datatransport/tc-123/import",
		TenantId:      "admin",
		Status:        "Running",
		StartTime:     time.Now().Unix(),
	}
	createIn := &dbpb.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &dbpb.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(context.TODO(), createIn, createOut)
	if err != nil {
		t.Errorf("TestDBServiceHandler_FindTrasnportRecordByID create record failed: %s", err.Error())
		return
	}

	in := &dbpb.DBFindTransportRecordByIDRequest{
		RecordId: "3333",
	}
	out := &dbpb.DBFindTransportRecordByIDResponse{}
	err = handler.FindTrasnportRecordByID(context.TODO(), in, out)
	if err != nil {
		t.Errorf("TestDBServiceHandler_FindTrasnportRecordByID failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_FindTrasnportRecordByID success, record: %v", out)
}

func TestDBServiceHandler_ListTrasnportRecord(t *testing.T) {
	record := &dbpb.TransportRecordDTO{
		ID:            "4444",
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "/tmp/tiem/datatransport/tc-123/import",
		TenantId:      "admin",
		Status:        "Running",
		StartTime:     time.Now().Unix(),
	}
	createIn := &dbpb.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &dbpb.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(context.TODO(), createIn, createOut)
	if err != nil {
		t.Errorf("TestDBServiceHandler_ListTrasnportRecord create record failed: %s", err.Error())
		return
	}

	in := &dbpb.DBListTransportRecordRequest{
		ClusterId: "tc-123",
	}
	out := &dbpb.DBListTransportRecordResponse{}
	err = handler.ListTrasnportRecord(context.TODO(), in, out)
	if err != nil {
		t.Errorf("TestDBServiceHandler_ListTrasnportRecord failed: %s", err.Error())
		return
	}
	t.Logf("TestDBServiceHandler_ListTrasnportRecord success, record: %v", out)
}
