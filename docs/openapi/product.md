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

##### message.QueryProductsInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | [ [structs.ProductConfigInfo](#structsproductconfiginfo) ] |  | No |

##### structs.ProductConfigInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| components | [ [structs.ProductComponentPropertyWithZones](#structsproductcomponentpropertywithzones) ] |  | No |
| productId | string |  | No |
| productName | string |  | No |
| versions | [ [structs.SpecificVersionProduct](#structsspecificversionproduct) ] |  | No |

#### Example
request
```
curl -X 'GET' \
  'http://localhost:4116/api/v1/products/?productIDs=TiDB' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken'
```
response
```JSON
{
  "code": 0,
  "data": {
    "products": [
      {
        "components": [
          {
            "availableZones": [
              {
                "specs": [
                  {
                    "cpu": 0,
                    "diskType": "string",
                    "id": "string",
                    "memory": 0,
                    "name": "string",
                    "zoneId": "string",
                    "zoneName": "string"
                  }
                ],
                "zoneId": "string",
                "zoneName": "string"
              }
            ],
            "endPort": 0,
            "id": "string",
            "maxInstance": 0,
            "maxPort": 0,
            "minInstance": 0,
            "name": "string",
            "purposeType": "string",
            "startPort": 0,
            "suggestedInstancesCount": [
              0
            ]
          }
        ],
        "productId": "string",
        "productName": "string",
        "versions": [
          {
            "arch": "string",
            "productId": "string",
            "version": "string"
          }
        ]
      }
    ]
  },
  "message": "string"
}
```

### POST
#### Summary

update products

#### Description

update products

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| UpdateProductsInfoReq | body | update products info request parameter | Yes | [message.UpdateProductsInfoReq](#messageupdateproductsinforeq) |

##### message.UpdateProductsInfoReq

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | [ [structs.ProductConfigInfo](#structsproductconfiginfo) ] |  | Yes |

##### structs.ProductConfigInfo

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

##### message.UpdateProductsInfoResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message.UpdateProductsInfoResp | object |  |  |

#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/products/' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{
  "products": [
    {
      "components": [
        {
          "availableZones": [
            {
              "specs": [
                {
                  "cpu": 0,
                  "diskType": "string",
                  "id": "string",
                  "memory": 0,
                  "name": "string",
                  "zoneId": "string",
                  "zoneName": "string"
                }
              ],
              "zoneId": "string",
              "zoneName": "string"
            }
          ],
          "endPort": 0,
          "id": "string",
          "maxInstance": 0,
          "maxPort": 0,
          "minInstance": 0,
          "name": "string",
          "purposeType": "string",
          "startPort": 0,
          "suggestedInstancesCount": [
            0
          ]
        }
      ],
      "productId": "string",
      "productName": "string",
      "versions": [
        {
          "arch": "string",
          "productId": "string",
          "version": "string"
        }
      ]
    }
  ]
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

##### message.QueryAvailableProductsResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | object | arch version | No |

#### Example
request
```
curl -X 'GET' \
  'http://localhost:4116/api/v1/products/available?vendorId=Local' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken'
```

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

##### message.QueryProductDetailResp

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| products | object |  | No |

#### structs.ProductDetail

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string | The ID of the product consists of the product ID | No |
| name | string | The name of the product consists of the product name and the version | No |
| versions | object | Organize product information by version | No |

#### Example
request
```
curl -X 'GET' \
'http://localhost:4116/api/v1/products/detail?internalProduct=0&productId=TiDB&regionId=region1&vendorId=Local' \
-H 'accept: application/json' \
-H 'Authorization: Bearer mytoken'
```

## /vendors/

### GET
#### Summary

query vendors

#### Description

query vendors

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| vendorIDs | query | vendor id collection | No | [ string ] |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### Example
request
```
curl -X 'GET' \
  'http://localhost:4116/api/v1/vendors/' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken'
```
response
```JSON
{
  "code": 0,
  "data": {
    "vendors": [
      {
        "id": "string",
        "name": "string",
        "regions": [
          {
            "id": "string",
            "name": "string",
            "zones": [
              {
                "comment": "string",
                "zoneId": "string",
                "zoneName": "string"
              }
            ]
          }
        ],
        "specs": [
          {
            "cpu": 0,
            "diskType": "string",
            "id": "string",
            "memory": 0,
            "name": "string",
            "purposeType": "string"
          }
        ]
      }
    ]
  },
  "message": "string"
}
```

### POST
#### Summary

update vendors

#### Description

update vendors

#### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| UpdateVendorInfoReq | body | update vendor info request parameter | Yes | [message.UpdateVendorInfoReq](#messageupdatevendorinforeq) |

#### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [controller.CommonResult](#controllercommonresult) & object |
| 401 | Unauthorized | [controller.CommonResult](#controllercommonresult) |
| 403 | Forbidden | [controller.CommonResult](#controllercommonresult) |
| 500 | Internal Server Error | [controller.CommonResult](#controllercommonresult) |

#### Example
request
```
curl -X 'POST' \
  'http://localhost:4116/api/v1/vendors/' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken' \
  -H 'Content-Type: application/json' \
  -d '{
  "vendors": [
    {
      "id": "Local",
      "name": "string",
      "regions": [
        {
          "id": "region1",
          "name": "string",
          "zones": [
            {
              "comment": "string",
              "zoneId": "zone1",
              "zoneName": "zone1"
            }
          ]
        }
      ],
      "specs": [
        {
          "cpu": 4,
          "diskType": "string",
          "id": "large",
          "memory": 8,
          "name": "string",
          "purposeType": "string"
        }
      ]
    }
  ]
}'
```
response
```JSON
{
  "code": 0,
  "data": {
    "vendors": [
      {
        "id": "string",
        "name": "string",
        "regions": [
          {
            "id": "string",
            "name": "string",
            "zones": [
              {
                "comment": "string",
                "zoneId": "string",
                "zoneName": "string"
              }
            ]
          }
        ],
        "specs": [
          {
            "cpu": 0,
            "diskType": "string",
            "id": "string",
            "memory": 0,
            "name": "string",
            "purposeType": "string"
          }
        ]
      }
    ]
  },
  "message": "string"
}
```

## /vendors/available

### GET
#### Summary

query available vendors and regions

#### Description

query available vendors and regions

#### Responses

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

#### Example
request
```
curl -X 'GET' \
  'http://localhost:4116/api/v1/vendors/available' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer mytoken'
```
response
```JSON
{
  "code": 0,
  "data": {
    "vendors": {
      "vendor1": {
        "id": "vendor1",
        "name": "string",
        "regions": {
          "region1": {
            "id": "string",
            "name": "string"
          },
          "region2": {
            "id": "string",
            "name": "string"
          },
          "region3": {
            "id": "string",
            "name": "string"
          }
        }
      },
      "vendor2": {
        "id": "vendor2",
        "name": "string",
        "regions": {
          "region4": {
            "id": "string",
            "name": "string"
          },
          "region6": {
            "id": "string",
            "name": "string"
          },
          "region5": {
            "id": "string",
            "name": "string"
          }
        }
      }
    }
  },
  "message": "string"
}
```
curl -X 'GET' \
'http://localhost:4116/api/v1/vendors/available' \
-H 'accept: application/json' \
-H 'Authorization: Bearer mytoken'

## CommonModel

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
