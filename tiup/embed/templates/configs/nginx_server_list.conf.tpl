{{- range $idx, $addr := .APIServers}}
server {{$addr}} weight=1 max_fails=2 fail_timeout=10s;
{{- end}}