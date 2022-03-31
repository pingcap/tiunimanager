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

/*******************************************************************************
 * @File: executor.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/29 11:10
*******************************************************************************/

package log

import (
	ctx "context"
	"encoding/json"

	"fmt"

	"github.com/pingcap-inc/tiem/deployment"

	"github.com/pingcap-inc/tiem/common/structs"

	"github.com/pingcap-inc/tiem/common/constants"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	workflow "github.com/pingcap-inc/tiem/workflow2"
	"gopkg.in/yaml.v2"
)

func collectorClusterLogConfig(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin collector cluster log config executor method")
	defer framework.LogWithContext(ctx).Info("end collector cluster log config executor method")

	var clusterMeta meta.ClusterMeta
	err := ctx.GetData(contextClusterMeta, &clusterMeta)
	if err != nil {
		return err
	}

	// get current cluster hosts
	hosts := listClusterHosts(&clusterMeta)
	framework.LogWithContext(ctx).Infof("cluster [%s] list host: %v", clusterMeta.Cluster.ID, hosts)

	node.Record("get instance log info")
	for hostID, hostIP := range hosts {
		instances, err := meta.QueryInstanceLogInfo(ctx, hostID, []string{}, []string{})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("cluster [%s] query metas failed. err: %v", clusterMeta.Cluster.ID, err)
			return err
		}

		for _, i := range instances {
			node.Record(fmt.Sprintf("HostIP: %s, log dir : %s", i.IP, i.LogDir))
		}
		collectorConfigs, err := buildCollectorClusterLogConfig(ctx, instances)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("build collector cluster log config err： %v", err)
			return err
		}
		bs, err := yaml.Marshal(collectorConfigs)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("marshal yaml err： %v", err)
			return err
		}
		collectorYaml := string(bs)

		// Get the deploy info of push
		clusterComponentType, clusterName, home, deployDir, err := getDeployInfo(&clusterMeta, ctx, hostIP)
		if err != nil {
			return err
		}
		transferTaskId, err := deployment.M.Push(ctx, clusterComponentType, clusterName, collectorYaml,
			deployDir+"/conf/input_tidb.yml", home, node.ParentID, []string{"-N", hostIP}, 0)
		framework.LogWithContext(ctx).Infof("got transferTaskId: %s", transferTaskId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("invoke tiup transfer err： %v", err)
			return err
		}
		node.OperationID = transferTaskId
	}

	return nil
}

// getDeployInfo
// @Description: Get the deploy info of push
// @Parameter clusterMeta
// @Parameter ctx
// @Parameter hostIP
// @return secondparty.TiUPComponentTypeStr
// @return string
// @return string
// @return error
func getDeployInfo(clusterMeta *meta.ClusterMeta, ctx *workflow.FlowContext, hostIP string) (deployment.TiUPComponentType, string, string, string, error) {
	deployDir := "/tiem-test/filebeat"
	clusterComponentType := deployment.TiUPComponentTypeCluster
	home := framework.GetTiupHomePathForTidb()
	clusterName := clusterMeta.Cluster.ID
	if framework.Current.GetClientArgs().EMClusterName != "" {
		deployDir = "/em-deploy/filebeat-0"
		clusterComponentType = deployment.TiUPComponentTypeEM
		home = framework.GetTiupHomePathForTiem()
		clusterName = framework.Current.GetClientArgs().EMClusterName

		// Parse EM topology structure to get filebeat deploy dir
		result, err := deployment.M.Display(ctx, clusterComponentType, clusterName, framework.GetTiupHomePathForTiem(), []string{"--json"}, 0)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("invoke tiup cluster display err： %v", err)
			return "", "", "", "", err
		}
		emTopo := new(structs.EMMetaTopo)
		err = json.Unmarshal([]byte(result), &emTopo)
		if err != nil {
			return "", "", "", "", err
		}
		for _, instance := range emTopo.Instances {
			if instance.Role == "filebeat" && instance.Host == hostIP {
				deployDir = instance.DeployDir
				break
			}
		}
	}
	return clusterComponentType, clusterName, home, deployDir, nil
}

// buildCollectorClusterLogConfig
// @Description: build collector cluster log config
// @Parameter ctx
// @Parameter host
// @Parameter clusters
// @return []CollectorClusterLogConfig
// @return error
func buildCollectorClusterLogConfig(ctx ctx.Context, clusterInfos []*meta.InstanceLogInfo) ([]CollectorClusterLogConfig, error) {
	framework.LogWithContext(ctx).Info("begin collector cluster log config executor method")
	defer framework.LogWithContext(ctx).Info("end collector cluster log config executor method")

	// Construct the structure of multiple clusters corresponding instances
	clusterInstances := make(map[string][]*meta.InstanceLogInfo, 0)
	for _, instance := range clusterInfos {
		insts := clusterInstances[instance.ClusterID]
		if insts == nil {
			clusterInstances[instance.ClusterID] = []*meta.InstanceLogInfo{instance}
		} else {
			insts = append(insts, instance)
			clusterInstances[instance.ClusterID] = insts
		}
	}

	// build collector cluster log configs
	configs := make([]CollectorClusterLogConfig, 0)
	for clusterID, instances := range clusterInstances {
		cfg := CollectorClusterLogConfig{Module: "tidb"}

		for _, server := range instances {
			// TiDB modules
			if server.InstanceType == constants.ComponentIDTiDB {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/tidb.log"
				// multiple instances of the same cluster
				if len(cfg.TiDB.Var.Paths) > 0 {
					cfg.TiDB.Var.Paths = append(cfg.TiDB.Var.Paths, logPath)
					continue
				}
				cfg.TiDB = buildCollectorModuleDetail(clusterID, server.IP, logPath)
			}

			// PD modules
			if server.InstanceType == constants.ComponentIDPD {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/pd.log"
				// multiple instances of the same cluster
				if len(cfg.PD.Var.Paths) > 0 {
					cfg.PD.Var.Paths = append(cfg.PD.Var.Paths, logPath)
					continue
				}
				cfg.PD = buildCollectorModuleDetail(clusterID, server.IP, logPath)
			}

			// TiKV modules
			if server.InstanceType == constants.ComponentIDTiKV {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/tikv.log"
				// multiple instances of the same cluster
				if len(cfg.TiKV.Var.Paths) > 0 {
					cfg.TiKV.Var.Paths = append(cfg.TiKV.Var.Paths, logPath)
					continue
				}
				cfg.TiKV = buildCollectorModuleDetail(clusterID, server.IP, logPath)
			}

			// TiFlash modules
			if server.InstanceType == constants.ComponentIDTiFlash {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/tiflash.log"
				// multiple instances of the same cluster
				if len(cfg.TiFlash.Var.Paths) > 0 {
					cfg.TiFlash.Var.Paths = append(cfg.TiFlash.Var.Paths, logPath)
					continue
				}
				cfg.TiFlash = buildCollectorModuleDetail(clusterID, server.IP, logPath)
			}

			// CDC modules
			if server.InstanceType == constants.ComponentIDCDC {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/cdc.log"
				// multiple instances of the same cluster
				if len(cfg.CDC.Var.Paths) > 0 {
					cfg.CDC.Var.Paths = append(cfg.CDC.Var.Paths, logPath)
					continue
				}
				cfg.CDC = buildCollectorModuleDetail(clusterID, server.IP, logPath)
			}
		}

		configs = append(configs, cfg)
	}
	return configs, nil
}

// buildCollectorModuleDetail
// @Description: build collector module detail struct
// @Parameter clusterId
// @Parameter host
// @Parameter logDir
// @return CollectorModuleDetail
func buildCollectorModuleDetail(clusterID string, host, logDir string) CollectorModuleDetail {
	return CollectorModuleDetail{
		Enabled: true,
		Var: CollectorModuleVar{
			Paths: []string{logDir},
		},
		Input: CollectorModuleInput{
			Fields: CollectorModuleFields{
				Type:      "tidb",
				ClusterId: clusterID,
				Ip:        host,
			},
			FieldsUnderRoot: true,
			IncludeLines:    nil,
			ExcludeLines:    nil,
		},
	}
}

// listClusterHosts
// @Description: List the hosts after cluster de-duplication
// @Parameter clusterMeta
// @return map[string]string
func listClusterHosts(clusterMeta *meta.ClusterMeta) map[string]string {
	hosts := make(map[string]string, 0)
	for _, instances := range clusterMeta.Instances {
		for _, instance := range instances {
			hosts[instance.HostID] = instance.HostIP[0]
		}
	}
	return hosts
}

// defaultEnd
// @Description: default end
func defaultEnd(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin default end executor method")
	defer framework.LogWithContext(ctx).Info("end default end executor method")

	return nil
}
