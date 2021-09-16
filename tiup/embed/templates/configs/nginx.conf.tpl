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

    access_log  {{.Logdir}}/access.log  main;

    upstream openapi-servers {
        least_conn;

        ### 1. 替换web_servers的host和port
        upsync {{.Host}}:{{.Port}}/etcd/v2/keys/micro/registry/openapi-server upsync_timeout=6m upsync_interval=500ms upsync_type=etcd strong_dependency=off;
        ### 2. 替换web_servers部署目录地址
        upsync_dump_path {{.DeployDir}}/conf/server_list.conf;
        include {{.DeployDir}}/conf/server_list.conf;
    }

    upstream etcdcluster {
        ### 3. 替换etcd服务的地址，一行一个，多个etcd需要插入多行
{{- range $idx, $addr := .RegistryEndpoints}}
        server {{.addr}} weight=1 fail_timeout=10 max_fails=3;
{{- end}}
    }

    gzip  on;

    server {
        ### 4. 监听端口，取web_servers的port
        listen {{.Port}};
        server_name  {{.ServerName}};

        location / {
            root   html;
            index  index.html index.htm;
        }

        location ^~ /api {
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
