[Unit]
Description={{.ServiceName}} service
After=syslog.target network.target remote-fs.target nss-lookup.target

[Service]
{{- if ne .Type ""}}
Type={{.Type}}
{{- end}}
{{- if ne .JavaHome ""}}
Environment="JAVA_HOME={{.JavaHome}}"
{{- end}}
User={{.User}}
ExecStart={{.DeployDir}}/scripts/run_{{.ServiceName}}.sh

{{- if .Restart}}
Restart={{.Restart}}
{{else}}
Restart=always
{{end}}
RestartSec=15s

[Install]
WantedBy=multi-user.target
