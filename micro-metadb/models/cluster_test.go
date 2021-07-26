package models

import (
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
	})

	t.Run("empty clusterId", func(t *testing.T) {
		_, _, err := UpdateClusterDemand("", "aaa", "111")
		if err == nil {
			t.Errorf("UpdateClusterDemand() error = %v", err)
		}

	})
}
