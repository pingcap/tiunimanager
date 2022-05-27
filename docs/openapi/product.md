## /products/

### GET
#### Summary

query products by ids

#### Description

query products by ids

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| productIDs | query | product id collection | No | [ string ] |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### message.QueryProductsInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | [ [structs.ProductConfigInfo](#structsproductconfiginfo) ] |  | No |

#### structs.ProductConfigInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| components | [ [structs.ProductComponentPropertyWithZones](#structsproductcomponentpropertywithzones) ] |  | No |
| productId | string |  | No |
| productName | string |  | No |
| versions | [ [structs.SpecificVersionProduct](#structsspecificversionproduct) ] |  | No |

### POST
#### Summary

update products

#### Description

update products

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| UpdateProductsInfoReq | body | update products info request parameter | Yes | [message.UpdateProductsInfoReq](#messageupdateproductsinforeq) |

#### message.UpdateProductsInfoReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | [ [structs.ProductConfigInfo](#structsproductconfiginfo) ] |  | Yes |

#### structs.ProductConfigInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| components | [ [structs.ProductComponentPropertyWithZones](#structsproductcomponentpropertywithzones) ] |  | No |
| productId | string |  | No |
| productName | string |  | No |
| versions | [ [structs.SpecificVersionProduct](#structsspecificversionproduct) ] |  | No |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### message.UpdateProductsInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateProductsInfoResp | object |  |  |

## /products/available

### GET
#### Summary

queries all products' information

#### Description

queries all products' information

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| internalProduct | query |  | No | integer |
| status | query |  | No | string |
| vendorId | query |  | No | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |


#### message.QueryAvailableProductsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | object | arch version | No |

## /products/detail

### GET
#### Summary

query all product detail

#### Description

query all product detail

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| internalProduct | query |  | No | integer |
| productId | query |  | No | string |
| regionId | query |  | No | string |
| status | query |  | No | string |
| vendorId | query |  | No | string |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### message.QueryProductDetailResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | object |  | No |

#### structs.ProductDetail

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | The ID of the product consists of the product ID | No |
| name | string | The name of the product consists of the product name and the version | No |
| versions | object | Organize product information by version | No |

### /vendors/

#### GET
##### Summary

query vendors

##### Description

query vendors

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| vendorIDs | query | vendor id collection | No | [ string ] |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

#### POST
##### Summary

update vendors

##### Description

update vendors

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| UpdateVendorInfoReq | body | update vendor info request parameter | Yes | [message.UpdateVendorInfoReq](#messageupdatevendorinforeq) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /vendors/available

#### GET
##### Summary

query available vendors and regions

##### Description

query available vendors and regions

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

##### message.QueryAvailableVendorsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| vendors | object |  | No |

##### message.QueryVendorInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| vendors | [ [structs.VendorConfigInfo](#structsvendorconfiginfo) ] |  | No |

#### controller.CommonResult

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| data | object |  | No |
| message | string |  | No |

#### controller.Page

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| page | integer |  | No |
| pageSize | integer |  | No |
| total | integer |  | No |

#### controller.ResultWithPage

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| data | object |  | No |
| message | string |  | No |
| page | [controller.Page](#controllerpage) |  | No |
