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
 * @File: metaconfig.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/11
*******************************************************************************/

package meta

import (
	"bytes"
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/platform/config"
	resourceTemplate "github.com/pingcap-inc/tiem/resource/template"
	"text/template"
)

// GenerateTopologyConfig
// @Description generate yaml config based on cluster topology
// @Return		yaml config
// @Return		error
func (p *ClusterMeta) GenerateTopologyConfig(ctx context.Context) (string, error) {
	if p.Cluster == nil || len(p.Instances) == 0 {
		return "", errors.NewError(errors.TIEM_PARAMETER_INVALID, "cluster topology is empty, please check it!")
	}

	t, err := template.New("topology").Parse(resourceTemplate.ClusterTopology)
	if err != nil {
		return "", errors.NewError(errors.TIEM_PARAMETER_INVALID, err.Error())
	}

	topology := new(bytes.Buffer)
	if err = t.Execute(topology, NewClusterMetaRenderData(ctx, *p)); err != nil {
		return "", errors.NewError(errors.TIEM_UNRECOGNIZED_ERROR, err.Error())
	}
	framework.LogWithContext(ctx).Infof("generate topology config: %s", topology.String())

	return topology.String(), nil
}

func NewClusterMetaRenderData(ctx context.Context, meta ClusterMeta) *ClusterMetaRenderData {
	globalUser := framework.Current.GetClientArgs().DeployUser
	if len(globalUser) == 0 {
		globalUser = "tidb"
	}
	globalGroup := framework.Current.GetClientArgs().DeployGroup
	if len(globalGroup) == 0 {
		globalGroup = globalUser
	}
	sshConfigPort, err := models.GetConfigReaderWriter().GetConfig(ctx, constants.ConfigKeyDefaultSSHPort)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get config ConfigKeyDefaultSSHPort failed, err = %s", err.Error())
		sshConfigPort = &config.SystemConfig{
			ConfigValue: "22",
		}
	}

	// init grafana user and password
	grafanaUser := framework.Current.GetClientArgs().DeployUser
	if len(grafanaUser) == 0 {
		grafanaUser = "admin"
	}
	grafanaPassword := GetRandomString(10)

	return &ClusterMetaRenderData{
		ClusterMeta:     meta,
		GlobalUser:      globalUser,
		GlobalGroup:     globalGroup,
		GlobalSSHPort:   sshConfigPort.ConfigValue,
		GrafanaUser:     grafanaUser,
		GrafanaPassword: grafanaPassword,
	}
}

type ClusterMetaRenderData struct {
	ClusterMeta
	GlobalUser      string
	GlobalGroup     string
	GlobalSSHPort   string
	GrafanaUser     string
	GrafanaPassword string
}
