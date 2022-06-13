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

package rbac

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/models"
	mock_account "github.com/pingcap/tiunimanager/test/mockmodels/mockaccount"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRBACManager_CreateRole_case1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountRW := mock_account.NewMockReaderWriter(ctrl)
	models.SetAccountReaderWriter(accountRW)
	accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	rbacService := GetRBACService()
	_, errCreate := rbacService.CreateRole(context.TODO(), message.CreateRoleReq{Role: "testrole"}, false)
	assert.Nil(t, errCreate)

	resp, errQuery := rbacService.QueryRoles(context.TODO(), message.QueryRolesReq{})
	assert.Nil(t, errQuery)
	assert.Equal(t, true, checkContainRole("testrole", resp.Roles))
}

func TestRBACManager_CreateRole_case2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountRW := mock_account.NewMockReaderWriter(ctrl)
	models.SetAccountReaderWriter(accountRW)
	accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	rbacService := GetRBACService()
	_, errCreate := rbacService.CreateRole(context.TODO(), message.CreateRoleReq{Role: ""}, false)
	assert.NotNil(t, errCreate)

	_, errCreate2 := rbacService.CreateRole(context.TODO(), message.CreateRoleReq{Role: string(constants.RbacRoleAdmin)}, false)
	assert.NotNil(t, errCreate2)
}

func TestRBACManager_DeleteRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountRW := mock_account.NewMockReaderWriter(ctrl)
	models.SetAccountReaderWriter(accountRW)
	accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	rbacService := GetRBACService()
	_, errDelete := rbacService.DeleteRole(context.TODO(), message.DeleteRoleReq{Role: "testrole"}, false)
	assert.Nil(t, errDelete)

	_, errDelete2 := rbacService.DeleteRole(context.TODO(), message.DeleteRoleReq{Role: string(constants.RbacRoleAdmin)}, false)
	assert.NotNil(t, errDelete2)

	resp, errQuery := rbacService.QueryRoles(context.TODO(), message.QueryRolesReq{})
	assert.Nil(t, errQuery)
	assert.Equal(t, false, checkContainRole("testrole", resp.Roles))
}

func TestRBACManager_BindRolesForUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountRW := mock_account.NewMockReaderWriter(ctrl)
	models.SetAccountReaderWriter(accountRW)
	accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	rbacService := GetRBACService()
	_, errBind1 := rbacService.BindRolesForUser(context.TODO(), message.BindRolesForUserReq{UserID: "testuser", Roles: []string{"testrole"}})
	assert.NotNil(t, errBind1)

	_, errBind2 := rbacService.BindRolesForUser(context.TODO(), message.BindRolesForUserReq{UserID: string(constants.RbacRoleAdmin), Roles: []string{"testrole"}})
	assert.NotNil(t, errBind2)

	_, errBind3 := rbacService.BindRolesForUser(context.TODO(), message.BindRolesForUserReq{UserID: "", Roles: []string{"testrole"}})
	assert.NotNil(t, errBind3)
}

func TestRBACManager_AddPermissionsForRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountRW := mock_account.NewMockReaderWriter(ctrl)
	models.SetAccountReaderWriter(accountRW)
	accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	rbacService := GetRBACService()
	_, errAdd1 := rbacService.AddPermissionsForRole(context.TODO(), message.AddPermissionsForRoleReq{Role: "", Permissions: []structs.RbacPermission{}}, false)
	assert.NotNil(t, errAdd1)

	_, errAdd2 := rbacService.AddPermissionsForRole(context.TODO(), message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{}}, false)
	assert.NotNil(t, errAdd2)

	_, errAdd3 := rbacService.AddPermissionsForRole(context.TODO(), message.AddPermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{{Resource: "testresource", Action: "testaction"}}}, true)
	assert.NotNil(t, errAdd3)
}

func TestRBACManager_DeletePermissionsForRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountRW := mock_account.NewMockReaderWriter(ctrl)
	models.SetAccountReaderWriter(accountRW)
	accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	rbacService := GetRBACService()
	_, errDel1 := rbacService.DeletePermissionsForRole(context.TODO(), message.DeletePermissionsForRoleReq{Role: "", Permissions: []structs.RbacPermission{}})
	assert.NotNil(t, errDel1)

	_, errDel2 := rbacService.DeletePermissionsForRole(context.TODO(), message.DeletePermissionsForRoleReq{Role: string(constants.RbacRoleAdmin), Permissions: []structs.RbacPermission{}})
	assert.NotNil(t, errDel2)

	_, errDel3 := rbacService.DeletePermissionsForRole(context.TODO(), message.DeletePermissionsForRoleReq{Role: "testrole", Permissions: []structs.RbacPermission{{Resource: "testresource", Action: "testaction"}}})
	assert.NotNil(t, errDel3)
}

func TestRBACManager_CheckPermissionForUser_case1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountRW := mock_account.NewMockReaderWriter(ctrl)
	models.SetAccountReaderWriter(accountRW)
	accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	rbacService := GetRBACService()
	resp, err := rbacService.CheckPermissionForUser(context.TODO(), message.CheckPermissionForUserReq{UserID: "user", Permissions: []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionRead)}}})
	assert.Nil(t, err)
	assert.Equal(t, resp.Result, false)
}

func TestRBACManager_CheckPermissionForUser_case2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountRW := mock_account.NewMockReaderWriter(ctrl)
	models.SetAccountReaderWriter(accountRW)
	accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	rbacService := GetRBACService()
	_, errCreate := rbacService.CreateRole(context.TODO(), message.CreateRoleReq{Role: "testrole"}, false)
	assert.Nil(t, errCreate)

	_, errAdd := rbacService.AddPermissionsForRole(context.TODO(), message.AddPermissionsForRoleReq{Role: "testrole", Permissions: []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionAll)}}}, false)
	assert.Nil(t, errAdd)

	_, errBind := rbacService.BindRolesForUser(context.TODO(), message.BindRolesForUserReq{UserID: "testuser", Roles: []string{"testrole"}})
	assert.Nil(t, errBind)

	query, errQuery := rbacService.QueryRoles(context.TODO(), message.QueryRolesReq{UserID: "testuser"})
	assert.Nil(t, errQuery)
	assert.Equal(t, true, checkContainRole("testrole", query.Roles))

	resp, errCheck := rbacService.CheckPermissionForUser(context.TODO(), message.CheckPermissionForUserReq{UserID: "testuser", Permissions: []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionAll)}}})
	assert.Nil(t, errCheck)
	assert.Equal(t, resp.Result, true)

	_, errUnBind := rbacService.UnbindRoleForUser(context.TODO(), message.UnbindRoleForUserReq{UserID: "testuser", Role: "testrole"})
	assert.Nil(t, errUnBind)

	_, errDelPermission := rbacService.DeletePermissionsForRole(context.TODO(), message.DeletePermissionsForRoleReq{Role: "testrole", Permissions: []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionAll)}}})
	assert.Nil(t, errDelPermission)

	resp, errCheck2 := rbacService.CheckPermissionForUser(context.TODO(), message.CheckPermissionForUserReq{UserID: "testuser", Permissions: []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionAll)}}})
	assert.Nil(t, errCheck2)
	assert.Equal(t, resp.Result, false)

	_, errDelRole := rbacService.DeleteRole(context.TODO(), message.DeleteRoleReq{Role: "testrole"}, false)
	assert.Nil(t, errDelRole)
}

func TestRBACManager_QueryPermissionsForUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountRW := mock_account.NewMockReaderWriter(ctrl)
	models.SetAccountReaderWriter(accountRW)
	accountRW.EXPECT().GetUserByName(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	rbacService := GetRBACService()
	_, errCreate := rbacService.CreateRole(context.TODO(), message.CreateRoleReq{Role: "testrole"}, false)
	assert.Nil(t, errCreate)

	_, errAdd := rbacService.AddPermissionsForRole(context.TODO(), message.AddPermissionsForRoleReq{Role: "testrole", Permissions: []structs.RbacPermission{{Resource: string(constants.RbacResourceCluster), Action: string(constants.RbacActionAll)}}}, false)
	assert.Nil(t, errAdd)

	_, errBind := rbacService.BindRolesForUser(context.TODO(), message.BindRolesForUserReq{UserID: "testuser", Roles: []string{"testrole"}})
	assert.Nil(t, errBind)

	resp, errQuery := rbacService.QueryPermissionsForUser(context.TODO(), message.QueryPermissionsForUserReq{UserID: "testuser"})
	assert.Nil(t, errQuery)
	assert.Equal(t, "testuser", resp.UserID)
	assert.Equal(t, string(constants.RbacResourceCluster), resp.Permissions[0].Resource)
	assert.Equal(t, string(constants.RbacActionAll), resp.Permissions[0].Action)
}

func checkContainRole(role string, roles []string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
