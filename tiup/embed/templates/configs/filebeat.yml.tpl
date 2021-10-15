setup:
  ilm:
    enabled: false
  template:
    enabled: true
    name: "${indexPrefix:tiem}"
    pattern: "${indexPrefix:tiem}-*"
    overwrite: true
setup.template.append_fields:
  - name: msg
    type: text

processors:
  - rate_limit:
      limit: "200/s"

logging.level: info
logging.to_files: true
logging.files:
  path: {{.LogDir}}
  name: filebeat
  keepfiles: 7
  permissions: 0644

filebeat.config.modules:
  enable: true
  path: ${path.config}/conf/*.yml
  reload.enabled: true
  reload.period: 5s

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
      include_lines:  [  ]
      exclude_lines: [  ]
      fields_under_root: true
    - type: log
      enabled: true
      paths:
{{- range .AuditLogs}}
        - '{{.}}'
{{- end}}
      fields:
        type: audit
      json:
        keys_under_root: true  
        overwrite_keys: true
        message_key: msg
      #include_lines:  [  ]
      exclude_lines: [  ]
      fields_under_root: true
  
output.elasticsearch:
  hosts: "${esAddress:{{.ElasticSearchHost}}}"
  indices:
    - index: "${indexPrefix:tiem}-system-logs-%{+yyyy.MM.dd}"
      when.equals:
        type: "logs"
    - index: "${indexPrefix:tiem}-system-audit-%{+yyyy.MM.dd}"
      when.equals:
        type: "audit"
    - index: "${indexPrefix:tiem}-tidb-cluster-%{+yyyy.MM.dd}"
      when.equals:
        type: "tidb"
    - index: "${indexPrefix:tiem}-tidb-slowlog-%{+yyyy.MM.dd}"
      when.equals:
        type: "tidb_slowlog"