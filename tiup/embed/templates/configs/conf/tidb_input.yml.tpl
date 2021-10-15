{{- range $idx, $obj := .TidbDeployData}}
- type: log
  enabled: true
  paths:
    -{{ $obj.logDir }}/*.log
  fields:
    type: tidb
    clusterId: "{{ $obj.clusterId }}"
    clusterName: "{{ $obj.clusterName }}"
    host: "{{ $obj.host }}"
    mod: "{{ $obj.mod }}"   # TiDB/TiKV/PD
  json:
    keys_under_root: true  
    overwrite_keys: true
    message_key: msg
  #include_lines:  [ 'tidb' ]
  exclude_lines: [  ]
  fields_under_root: true
{{- end}}
