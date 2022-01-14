/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
	"fmt"
	"text/template"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	resourceTemplate "github.com/pingcap-inc/tiem/resource/template"
)

func (p *FileHostInitiator) Verify2(ctx context.Context, h *structs.HostInfo) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("apply and verify host %v begins", *h)
	tempateInfo := templateCheckHost{}
	tempateInfo.buildCheckHostTemplateItems(h)

	templateStr, err := tempateInfo.generateTopologyConfig(ctx)
	if err != nil {
		return err
	}
	ignoreWarnings, ok := ctx.Value(rp_consts.ContextIgnoreWarnings).(bool)
	if !ok {
		return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, "get ignore warning flag from context failed")
	}
	log.Infof("apply and check cluster ignore warning (%v) on %s", ignoreWarnings, templateStr)

	if rp_consts.SecondPartyReady {
		resultStr, err := p.secondPartyServ.Check(ctx, secondparty.TiEMComponentTypeStr, templateStr, rp_consts.DefaultTiupTimeOut,
			[]string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa", "--apply", "--format", "json"})
		if err != nil {
			errMsg := fmt.Sprintf("call second serv to check host %s %s [%v] failed, %v", h.HostName, h.IP, templateStr, err)
			return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
		}
		framework.LogWithContext(ctx).Infof("check host %s %s for %v done", h.HostName, h.IP, tempateInfo)

		// deal with the result
		var results checkHostResults
		(&results).buildFromJson(resultStr)
		sortedResult := results.analyzeCheckResults()

		pass := sortedResult["Pass"]
		fails := sortedResult["Fail"]
		warnings := sortedResult["Warn"]

		if len(*fails) > 0 {
			errMsg := fmt.Sprintf("check host %s %s has %d fails, %v", h.HostName, h.IP, len(*fails), *fails)
			return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
		}

		if len(*warnings) > 0 && !ignoreWarnings {
			errMsg := fmt.Sprintf("check host %s %s has %d warnings, %v", h.HostName, h.IP, len(*warnings), *warnings)
			return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
		}

		log.Infof("check host %s %s has %d warnings and %d pass", h.HostName, h.IP, len(*warnings), len(*pass))
	}

	return nil
}

// template info to parse em cluster scale out yaml template file
type templateScaleOut struct {
	HostIPs []string
}

func (p *templateScaleOut) generateTopologyConfig(ctx context.Context) (string, error) {
	t, err := template.New("import_topology.yaml").Parse(resourceTemplate.EMClusterScaleOut)
	if err != nil {
		return "", errors.NewError(errors.TIEM_PARAMETER_INVALID, err.Error())
	}

	topology := new(bytes.Buffer)
	if err = t.Execute(topology, p); err != nil {
		return "", errors.NewError(errors.TIEM_UNRECOGNIZED_ERROR, err.Error())
	}
	framework.LogWithContext(ctx).Infof("generate topology config: %s", topology.String())

	return topology.String(), nil
}

// template info to parse cluster check yaml template file
type templateCheckHost struct {
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
		return "", errors.NewError(errors.TIEM_PARAMETER_INVALID, err.Error())
	}

	topology := new(bytes.Buffer)
	if err = t.Execute(topology, p); err != nil {
		return "", errors.NewError(errors.TIEM_UNRECOGNIZED_ERROR, err.Error())
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

// hostCheckResult represents the check result of each node
type checkHostResult struct {
	Node    string `json:"node"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
