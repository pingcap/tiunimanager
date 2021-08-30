package models

import (
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
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
		name        string
		args        args
		wantErr     bool
		wants       []func (args args, cluster *ClusterDO) bool
	}{
		{"normal create", args{TenantId: "111", ClusterName: "testCluster", OwnerId: "111"}, false, []func (args args, cluster *ClusterDO) bool{
			func (args args, cluster *ClusterDO) bool{ return args.TenantId == cluster.TenantId},
			func (args args, cluster *ClusterDO) bool{ return args.ClusterName == cluster.Name },
			func (args args, cluster *ClusterDO) bool{ return cluster.ID != ""},
			func (args args, cluster *ClusterDO) bool{ return cluster.Code == "testCluster"},
		}},
		{"chinese name", args{TenantId: "111", ClusterName: "中文测试集群", OwnerId: "111"}, false, []func (args args, cluster *ClusterDO) bool{
			func (args args, cluster *ClusterDO) bool{ return args.TenantId == cluster.TenantId},
			func (args args, cluster *ClusterDO) bool{ return args.ClusterName == cluster.Name },
			func (args args, cluster *ClusterDO) bool{ return cluster.ID != ""},
			func (args args, cluster *ClusterDO) bool{ return cluster.Code == "_zhong_wen_ce_shi_ji_qun"},
		}},
		{"error", args{TenantId: "111", ClusterName: "中文测试集群"}, true,[]func (args args, cluster *ClusterDO) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCluster, err := CreateCluster(tt.args.ClusterName, tt.args.DbPassword, tt.args.ClusterType, tt.args.ClusterVersion, tt.args.Tls, tt.args.Tags, tt.args.OwnerId, tt.args.TenantId)
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
	t.Run("normal", func(t *testing.T) {
		cluster := &ClusterDO{
			Entity : Entity{TenantId: "111"},
			OwnerId: "111",
		}
		MetaDB.Create(cluster)

		newCluster, err := DeleteCluster(cluster.ID)

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
		_, err := DeleteCluster("TestDeleteClusterId")

		if err == nil {
			t.Errorf("DeleteCluster() error = %v", err)
		}
	})
	t.Run("empty clusterId", func(t *testing.T) {
		_, err := DeleteCluster("")

		if err == nil {
			t.Errorf("DeleteCluster() error = %v", err)
		}
	})
}

func TestUpdateClusterDemand(t *testing.T) {
	t.Run("normal update demand", func(t *testing.T) {
		cluster := &ClusterDO{
			Entity : Entity{TenantId: "111"},
			OwnerId: "ttt",
		}
		MetaDB.Create(cluster)

		demandId := cluster.CurrentDemandId

		cluster, demand, err := UpdateClusterDemand(cluster.ID, "aaa", cluster.TenantId)
		if err != nil {
			t.Errorf("UpdateClusterDemand() error = %v", err)
		}

		if demand == nil || demand.ID == 0 {
			t.Errorf("UpdateClusterDemand() demand = %v", demand)
		}

		if cluster.CurrentDemandId == 0 || cluster.CurrentDemandId <= demandId {
			t.Errorf("UpdateClusterDemand() new demand id = %v", cluster.CurrentDemandId)
		}

		if cluster.ID == ""{
			t.Errorf("UpdateClusterDemand() cluster.ID empty")
		}
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, _, err := UpdateClusterDemand("", "aaa", "111")
		if err == nil {
			t.Errorf("UpdateClusterDemand() error = %v", err)
		}

	})
}

func TestUpdateClusterFlowId(t *testing.T) {
	t.Run("normal update demand", func(t *testing.T) {
		cluster := &ClusterDO{
			Entity : Entity{TenantId: "111"},
			OwnerId: "ttt",
		}
		MetaDB.Create(cluster)

		flowId := uint(111)

		cluster, err := UpdateClusterFlowId(cluster.ID, flowId)

		if err != nil {
			t.Errorf("UpdateClusterFlowId() error = %v", err)
		}

		if cluster.CurrentFlowId != flowId {
			t.Errorf("UpdateClusterFlowId() want flowId = %v, got = %v", flowId, cluster.CurrentFlowId)
		}

		if cluster.ID == ""{
			t.Errorf("UpdateClusterFlowId() cluster.ID empty")
		}
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, err := UpdateClusterFlowId("", 111)
		if err == nil {
			t.Errorf("UpdateClusterFlowId() error = %v", err)
		}

	})
}

func TestUpdateClusterStatus(t *testing.T) {
	t.Run("normal update status", func(t *testing.T) {
		cluster := &ClusterDO{
			Entity : Entity{TenantId: "111"},
			OwnerId: "ttt",
		}
		MetaDB.Create(cluster)

		status := int8(2)

		cluster, err := UpdateClusterStatus(cluster.ID, status)

		if err != nil {
			t.Errorf("UpdateClusterStatus() error = %v", err)
		}

		if cluster.Status != status {
			t.Errorf("UpdateClusterFlowId() want status = %v, got = %v", status, cluster.Status)
		}


		if cluster.ID == ""{
			t.Errorf("UpdateClusterFlowId() cluster.ID empty")
		}
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, err := UpdateClusterStatus("", 2)
		if err == nil {
			t.Errorf("UpdateClusterFlowId() error = %v", err)
		}

	})
}

func TestUpdateTiUPConfig(t *testing.T) {
	t.Run("normal update config", func(t *testing.T) {
		cluster := &ClusterDO{
			Entity : Entity{TenantId: "111"},
			OwnerId: "ttt",
		}
		MetaDB.Create(cluster)

		currentConfigId := cluster.CurrentTiupConfigId

		cluster, err := UpdateTiUPConfig(cluster.ID, "aaa", cluster.TenantId)
		if err != nil {
			t.Errorf("UpdateTiUPConfig() error = %v", err)
		}

		if cluster.CurrentTiupConfigId == 0 || cluster.CurrentTiupConfigId <= currentConfigId {
			t.Errorf("UpdateTiUPConfig() new config id = %v, current config id = %v", cluster.CurrentTiupConfigId, currentConfigId)
		}

		if cluster.ID == "" {
			t.Errorf("UpdateTiUPConfig() cluster.ID empty")
		}
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, _, err := UpdateClusterDemand("", "aaa", "111")
		if err == nil {
			t.Errorf("UpdateClusterDemand() error = %v", err)
		}

	})
}

func TestListClusters(t *testing.T) {
	MetaDB.Create(&ClusterDO{
		Entity : Entity{TenantId: "111"},
		Type:    "test_type_1",
		Name:    "test_cluster_name",
		Tags:    ",tag,",
		OwnerId: "ttt",
	})
	MetaDB.Create(&ClusterDO{
		Entity : Entity{TenantId: "111"},
		Type:    "test_type_1",
		Name:    "1111test",
		Tags:    "tag,",
		OwnerId: "ttt",
	})
	MetaDB.Create(&ClusterDO{
		Entity : Entity{TenantId: "111"},
		Type:    "whatever",
		Name:    "test_cluster_name",
		Tags:    ",tag",
		OwnerId: "ttt",
	})
	cluster := &ClusterDO{
		Entity : Entity{TenantId: "111"},
		Type:    "test_type_1",
		Name:    "whatever",
		Tags:    "1,tag,2",
		OwnerId: "ttt",
	}
	MetaDB.Create(cluster)

	MetaDB.Create(&ClusterDO{
		Entity : Entity{TenantId: "111"},
		OwnerId: "ttt",
	})

	t.Run("cluster id", func(t *testing.T) {
		clusters,total,err := ListClusters(cluster.ID, "", "", "", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 1 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 1)
		}

		if len(clusters) != 1 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 1)
		}

		if clusters[0].ID != cluster.ID || clusters[0].Name != cluster.Name {
			t.Errorf("ListClusters() clusters = %v, want = %v", clusters[0], cluster)
		}
	})

	t.Run("cluster name", func(t *testing.T) {
		clusters,total,err := ListClusters("", "test_cluster_name", "", "", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 2 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 2)
		}

		if len(clusters) != 2 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 2)
		}

		for _,v := range clusters {
			if !strings.Contains(v.Name, "test") {
				t.Errorf("ListClusters() clusters = %v, want cluster name contains = %v", v, "test")
			}
		}
	})

	t.Run("cluster type", func(t *testing.T) {
		clusters,total,err := ListClusters("", "", "test_type_1", "", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 3)
		}

		if len(clusters) != 3 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 3)
		}

		for _,v := range clusters {
			if v.Type != "test_type_1" {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, "type1")
			}
		}
	})

	t.Run("cluster status", func(t *testing.T) {
		UpdateClusterStatus(cluster.ID, 9)
		clusters,total,err := ListClusters("", "", "", "0", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total < 4 {
			t.Errorf("ListClusters() total = %v, want %v at least", total, 4)
		}

		if len(clusters) < 4 {
			t.Errorf("ListClusters() clusters len = %v, want %v at least", len(clusters), 4)
		}

		for _,v := range clusters {
			if v.Status != 0 {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, 0)
			}
		}
	})

	t.Run("cluster tag", func(t *testing.T) {
		clusters,total,err := ListClusters("", "", "", "", "tag", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 2 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 2)
		}

		if len(clusters) != 2 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 2)
		}

		for _,v := range clusters {
			if !strings.Contains(v.Tags, "tag") {
				t.Errorf("ListClusters() clusters = %v, want cluster tag contains = %v", clusters[0], "tag")
			}
		}
	})

	t.Run("cluster empty", func(t *testing.T) {
		clusters,total,err := ListClusters("", "", "", "", "", 0, 5)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total < 5 {
			t.Errorf("ListClusters() total = %v, want %v at least", total, 5)
		}

		if len(clusters) != 5 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 5)
		}

		for _,v := range clusters {
			if v.ID == "" {
				t.Errorf("ListClusters() clusters = %v", v)
			}
		}
	})

	t.Run("cluster offset", func(t *testing.T) {
		clusters,total,err := ListClusters("", "", "test_type_1", "", "", 1, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 3)
		}

		if len(clusters) != 2 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 2)
		}

		for _,v := range clusters {
			if v.Type != "test_type_1" {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, "type1")
			}
		}
	})

	t.Run("cluster limit", func(t *testing.T) {
		clusters,total,err := ListClusters("", "", "test_type_1", "", "", 1, 1)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 3)
		}

		if len(clusters) != 1 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 1)
		}

		for _,v := range clusters {
			if v.Type != "test_type_1" {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, "type1")
			}
		}
	})

	t.Run("cluster offset 10", func(t *testing.T) {
		clusters,total,err := ListClusters("", "", "test_type_1", "", "", 10, 10)

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
		clusters,total,err := ListClusters("ffffff", "", "type1", "", "", 0, 10)

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
	cluster1 := &ClusterDO{
		Entity : Entity{TenantId: "111"},
		Type:    "someType",
		Name:    "cluster1",
		Tags:    "",
		OwnerId: "me",
	}
	MetaDB.Create(cluster1)
	f, _ := CreateFlow("flow1", "flow1", cluster1.ID)

	cluster1 ,_,_ = UpdateClusterDemand(cluster1.ID, "demand1", "111")
	cluster1 ,_ = UpdateClusterFlowId(cluster1.ID, f.ID)
	cluster1 ,_ = UpdateTiUPConfig(cluster1.ID, "tiup1", "111")

	for i := 0; i < 10; i++ {
		MetaDB.Create(&ClusterDO{
			Entity : Entity{TenantId: "111"},
			Type:    "someType",
			Name:    "otherCluster",
			Tags:    "1,tag,2",
			OwnerId: "me",
		})
	}
	cluster3 := &ClusterDO {
		Entity : Entity{TenantId: "111"},
		Type:    "otherType",
		Name:    "whatever",
		Tags:    "",
		OwnerId: "me",
	}
	MetaDB.Create(cluster3)

	t.Run("normal", func(t *testing.T) {
		results,total,err := ListClusterDetails("", "", "someType", "", "", 0, 10)

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
		if results[0].TiUPConfig.Content != "tiup1" {
			t.Errorf("ListClusters() TiUPConfig = %v, want = %v", results[0].TiUPConfig.Content, "tiup1")
		}
		if results[0].Flow.Name != "flow1" {
			t.Errorf("ListClusters() Flow = %v, want = %v", results[0].Flow.Name, "flow1")
		}
	})

}

func TestSaveBackupRecord(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		record := &dbPb.DBBackupRecordDTO{
			TenantId:	"111",
			ClusterId:  "111",
			StartTime:  time.Now().Unix(),
			EndTime:    time.Now().Unix(),
			BackupRange:"FULL",
			BackupType: "ALL",
			OperatorId: "operator1",
			FilePath:   "path1",
			FlowId: 	1,
			Size:		0,
		}
		gotDo, err := SaveBackupRecord(record)
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
	t.Run("normal", func(t *testing.T) {
		gotDo, err := SaveRecoverRecord("111", "111", "operator1", 1,1)
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
	rcd := &dbPb.DBBackupRecordDTO{
		TenantId:	"111",
		ClusterId:  "111",
		StartTime:  time.Now().Unix(),
		EndTime:    time.Now().Unix(),
		BackupRange:"FULL",
		BackupType: "ALL",
		OperatorId: "operator1",
		FilePath:   "path1",
		FlowId: 	1,
		Size:		0,
	}
	record, _ :=SaveBackupRecord(rcd)
	t.Run("normal", func(t *testing.T) {
		got, err := DeleteBackupRecord(record.ID)
		if err != nil {
			t.Errorf("DeleteBackupRecord() error = %v", err)
			return
		}
		if got.ID != record.ID {
			t.Errorf("DeleteBackupRecord() error, want id = %v, got = %v", record.ID, got.ID)
			return
		}
		if !got.DeletedAt.Valid {
			t.Errorf("DeleteBackupRecord() error, DeletedAt valid")
			return
		}
	})
	t.Run("no record", func(t *testing.T) {
		_, err := DeleteBackupRecord(999999)
		if err == nil {
			t.Errorf("DeleteBackupRecord() want error")
			return
		}
	})

}

func TestListBackupRecords(t *testing.T) {
	flow, _ := CreateFlow("backup", "backup", "111")
	record := &dbPb.DBBackupRecordDTO{
		TenantId:	"111",
		ClusterId:  "111",
		StartTime:  time.Now().Unix(),
		EndTime:    time.Now().Unix(),
		BackupRange:"FULL",
		BackupType: "ALL",
		OperatorId: "operator1",
		FilePath:   "path1",
		FlowId: 	int64(flow.ID),
		Size:		0,
	}
	SaveBackupRecord(record)
	SaveBackupRecord(record)
	SaveBackupRecord(record)
	SaveBackupRecord(record)
	SaveBackupRecord(record)
	SaveBackupRecord(record)
	SaveBackupRecord(record)

	t.Run("normal", func(t *testing.T) {
		dos , total , err := ListBackupRecords("11111", 2,2)
		if err != nil {
			t.Errorf("ListBackupRecords() error = %v", err)
			return
		}
		if total != 5 {
			t.Errorf("ListBackupRecords() error, want total = %v, got = %v", 5, total)
			return
		}

		if len(dos) != 2 {
			t.Errorf("ListBackupRecords() error, want length = %v, got = %v", 2, len(dos))
			return
		}

		if dos[1].BackupRecordDO.ClusterId != "11111" {
			t.Errorf("ListBackupRecords() error, want ClusterId = %v, got = %v", "111", dos[1].BackupRecordDO.ClusterId)
			return
		}


		if dos[0].BackupRecordDO.ID <= dos[1].BackupRecordDO.ID {
			t.Errorf("ListBackupRecords() error, want order by id desc, got = %v", dos)
			return
		}

		if int64(dos[0].Flow.ID) != dos[0].BackupRecordDO.FlowId {
			t.Errorf("ListBackupRecords() error, want FlowId = %v, got = %v", dos[0].BackupRecordDO.FlowId,  dos[0].Flow.ID)
			return
		}
	})
}

func TestSaveParameters(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		gotDo, err := SaveParameters("111", "111", "someone", 1,"content1")
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
	SaveParameters("111", "111", "someone", 1,"content1")
	SaveParameters("111", "111", "someone", 1,"content2")
	SaveParameters("111", "111", "someone", 1,"wanted")
	SaveParameters("111", "222", "someone", 1,"content4")

	t.Run("normal", func(t *testing.T) {
		gotDo, err := GetCurrentParameters("111")
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
	cluster, _ := CreateCluster("TestFetchCluster", "tt.args.DbPassword", "tidb", "v1", true, "","TestFetchCluster.ownerId", defaultTenantId)
	t.Run("normal", func(t *testing.T) {
		gotResult, err := FetchCluster(cluster.ID)
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
		_, err := FetchCluster("what ever")
		if err == nil {
			t.Errorf("FetchCluster() want error")
			return
		}
	})
	t.Run("with demand", func(t *testing.T) {
		cluster, demand, _ :=UpdateClusterDemand(cluster.ID, "demand content", defaultTenantId)
		gotResult, err := FetchCluster(cluster.ID)
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
		cluster, demand, _ := UpdateClusterDemand(cluster.ID, "demand content", defaultTenantId)
		MetaDB.Delete(demand)
		_, err := FetchCluster(cluster.ID)
		if err == nil {
			t.Errorf("FetchCluster() want error")
			return
		}
		UpdateClusterDemand(cluster.ID, "demand content", defaultTenantId)
	})
	t.Run("with config", func(t *testing.T) {
		cluster, _ := UpdateTiUPConfig(cluster.ID, "config content", defaultTenantId)
		gotResult, err := FetchCluster(cluster.ID)
		if err != nil {
			t.Errorf("FetchCluster() error = %v", err)
			return
		}
		if gotResult.Cluster.ID != cluster.ID {
			t.Errorf("FetchCluster() want id = %v, got = %v", cluster.ID, gotResult.Cluster.ID)
			return
		}
		if gotResult.Cluster.CurrentTiupConfigId != cluster.CurrentTiupConfigId {
			t.Errorf("FetchCluster() want CurrentTiupConfigId = %v, got = %v", cluster.CurrentTiupConfigId, gotResult.Cluster.CurrentTiupConfigId)
			return
		}
		if gotResult.TiUPConfig.ID != cluster.CurrentTiupConfigId {
			t.Errorf("FetchCluster() want TiUPConfig id = %v, got = %v", cluster.CurrentTiupConfigId, gotResult.TiUPConfig.ID)
			return
		}
		if gotResult.TiUPConfig.Content != "config content" {
			t.Errorf("FetchCluster() want TiUPConfig content = %v, got = %v", "config content", gotResult.TiUPConfig.Content)
			return
		}
	})
	t.Run("with config err", func(t *testing.T) {
		cluster, _ := UpdateTiUPConfig(cluster.ID, "config content", defaultTenantId)
		MetaDB.Model(&TiUPConfigDO{}).Where("id = ?", cluster.CurrentTiupConfigId).Delete(&TiUPConfigDO{})
		_, err := FetchCluster(cluster.ID)
		if err == nil {
			t.Errorf("FetchCluster() want error")
			return
		}
		UpdateTiUPConfig(cluster.ID, "config content", defaultTenantId)
	})
	t.Run("with flow", func(t *testing.T) {
		flow, _ := CreateFlow("whatever", "whatever", "whatever")
		cluster, _ := UpdateClusterFlowId(cluster.ID, flow.ID)
		gotResult, err := FetchCluster(cluster.ID)
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
		cluster, _ := UpdateClusterFlowId(cluster.ID, 555555)
		_, err := FetchCluster(cluster.ID)
		if err == nil {
			t.Errorf("FetchCluster() want error")
			return
		}
		UpdateClusterFlowId(cluster.ID, 0)
	})

}
