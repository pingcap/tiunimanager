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
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDBServiceHandler_CreateTransportRecord(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	record := &dbpb.TransportRecordDTO{
		RecordId:      1111,
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "/tmp/tiem/datatransport/tc-123/import",
		TenantId:      "admin",
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
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	record := &dbpb.TransportRecordDTO{
		RecordId:      2222,
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "/tmp/tiem/datatransport/tc-123/import",
		TenantId:      "admin",
		StartTime:     time.Now().Unix(),
	}
	createIn := &dbpb.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &dbpb.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(context.TODO(), createIn, createOut)
	assert.NoError(t, err)

	record.EndTime = time.Now().Unix()

	updateIn := &dbpb.DBUpdateTransportRecordRequest{
		Record: record,
	}
	updateOut := &dbpb.DBUpdateTransportRecordResponse{}
	err = handler.UpdateTransportRecord(context.TODO(), updateIn, updateOut)
	assert.NoError(t, err)
}

func TestDBServiceHandler_FindTrasnportRecordByID(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	record := &dbpb.TransportRecordDTO{
		RecordId:      3333,
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "/tmp/tiem/datatransport/tc-123/import",
		TenantId:      "admin",
		StartTime:     time.Now().Unix(),
	}
	createIn := &dbpb.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &dbpb.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(context.TODO(), createIn, createOut)
	assert.NoError(t, err)

	in := &dbpb.DBFindTransportRecordByIDRequest{
		RecordId: 3333,
	}
	out := &dbpb.DBFindTransportRecordByIDResponse{}
	err = handler.FindTrasnportRecordByID(context.TODO(), in, out)
	assert.NoError(t, err)
}

func TestDBServiceHandler_ListTrasnportRecord(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	record := &dbpb.TransportRecordDTO{
		RecordId:      4444,
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "/tmp/tiem/datatransport/tc-123/import",
		TenantId:      "admin",
		StartTime:     time.Now().Unix(),
	}
	createIn := &dbpb.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &dbpb.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(context.TODO(), createIn, createOut)
	assert.NoError(t, err)

	in := &dbpb.DBListTransportRecordRequest{
		ClusterId: "tc-123",
		Page: &dbpb.DBPageDTO{
			Page:     1,
			PageSize: 10,
		},
	}
	out := &dbpb.DBListTransportRecordResponse{}
	err = handler.ListTrasnportRecord(context.TODO(), in, out)
	assert.NoError(t, err)
}

func TestDBServiceHandler_DeleteTransportRecord(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	record := &dbpb.TransportRecordDTO{
		RecordId:      3333,
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "/tmp/tiem/datatransport/tc-123/import",
		TenantId:      "admin",
		StartTime:     time.Now().Unix(),
	}
	createIn := &dbpb.DBCreateTransportRecordRequest{
		Record: record,
	}
	createOut := &dbpb.DBCreateTransportRecordResponse{}

	err := handler.CreateTransportRecord(context.TODO(), createIn, createOut)
	assert.NoError(t, err)

	in := &dbpb.DBDeleteTransportRequest{
		RecordId: 3333,
	}
	out := &dbpb.DBDeleteTransportResponse{}
	err = handler.DeleteTransportRecord(context.TODO(), in, out)
	assert.NoError(t, err)
}
