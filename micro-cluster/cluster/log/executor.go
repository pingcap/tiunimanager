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

	"github.com/pingcap-inc/tiem/common/constants"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/workflow"
	"gopkg.in/yaml.v2"
)

func collectorClusterLogConfig(node *workflowModel.WorkFlowNode, ctx *workflow.FlowContext) error {
	framework.LogWithContext(ctx).Info("begin collector cluster log config executor method")
	defer framework.LogWithContext(ctx).Info("end collector cluster log config executor method")

	clusterMeta := ctx.GetData(contextClusterMeta).(*handler.ClusterMeta)

	// get current cluster hosts
	hosts := listClusterHosts(clusterMeta)
	framework.LogWithContext(ctx).Infof("cluster [%s] list host: %v", clusterMeta.Cluster.ID, hosts)

	for hostID, hostIP := range hosts {
		instances, err := handler.QueryInstanceLogInfo(ctx, hostID, []string{}, []string{})
		if err != nil {
			framework.LogWithContext(ctx).Errorf("cluster [%s] query metas failed. err: %v", clusterMeta.Cluster.ID, err)
			return err
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
		// todo: When the tiem scale-out and scale-in is complete, change to take the filebeat deployDir from the tiem topology
		deployDir := "/tiem-test/filebeat"
		clusterComponentType := secondparty.ClusterComponentTypeStr
		clusterName := clusterMeta.Cluster.ID
		if framework.Current.GetClientArgs().EMClusterName != "" {
			deployDir = "/tiem-deploy/filebeat-0"
			clusterComponentType = secondparty.TiEMComponentTypeStr
			clusterName = framework.Current.GetClientArgs().EMClusterName
		}
		transferTaskId, err := secondparty.Manager.Transfer(ctx, clusterComponentType,
			clusterName, collectorYaml, deployDir+"/conf/input_tidb.yml",
			0, []string{"-N", hostIP}, node.ID)
		framework.LogWithContext(ctx).Infof("got transferTaskId: %s", transferTaskId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("collectorClusterLogConfig invoke tiup transfer err： %v", err)
			return err
		}
	}

	return nil
}

// buildCollectorClusterLogConfig
// @Description: build collector cluster log config
// @Parameter ctx
// @Parameter host
// @Parameter clusters
// @return []CollectorClusterLogConfig
// @return error
func buildCollectorClusterLogConfig(ctx ctx.Context, clusterInfos []*handler.InstanceLogInfo) ([]CollectorClusterLogConfig, error) {
	framework.LogWithContext(ctx).Info("begin collector cluster log config executor method")
	defer framework.LogWithContext(ctx).Info("end collector cluster log config executor method")

	// Construct the structure of multiple clusters corresponding instances
	clusterInstances := make(map[string][]*handler.InstanceLogInfo, 0)
	for _, instance := range clusterInfos {
		insts := clusterInstances[instance.ClusterID]
		if insts == nil {
			clusterInstances[instance.ClusterID] = []*handler.InstanceLogInfo{instance}
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

			// TiCDC modules
			if server.InstanceType == constants.ComponentIDCDC {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/cdc.log"
				// multiple instances of the same cluster
				if len(cfg.TiCDC.Var.Paths) > 0 {
					cfg.TiCDC.Var.Paths = append(cfg.TiCDC.Var.Paths, logPath)
					continue
				}
				cfg.TiCDC = buildCollectorModuleDetail(clusterID, server.IP, logPath)
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
func listClusterHosts(clusterMeta *handler.ClusterMeta) map[string]string {
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
