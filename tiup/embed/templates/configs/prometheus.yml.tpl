global:
  scrape_interval:     15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: node_exporter
    scrape_interval: 10s
    static_configs
      - targets:
          - {{.host}}:{{.node_exporter_port}} ### 1. 遍历每台部署了node_exporter的实例[host+node_exporter_port]组成一个target，遍历显示
          -
          -

  - job_name: tiem
    static_configs
      - targets: 
          - {{.host}}:{{.metrics_port}} ### 2. 将tiem的三个组件metadb/cluster/api的每一个实例的[host+metrics_port]组成一个target，遍历显示
          -
          -

  - job_name: 'tidb_cluster'
    scrape_interval: 15s
    params:
     'match[]':
       - '{job=~".*"}'
    http_sd_configs:
        ### 3. 可否直接通过tiup获取tiem_api_servers的地址？比如：{{.tiem_api_servers.host}}:{{.tiem_api_servers.port}}
      - url: "http://{{.tiem_api_servers.host}}:{{.tiem_api_servers.port}}/api/sd/metrics?address={{.monitoring_servers.host}}:{{.monitoring_servers.port}}"
        refresh_interval: 15s
