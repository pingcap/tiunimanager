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

package handler

import (
	"context"
	"encoding/json"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Replication struct {
	Replicas             int    `json:"count"`
	EnablePlacementRules string `json:"enable-placement-rules"`
}

func Contain(list interface{}, target interface{}, ) bool {
	if reflect.TypeOf(list).Kind() == reflect.Slice || reflect.TypeOf(list).Kind() == reflect.Array {
		listValue := reflect.ValueOf(list)
		for i := 0; i < listValue.Len(); i++ {
			if target == listValue.Index(i).Interface() {
				return true
			}
		}
	}
	if reflect.TypeOf(target).Kind() == reflect.String && reflect.TypeOf(list).Kind() == reflect.String {
		return strings.Contains(list.(string), target.(string))
	}
	return false
}

// ScaleOutPreCheck
// @Description when scale out TiFlash, check placement rules
// @Parameter	cluster meta
// @Parameter	computes
// @Return		error
func ScaleOutPreCheck(ctx context.Context, meta *ClusterMeta, computes []structs.ClusterResourceParameterCompute) error {
	if len(computes) <= 0 || meta == nil {
		return framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "cluster resource parameter is empty!")
	}

	for _, component := range computes {
		if component.Type == string(constants.ComponentIDTiFlash) {
			var pdID string
			for componentType, instances := range meta.Instances {
				if componentType == string(constants.ComponentIDPD) {
					pdID = strings.Join([]string{instances[0].HostIP[0],
						strconv.Itoa(int(instances[0].Ports[0]))}, ":")
					break
				}
			}

			config, err := secondparty.Manager.ClusterComponentCtl(ctx, secondparty.CTLComponentTypeStr,
				meta.Cluster.Version, spec.ComponentPD, []string{"-u", pdID, "config", "show", "replication"})
			if err != nil {
				return err
			}
			replication := &Replication{}
			if err = json.Unmarshal([]byte(config), replication); err != nil {
				return framework.WrapError(common.TIEM_UNMARSHAL_ERROR, "", err)
			}
			if replication.EnablePlacementRules == "false" {
				return framework.NewTiEMError(common.TIEM_CHECK_PLACEMENT_RULES_ERROR,
					"enable-placement-rules is false, can not scale out TiFlash")
			}
			break
		}
	}

	return nil
}

// ScaleInPreCheck
// @Description When scale in TiFlash, ensure the number of remaining TiFlash instances is
//				greater than or equal to the maximum number of copies of all data tables
// @Parameter	cluster meta
// @Parameter	instance which will be deleted
// @Return		error
func ScaleInPreCheck(ctx context.Context, meta *ClusterMeta, instance *management.ClusterInstance) error {
	if meta == nil || instance == nil {
		return framework.NewTiEMError(common.TIEM_PARAMETER_INVALID, "parameter is nil!")
	}

	if instance.Type == string(constants.ComponentIDTiFlash) {
		var pdID string
		for componentType, instances := range meta.Instances {
			if componentType == string(constants.ComponentIDPD) {
				pdID = strings.Join([]string{instances[0].HostIP[0],
					strconv.Itoa(int(instances[0].Ports[0]))}, ":")
				break
			}
		}

		config, err := secondparty.Manager.ClusterComponentCtl(ctx, secondparty.CTLComponentTypeStr,
			meta.Cluster.Version, spec.ComponentPD, []string{"-u", pdID, "config", "placement-rules", "show"})
		if err != nil {
			return err
		}
		replications := make([]Replication, 0)
		if err = json.Unmarshal([]byte(config), replications); err != nil {
			return framework.WrapError(common.TIEM_UNMARSHAL_ERROR, "", err)
		}
		maxReplicas := 0
		for _, replica := range replications {
			if replica.Replicas > maxReplicas {
				maxReplicas = replica.Replicas
			}
		}
		if len(meta.Instances[string(constants.ComponentIDTiFlash)])-1 < maxReplicas {
			return framework.NewTiEMError(common.TIEM_CHECK_TIFLASH_MAX_REPLICAS_ERROR,
				"the number of remaining TiFlash instances is less than the maximum copies of data tables")
		}
	}
	return nil
}

func WaitWorkflow(workflowID string, interval time.Duration) error {
	return nil
}
