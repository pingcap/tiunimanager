/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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

package hostinitiator

import (
	"bytes"
	"context"
	"encoding/json"
	"text/template"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	resourceTemplate "github.com/pingcap/tiunimanager/resource/template"
)

type HostAddr struct {
	HostIP  string
	SSHPort int
}

// template info to parse em cluster scale out yaml template file
type templateScaleOut struct {
	HostAddrs []HostAddr
}

func (p *templateScaleOut) generateTopologyConfig(ctx context.Context) (string, error) {
	t, err := template.New("import_topology").Parse(resourceTemplate.EMClusterScaleOut)
	if err != nil {
		return "", errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, err.Error())
	}

	topology := new(bytes.Buffer)
	if err = t.Execute(topology, p); err != nil {
		return "", errors.NewError(errors.TIUNIMANAGER_UNRECOGNIZED_ERROR, err.Error())
	}
	framework.LogWithContext(ctx).Infof("generate topology config: %s", topology.String())

	return topology.String(), nil
}

// template info to parse cluster check yaml template file
type templateCheckHost struct {
	GlobalSSHPort            int
	GlobalUser               string
	GlobalGroup              string
	GlobalArch               string
	TemplateItemsForCompute  []checkHostTemplateItem
	TemplateItemsForSchedule []checkHostTemplateItem
	TemplateItemsForStorage  []checkHostTemplateItem
}

type checkHostTemplateItem struct {
	HostIP    string
	DataDir   string
	DeployDir string
	Port1     int
	Port2     int
}

func (p *templateCheckHost) buildCheckHostTemplateItems(h *structs.HostInfo) {
	p.GlobalUser = framework.GetCurrentDeployUser()
	p.GlobalGroup = framework.GetCurrentDeployGroup()
	p.GlobalSSHPort = int(h.SSHPort)
	p.GlobalArch = constants.GetArchAlias(constants.ArchType(h.Arch))
	p.TemplateItemsForCompute = make([]checkHostTemplateItem, 0)
	p.TemplateItemsForSchedule = make([]checkHostTemplateItem, 0)
	p.TemplateItemsForStorage = make([]checkHostTemplateItem, 0)

	// define below start port for each component just to work around the port conflict in check yaml topology
	var tidbPort = 10000
	var tidbStatusPort = 11000
	var tikvPort = 12000
	var tikvStatusPort = 13000
	var pdPeerPort = 14000
	var pdClientPort = 15000

	purposes := h.GetPurposes()
	for _, purpose := range purposes {
		if purpose == string(constants.PurposeCompute) {
			for _, disk := range h.Disks {
				if disk.Status != string(constants.DiskAvailable) {
					continue
				}
				p.TemplateItemsForCompute = append(p.TemplateItemsForCompute, checkHostTemplateItem{
					HostIP: h.IP,
					// Only DeployDir for tidb
					DeployDir: disk.Path,
					Port1:     tidbPort,
					Port2:     tidbStatusPort,
				})
				tidbPort++
				tidbStatusPort++
			}
		}
		if purpose == string(constants.PurposeSchedule) {
			for _, disk := range h.Disks {
				if disk.Status != string(constants.DiskAvailable) {
					continue
				}
				p.TemplateItemsForSchedule = append(p.TemplateItemsForSchedule, checkHostTemplateItem{
					HostIP:    h.IP,
					DataDir:   disk.Path,
					DeployDir: disk.Path,
					Port1:     pdPeerPort,
					Port2:     pdClientPort,
				})
				pdPeerPort++
				pdClientPort++
			}
		}
		if purpose == string(constants.PurposeStorage) {
			for _, disk := range h.Disks {
				if disk.Status != string(constants.DiskAvailable) {
					continue
				}
				p.TemplateItemsForStorage = append(p.TemplateItemsForStorage, checkHostTemplateItem{
					HostIP:    h.IP,
					DataDir:   disk.Path,
					DeployDir: disk.Path,
					Port1:     tikvPort,
					Port2:     tikvStatusPort,
				})
				tikvPort++
				tikvStatusPort++
			}
		}
	}
}

func (p *templateCheckHost) generateTopologyConfig(ctx context.Context) (string, error) {
	t, err := template.New("checkHost_topology").Parse(resourceTemplate.EMClusterCheck)
	if err != nil {
		return "", errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, err.Error())
	}

	topology := new(bytes.Buffer)
	if err = t.Execute(topology, p); err != nil {
		return "", errors.NewError(errors.TIUNIMANAGER_UNRECOGNIZED_ERROR, err.Error())
	}
	framework.LogWithContext(ctx).Infof("generate topology config: %s", topology.String())

	return topology.String(), nil
}

type checkHostResults struct {
	Result []checkHostResult `json:"result"`
}

func (results *checkHostResults) buildFromJson(resultStr string) (err error) {
	return json.Unmarshal([]byte(resultStr), results)
}

func (results checkHostResults) analyzeCheckResults() (sortedResult map[string]*[]checkHostResult) {
	sortedResult = make(map[string]*[]checkHostResult)
	for i := range results.Result {
		if res, ok := sortedResult[results.Result[i].Status]; ok {
			*res = append(*res, results.Result[i])
		} else {
			sortedResult[results.Result[i].Status] = &[]checkHostResult{results.Result[i]}
		}
	}
	return
}

type ResultStatus string

const (
	Pass ResultStatus = "Pass"
	Warn ResultStatus = "Warn"
	Fail ResultStatus = "Fail"
)

// hostCheckResult represents the check result of each node
type checkHostResult struct {
	Node    string `json:"node"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
