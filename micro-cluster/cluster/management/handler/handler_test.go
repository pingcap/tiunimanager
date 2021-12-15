/*******************************************************************************
 * @File: handler_test
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/9
*******************************************************************************/

package handler

import (
	ctx "context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestClusterMeta_AddInstances(t *testing.T) {
	meta := &ClusterMeta{}
	{
		computes := make([]structs.ClusterResourceParameterCompute, 0)
		err := meta.AddInstances(ctx.TODO(), computes)
		assert.Error(t, err)
	}

	{
		computes := []structs.ClusterResourceParameterCompute{
			{
				Type:  "TiDB",
				Count: 2,
				Resource: []structs.ClusterResourceParameterComputeResource{
					{
						Zone:         "TEST_Region1,TEST_Zone1",
						Spec:         "4C8G",
						DiskType:     "SSD",
						DiskCapacity: 16,
						Count:        1,
					},
					{
						Zone:         "TEST_Region1,TEST_Zone2",
						Spec:         "8C16G",
						DiskType:     "SSD",
						DiskCapacity: 32,
						Count:        1,
					},
				},
			},
			{
				Type:  "TiKV",
				Count: 1,
				Resource: []structs.ClusterResourceParameterComputeResource{
					{
						Zone:         "TEST_Region1,TEST_Zone2",
						Spec:         "8C16G",
						DiskType:     "SSD",
						DiskCapacity: 256,
						Count:        1,
					},
				},
			},
			{
				Type:  "PD",
				Count: 1,
				Resource: []structs.ClusterResourceParameterComputeResource{
					{
						Zone:         "TEST_Region1,TEST_Zone1",
						Spec:         "4C8G",
						DiskType:     "SSD",
						DiskCapacity: 16,
						Count:        1,
					},
				},
			},
		}

		err := meta.AddInstances(ctx.TODO(), computes)
		assert.Error(t, err)

		meta.Cluster = &management.Cluster{
			Entity:  common.Entity{ID: "testCluster"},
			Version: "v5.2.0",
		}
		err = meta.AddInstances(ctx.TODO(), computes)
		assert.NoError(t, err)
		assert.Equal(t, len(meta.Instances), 3)
		assert.Equal(t, len(meta.Instances["TiDB"]), 2)
		assert.Equal(t, len(meta.Instances["TiKV"]), 1)
		assert.Equal(t, len(meta.Instances["PD"]), 1)
		assert.Equal(t, meta.Instances["TiDB"][0].ClusterID, "testCluster")
		assert.Equal(t, meta.Instances["TiDB"][0].Type, "TiDB")
		assert.Equal(t, meta.Instances["TiDB"][0].Version, "v5.2.0")
		assert.Equal(t, meta.Instances["PD"][0].Status, string(constants.ClusterInstanceInitializing))
	}
}

func TestClusterMeta_GenerateTopologyConfig(t *testing.T) {
	meta := &ClusterMeta{}
	_, err := meta.GenerateTopologyConfig(ctx.TODO())
	assert.Error(t, err)

	{
		meta.Cluster = &management.Cluster{
			Entity:          common.Entity{ID: "testCluster", Status: string(constants.ClusterInitializing)},
			CpuArchitecture: constants.ArchArm64,
			Version:         "v5.3.0",
			Type:            "TiDB",
		}
		meta.Instances = map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity:   common.Entity{ID: "testInstance01", Status: string(constants.ClusterInstanceInitializing)},
					HostIP:   []string{"127.0.0.1"},
					DiskPath: "/test",
					Ports:    []int32{4001, 4002},
				},
			},
			"TiKV": {
				{
					Entity:   common.Entity{ID: "testInstance02", Status: string(constants.ClusterInstanceInitializing)},
					HostIP:   []string{"127.1.0.1"},
					DiskPath: "/test",
					Ports:    []int32{5001, 5002},
				},
			},
			"PD": {
				{
					Entity:   common.Entity{ID: "testInstance03", Status: string(constants.ClusterInstanceInitializing)},
					HostIP:   []string{"127.2.0.1"},
					DiskPath: "/test",
					Ports:    []int32{6001, 6002, 6003, 6004, 6005, 6006},
				},
			},
			"TiFlash": {
				{
					Entity:   common.Entity{ID: "testInstance04", Status: string(constants.ClusterInstanceInitializing)},
					HostIP:   []string{"127.3.0.1"},
					DiskPath: "/test",
					Ports:    []int32{7001, 7002, 7003, 7004, 7005, 7006},
				},
			},
			"TiCDC": {
				{
					Entity:   common.Entity{ID: "testInstance05", Status: string(constants.ClusterInstanceInitializing)},
					HostIP:   []string{"127.4.0.1"},
					DiskPath: "/test",
					Ports:    []int32{8001},
				},
			},
		}
		meta.NodeExporterPort = 1000
		meta.BlackboxExporterPort = 1001
		got, err := meta.GenerateTopologyConfig(ctx.TODO())
		assert.NoError(t, err)
		config := spec.Specification{}
		err = yaml.Unmarshal([]byte(got), &config)
		assert.NoError(t, err)
		assert.Equal(t, len(config.TiDBServers), 1)
		assert.Equal(t, len(config.TiKVServers), 1)
		assert.Equal(t, len(config.PDServers), 1)
		assert.Equal(t, len(config.TiFlashServers), 1)
		assert.Equal(t, len(config.CDCServers), 1)
		assert.Equal(t, config.GlobalOptions.Arch, "arm64")

		meta.Cluster.Status = string(constants.ClusterRunning)
		meta.Instances["TiDB"] = append(meta.Instances["TiDB"], &management.ClusterInstance{
			Entity:   common.Entity{ID: "testInstance011", Status: string(constants.ClusterInstanceRunning)},
			HostIP:   []string{"127.0.0.2"},
			DiskPath: "/test",
			Ports:    []int32{4003, 4004},
		})
		got, err = meta.GenerateTopologyConfig(ctx.TODO())
		fmt.Println(got)
		assert.NoError(t, err)
		config = spec.Specification{}
		err = yaml.Unmarshal([]byte(got), &config)
		assert.NoError(t, err)
		assert.Equal(t, len(config.TiDBServers), 1)
		assert.Equal(t, len(config.TiKVServers), 1)
		assert.Equal(t, len(config.PDServers), 1)
		assert.Equal(t, len(config.TiFlashServers), 1)
		assert.Equal(t, len(config.CDCServers), 1)
		assert.Equal(t, config.GlobalOptions.Arch, "")
	}
}
