---
server.name: {{.Name}}
server.host: {{.IP}}
server.port: {{.Port}}
elasticsearch.hosts:
{{- range .ElasticSearchAddresses}}
  - "http://{{.}}"
{{- end}}
monitoring.ui.container.elasticsearch.enabled: true

# elasticsearch.username: elastic
# elasticsearch.password: elastic
