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

package template

var ClusterTopology = `
{{ if eq .Cluster.Status "Initializing" }}
global:
  user: tidb
  group: tidb
  ssh_port: 22
  enable_tls: false
  deploy_dir: {{ .Cluster.ID }}/tidb-deploy
  data_dir: {{ .Cluster.ID }}/tidb-data
  log_dir: {{ .Cluster.ID }}/tidb-log
  os: linux
  {{ if eq .Cluster.CpuArchitecture "X86_64" }}
  arch: amd64
  {{ else if eq .Cluster.CpuArchitecture "ARM64" }}
  arch: arm64
  {{ else }}
  arch: amd64
  {{ end }}
monitored:
  node_exporter_port: {{ .NodeExporterPort }}
  blackbox_exporter_port: {{ .BlackboxExporterPort }}
{{ end }}
{{ range $key, $instances := .Instances }}
{{ if and (eq $key "TiDB") (len $instances) }}
tidb_servers:
  {{ range $instances }}
  {{ if eq .Status "Initializing" }}
  - host: {{ index .HostIP 0 }}
    port: {{ index .Ports 0 }}
    status_port: {{ index .Ports 1 }}
    deploy_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/tidb-deploy
  {{ end }}
  {{ end }}
{{ else if and (eq $key "TiKV") (len $instances) }}
tikv_servers:
  {{ range $instances }}
  {{ if eq .Status "Initializing" }}
  - host: {{ index .HostIP 0 }}
    port: {{ index .Ports 0 }}
    status_port: {{ index .Ports 1 }}
    deploy_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/tikv-deploy
    data_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/tikv-data
  {{ end }}
  {{ end }}
{{ else if and (eq $key "TiFlash") (len $instances) }}
tiflash_servers:
  {{ range $instances }}
  {{ if eq .Status "Initializing" }}
  - host: {{ index .HostIP 0 }}
    tcp_port: {{ index .Ports 0 }}
    http_port: {{ index .Ports 1 }}
    flash_service_port: {{ index .Ports 2 }}
    flash_proxy_port: {{ index .Ports 3 }}
    flash_proxy_status_port: {{ index .Ports 4 }}
    metrics_port: {{ index .Ports 5 }}
    deploy_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/tiflash-deploy
    data_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/tiflash-data
    config:
      server.labels:
		vendorregoin: {{ $.Cluster.Region }}
		zone: {{ .Zone }}
		rack: {{ .Rack }}
		host: {{ .HostID }}
  {{ end }}
  {{ end }}
{{ else if and (eq $key "CDC") (len $instances) }}
cdc_servers:
  {{ range $instances }}
  {{ if eq .Status "Initializing" }}
  - host: {{ index .HostIP 0 }}
    port: {{ index .Ports 0 }}
    deploy_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/cdc-deploy
    data_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/cdc-data
  {{ end }}
  {{ end }}
{{ else if and (eq $key "Grafana") (len $instances) }}
grafana_servers:
  {{ range $instances }}
  {{ if eq .Status "Initializing" }}
  - host: {{ index .HostIP 0 }}
    port: {{ index .Ports 0}}
    deploy_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/grafana-deploy
    anonymous_enable: true
    default_theme: light
    org_name: Main Org.
    org_role: Viewer
    config:
      security.allow_embedding: true
  {{ end }}
  {{ end }}
{{ else if and (eq $key "Prometheus") (len $instances) }}
monitoring_servers:
  {{ range $instances }}
  {{ if eq .Status "Initializing" }}
  - host: {{ index .HostIP 0 }}
    port: {{ index .Ports 0}}
    deploy_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/prometheus-deploy
    data_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/prometheus-data
  {{ end }}
  {{ end }}
{{ else if and (eq $key "AlertManger") (len $instances) }}
alertmanager_servers:
  {{ range $instances }}
  {{ if eq .Status "Initializing" }}
  - host: {{ index .HostIP 0 }}
    web_port: {{ index .Ports 0}}
    cluster_port: {{ index .Ports 1}}
    deploy_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/alertmanager-deploy
    data_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/alertmanager-data
  {{ end }}
  {{ end }}
{{ else if and (eq $key "PD") (len $instances) }}
pd_servers:
  {{ range $instances }}
  {{ if eq .Status "Initializing" }}
  - host: {{ index .HostIP 0 }}
    client_port: {{ index .Ports 0 }}
    peer_port: {{ index .Ports 1 }}
    deploy_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/pd-deploy
    data_dir: {{ .DiskPath }}/{{ $.Cluster.ID }}/pd-data
  {{ end }}
  {{ end }}
{{ end }}
{{ end }}
`
