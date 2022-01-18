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
	gormadapter "github.com/casbin/gorm-adapter/v3"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type RBACReadWrite struct {
	dbCommon.GormDB
}

func NewRBACReadWrite(db *gorm.DB) *RBACReadWrite {
	m := RBACReadWrite{
		dbCommon.WrapDB(db),
	}
	return &m
}

func (m *RBACReadWrite) GetRBACAdapter(ctx context.Context) (*gormadapter.Adapter, error) {
	return gormadapter.NewAdapterByDBWithCustomTable(m.DB(ctx), &RBAC{}, "rbac")
}
