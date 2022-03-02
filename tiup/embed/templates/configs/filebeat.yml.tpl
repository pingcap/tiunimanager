setup:
  ilm:
    enabled: false
  template:
    enabled: true
    name: "${indexPrefix:em}"
    pattern: "${indexPrefix:em}-*"
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
  name: filebeat.log
  keepfiles: 7
  permissions: 0644

filebeat.config.modules:
  enable: true
  path: {{.DeployDir}}/conf/input_*.yml
  reload.enabled: true
  reload.period: 5s

{{- if .HasAllLogs }}
filebeat:
  inputs:
{{- if .HasGeneralLogs }}
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
{{- end}}
{{- if .HasAuditLogs }}
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
{{- end}}
{{- end}}
  
output.elasticsearch:
  hosts: "${esAddress:{{.ElasticSearchHost}}}"
  indices:
    - index: "${indexPrefix:em}-system-logs-%{+yyyy.MM.dd}"
      when.equals:
        type: "logs"
    - index: "${indexPrefix:em}-system-audit-%{+yyyy.MM.dd}"
      when.equals:
        type: "audit"
    - index: "${indexPrefix:em}-database-cluster-%{+yyyy.MM.dd}"
      when.equals:
        type: "tidb"
    - index: "${indexPrefix:em}-database-slowlog-%{+yyyy.MM.dd}"
      when.equals:
        type: "tidb_slowlog"