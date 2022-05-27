## /user/login

### POST
#### Summary

login

#### Description

login

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| loginInfo | body | login info | Yes | [message.LoginReq](#messageloginreq) |

##### message.LoginReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userName | string |  | Yes |
| userPassword | string |  | Yes |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [message.LoginResp](#messageloginresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### message.LoginResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| passwordExpired | boolean |  | No |
| tenantId | string |  | No |
| token | string |  | No |
| userId | string |  | No |

#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/user/login' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "userName": "admin",
  "userPassword": "admin"
}'
```
response
```JSON
{
  "code": 0,
  "data": {
    "passwordExpired": false,
    "tenantId": "1111",
    "token": "mytoken",
    "userId": "22222"
  },
  "message": "string"
}
```

## /user/logout

### POST
#### Summary

logout

#### Description

logout

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [message.LogoutResp](#messagelogoutresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### message.LogoutResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userId | string |  | No |

#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/user/logout' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -d ''
```
response
```JSON
{
  "code": 0,
  "data": {
    "userId": "string"
  },
  "message": "string"
}
```

## /users/

### GET
#### Summary

queries all user profile

#### Description

query all user profile

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| queryUserRequest | body | query user profile request parameter | Yes | [message.QueryUserReq](#messagequeryuserreq) |

##### message.QueryUserReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| page | integer | Current page location | No |
| pageSize | integer | Number of this request | No |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.ResultWithPage](#controllerresultwithpage) & [message.QueryUserResp](#messagequeryuserresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### message.QueryUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [][structs.UserInfo](#structsuserinfo) |  | No |

#### Example
request
```
curl -X 'GET' \
  'http://localhost:4116/api/v1/users/' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{
  "page": 1,
  "pageSize": 10
}'
```
response
```JSON
{
  "code": 0,
  "data": {
    "users": {
      "additionalProp1": {
        "createAt": "string",
        "creator": "string",
        "defaultTenantId": "string",
        "email": "string",
        "id": "string",
        "names": [
          "string"
        ],
        "nickname": "string",
        "phone": "string",
        "status": "string",
        "tenantIds": [
          "string"
        ],
        "updateAt": "string"
      },
      "additionalProp2": {
        "createAt": "string",
        "creator": "string",
        "defaultTenantId": "string",
        "email": "string",
        "id": "string",
        "names": [
          "string"
        ],
        "nickname": "string",
        "phone": "string",
        "status": "string",
        "tenantIds": [
          "string"
        ],
        "updateAt": "string"
      },
      "additionalProp3": {
        "createAt": "string",
        "creator": "string",
        "defaultTenantId": "string",
        "email": "string",
        "id": "string",
        "names": [
          "string"
        ],
        "nickname": "string",
        "phone": "string",
        "status": "string",
        "tenantIds": [
          "string"
        ],
        "updateAt": "string"
      }
    }
  },
  "message": "string",
  "page": {
    "page": 0,
    "pageSize": 0,
    "total": 0
  }
}
```

### POST
#### Summary

create user

#### Description

create user

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| createUserReq | body | create user request parameter | Yes | [message.CreateUserReq](#messagecreateuserreq) |

#### message.CreateUserReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| email | string |  | Yes |
| name | string |  | Yes |
| nickname | string |  | No |
| password | string |  | Yes |
| phone | string |  | No |
| tenantId | string |  | No |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/users/' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{
  "email": "string",
  "name": "me",
  "nickname": "string",
  "password": "12121212",
  "phone": "string",
  "tenantId": "string"
}'
```
response
```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

## /users/{userId}

### DELETE
#### Summary

delete user

#### Description

delete user

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path | user id | Yes | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### Example
request
```
curl -X 'DELETE' \
  'http://localhost:4116/api/v1/users/121212' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{}'
```
response
```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

### GET
#### Summary

get user profile

#### Description

get user profile

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path | user id | Yes | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & [message.GetUserResp](#messagegetuserresp) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### message.GetUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [structs.UserInfo](#structsuserinfo) |  | No |

##### structs.UserInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| createAt | string |  | No |
| creator | string |  | No |
| defaultTenantId | string |  | No |
| email | string |  | No |
| id | string |  | No |
| names | [ string ] |  | No |
| nickname | string |  | No |
| phone | string |  | No |
| status | string |  | No |
| tenantIds | [ string ] |  | No |
| updateAt | string |  | No |

#### Example
request
```
curl -X 'GET' \
  'http://localhost:4116/api/v1/users/121212' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken'
```
response
```JSON
{
  "code": 0,
  "data": {
    "user": {
      "createAt": "string",
      "creator": "string",
      "defaultTenantId": "string",
      "email": "string",
      "id": "string",
      "names": [
        "string"
      ],
      "nickname": "string",
      "phone": "string",
      "status": "string",
      "tenantIds": [
        "string"
      ],
      "updateAt": "string"
    }
  },
  "message": "string"
}
```

## /users/{userId}/password

### POST
#### Summary

update user password

#### Description

update user password

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path | user id | Yes | string |
| UpdateUserPasswordRequest | body | query user password request parameter | Yes | [message.UpdateUserPasswordReq](#messageupdateuserpasswordreq) |

##### message.UpdateUserPasswordReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| password | string |  | Yes |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/users/121212/password' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": "121212",
  "password": "newpassword"
}'
```
response
```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

## /users/{userId}/update_profile

### POST
#### Summary

update user profile

#### Description

update user profile

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path | user id | Yes | string |
| updateUserProfileRequest | body | query user profile request parameter | Yes | [message.UpdateUserProfileReq](#messageupdateuserprofilereq) |

##### message.UpdateUserProfileReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| email | string |  | No |
| nickname | string |  | No |
| phone | string |  | No |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/users/121212/update_profile' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{
  "email": "string",
  "nickname": "newNick",
  "phone": "string"
}'
```
response
```JSON
{
  "code": 0,
  "data": {},
  "message": "string"
}
```

## commonModel

### controller.CommonResult

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| data | object |  | No |
| message | string |  | No |

### controller.Page

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| page | integer |  | No |
| pageSize | integer |  | No |
| total | integer |  | No |

### controller.ResultWithPage

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| data | object |  | No |
| message | string |  | No |
| page | [controller.Page](#controllerpage) |  | No |
