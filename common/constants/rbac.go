/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package constants

// RbacAction Definition rbac permission action enum
type RbacAction string

var RbacActionMap = map[string]RbacAction{
	string(RbacActionAll):    RbacActionAll,
	string(RbacActionRead):   RbacActionRead,
	string(RbacActionCreate): RbacActionCreate,
	string(RbacActionUpdate): RbacActionUpdate,
	string(RbacActionDelete): RbacActionDelete,
}

const (
	RbacActionAll    RbacAction = "*"
	RbacActionRead   RbacAction = "read"
	RbacActionCreate RbacAction = "create"
	RbacActionUpdate RbacAction = "update"
	RbacActionDelete RbacAction = "delete"
)

// RbacResource Definition rbac resource enum
type RbacResource string

var RbacResourceMap = map[string]RbacResource{
	string(RbacResourceCluster):   RbacResourceCluster,
	string(RbacResourceResource):  RbacResourceResource,
	string(RbacResourceParameter): RbacResourceParameter,
	string(RbacResourceUser):      RbacResourceUser,
	string(RbacResourceCDC):       RbacResourceCDC,
	string(RbacResourceProduct):   RbacResourceProduct,
	string(RbacResourceSystem):    RbacResourceSystem,
	string(RbacResourceWorkflow):  RbacResourceWorkflow,
}

const (
	RbacResourceCluster   RbacResource = "cluster"
	RbacResourceResource  RbacResource = "resource"
	RbacResourceParameter RbacResource = "parameter"
	RbacResourceUser      RbacResource = "user"
	RbacResourceCDC       RbacResource = "cdc"
	RbacResourceProduct   RbacResource = "product"
	RbacResourceSystem    RbacResource = "system"
	RbacResourceWorkflow  RbacResource = "workflow"
)

// RbacRole Definition rbac role enum
type RbacRole string

var RbacRoleMap = map[string]RbacRole{
	string(RbacRoleAdmin): RbacRoleAdmin,
}

const (
	RbacRoleAdmin RbacRole = "RBAC_ROLE_ADMIN"
)
