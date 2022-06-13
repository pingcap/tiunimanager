/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 ******************************************************************************/

/*******************************************************************************
 * @File: metadisplay_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/23
*******************************************************************************/

package meta

import (
	"context"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/models/cluster/management"
	"github.com/pingcap/tiunimanager/models/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClusterMeta_GetDBUserNamePassword(t *testing.T) {
	meta := &ClusterMeta {
		Cluster: &management.Cluster{
			Version: "v5.2.2",
		},
		DBUsers: map[string]*management.DBUser {
			string(constants.Root):                      {Name: "root", Password: common.PasswordInExpired{Val: "aaa"}},
			string(constants.DBUserBackupRestore):       {Name: "br"},
			string(constants.DBUserParameterManagement): {Name: "parameter"},
			string(constants.DBUserCDCDataSync):         {Name: "cdc"},
			string(constants.DBUserGrafana):             {Name: "grafana"},
		},
	}
	t.Run("normal", func(t *testing.T) {
		account, err := meta.GetDBUserNamePassword(context.TODO(), constants.Root)
		assert.NoError(t, err)
		assert.Equal(t, "root", account.Name)
		assert.Equal(t, "aaa", account.Password.Val)

		account, err = meta.GetDBUserNamePassword(context.TODO(), constants.DBUserBackupRestore)
		assert.NoError(t, err)
		assert.Equal(t, "br", account.Name)
		account, err = meta.GetDBUserNamePassword(context.TODO(), constants.DBUserParameterManagement)
		assert.NoError(t, err)
		assert.Equal(t, "parameter", account.Name)
		account, err = meta.GetDBUserNamePassword(context.TODO(), constants.DBUserCDCDataSync)
		assert.NoError(t, err)
		assert.Equal(t, "cdc", account.Name)
		account, err = meta.GetDBUserNamePassword(context.TODO(), constants.DBUserGrafana)
		assert.NoError(t, err)
		assert.Equal(t, "grafana", account.Name)
	})
	t.Run("unknown type", func(t *testing.T) {
		_, err := meta.GetDBUserNamePassword(context.TODO(), "unknown")
		assert.Error(t, err)
	})
	t.Run("unknown version", func(t *testing.T) {
		meta.Cluster.Version = "vvvv"
		defer func() {
			meta.Cluster.Version = "v5.2.2"
		}()
		_, err := meta.GetDBUserNamePassword(context.TODO(), constants.DBUserBackupRestore)
		assert.Error(t, err)
	})
	t.Run("5.0.0 br", func(t *testing.T) {
		meta.Cluster.Version = "v5.0.0"
		defer func() {
			meta.Cluster.Version = "v5.2.2"
		}()
		account, err := meta.GetDBUserNamePassword(context.TODO(), constants.DBUserBackupRestore)
		assert.NoError(t, err)
		assert.Equal(t, "root", account.Name)
	})
	t.Run("5.2.1 cdc", func(t *testing.T) {
		meta.Cluster.Version = "v5.2.1"
		defer func() {
			meta.Cluster.Version = "v5.2.2"
		}()
		_, err := meta.GetDBUserNamePassword(context.TODO(), constants.DBUserCDCDataSync)
		assert.Error(t, err)
	})
}
