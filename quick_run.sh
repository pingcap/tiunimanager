nohup ./bin/metadb-server &
sleep 2
nohup ./bin/cluster-server --metrics-port=4122 &
nohup ./bin/openapi-server --metrics-port=4123 &
