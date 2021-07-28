package models

import (
	"strings"
	"testing"
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
			func (args args, cluster *ClusterDO) bool{ return args.ClusterName == cluster.ClusterName},
			func (args args, cluster *ClusterDO) bool{ return cluster.ID != ""},
		}},
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
	t.Run("normal delete", func(t *testing.T) {
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
		ClusterType: "type1",
		ClusterName: "test",
		Tags: ",tag,",
		OwnerId: "ttt",
	})
	MetaDB.Create(&ClusterDO{
		Entity : Entity{TenantId: "111"},
		ClusterType: "type1",
		ClusterName: "1111test",
		Tags: "tag,",
		OwnerId: "ttt",
	})
	MetaDB.Create(&ClusterDO{
		Entity : Entity{TenantId: "111"},
		ClusterType: "whatever",
		ClusterName: "test1111",
		Tags: ",tag",
		OwnerId: "ttt",
	})
	cluster := &ClusterDO{
		Entity : Entity{TenantId: "111"},
		ClusterType: "type1",
		ClusterName: "whatever",
		Tags: "1,tag,2",
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

		if clusters[0].ID != cluster.ID || clusters[0].ClusterName != cluster.ClusterName{
			t.Errorf("ListClusters() clusters = %v, want = %v", clusters[0], cluster)
		}
	})

	t.Run("cluster name", func(t *testing.T) {
		clusters,total,err := ListClusters("", "test", "", "", "", 0, 10)

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
			if !strings.Contains(v.ClusterName, "test") {
				t.Errorf("ListClusters() clusters = %v, want cluster name contains = %v", v, "test")
			}
		}
	})

	t.Run("cluster type", func(t *testing.T) {
		clusters,total,err := ListClusters("", "", "type1", "", "", 0, 10)

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
			if v.ClusterType != "type1" {
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
		if total != 4 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 4)
		}

		if len(clusters) != 4 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 4)
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
		clusters,total,err := ListClusters("", "", "", "", "", 0, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 5 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 5)
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
		clusters,total,err := ListClusters("", "", "type1", "", "", 1, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 2)
		}

		if len(clusters) != 2 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 2)
		}

		for _,v := range clusters {
			if v.ClusterType != "type1" {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, "type1")
			}
		}
	})

	t.Run("cluster limit", func(t *testing.T) {
		clusters,total,err := ListClusters("", "", "type1", "", "", 1, 1)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 2)
		}

		if len(clusters) != 1 {
			t.Errorf("ListClusters() clusters len = %v, want = %v", len(clusters), 1)
		}

		for _,v := range clusters {
			if v.ClusterType != "type1" {
				t.Errorf("ListClusters() clusters = %v, wantClusterType = %v", v, "type1")
			}
		}
	})

	t.Run("cluster offset 10", func(t *testing.T) {
		clusters,total,err := ListClusters("", "", "type1", "", "", 10, 10)

		if err != nil {
			t.Errorf("ListClusters() error = %v", err)
		}
		if total != 3 {
			t.Errorf("ListClusters() total = %v, want = %v", total, 2)
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
		ClusterType: "someType",
		ClusterName: "cluster1",
		Tags: "",
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
			ClusterType: "someType",
			ClusterName: "otherCluster",
			Tags: "1,tag,2",
			OwnerId: "me",
		})
	}
	cluster3 := &ClusterDO {
		Entity : Entity{TenantId: "111"},
		ClusterType: "otherType",
		ClusterName: "whatever",
		Tags: "",
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
			results[0].Cluster.ClusterName != cluster1.ClusterName {
			t.Errorf("ListClusters() clusters = %v, want = %v", results[0], cluster1)
		}

		if results[0].DemandRecord.Content != "demand1" {
			t.Errorf("ListClusters() DemandRecord = %v, want = %v", results[0].DemandRecord.Content, "demand1")
		}
		if results[0].TiUPConfig.Content != "tiup1" {
			t.Errorf("ListClusters() TiUPConfig = %v, want = %v", results[0].TiUPConfig.Content, "tiup1")
		}
		if results[0].Flow.FlowName != "flow1" {
			t.Errorf("ListClusters() Flow = %v, want = %v", results[0].Flow.FlowName, "flow1")
		}
	})

}