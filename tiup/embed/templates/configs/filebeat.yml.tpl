setup:
  ilm:
    enabled: false
  template:
    enabled: true
    name: "${indexPrefix:tiem-cluster}"
    pattern: "${indexPrefix:tiem-cluster}-*"
    overwrite: true

logging.level: info
logging.to_files: true
logging.files:
  path: {{.LogDir}}
  name: filebeat
  keepfiles: 7
  permissions: 0644

filebeat:
  inputs:
    - type: log
      enabled: true
      paths:
{{- range .GeneralLogs}}
        - '{{.}}'
{{- end}}
      fields:
        type: logs
      json:
        keys_under_root: true  
        overwrite_keys: true
        message_key: msg
      #include_lines:  [ 'log' ]
      exclude_lines: [  ]
      fields_under_root: true
    - type: log
      enabled: true
      paths:
{{- range .AuditLogs}}
        - '{{.}}'
{{- end}}
      fields:
        type: audits
      json:
        keys_under_root: true  
        overwrite_keys: true
        message_key: msg
      #include_lines:  [ 'audits' ]
      exclude_lines: [  ]
      fields_under_root: true

filebeat.config.inputs:
  enabled: true
  path: conf/*.yml
  reload.enabled: true
  reload.period: 5s
  
output.elasticsearch:
  ### 3. 读取拓扑文件配置的es的主机列表，如果是多个es实例，需要用逗号进行拼装，例如：192.168.1.101:9200,192.168.1.102:9200,192.168.1.103:9200
  hosts: "${esAddress:{{.ElasticSearchHost}}}"
  indices:
    - index: "${indexPrefix:tiem}-logs-%{+yyyy.MM.dd}"
      when.equals:
        type: "logs"
    - index: "${indexPrefix:tiem}-audits-%{+yyyy.MM.dd}"
      when.equals:
        type: "audits"
    - index: "${indexPrefix:tiem}-tidb-cluster-%{+yyyy.MM.dd}"
      when.equals:
        type: "tidb"