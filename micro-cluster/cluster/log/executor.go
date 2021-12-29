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

	"github.com/pingcap-inc/tiem/message/cluster"

	"github.com/pingcap-inc/tiem/models/cluster/management"

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

	_, total, err := handler.Query(ctx, cluster.QueryClustersReq{})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("cluster [%s] query metas failed. err: %v", clusterMeta.Cluster.ID, err)
		return err
	}
	framework.LogWithContext(ctx).Infof("cluster [%s] query metas total: %d", clusterMeta.Cluster.ID, total)

	for _, host := range hosts {
		collectorConfigs, err := buildCollectorClusterLogConfig(ctx, host, nil)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("build collector cluster log config err： %v", err)
			break
		}
		bs, err := yaml.Marshal(collectorConfigs)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("marshal yaml err： %v", err)
			break
		}
		collectorYaml := string(bs)
		// todo: When the tiem scale-out and scale-in is complete, change to take the filebeat deployDir from the tiem topology
		deployDir := "/tiem-test/filebeat"
		transferTaskId, err := secondparty.Manager.Transfer(ctx, secondparty.ClusterComponentTypeStr,
			clusterMeta.Cluster.Name, collectorYaml, deployDir+"/conf/input_tidb.yml",
			0, []string{"-N", host}, node.ID)
		framework.LogWithContext(ctx).Infof("got transferTaskId: %s", transferTaskId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("collectorClusterLogConfig invoke tiup transfer err： %v", err)
			break
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
func buildCollectorClusterLogConfig(ctx ctx.Context, host string, metas []*management.Result) ([]CollectorClusterLogConfig, error) {
	framework.LogWithContext(ctx).Info("begin collector cluster log config executor method")
	defer framework.LogWithContext(ctx).Info("end collector cluster log config executor method")

	// build collector cluster log configs
	configs := make([]CollectorClusterLogConfig, 0)
	//for _, meta := range metas {
	//	if meta == nil || meta.Cluster == nil || meta.Instances == nil {
	//		framework.LogWithContext(ctx).Warnf("build collector cluster log meta is nil")
	//		continue
	//	}
	//	cfg := CollectorClusterLogConfig{Module: "tidb"}
	//
	//	// TiDB modules
	//	for _, server := range meta.Instances[""] {
	//		if server.Host == host {
	//			if len(server.LogDir) == 0 {
	//				continue
	//			}
	//			logPath := server.LogDir + "/tidb.log"
	//			// multiple instances of the same cluster
	//			if len(cfg.TiDB.Var.Paths) > 0 {
	//				cfg.TiDB.Var.Paths = append(cfg.TiDB.Var.Paths, logPath)
	//				continue
	//			}
	//			cfg.TiDB = buildCollectorModuleDetail(aggregation, server.Host, logPath)
	//		}
	//	}
	//
	//	// PD modules
	//	for _, server := range spec.PDServers {
	//		if server.Host == host {
	//			if len(server.LogDir) == 0 {
	//				continue
	//			}
	//			logPath := server.LogDir + "/pd.log"
	//			// multiple instances of the same cluster
	//			if len(cfg.PD.Var.Paths) > 0 {
	//				cfg.PD.Var.Paths = append(cfg.PD.Var.Paths, logPath)
	//				continue
	//			}
	//			cfg.PD = buildCollectorModuleDetail(aggregation, server.Host, logPath)
	//		}
	//	}
	//
	//	// TiKV modules
	//	for _, server := range spec.TiKVServers {
	//		if server.Host == host {
	//			if len(server.LogDir) == 0 {
	//				continue
	//			}
	//			logPath := server.LogDir + "/tikv.log"
	//			// multiple instances of the same cluster
	//			if len(cfg.TiKV.Var.Paths) > 0 {
	//				cfg.TiKV.Var.Paths = append(cfg.TiKV.Var.Paths, logPath)
	//				continue
	//			}
	//			cfg.TiKV = buildCollectorModuleDetail(aggregation, server.Host, logPath)
	//		}
	//	}
	//
	//	// TiFlash modules
	//	for _, server := range spec.TiFlashServers {
	//		if server.Host == host {
	//			if len(server.LogDir) == 0 {
	//				continue
	//			}
	//			logPath := server.LogDir + "/tiflash.log"
	//			// multiple instances of the same cluster
	//			if len(cfg.TiFlash.Var.Paths) > 0 {
	//				cfg.TiFlash.Var.Paths = append(cfg.TiFlash.Var.Paths, logPath)
	//				continue
	//			}
	//			cfg.TiFlash = buildCollectorModuleDetail(aggregation, server.Host, logPath)
	//		}
	//	}
	//
	//	// TiCDC modules
	//	for _, server := range spec.CDCServers {
	//		if server.Host == host {
	//			if len(server.LogDir) == 0 {
	//				continue
	//			}
	//			logPath := server.LogDir + "/ticdc.log"
	//			// multiple instances of the same cluster
	//			if len(cfg.TiCDC.Var.Paths) > 0 {
	//				cfg.TiCDC.Var.Paths = append(cfg.TiCDC.Var.Paths, logPath)
	//				continue
	//			}
	//			cfg.TiCDC = buildCollectorModuleDetail(aggregation, server.Host, logPath)
	//		}
	//	}
	//
	//	configs = append(configs, cfg)
	//}
	return configs, nil
}

// buildCollectorModuleDetail
// @Description: build collector module detail struct
// @Parameter clusterId
// @Parameter host
// @Parameter logDir
// @return CollectorModuleDetail
func buildCollectorModuleDetail(clusterId string, host, logDir string) CollectorModuleDetail {
	return CollectorModuleDetail{
		Enabled: true,
		Var: CollectorModuleVar{
			Paths: []string{logDir},
		},
		Input: CollectorModuleInput{
			Fields: CollectorModuleFields{
				Type:      "tidb",
				ClusterId: clusterId,
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
// @return []string
func listClusterHosts(clusterMeta *handler.ClusterMeta) []string {
	kv := make(map[string]byte, 0)
	for _, instances := range clusterMeta.Instances {
		for _, instance := range instances {
			kv[instance.HostIP[0]] = 1
		}
	}

	hosts := make([]string, 0, len(kv))
	for host := range kv {
		hosts = append(hosts, host)
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
