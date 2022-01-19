/*
 * Copyright (c)  2022 PingCAP, Inc.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*******************************************************************************
 * @File: deploymentInterface
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/31
*******************************************************************************/

package deployment

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

// TiUPComponentType Type of TiUP component, e.g. cluster/dm/tiem
type TiUPComponentType string

const (
	TiUPComponentTypeCluster = "cluster"
	TiUPComponentTypeDM      = "dm"
	TiUPComponentTypeTiEM    = "tiem"
	TiUPComponentTypeCtrl    = "ctl"
	TiUPComponentTypeDefault = "default"
)

type TiUPMeta struct {
	Component string
	Servers   []*interface{}
}

var component = map[constants.EMProductComponentIDType]TiUPMeta{
	constants.ComponentIDTiDB:             {spec.ComponentTiDB, nil},
	constants.ComponentIDTiKV:             {spec.ComponentTiKV, nil},
	constants.ComponentIDTiFlash:          {spec.ComponentTiFlash, nil},
	constants.ComponentIDPD:               {spec.ComponentPD, nil},
	constants.ComponentIDCDC:              {spec.ComponentCDC, nil},
	constants.ComponentIDGrafana:          {spec.ComponentGrafana, nil},
	constants.ComponentIDPrometheus:       {spec.ComponentPrometheus, nil},
	constants.ComponentIDAlertManger:      {spec.ComponentAlertmanager, nil},
	constants.ComponentIDNodeExporter:     {spec.ComponentNodeExporter, nil},
	constants.ComponentIDBlackboxExporter: {spec.ComponentBlackboxExporter, nil},
}

const (
	CMDDeploy       = "deploy"
	CMDYes          = "--yes"
	CMDScaleOut     = "scale-out"
	CMDScaleIn      = "scale-in"
	CMDNode         = "--node"
	CMDStart        = "start"
	CMDRestart      = "restart"
	CMDStop         = "stop"
	CMDList         = "list"
	CMDDestroy      = "destroy"
	CMDDisplay      = "display"
	CMDUpgrade      = "upgrade"
	CMDShowConfig   = "show-config"
	CMDEditConfig   = "edit-config"
	CMDReload       = "reload"
	CMDExec         = "exec"
	CMDDumpling     = "dumpling"
	CMDLightning    = "tidb-lightning"
	CMDTopologyFile = "--topology-file"
	CMDPush         = "push"
)

type Interface interface {
	// Deploy
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param version
	// @param configYaml
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Deploy(ctx context.Context, componentType TiUPComponentType, clusterID, version, configYaml, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// Start
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Start(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// Restart
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Restart(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// Upgrade
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param version
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Upgrade(ctx context.Context, componentType TiUPComponentType, clusterID, version, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// ScaleIn
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param nodeID
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	ScaleIn(ctx context.Context, componentType TiUPComponentType, clusterID, nodeID, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// ScaleOut
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param configYaml
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	ScaleOut(ctx context.Context, componentType TiUPComponentType, clusterID, configYaml, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// Destroy
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Destroy(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// Reload
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Reload(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// EditClusterConfig
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param home
	// @param workFlowID
	// @param configs
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	EditClusterConfig(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, configs map[string]map[string]interface{}, args []string, timeout int) (ID string, err error)
	// EditInstanceConfig
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param component
	// @param host
	// @param home
	// @param workFlowID
	// @param config
	// @param args[]
	// @param port
	// @param timeout
	// @return ID
	// @return err
	EditInstanceConfig(ctx context.Context, componentType TiUPComponentType, clusterID, component, host, home, workFlowID string, config map[string]interface{}, args []string, port, timeout int) (ID string, err error)
	// List
	// @Description:
	// @param ctx
	// @param componentType
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return result
	// @return err
	List(ctx context.Context, componentType TiUPComponentType, home, workFlowID string, args []string, timeout int) (result string, err error)
	// Display
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return result
	// @return err
	Display(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (result string, err error)
	// ShowConfig
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return spec
	// @return err
	ShowConfig(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (spec *spec.Specification, err error)
	// Dumpling
	// @Description:
	// @param ctx
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Dumpling(ctx context.Context, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// Lightning
	// @Description:
	// @param ctx
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Lightning(ctx context.Context, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// Push
	// @Description:
	// @param ctx
	// @param componentType
	// @param clusterID
	// @param collectorYaml
	// @param remotePath
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Push(ctx context.Context, componentType TiUPComponentType, clusterID, collectorYaml, remotePath, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// Ctl
	// @Description:
	// @param ctx
	// @param componentType
	// @param version
	// @param component
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return ID
	// @return err
	Ctl(ctx context.Context, componentType TiUPComponentType, version, component, home, workFlowID string, args []string, timeout int) (ID string, err error)
	// Exec
	// @Description:
	// @param ctx
	// @param componentType
	// @param home
	// @param workFlowID
	// @param args[]
	// @param timeout
	// @return result
	// @return err
	Exec(ctx context.Context, componentType TiUPComponentType, home, workFlowID string, args []string, timeout int) (result string, err error)
}