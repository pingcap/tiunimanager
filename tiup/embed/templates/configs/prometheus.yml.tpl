global:
  scrape_interval:     15s
  evaluation_interval: 15s
  external_labels:
    cluster: '{{.ClusterName}}'
    monitor: "prometheus"

{{- if .AlertmanagerAddrs}}
alerting:
  alertmanagers:
  - static_configs:
    - targets:
{{- range .AlertmanagerAddrs}}
      - '{{.}}'
{{- end}}
{{- end}}

scrape_configs:
  - job_name: node_exporter
    honor_labels: true # don't overwrite job & instance labels
    scrape_interval: 10s
    static_configs:
      - targets:
{{- range .NodeExporterAddrs}}
        - '{{.}}'
{{- end}}

{{- range $addr := .BlackboxExporterAddrs}}
  - job_name: "blackbox_exporter_{{$addr}}_icmp"
    scrape_interval: 6s
    metrics_path: /probe
    params:
      module: [icmp]
    static_configs:
    - targets:
    {{- range $.MonitoredServers}}
      - '{{.}}'
    {{- end}}
    relabel_configs:
      - source_labels: [__address__]
        regex: (.*)(:80)?
        target_label: __param_target
        replacement: ${1}
      - source_labels: [__param_target]
        regex: (.*)
        target_label: ping
        replacement: ${1}
      - source_labels: []
        regex: .*
        target_label: __address__
        replacement: {{$addr}}
{{- end}}

  - job_name: em
    static_configs:
      - targets: 
{{- range .TiEMAPIServers}}
        - '{{.}}'
{{- end}}
{{- range .TiEMMetaDBServers}}
        - '{{.}}'
{{- end}}
{{- range .TiEMClusterServers}}
        - '{{.}}'
{{- end}}
{{- range .TiEMFileServers}}
        - '{{.}}'
{{- end}}

#  - job_name: 'tidb_cluster'
#    scrape_interval: 15s
#    params:
#     'match[]':
#       - '{job=~".*"}'
#    http_sd_configs:
#      - url: "http://".tiem_api_servers.host:".tiem_api_servers.port"/api/sd/metrics?address=".monitoring_servers.host":".monitoring_servers.port""
#        refresh_interval: 15s
