## 1. Make `cert` directory

make sure that you are in the project's root path

```shell
mkdir -p bin/cert
cd bin/cert/
```

## 2. Generate server.crt, server.csr, server.key

make sure that you are in the directory `<project root path>/bin/cert`

```shell
openssl genrsa -out rootCA.key 2048
openssl req -x509 -new -nodes -key rootCA.key -days 1024 -out rootCA.pem -subj '/CN=em-server/O=PingCAP/C=CN'
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj '/CN=em-server/O=PingCAP/C=CN'
openssl x509 -req -in server.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out server.crt -days 3650
```

## 3. Generate aes.key

make sure that you are in the directory `<project root path>/bin/cert`

```shell
echo $RANDOM | md5sum | head -c 32 > aes.key
```

## 4. Generate etcd-ca.pem, etcd-server-key.pem, etcd-server.pem

> You may need to generate executable `cfssl` and `cfssljson` by following [cfssl](https://github.com/cloudflare/cfssl/) before execute the commands below.

make sure that you are in the directory `<project root path>/bin/cert`

```shell
cfsslBinPath=<please fill it>
${cfsslBinPath}/cfssl gencert -initca ../../build_helper/ca-csr.json | ${cfsslBinPath}/cfssljson  -bare etcd-ca -
${cfsslBinPath}/cfssl gencert -ca=etcd-ca.pem -ca-key=etcd-ca-key.pem --config=../../build_helper/ca-config.json -profile=server ../../build_helper/server.json | ${cfsslBinPath}/cfssljson -bare etcd-server
```