nohup ./bin/cluster-server --host=127.0.0.1 --port=4101 --tracer-address=127.0.0.1:6831 --metrics-port=4104 --registry-address=127.0.0.1:4106 --elasticsearch-address=127.0.0.1:9200 --skip-host-init=true  --em-version=InTesting --deploy-user=tidb --deploy-group=tidb >> logs/system.log 2>&1 &
sleep 3
nohup ./bin/openapi-server --host=127.0.0.1 --port=4116 --tracer-address=127.0.0.1:6831 --metrics-port=4103 --registry-address=127.0.0.1:4106 >> system.log &
sleep 3
nohup ./bin/file-server --host=127.0.0.1 --port=4102 --tracer-address=127.0.0.1:6831 --metrics-port=4105 --registry-address=127.0.0.1:4106 >> system.log &
