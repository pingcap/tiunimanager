# Preface

This doc will introduce the procedure to deploy a TiDB cluster using TiUniManager by calling OpenAPI. You can also do most of the following steps through [TiUniManager UI](https://github.com/pingcap/tiunimanager-ui) except for [configuring SSH login without password](#configure-ssh-login-without-password).

## Change `admin`'s Password

For security issue, the admin's default password is expired, which means you need to change it before calling other APIs.

1. Login using default password to get the token

    ```shell
    curl -X 'POST' \
        'http://127.0.0.1:4100/api/v1/user/login' \
        -H 'accept: application/json' \
        -H 'Content-Type: application/json' \
        -d '{
        "userName": "admin",
        "userPassword": "admin"
    }'
    ```

2. Get `admin`'s user ID

    make sure that you are in the project's root path

    1. Connect to meta DB of TiUniManager

        ```shell
        sqlite3 em.db
        ```

    2. Get the ID

        ```sqlite
        sqlite> select ID from users where name='admin';
        ```

3. Update password

    ```shell
    curl -X 'POST' \
        'http://127.0.0.1:4100/api/v1/users/<ADMIN_ID>/password' \
        -H 'accept: application/json' \
        -H 'Authorization: Bearer <TOKEN>' \
        -H 'Content-Type: application/json' \
        -d '{
        "password": "<NEW PASSWORD>"
    }'
    ```

4. Login using new password to get the token

    Now you may need to use this token to call the APIs for the following steps.

    ```shell
    curl -X 'POST' \
        'http://127.0.0.1:4100/api/v1/user/login' \
        -H 'accept: application/json' \
        -H 'Content-Type: application/json' \
        -d '{
        "userName": "admin",
        "userPassword": "<NEW PASSWORD>"
    }'
    ```

## Configure SSH login without password

> **NOTICE：**
> 
> The main purpose of this step is decribed as below:
> 
> - SSH login from TiUniManager host to all the resource hosts without password
> - Store the identity file as `/home/tidb/.ssh/tiup_rsa`

1. All resource hosts

    Add user `tidb` and set password

    ```shell
    useradd tidb
    passwd tidb
    ```

2. TiUniManager host

    1. Generate key pair

        ```shell
        ssh-keygen -t rsa
        ```

    2. Copy the public key to all resouce hosts

        ```shell
        ssh-copy-id -i ~/.ssh/id_rsa.pub tidb@<RESOURCE IP>
        ```

    3. Store identity file

        ```shell
        mkdir -p /home/tidb/.ssh
        cp ~/.ssh/id_rsa /home/tidb/.ssh/tiup_rsa
        ```

    4. Verify login without password

        ```shell
        ssh tidb@<RESOURCE IP> -i /home/tidb/.ssh/tiup_rsa
        ```       

## Import hosts

Before calling the OpenAPI, please generate the host file. You can find the template file in `<project's root path>/resource/hostInfo_template.xlsx`.

```shell
curl -X 'POST' \
  'http://127.0.0.1:4100/api/v1/resources/hosts' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: multipart/form-data' \
  -F 'hostReserved=false' \
  -F 'file=@<PATH TO HOST FILE>;type=application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
```

Since it is an async task, you may need to check the status before continuing.

```shell
curl -X GET "http://127.0.0.1:4100/api/v1/workflow/<WORKFLOW ID>" -H "Authorization: Bearer <TOKEN>"
```

## Create TiDB Cluster

```shell
curl -X 'POST' \
  'http://127.0.0.1:4100/api/v1/clusters/' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{
  "clusterName": "tidb-test",
  "clusterType": "TiDB",
  "clusterVersion": "v5.2.2",
  "copies": 1,
  "cpuArchitecture": "X86_64",
  "dbPassword": "<YOU PASSWORD>",
  "exclusive": false,
  "parameterGroupID": "",
  "vendor": "Local",
  "region": "Region1",
  "resourceParameters": {
    "instanceResource": [
      {
        "componentType": "TiDB",
        "resource": [
          {
            "count": 1,
            "specCode": "4C8G",
            "zoneCode": "Region1,Zone1_1"
          }
        ],
        "totalNodeCount": 1
      },
      {
        "componentType": "TiKV",
        "resource": [
          {
            "count": 1,
            "specCode": "4C8G",
            "zoneCode": "Region1,Zone1_1"
          }
        ],
        "totalNodeCount": 1
      },
      {
        "componentType": "PD",
        "resource": [
          {
            "count": 1,
            "specCode": "4C8G",
            "zoneCode": "Region1,Zone1_1"
          }
        ],
        "totalNodeCount": 1
      }
    ]
  },
  "tags": [
    "tag1"
  ]
}'
```

Since it is an async task, you may need to check the status before continuing.

```shell
curl -X GET "http://127.0.0.1:4100/api/v1/workflow/<WORKFLOW ID>" -H "Authorization: Bearer <TOKEN>"
```
