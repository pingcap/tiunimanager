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

/*******************************************************************************
 * @File: collector.go
 * @Description: collector tidb log config
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/8 15:21
*******************************************************************************/

package domain

import ctx "context"

// CollectorTiDBLogConfig
// @Description: collector tidb log config struct
type CollectorTiDBLogConfig struct {
	Module  string                `json:"module" yaml:"module"`
	TiDB    CollectorModuleDetail `json:"tidb" yaml:"tidb"`
	PD      CollectorModuleDetail `json:"pd" yaml:"pd"`
	TiKV    CollectorModuleDetail `json:"tikv" yaml:"tikv"`
	TiFlash CollectorModuleDetail `json:"tiflash" yaml:"tiflash"`
	TiCDC   CollectorModuleDetail `json:"ticdc" yaml:"ticdc"`
}

type CollectorModuleDetail struct {
	Enabled bool                 `json:"enabled" yaml:"enabled"`
	Var     CollectorModuleVar   `json:"var" yaml:"var"`
	Input   CollectorModuleInput `json:"input" yaml:"input"`
}

type CollectorModuleVar struct {
	Paths []string `json:"paths" yaml:"paths"`
}

type CollectorModuleInput struct {
	Fields          CollectorModuleFields `json:"fields" yaml:"fields"`
	FieldsUnderRoot bool                  `json:"fields_under_root" yaml:"fields_under_root"`
	IncludeLines    []string              `json:"include_lines" yaml:"include_lines"`
	ExcludeLines    []string              `json:"exclude_lines" yaml:"exclude_lines"`
}

type CollectorModuleFields struct {
	Type      string `json:"type" yaml:"type"`
	ClusterId string `json:"clusterId" yaml:"clusterId"`
	Ip        string `json:"ip" yaml:"ip"`
}

// buildCollectorTiDBLogConfig
// @Description: build collector TiDB log config
// @Parameter ctx
// @Parameter host
// @Parameter clusters
// @return []CollectorTiDBLogConfig
// @return error
func buildCollectorTiDBLogConfig(ctx ctx.Context, host string, clusters []*ClusterAggregation) ([]CollectorTiDBLogConfig, error) {
	getLoggerWithContext(ctx).Info("begin buildCollectorTiDBLogConfig")
	defer getLoggerWithContext(ctx).Info("end buildCollectorTiDBLogConfig")

	// build collector tidb log configs
	configs := make([]CollectorTiDBLogConfig, 0)
	for _, aggregation := range clusters {
		cfg := CollectorTiDBLogConfig{Module: "tidb"}
		spec := aggregation.CurrentTopologyConfigRecord.ConfigModel

		// TiDB modules
		for _, server := range spec.TiDBServers {
			if server.Host == host {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/tidb.log"
				// multiple instances of the same cluster
				if len(cfg.TiDB.Var.Paths) > 0 {
					cfg.TiDB.Var.Paths = append(cfg.TiDB.Var.Paths, logPath)
					continue
				}
				cfg.TiDB = buildCollectorModuleDetail(aggregation, server.Host, logPath)
			}
		}

		// PD modules
		for _, server := range spec.PDServers {
			if server.Host == host {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/pd.log"
				// multiple instances of the same cluster
				if len(cfg.PD.Var.Paths) > 0 {
					cfg.PD.Var.Paths = append(cfg.PD.Var.Paths, logPath)
					continue
				}
				cfg.PD = buildCollectorModuleDetail(aggregation, server.Host, logPath)
			}
		}

		// TiKV modules
		for _, server := range spec.TiKVServers {
			if server.Host == host {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/tikv.log"
				// multiple instances of the same cluster
				if len(cfg.TiKV.Var.Paths) > 0 {
					cfg.TiKV.Var.Paths = append(cfg.TiKV.Var.Paths, logPath)
					continue
				}
				cfg.TiKV = buildCollectorModuleDetail(aggregation, server.Host, logPath)
			}
		}

		// TiFlash modules
		for _, server := range spec.TiFlashServers {
			if server.Host == host {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/tiflash.log"
				// multiple instances of the same cluster
				if len(cfg.TiFlash.Var.Paths) > 0 {
					cfg.TiFlash.Var.Paths = append(cfg.TiFlash.Var.Paths, logPath)
					continue
				}
				cfg.TiFlash = buildCollectorModuleDetail(aggregation, server.Host, logPath)
			}
		}

		// TiCDC modules
		for _, server := range spec.CDCServers {
			if server.Host == host {
				if len(server.LogDir) == 0 {
					continue
				}
				logPath := server.LogDir + "/ticdc.log"
				// multiple instances of the same cluster
				if len(cfg.TiCDC.Var.Paths) > 0 {
					cfg.TiCDC.Var.Paths = append(cfg.TiCDC.Var.Paths, logPath)
					continue
				}
				cfg.TiCDC = buildCollectorModuleDetail(aggregation, server.Host, logPath)
			}
		}

		configs = append(configs, cfg)
	}
	return configs, nil
}

// buildCollectorModuleDetail
// @Description: build collector module detail struct
// @Parameter aggregation
// @Parameter host
// @Parameter logDir
// @return CollectorModuleDetail
func buildCollectorModuleDetail(aggregation *ClusterAggregation, host, logDir string) CollectorModuleDetail {
	return CollectorModuleDetail{
		Enabled: true,
		Var: CollectorModuleVar{
			Paths: []string{logDir},
		},
		Input: CollectorModuleInput{
			Fields: CollectorModuleFields{
				Type:      "tidb",
				ClusterId: aggregation.Cluster.Id,
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
// @Parameter aggregation
// @return []string
func listClusterHosts(aggregation *ClusterAggregation) []string {
	spec := aggregation.CurrentTopologyConfigRecord.ConfigModel
	kv := make(map[string]byte, 0)
	for _, server := range spec.TiDBServers {
		kv[server.Host] = 1
	}
	for _, server := range spec.PDServers {
		kv[server.Host] = 1
	}
	for _, server := range spec.TiKVServers {
		kv[server.Host] = 1
	}
	for _, server := range spec.TiFlashServers {
		kv[server.Host] = 1
	}
	for _, server := range spec.CDCServers {
		kv[server.Host] = 1
	}
	hosts := make([]string, 0, len(kv))
	for host := range kv {
		hosts = append(hosts, host)
	}
	return hosts
}
