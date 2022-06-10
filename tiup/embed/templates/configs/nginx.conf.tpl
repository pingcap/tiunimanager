worker_processes  1;

events {
    worker_connections  1024;
}

http {
    include       mime.types;
    default_type  application/octet-stream;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  {{.LogDir}}/access.log  main;

    upstream openapi-servers {
        least_conn;

        upsync {{.IP}}:{{.Port}}/etcd/v2/keys/micro/registry/openapi-server upsync_timeout=6m upsync_interval=500ms upsync_type=etcd strong_dependency=off;
        upsync_dump_path {{.DeployDir}}/conf/server_list.conf;
        include {{.DeployDir}}/conf/server_list.conf;
    }

    upstream file-servers {
        least_conn;

        upsync {{.IP}}:{{.Port}}/etcd/v2/keys/micro/registry/file-server upsync_timeout=6m upsync_interval=500ms upsync_type=etcd strong_dependency=off;
        upsync_dump_path {{.DeployDir}}/conf/file_server_list.conf;
        include {{.DeployDir}}/conf/file_server_list.conf;
    }

    upstream etcdcluster {
{{- range $idx, $addr := .RegistryEndpoints}}
        server {{$addr}} weight=1 fail_timeout=10 max_fails=3;
{{- end}}
    }

    gzip  on;

    server {
        listen {{.Port}};
        server_name  {{.ServerName}};

        location / {
            root   html;
            index  index.html index.htm;
            try_files $uri $uri/ /index.html;
        }

{{- if not .EnableHttps }}
        location ^~ /api {
            proxy_pass http://openapi-servers;
        }

        location ~ ^/(swagger|web)/ {
            proxy_pass http://openapi-servers;
        }

        location ^~ /fs {
            proxy_pass http://file-servers;
        }
{{- end}}

        location ^~/etcd/ {
            proxy_pass https://etcdcluster/;
            proxy_ssl_trusted_certificate {{.DeployDir}}/cert/etcd-ca.pem;
            proxy_ssl_certificate         {{.DeployDir}}/cert/etcd-server.pem;
            proxy_ssl_certificate_key     {{.DeployDir}}/cert/etcd-server-key.pem;
            proxy_ssl_verify              off;
            allow 127.0.0.1;
            allow {{.IP}};
            deny all;
        }

        location /grafana/ {
            proxy_set_header X-WEBAUTH-USER admin;
            proxy_pass http://{{.GrafanaAddress}}/;
        }

        location ~ /grafanas {
            rewrite ^/grafanas-([^/]+)-(\d+)/(.*) /$3 break;
            proxy_set_header X-WEBAUTH-USER admin;
            proxy_pass http://$1:$2/$3$is_args$args;
        }

        location ~ ^/env {
            default_type application/json;
        {{- if ne .KibanaAddress "" }}
            return 200 '{"protocol": "{{.Protocol}}", "tlsPort": {{.TlsPort}}, "service": {"grafana": "http://{{.IP}}:{{.Port}}/grafana/d/em000001/em-server?orgId=1&refresh=10s&kiosk=tv", "kibana": "http://{{.KibanaAddress}}/app/discover", "alert": "http://{{.AlertManagerAddress}}", "tracer": "http://{{.TracerAddress}}"}}';
        {{- else}}
            return 200 '{"protocol": "{{.Protocol}}", "tlsPort": {{.TlsPort}}, "service": {"grafana": "http://{{.IP}}:{{.Port}}/grafana/d/em000001/em-server?orgId=1&refresh=10s&kiosk=tv", "kibana": "", "alert": "http://{{.AlertManagerAddress}}", "tracer": "http://{{.TracerAddress}}"}}';
        {{- end}}
        }

        location = /upstream_show {
            upstream_show;
        }
    }

{{- if .EnableHttps }}

    server {
        listen {{.TlsPort}} ssl;
        server_name  {{.ServerName}};

        ssl_certificate  {{.DeployDir}}/cert/server.crt;
        ssl_certificate_key {{.DeployDir}}/cert/server.key;
        server_tokens off;

        fastcgi_param   HTTPS               on;
        fastcgi_param   HTTP_SCHEME         https;

        location ^~ /api {
            proxy_pass https://openapi-servers;
        }

        location ~ ^/(swagger|system|web)/ {
            proxy_pass https://openapi-servers;
        }

        location ^~ /fs {
            proxy_pass http://file-servers;
        }
    }
{{- end}}
}
