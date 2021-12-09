/*******************************************************************************
 * @File: handler.go
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/8
*******************************************************************************/

package handler

import (
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
)

type ClusterMeta struct {
	Cluster         *structs.ClusterInfo
	ClusterTopology map[string][]*structs.ClusterInstanceInfo
}

// AddInstances
// @Description add new instances into cluster topology, then alloc host ip, ports and disks for new instances
// @Parameter	computes
// @Return		error
func (meta *ClusterMeta) AddInstances(computes []structs.ClusterResourceParameterCompute) error {
	if len(computes) <= 0 {
		return framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "AddInstances error: parameter is empty!")
	}
	for _, component := range computes {
		for _, instance := range component.Resource {

		}
	}
	return nil
}

// GenerateTopologyConfig
// @Description generate yaml config based on cluster topology
// @Return		yaml config
// @Return		error
func (meta *ClusterMeta) GenerateTopologyConfig() (string, error){
	return "", nil
}

// Persist
// @Description persist new instances and cluster info into db
// @Return		error
func (meta *ClusterMeta) Persist() error {
	return nil
}

