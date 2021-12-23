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
 ******************************************************************************/

package service

import (
	"context"
	"database/sql"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	"reflect"
	"testing"
	"time"
)

func Test_unix2NullTime(t *testing.T) {
	now := time.Now()
	type args struct {
		unix int64
	}
	tests := []struct {
		name string
		args args
		want sql.NullTime
	}{
		{"normal", args{now.Unix()}, sql.NullTime{time.Unix(now.Unix(), 0), true}},
		{"normal", args{0}, sql.NullTime{time.Unix(0, 0), false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unix2NullTime(tt.args.unix); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unix2NullTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBServiceHandler_CreateClusterRelation(t *testing.T) {
	data := []*models.ClusterRelation{
		{
			Record:           models.Record{TenantId: "111"},
			SubjectClusterId: "1",
			ObjectClusterId:  "2",
			RelationType:     uint32(common.SlaveTo),
		},
		{
			Record:           models.Record{TenantId: ""},
			SubjectClusterId: "",
			ObjectClusterId:  "",
			RelationType:     uint32(common.StandBy),
		},
		{
			Record:           models.Record{TenantId: "111"},
			SubjectClusterId: "1",
			ObjectClusterId:  "6",
			RelationType:     uint32(common.CloneFrom),
		},
	}
	dataDTOs := make([]*dbpb.DBClusterRelationDTO, len(data))
	for i, v := range data {
		dataDTOs[i] = convertToClusterRelationDTO(v)
	}
	req1 := &dbpb.DBCreateClusterRelationRequest{Relation: dataDTOs[0]}
	req2 := &dbpb.DBCreateClusterRelationRequest{Relation: dataDTOs[1]}
	resp := &dbpb.DBCreateClusterRelationResponse{}
	t.Run("normal", func(t *testing.T) {
		err := handler.CreateClusterRelation(context.TODO(), req1, resp)
		if err != nil {
			t.Errorf("CreateClusterRelation() error = %v", err)
			return
		}
	})
	t.Run("empty parameter", func(t *testing.T) {
		err := handler.CreateClusterRelation(context.TODO(), req2, resp)
		if err == nil {
			t.Errorf("CreateClusterRelation() error = %v", err)
		}
	})
}

func TestDBServiceHandler_SwapClusterRelation(t *testing.T) {
	data1 := models.ClusterRelation{
		Record:           models.Record{TenantId: "111"},
		SubjectClusterId: "1",
		ObjectClusterId:  "2",
		RelationType:     uint32(common.SlaveTo),
	}
	data2 := models.ClusterRelation{
		Record:           models.Record{TenantId: "111"},
		SubjectClusterId: "1",
		ObjectClusterId:  "6",
		RelationType:     uint32(common.CloneFrom),
	}
	clusterManager := handler.Dao().ClusterManager()
	clusterManager.CreateClusterRelation(context.TODO(), data1)
	clusterManager.CreateClusterRelation(context.TODO(), data2)
	t.Run("normal", func(t *testing.T) {
		req := &dbpb.DBSwapClusterRelationRequest{Id: 1}
		resp := &dbpb.DBSwapClusterRelationResponse{}
		err := handler.SwapClusterRelation(context.TODO(), req, resp)
		if err != nil {
			t.Errorf("SwapClusterRelation() error = %v", err)
		}
	})

	t.Run("no record", func(t *testing.T) {
		req := &dbpb.DBSwapClusterRelationRequest{Id: 5}
		resp := &dbpb.DBSwapClusterRelationResponse{}
		err := handler.SwapClusterRelation(context.TODO(), req, resp)
		if err == nil {
			t.Errorf("SwapClusterRelation() error = %v", err)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		req := &dbpb.DBSwapClusterRelationRequest{Id: 0}
		resp := &dbpb.DBSwapClusterRelationResponse{}
		err := handler.SwapClusterRelation(context.TODO(), req, resp)
		if err == nil {
			t.Errorf("SwapClusterRelation() error = %v", err)
		}
	})
}

func TestDBServiceHandler_ListClusterRelation(t *testing.T) {
	data1 := models.ClusterRelation{
		Record:           models.Record{TenantId: "111"},
		SubjectClusterId: "1212",
		ObjectClusterId:  "2323",
		RelationType:     uint32(common.SlaveTo),
	}
	data2 := models.ClusterRelation{
		Record:           models.Record{TenantId: "111"},
		SubjectClusterId: "1212",
		ObjectClusterId:  "6767",
		RelationType:     uint32(common.CloneFrom),
	}
	clusterManager := handler.Dao().ClusterManager()
	clusterManager.CreateClusterRelation(context.TODO(), data1)
	clusterManager.CreateClusterRelation(context.TODO(), data2)
	t.Run("normal", func(t *testing.T) {
		req := &dbpb.DBListClusterRelationRequest{SubjectClusterId: "1212"}
		resp := &dbpb.DBListClusterRelationResponse{}
		err := handler.ListClusterRelation(context.TODO(), req, resp)
		if err != nil {
			t.Errorf("ListClusterRelation() error = %v", err)
		}
		if len(resp.Relation) != 2 {
			t.Errorf("ListClusterRelation() result len = %v, want = %v", len(resp.Relation), 2)
		}
	})

	t.Run("no record", func(t *testing.T) {
		req := &dbpb.DBListClusterRelationRequest{SubjectClusterId: "2323"}
		resp := &dbpb.DBListClusterRelationResponse{}
		_ = handler.ListClusterRelation(context.TODO(), req, resp)

		if len(resp.Relation) != 0 {
			t.Errorf("ListClusterRelation() result len = %v, want = %v", len(resp.Relation), 0)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		req := &dbpb.DBListClusterRelationRequest{SubjectClusterId: ""}
		resp := &dbpb.DBListClusterRelationResponse{}
		err := handler.ListClusterRelation(context.TODO(), req, resp)
		if err == nil {
			t.Errorf("SwapClusterRelation() error = %v", err)
		}
	})
}
