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

        location ^~ /api {
            proxy_pass http://openapi-servers;
        }

        location ~ ^/(swagger|system|web)/ {
            proxy_pass http://openapi-servers;
        }

        location ^~/etcd/ {
            proxy_pass http://etcdcluster/;
        }

        location = /upstream_show {
            upstream_show;
        }
    }
}
