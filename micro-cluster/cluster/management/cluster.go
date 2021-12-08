/*******************************************************************************
 * @File: cluster.go
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/7
*******************************************************************************/

package management

import "github.com/pingcap-inc/tiem/common/structs"

type ClusterContext struct {
	Cluster         *structs.ClusterInfo
	ClusterTopology map[string][]structs.ClusterInstanceInfo
}

