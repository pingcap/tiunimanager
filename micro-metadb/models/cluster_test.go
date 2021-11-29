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

package models

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"strings"
	"testing"
	"time"
)

func TestCreateCluster(t *testing.T) {
	type args struct {
		ClusterName    string
		DbPassword     string
		ClusterType    string
		ClusterVersion string
		Tls            bool
		Tags           string
		OwnerId        string
		TenantId       string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(args args, cluster *Cluster) bool
	}{
		{"normal create", args{TenantId: "111", ClusterName: "testCluster", OwnerId: "111"}, false, []func(args args, cluster *Cluster) bool{
			func(args args, cluster *Cluster) bool { return args.TenantId == cluster.TenantId },
			func(args args, cluster *Cluster) bool { return args.ClusterName == cluster.Name },
			func(args args, cluster *Cluster) bool { return cluster.ID != "" },
			func(args args, cluster *Cluster) bool { return cluster.Code == "testCluster" },
		}},
		{"chinese name", args{TenantId: "111", ClusterName: "中文测试集群", OwnerId: "111"}, false, []func(args args, cluster *Cluster) bool{
			func(args args, cluster *Cluster) bool { return args.TenantId == cluster.TenantId },
			func(args args, cluster *Cluster) bool { return args.ClusterName == cluster.Name },
			func(args args, cluster *Cluster) bool { return cluster.ID != "" },
			func(args args, cluster *Cluster) bool { return cluster.Code == "_zhong_wen_ce_shi_ji_qun" },
		}},
		{"error", args{TenantId: "111", ClusterName: "中文测试集群"}, true, []func(args args, cluster *Cluster) bool{}},
	}
	clusterTbl := Dao.ClusterManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCluster, err := clusterTbl.CreateCluster(context.TODO(), Cluster{Entity: Entity{TenantId: tt.args.TenantId},
				Name:       tt.args.ClusterName,
				DbPassword: tt.args.DbPassword,
				Type:       tt.args.ClusterType,
				Version:    tt.args.ClusterVersion,
				Tls:        tt.args.Tls,
				Tags:       tt.args.Tags,
				OwnerId:    tt.args.OwnerId,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, gotCluster) {
					t.Errorf("CreateCluster() test error, testname = %v, assert %v, args = %v, gotCluster = %v", tt.name, i, tt.args, gotCluster)
				}
			}

		})
	}
}

func TestDeleteCluster(t *testing.T) {
	clusterTbl := Dao.ClusterManager()
	t.Run("normal", func(t *testing.T) {
		cluster := &Cluster{
			Entity:  Entity{TenantId: "111"},
			OwnerId: "111",
		}
		MetaDB.Create(cluster)
		defer MetaDB.Delete(cluster)

		newCluster, err := clusterTbl.DeleteCluster(context.TODO(), cluster.ID)

		if err != nil {
			t.Errorf("DeleteCluster() error = %v", err)
		}

		if !newCluster.DeletedAt.Valid {
			t.Errorf("DeleteCluster() DeletedAt = %v", newCluster.DeletedAt)
		}

		err = MetaDB.Find(newCluster).Where("id = ?", newCluster.ID).Error

		if err != nil {
			t.Errorf("DeleteCluster() error = %v", err)
		}

		if !newCluster.DeletedAt.Valid {
			t.Errorf("DeleteCluster() DeletedAt = %v", newCluster.DeletedAt)
		}
	})

	t.Run("no record", func(t *testing.T) {
		_, err := clusterTbl.DeleteCluster(context.TODO(), "TestDeleteClusterId")

		if err == nil {
			t.Errorf("DeleteCluster() error = %v", err)
		}
	})
	t.Run("empty clusterId", func(t *testing.T) {
		_, err := clusterTbl.DeleteCluster(context.TODO(), "")

		if err == nil {
			t.Errorf("DeleteCluster() error = %v", err)
		}
	})
}

func TestUpdateClusterDemand(t *testing.T) {
	clusterTbl := Dao.ClusterManager()
	t.Run("normal update demand", func(t *testing.T) {
		cluster := &Cluster{
			Entity:  Entity{TenantId: "111"},
			OwnerId: "ttt",
		}
		MetaDB.Create(cluster)
		defer MetaDB.Delete(cluster)

		demandId := cluster.CurrentDemandId

		cluster, demand, err := clusterTbl.UpdateComponentDemand(context.TODO(), cluster.ID, "aaa", cluster.TenantId)
		if err != nil {
			t.Errorf("UpdateComponentDemand() error = %v", err)
		}

		if demand == nil || demand.ID == 0 {
			t.Errorf("UpdateComponentDemand() demand = %v", demand)
		}

		if cluster.CurrentDemandId == 0 || cluster.CurrentDemandId <= demandId {
			t.Errorf("UpdateComponentDemand() new demand id = %v", cluster.CurrentDemandId)
		}

		if cluster.ID == "" {
			t.Errorf("UpdateComponentDemand() cluster.ID empty")
		}
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, _, err := clusterTbl.UpdateComponentDemand(context.TODO(), "", "aaa", "111")
		if err == nil {
			t.Errorf("UpdateComponentDemand() error = %v", err)
		}

	})
}

func TestUpdateClusterFlowId(t *testing.T) {
	clusterTbl := Dao.ClusterManager()
	t.Run("normal update demand", func(t *testing.T) {
		cluster := &Cluster{
			Entity:  Entity{TenantId: "111"},
			OwnerId: "ttt",
		}
		MetaDB.Create(cluster)
		defer MetaDB.Delete(cluster)

		flowId := uint(111)

		cluster, err := clusterTbl.UpdateClusterFlowId(context.TODO(), cluster.ID, flowId)

		if err != nil {
			t.Errorf("UpdateClusterFlowId() error = %v", err)
		}

		if cluster.CurrentFlowId != flowId {
			t.Errorf("UpdateClusterFlowId() want flowId = %v, got = %v", flowId, cluster.CurrentFlowId)
		}

		if cluster.ID == "" {
			t.Errorf("UpdateClusterFlowId() cluster.ID empty")
		}
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, err := clusterTbl.UpdateClusterFlowId(context.TODO(), "", 111)
		if err == nil {
			t.Errorf("UpdateClusterFlowId() error = %v", err)
		}

	})
}

func TestUpdateClusterStatus(t *testing.T) {
	clusterTbl := Dao.ClusterManager()
	t.Run("normal update status", func(t *testing.T) {
		cluster := &Cluster{
			Entity:  Entity{TenantId: "111"},
			OwnerId: "ttt",
		}
		MetaDB.Create(cluster)
		defer MetaDB.Delete(cluster)

		status := int8(2)

		cluster, err := clusterTbl.UpdateClusterStatus(context.TODO(), cluster.ID, status)

		if err != nil {
			t.Errorf("UpdateClusterStatus() error = %v", err)
		}

		if cluster.Status != status {
			t.Errorf("UpdateClusterFlowId() want status = %v, got = %v", status, cluster.Status)
		}

		if cluster.ID == "" {
			t.Errorf("UpdateClusterFlowId() cluster.ID empty")
		}
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, err := clusterTbl.UpdateClusterStatus(context.TODO(), "", 2)
		if err == nil {
			t.Errorf("UpdateClusterFlowId() error = %v", err)
		}

	})
}

func TestUpdateTopologyConfig(t *testing.T) {
	clusterTbl := Dao.ClusterManager()
	t.Run("normal update config", func(t *testing.T) {
		cluster := &Cluster{
			Entity:  Entity{TenantId: "111"},
			OwnerId: "ttt",
		}
		MetaDB.Create(cluster)
		defer MetaDB.Delete(cluster)

		currentConfigId := cluster.CurrentTopologyConfigId

		cluster, err := clusterTbl.UpdateTopologyConfig(context.TODO(), cluster.ID, "aaa", cluster.TenantId)
		if err != nil {
			t.Errorf("UpdateTopologyConfig() error = %v", err)
		}

		if cluster.CurrentTopologyConfigId == 0 || cluster.CurrentTopologyConfigId <= currentConfigId {
			t.Errorf("UpdateTopologyConfig() new config id = %v, current config id = %v", cluster.CurrentTopologyConfigId, currentConfigId)
		}

		if cluster.ID == "" {
			t.Errorf("UpdateTopologyConfig() cluster.ID empty")
		}
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, _, err := clusterTbl.UpdateComponentDemand(context.TODO(), "", "aaa", "111")
		if err == nil {
			t.Errorf("UpdateComponentDemand() error = %v", err)
		}

	})
}

func TestListClusters(t *testing.T) {
	c1 := &Cluster{
		Entity:  Entity{TenantId: "111"},
		Type:    "test_type_1",
		Name:    "test_cluster_name",
		Tags:    ",tag,",
		OwnerId: "ttt",
	}
	MetaDB.Create(c1)
	defer MetaDB.Delete(c1)

	c2 := &Cluster{
		Entity:  Entity{TenantId: "111"},
		Type:    "test_type_1",
		Name:    "1111test",
		Tags:    "tag,",
		OwnerId: "ttt",
	}
	MetaDB.Create(c2)
	defer MetaDB.Delete(c2)

	c3 := &Cluster{
		Entity:  Entity{TenantId: "111"},
		Type:    "whatever",
		Name:    "test_cluster_name",
		Tags:    ",tag",
		OwnerId: "ttt",
	}
	MetaDB.Create(c3)
	defer MetaDB.Delete(c3)

	c4 := &Cluster{
		Entity:  Entity{TenantId: "111"},
		Type:    "test_type_1",
		Name:    "whatever",
		Tags:    "1,tag,2",
		OwnerId: "ttt",
	}
	MetaDB.Create(c4)
	defer MetaDB.Delete(c4)

	c5 := &Cluster{
		Entity:  Entity{TenantId: "111"},
		OwnerId: "ttt",
	}
	MetaDB.Create(c5)
	defer MetaDB.Delete(c5)

	clusterTbl := Dao.ClusterManager()
	t.Run("cluster id", func(t *testing.T) {
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), c4.ID, "", "", "", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 1 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 1)
		}

		if len(clusters) != 1 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 1)
		}

		if clusters[0].ID != c4.ID || clusters[0].Name != c4.Name {
			t.Errorf("ListClusters() clusters = %v, want = %v", clusters[0], c4)
		}
	})

	t.Run("cluster name", func(t *testing.T) {
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), "", "test_cluster_name", "", "", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 2 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 2)
		}

		if len(clusters) != 2 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 2)
		}

		for _, v := range clusters {
			if !strings.Contains(v.Name, "test") {
				t.Errorf("ListClusters() clusters = %v, want cluster name contains = %v", v, "test")
			}
		}
	})

	t.Run("cluster type", func(t *testing.T) {
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), "", "", "test_type_1", "", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 3)
		}

		if len(clusters) != 3 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 3)
		}

		for _, v := range clusters {
			if v.Type != "test_type_1" {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, "type1")
			}
		}
	})

	t.Run("cluster status", func(t *testing.T) {
		clusterTbl.UpdateClusterStatus(context.TODO(), c4.ID, 9)
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), "", "", "", "0", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total < 4 {
			t.Errorf("ListClusters() total = %v, want %v at least", total, 4)
		}

		if len(clusters) < 4 {
			t.Errorf("ListClusters() clusters len = %v, want %v at least", len(clusters), 4)
		}

		for _, v := range clusters {
			if v.Status != 0 {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, 0)
			}
		}
	})

	t.Run("cluster tag", func(t *testing.T) {
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), "", "", "", "", "tag", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 2 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 2)
		}

		if len(clusters) != 2 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 2)
		}

		for _, v := range clusters {
			if !strings.Contains(v.Tags, "tag") {
				t.Errorf("ListClusters() clusters = %v, want cluster tag contains = %v", clusters[0], "tag")
			}
		}
	})

	t.Run("cluster empty", func(t *testing.T) {
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), "", "", "", "", "", 0, 5)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total < 5 {
			t.Errorf("ListClusters() total = %v, want %v at least", total, 5)
		}

		if len(clusters) != 5 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 5)
		}

		for _, v := range clusters {
			if v.ID == "" {
				t.Errorf("ListClusters() clusters = %v", v)
			}
		}
	})

	t.Run("cluster offset", func(t *testing.T) {
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), "", "", "test_type_1", "", "", 1, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}

		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 3)
		}

		if len(clusters) != 2 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 2)
		}

		for _, v := range clusters {
			if v.Type != "test_type_1" {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, "type1")
			}
		}
	})

	t.Run("cluster limit", func(t *testing.T) {
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), "", "", "test_type_1", "", "", 1, 1)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 3)
		}

		if len(clusters) != 1 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 1)
		}

		for _, v := range clusters {
			if v.Type != "test_type_1" {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, "type1")
			}
		}
	})

	t.Run("cluster offset 10", func(t *testing.T) {
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), "", "", "test_type_1", "", "", 10, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 3)
		}

		if len(clusters) != 0 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 0)
		}

	})

	t.Run("cluster no result", func(t *testing.T) {
		clusters, total, err := clusterTbl.ListClusters(context.TODO(), "ffffff", "", "type1", "", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 0 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 0)
		}

		if len(clusters) != 0 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 0)
		}

	})

}

func TestListClusterDetails(t *testing.T) {
	for i := 0; i < 10; i++ {
		MetaDB.Create(&Cluster{
			Entity:  Entity{TenantId: "111"},
			Type:    "someType",
			Name:    "otherCluster",
			Tags:    "1,tag,2",
			OwnerId: "me",
		})
	}
	cluster1 := &Cluster{
		Entity:  Entity{TenantId: "111"},
		Type:    "someType",
		Name:    "cluster1",
		Tags:    "",
		OwnerId: "me",
	}
	MetaDB.Create(cluster1)
	defer MetaDB.Delete(cluster1)

	f, _ := CreateFlow(MetaDB, "flow1", "flow1", cluster1.ID, "111")
	defer MetaDB.Delete(f)
	clusterTbl := Dao.ClusterManager()
	cluster1, _, _ = clusterTbl.UpdateComponentDemand(context.TODO(), cluster1.ID, "demand1", "111")
	cluster1, _ = clusterTbl.UpdateClusterFlowId(context.TODO(), cluster1.ID, f.ID)
	cluster1, _ = clusterTbl.UpdateTopologyConfig(context.TODO(), cluster1.ID, "tiup1", "111")

	cluster3 := &Cluster{
		Entity:  Entity{TenantId: "111"},
		Type:    "otherType",
		Name:    "whatever",
		Tags:    "",
		OwnerId: "me",
	}
	MetaDB.Create(cluster3)

	t.Run("normal", func(t *testing.T) {
		results, total, err := clusterTbl.ListClusterDetails(context.TODO(), "", "", "someType", "", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 11 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 11)
		}

		if len(results) != 10 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(results), 10)
		}

		if results[0].Cluster.ID != cluster1.ID ||
			results[0].Cluster.Name != cluster1.Name {
			t.Errorf("ListClusters() clusters = %v, want = %v", results[0], cluster1)
		}

		if results[0].DemandRecord.Content != "demand1" {
			t.Errorf("ListClusters() DemandRecord = %v, want = %v", results[0].DemandRecord.Content, "demand1")
		}
		if results[0].TopologyConfig.Content != "tiup1" {
			t.Errorf("ListClusters() TopologyConfig = %v, want = %v", results[0].TopologyConfig.Content, "tiup1")
		}
		if results[0].Flow.Name != "flow1" {
			t.Errorf("ListClusters() Flow = %v, want = %v", results[0].Flow.Name, "flow1")
		}
	})

}

func TestSaveBackupRecord(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	clusterTbl := Dao.ClusterManager()
	t.Run("normal", func(t *testing.T) {
		record := &dbpb.DBBackupRecordDTO{
			TenantId:     "111",
			ClusterId:    "111",
			StartTime:    time.Now().Unix(),
			EndTime:      time.Now().Unix(),
			BackupType:   "FULL",
			BackupMethod: "ALL",
			OperatorId:   "operator1",
			FilePath:     "path1",
			FlowId:       1,
			Size:         0,
		}
		gotDo, err := clusterTbl.SaveBackupRecord(context.TODO(), record)
		if err != nil {
			t.Errorf("SaveBackupRecord() error = %v", err)
			return
		}
		if gotDo.ID == 0 {
			t.Errorf("SaveBackupRecord() gotDoId == 0")
			return
		}
	})
}

func TestSaveRecoverRecord(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	clusterTbl := Dao.ClusterManager()
	t.Run("normal", func(t *testing.T) {
		gotDo, err := clusterTbl.SaveRecoverRecord(context.TODO(), "111", "111", "operator1", 1, 1)
		if err != nil {
			t.Errorf("SaveRecoverRecord() error = %v", err)
			return
		}
		if gotDo.ID == 0 {
			t.Errorf("SaveRecoverRecord() gotDoId == 0")
			return
		}
	})
}

func TestDeleteBackupRecord(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	clusterTbl := Dao.ClusterManager()
	rcd := &dbpb.DBBackupRecordDTO{
		TenantId:     "111",
		ClusterId:    "111",
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix(),
		BackupType:   "FULL",
		BackupMethod: "ALL",
		OperatorId:   "operator1",
		FilePath:     "path1",
		FlowId:       1,
		Size:         0,
	}
	record, _ := clusterTbl.SaveBackupRecord(context.TODO(), rcd)
	t.Run("normal", func(t *testing.T) {
		got, err := clusterTbl.DeleteBackupRecord(context.TODO(), record.ID)
		if err != nil {
			t.Errorf("DeleteBackupRecord() error = %v", err)
			return
		}
		if got.ID != record.ID {
			t.Errorf("DeleteBackupRecord() error, want id = %v, got = %v", record.ID, got.ID)
			return
		}
		//TODO
		/*if !got.DeletedAt.Valid {
			t.Errorf("DeleteBackupRecord() error, DeletedAt %v valid", got.DeletedAt)
			return
		}*/
	})
	t.Run("no record", func(t *testing.T) {
		_, err := clusterTbl.DeleteBackupRecord(context.TODO(), 999999)
		if err == nil {
			// TODO: Delete a non-existed record return no error by now
			//t.Errorf("DeleteBackupRecord() want error")
			return
		}
	})
}

func TestListBackupRecords(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.MetaDBService)
	brTbl := Dao.ClusterManager()
	flow, _ := CreateFlow(MetaDB, "backup", "backup", "111", "111")
	defer MetaDB.Delete(flow)
	record := &dbpb.DBBackupRecordDTO{
		TenantId:     "111",
		ClusterId:    "TestListBackupRecords",
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix(),
		BackupType:   "FULL",
		BackupMethod: "ALL",
		OperatorId:   "operator1",
		FilePath:     "path1",
		FlowId:       int64(flow.ID),
		Size:         0,
	}
	_, _ = brTbl.SaveBackupRecord(context.TODO(), record)
	_, _ = brTbl.SaveBackupRecord(context.TODO(), record)
	_, _ = brTbl.SaveBackupRecord(context.TODO(), record)
	_, _ = brTbl.SaveBackupRecord(context.TODO(), record)
	_, _ = brTbl.SaveBackupRecord(context.TODO(), record)
	_, _ = brTbl.SaveBackupRecord(context.TODO(), record)
	_, _ = brTbl.SaveBackupRecord(context.TODO(), record)

	t.Run("normal", func(t *testing.T) {
		dos, total, err := brTbl.ListBackupRecords(context.TODO(), "TestListBackupRecords", 0, 0, "", 2, 2)
		if err != nil {
			t.Errorf("ListBackupRecords() error = %v", err)
			return
		}
		if total != 7 {
			t.Errorf("ListBackupRecords() error, want total = %v, got = %v", 7, total)
			return
		}

		if len(dos) != 2 {
			t.Errorf("ListBackupRecords() error, want length = %v, got = %v", 2, len(dos))
			return
		}

		if dos[1].BackupRecord.ClusterId != "TestListBackupRecords" {
			t.Errorf("ListBackupRecords() error, want ClusterId = %v, got = %v", "111", dos[1].BackupRecord.ClusterId)
			return
		}

		if dos[0].BackupRecord.ID <= dos[1].BackupRecord.ID {
			t.Errorf("ListBackupRecords() error, want order by id desc, got = %v", dos)
			return
		}

		if int64(dos[0].Flow.ID) != dos[0].BackupRecord.FlowId {
			t.Errorf("ListBackupRecords() error, want FlowId = %v, got = %v", dos[0].BackupRecord.FlowId, dos[0].Flow.ID)
			return
		}
	})
}

func TestSaveParameters(t *testing.T) {
	brTbl := Dao.ClusterManager()
	t.Run("normal", func(t *testing.T) {
		gotDo, err := brTbl.SaveParameters(context.TODO(), "111", "111", "someone", 1, "content1")
		if err != nil {
			t.Errorf("SaveParameters() error = %v", err)
			return
		}
		if gotDo.ID == 0 {
			t.Errorf("SaveParameters() gotDoId == 0")
			return
		}
	})

}

func TestGetCurrentParameters(t *testing.T) {
	prTbl := Dao.ClusterManager()
	prTbl.SaveParameters(context.TODO(), "111", "111", "someone", 1, "content1")
	prTbl.SaveParameters(context.TODO(), "111", "111", "someone", 1, "content2")
	prTbl.SaveParameters(context.TODO(), "111", "111", "someone", 1, "wanted")
	prTbl.SaveParameters(context.TODO(), "111", "222", "someone", 1, "content4")

	t.Run("normal", func(t *testing.T) {
		gotDo, err := prTbl.GetCurrentParameters(context.TODO(), "111")
		if err != nil {
			t.Errorf("SaveRecoverRecord() error = %v", err)
			return
		}
		if gotDo.ID == 0 {
			t.Errorf("SaveRecoverRecord() gotDoId == 0")
			return
		}

		if gotDo.Content != "wanted" {
			t.Errorf("SaveRecoverRecord() gotDo.Content == %v", gotDo.Content)
			return
		}
	})
}

var defaultTenantId = "defaultTenantId"

func TestFetchCluster(t *testing.T) {
	clusterTbl := Dao.ClusterManager()
	cluster, _ := clusterTbl.CreateCluster(context.TODO(), Cluster{
		Entity: Entity{TenantId: defaultTenantId},
		Name: "TestFetchCluster",
		DbPassword: "tt.args.DbPassword",
		Type: "TiDB",
		Version: "v5.0.0",
		Tls: true,
		OwnerId: "TestFetchCluster.ownerId",
	})
	t.Run("normal", func(t *testing.T) {
		gotResult, err := clusterTbl.FetchCluster(context.TODO(), cluster.ID)
		if err != nil {
			t.Errorf("FetchCluster() error = %v", err)
			return
		}
		if gotResult.Cluster.ID != cluster.ID {
			t.Errorf("FetchCluster() want id = %v, got = %v", cluster.ID, gotResult.Cluster.ID)
			return
		}
	})
	t.Run("no result", func(t *testing.T) {
		_, err := clusterTbl.FetchCluster(context.TODO(), "what ever")
		if err == nil {
			t.Errorf("FetchCluster() want error")
			return
		}
	})
	t.Run("with demand", func(t *testing.T) {
		cluster, demand, _ := clusterTbl.UpdateComponentDemand(context.TODO(), cluster.ID, "demand content", defaultTenantId)
		gotResult, err := clusterTbl.FetchCluster(context.TODO(), cluster.ID)
		if err != nil {
			t.Errorf("FetchCluster() error = %v", err)
			return
		}
		if gotResult.Cluster.ID != cluster.ID {
			t.Errorf("FetchCluster() want id = %v, got = %v", cluster.ID, gotResult.Cluster.ID)
			return
		}
		if gotResult.Cluster.CurrentDemandId != demand.ID {
			t.Errorf("FetchCluster() want CurrentDemandId = %v, got = %v", cluster.CurrentDemandId, gotResult.Cluster.CurrentDemandId)
			return
		}
		if gotResult.DemandRecord.ID != demand.ID {
			t.Errorf("FetchCluster() want DemandRecord id = %v, got = %v", cluster.CurrentDemandId, gotResult.Cluster.CurrentDemandId)
			return
		}
		if gotResult.DemandRecord.Content != demand.Content {
			t.Errorf("FetchCluster() want DemandRecord content = %v, got = %v", demand.Content, gotResult.DemandRecord.Content)
			return
		}
	})
	t.Run("with demand err", func(t *testing.T) {
		cluster, demand, _ := clusterTbl.UpdateComponentDemand(context.TODO(), cluster.ID, "demand content", defaultTenantId)
		MetaDB.Delete(demand)
		_, err := clusterTbl.FetchCluster(context.TODO(), cluster.ID)
		if err == nil {
			t.Errorf("FetchCluster() want error")
			return
		}
		clusterTbl.UpdateComponentDemand(context.TODO(), cluster.ID, "demand content", defaultTenantId)
	})
	t.Run("with config", func(t *testing.T) {
		cluster, _ := clusterTbl.UpdateTopologyConfig(context.TODO(), cluster.ID, "config content", defaultTenantId)
		gotResult, err := clusterTbl.FetchCluster(context.TODO(), cluster.ID)
		if err != nil {
			t.Errorf("FetchCluster() error = %v", err)
			return
		}
		if gotResult.Cluster.ID != cluster.ID {
			t.Errorf("FetchCluster() want id = %v, got = %v", cluster.ID, gotResult.Cluster.ID)
			return
		}
		if gotResult.Cluster.CurrentTopologyConfigId != cluster.CurrentTopologyConfigId {
			t.Errorf("FetchCluster() want Current"+
				"TopologyConfigId = %v, got = %v", cluster.CurrentTopologyConfigId, gotResult.Cluster.CurrentTopologyConfigId)
			return
		}
		if gotResult.TopologyConfig.ID != cluster.CurrentTopologyConfigId {
			t.Errorf("FetchCluster() want TopologyConfig id = %v, got = %v", cluster.CurrentTopologyConfigId, gotResult.TopologyConfig.ID)
			return
		}
		if gotResult.TopologyConfig.Content != "config content" {
			t.Errorf("FetchCluster() want TopologyConfig content = %v, got = %v", "config content", gotResult.TopologyConfig.Content)
			return
		}
	})
	t.Run("with config err", func(t *testing.T) {
		cluster, _ := clusterTbl.UpdateTopologyConfig(context.TODO(), cluster.ID, "config content", defaultTenantId)
		MetaDB.Model(&TopologyConfig{}).Where("id = ?", cluster.CurrentTopologyConfigId).Delete(&TopologyConfig{})
		_, err := clusterTbl.FetchCluster(context.TODO(), cluster.ID)
		if err == nil {
			t.Errorf("FetchCluster() want error")
			return
		}
		clusterTbl.UpdateTopologyConfig(context.TODO(), cluster.ID, "config content", defaultTenantId)
	})
	t.Run("with flow", func(t *testing.T) {
		flow, _ := CreateFlow(MetaDB, "whatever", "whatever", "whatever", "111")
		defer MetaDB.Delete(flow)
		cluster, _ := clusterTbl.UpdateClusterFlowId(context.TODO(), cluster.ID, flow.ID)
		gotResult, err := clusterTbl.FetchCluster(context.TODO(), cluster.ID)
		if err != nil {
			t.Errorf("FetchCluster() error = %v", err)
			return
		}
		if gotResult.Cluster.ID != cluster.ID {
			t.Errorf("FetchCluster() want id = %v, got = %v", cluster.ID, gotResult.Cluster.ID)
			return
		}
		if gotResult.Cluster.CurrentFlowId != flow.ID {
			t.Errorf("FetchCluster() want CurrentFlowId = %v, got = %v", flow.ID, gotResult.Cluster.CurrentFlowId)
			return
		}
		if gotResult.Flow.ID != flow.ID {
			t.Errorf("FetchCluster() want flow id = %v, got = %v", flow.ID, gotResult.Flow.ID)
			return
		}
		if gotResult.Flow.StatusAlias != flow.StatusAlias {
			t.Errorf("FetchCluster() want flow StatusAlias = %v, got = %v", flow.StatusAlias, gotResult.Flow.StatusAlias)
			return
		}
	})
	t.Run("with flow error", func(t *testing.T) {
		cluster, _ := clusterTbl.UpdateClusterFlowId(context.TODO(), cluster.ID, 555555)
		_, err := clusterTbl.FetchCluster(context.TODO(), cluster.ID)
		if err == nil {
			t.Errorf("FetchCluster() want error")
			return
		}
		clusterTbl.UpdateClusterFlowId(context.TODO(), cluster.ID, 0)
	})

}

func TestDAOClusterManager_UpdateClusterInfo(t *testing.T) {
	clusterTbl := Dao.ClusterManager()
	t.Run("normal update info", func(t *testing.T) {
		cluster := &Cluster{
			Entity:  Entity{TenantId: "111"},
			OwnerId: "ttt",
		}
		MetaDB.Create(cluster)
		defer MetaDB.Delete(cluster)

		cluster, err := clusterTbl.UpdateClusterInfo(context.TODO(), cluster.ID, "newName", "TiDB", "v5.0.0", "213", true)

		if err != nil {
			t.Errorf("UpdateClusterInfo() error = %v", err)
		}

		if cluster.ID == "" {
			t.Errorf("UpdateClusterInfo() cluster.ID empty")
		}

		if cluster.Name != "newName" {
			t.Errorf("UpdateClusterInfo() want name = %s, got %s", "newName", cluster.Name)
		}
		if cluster.Type != "TiDB" {
			t.Errorf("UpdateClusterInfo() want type = %s, got %s", "TiDB", cluster.Type)
		}
		if cluster.Tls != true {
			t.Errorf("UpdateClusterInfo() want tls true, got false")
		}
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, err := clusterTbl.UpdateClusterInfo(context.TODO(), "", "newName", "TiDB", "v5.0.0", "213", true)
		if err == nil {
			t.Errorf("UpdateClusterInfo() want error, got nil")
		}

	})
}
