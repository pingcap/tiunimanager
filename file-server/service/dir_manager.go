/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/file-server/controller"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"path/filepath"
)

var DirMgr DirManager

type DirManager struct {
}

func InitDirManager() *DirManager {
	DirMgr = DirManager{}
	return &DirMgr
}

func (mgr *DirManager) GetImportPath(ctx context.Context, clusterId string) (string, error) {
	request := &message.GetSystemConfigReq{
		ConfigKey: constants.ConfigKeyImportShareStoragePath,
	}

	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("marshal request error: %s", err.Error())
		return "", fmt.Errorf("marshal request error: %s", err.Error())
	}

	rpcResp, err := client.ClusterClient.GetSystemConfig(ctx, &clusterpb.RpcRequest{Request: string(body)}, controller.DefaultTimeout)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call cluster service api failed %s", err.Error())
		return "", fmt.Errorf("call cluster service api failed %s", err.Error())
	}
	var resp message.GetSystemConfigResp
	err = json.Unmarshal([]byte(rpcResp.Response), &resp)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("unmarshal get system config rpc response error: %s", err.Error())
		return "", fmt.Errorf("unmarshal get system config rpc response error: %s", err.Error())
	}

	importPath := resp.ConfigValue
	if importPath == "" {
		framework.LogWithContext(ctx).Error("import path is empty")
		return "", fmt.Errorf("import path is empty")
	}
	framework.LogWithContext(ctx).Infof("get configKey %s, configValue %s success", constants.ConfigKeyImportShareStoragePath, importPath)
	importAbsDir, err := filepath.Abs(importPath)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("import dir %s is not vaild", importPath)
		return "", fmt.Errorf("import dir %s is not vaild", importPath)
	}

	importClusterPath := filepath.Join(importAbsDir, clusterId)
	if filepath.Dir(importClusterPath) != importAbsDir {
		framework.LogWithContext(ctx).Errorf("clusterId %s invaild, may cause path crossing", clusterId)
		return "", fmt.Errorf("clusterId %s invaild, may cause path crossing", clusterId)
	}

	return fmt.Sprintf("%s/%s/temp", importAbsDir, clusterId), nil
}
