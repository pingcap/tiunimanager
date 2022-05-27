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

#### message.LoginReq

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

#### message.LoginResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| passwordExpired | boolean |  | No |
| tenantId | string |  | No |
| token | string |  | No |
| userId | string |  | No |

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

#### message.LogoutResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userId | string |  | No |

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

#### message.QueryUserReq

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

#### message.QueryUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [][structs.UserInfo](#structsuserinfo) |  | No |

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

#### message.GetUserResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [structs.UserInfo](#structsuserinfo) |  | No |

#### structs.UserInfo

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

#### message.UpdateUserPasswordReq

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

#### message.UpdateUserProfileReq

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
