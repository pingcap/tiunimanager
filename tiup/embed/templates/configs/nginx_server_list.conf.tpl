        # 取tiem_api_servers地址，多个需要遍历插入多行记录
{{- range $idx, $addr := .APIServers}}
        server {{.addr}} weight=1 max_fails=2 fail_timeout=10s;
{{- end}}