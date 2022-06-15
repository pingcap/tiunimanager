/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package meta

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/deployment"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/micro-cluster/platform/config"
	"github.com/pingcap/tiunimanager/models"
	workflow "github.com/pingcap/tiunimanager/workflow2"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/models/cluster/management"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

const CheckMaxReplicaCmd = "SELECT MAX(replica_count) as max_replica_count FROM information_schema.tiflash_replica;"
const DefaultTiupTimeOut = 360
const LongTiupTimeOut = 600
const DefaultPDMaxCount = 7
const CheckInstanceStatusTimeout = 30 * 24 * time.Hour
const CheckInstanceStatusInterval = 10 * time.Second
const GetGCLifeTimeCmd = `SELECT VARIABLE_VALUE as gc_life_time FROM mysql.GLOBAL_VARIABLES WHERE VARIABLE_NAME="tidb_gc_life_time";`
const DefaultMaxGCLifeTime = "720h"

type PlacementRules struct {
	EnablePlacementRules string `json:"enable-placement-rules"`
}

type StoreInfos struct {
	Stores []StoreInfo `json:"stores"`
}

type StoreInfo struct {
	Store struct {
		ID        int    `json:"id"`
		Address   string `json:"address"`
		StateName string `json:"state_name"`
	} `json:"store"`
	Status struct {
		RegionCount int `json:"region_count"`
		LeaderCount int `json:"leader_count"`
	} `json:"status"`
}

type StoreStatus string

const (
	StoreUp         StoreStatus = "Up"
	StoreDisconnect StoreStatus = "Disconnect"
	StoreDown       StoreStatus = "Down"
	StoreOffline    StoreStatus = "Offline"
	StoreTombstone  StoreStatus = "Tombstone"
)

func Contain(list interface{}, target interface{}) bool {
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

// CompareTiDBVersion
// @Description compare TiDB versions, such as v5.1.0 and v4.0.0
// @Parameter   v1 and v2
// @Return      if v1 >= v2 return true
// @Return      error
func CompareTiDBVersion(v1, v2 string) (bool, error) {
	v1Nums := strings.Split(v1[1:], ".")
	v2Nums := strings.Split(v2[1:], ".")

	if len(v1Nums) != 3 || len(v2Nums) != 3 {
		return false, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID,
			"TiDB version format is invalid")
	}

	v1MajorVersionNumber, err := strconv.Atoi(v1Nums[0])
	if err != nil {
		return false, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "TiDB version %s format is invalid", v1)
	}
	v2MajorVersionNumber, err := strconv.Atoi(v2Nums[0])
	if err != nil {
		return false, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "TiDB version %s format is invalid", v2)
	}
	v1MinorVersionNumber, err := strconv.Atoi(v1Nums[1])
	if err != nil {
		return false, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "TiDB version %s format is invalid", v1)
	}
	v2MinorVersionNumber, err := strconv.Atoi(v2Nums[1])
	if err != nil {
		return false, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "TiDB version %s format is invalid", v2)
	}
	v1RevisionVersionNumber, err := strconv.Atoi(v1Nums[2])
	if err != nil {
		return false, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "TiDB version %s format is invalid", v1)
	}
	v2RevisionVersionNumber, err := strconv.Atoi(v2Nums[2])
	if err != nil {
		return false, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "TiDB version %s format is invalid", v2)
	}

	if v1MajorVersionNumber > v2MajorVersionNumber {
		return true, nil
	} else if v1MajorVersionNumber < v2MajorVersionNumber {
		return false, nil
	} else {
		if v1MinorVersionNumber > v2MinorVersionNumber {
			return true, nil
		} else if v1MinorVersionNumber < v2MinorVersionNumber {
			return false, nil
		} else {
			if v1RevisionVersionNumber >= v2RevisionVersionNumber {
				return true, nil
			} else {
				return false, nil
			}
		}
	}
}

// ScaleOutPreCheck
// @Description when scale out TiFlash, check placement rules
//				when scale out PD, suggest pd instances 1,3,5,7
// @Parameter	cluster meta
// @Parameter	computes
// @Return		error
func ScaleOutPreCheck(ctx context.Context, meta *ClusterMeta, computes []structs.ClusterResourceParameterCompute) error {
	if len(computes) <= 0 || meta == nil {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "parameter is invalid!")
	}

	tiupHomeForTidb := framework.GetTiupHomePathForTidb()
	for _, component := range computes {
		// Suggest PD instances 1, 3, 5, 7
		if component.Type == string(constants.ComponentIDPD) {
			pdCount := component.Count + len(meta.Instances[component.Type])
			if (pdCount%2 == 0) || pdCount > DefaultPDMaxCount {
				return errors.NewError(errors.TIUNIMANAGER_INVALID_TOPOLOGY, "Suggest PD instances [1, 3, 5, 7]")
			}
		}

		// check placement rules
		if component.Type == string(constants.ComponentIDTiFlash) {
			pdAddress := meta.GetPDClientAddresses()
			if len(pdAddress) <= 0 {
				return errors.NewError(errors.TIUNIMANAGER_PD_NOT_FOUND_ERROR, "cluster not found pd instance")
			}
			pdID := strings.Join([]string{pdAddress[0].IP, strconv.Itoa(pdAddress[0].Port)}, ":")

			config, err := deployment.M.Ctl(ctx, deployment.TiUPComponentTypeCtrl, meta.Cluster.Version, spec.ComponentPD,
				tiupHomeForTidb, []string{"-u", pdID, "config", "show", "replication"}, DefaultTiupTimeOut)
			if err != nil {
				return err
			}
			replication := &PlacementRules{}
			if err = json.Unmarshal([]byte(config), replication); err != nil {
				return errors.WrapError(errors.TIUNIMANAGER_UNMARSHAL_ERROR,
					fmt.Sprintf("parse placement rules error: %s", err.Error()), err)
			}
			if replication.EnablePlacementRules == "false" {
				return errors.NewError(errors.TIUNIMANAGER_CHECK_PLACEMENT_RULES_ERROR,
					"enable-placement-rules is false, can not scale out TiFlash, please check it!")
			}
		}
	}

	return nil
}

func CreateSQLLink(ctx context.Context, meta *ClusterMeta) (*sql.DB, error) {
	if meta == nil {
		return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "parameter is invalid")
	}
	address := meta.GetClusterConnectAddresses()
	if len(address) <= 0 {
		return nil, errors.NewError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, "component TiDB not found!")
	}
	rootUser, _ := meta.GetDBUserNamePassword(ctx, constants.Root)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql",
		rootUser.Name, rootUser.Password.Val, address[0].IP, address[0].Port))
	if err != nil {
		return nil, errors.WrapError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, err.Error(), err)
	}
	return db, nil
}

// ScaleInPreCheck
// @Description When scale in TiFlash, ensure the number of remaining TiFlash instances is
//				greater than or equal to the maximum number of copies of all data tables;
//				When scale in TiKV, ensure the number of remaining TiKV instances is greater than
//				or equal to the copies
// @Parameter	cluster meta
// @Parameter	instance which will be deleted
// @Return		error
func ScaleInPreCheck(ctx context.Context, meta *ClusterMeta, instance *management.ClusterInstance) error {
	if meta == nil || instance == nil {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "parameter is invalid!")
	}

	if meta.IsComponentRequired(ctx, instance.Type) {
		if len(meta.Instances[instance.Type]) <= 1 {
			errMsg := fmt.Sprintf("instance %s is unique in cluster %s, can not delete it", instance.ID, meta.Cluster.ID)
			framework.LogWithContext(ctx).Errorf(errMsg)
			return errors.NewError(errors.TIUNIMANAGER_DELETE_INSTANCE_ERROR, errMsg)
		}
	}

	if instance.Type == string(constants.ComponentIDTiKV) {
		if len(meta.Instances[instance.Type])-1 < meta.Cluster.Copies {
			errMsg := "the number of remaining TiKV instances is less than the copies"
			framework.LogWithContext(ctx).Errorf(errMsg)
			return errors.NewError(errors.TIUNIMANAGER_DELETE_INSTANCE_ERROR, errMsg)
		}
	}

	if instance.Type == string(constants.ComponentIDTiFlash) {
		db, err := CreateSQLLink(ctx, meta)
		if err != nil {
			return errors.WrapError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, err.Error(), err)
		}
		defer db.Close()
		var MaxReplicaCount sql.NullInt64
		err = db.QueryRow(CheckMaxReplicaCmd).Scan(&MaxReplicaCount)
		if err != nil {
			return errors.WrapError(errors.TIUNIMANAGER_SCAN_MAX_REPLICA_COUNT_ERROR, err.Error(), err)
		}
		if MaxReplicaCount.Valid {
			framework.LogWithContext(ctx).Infof("TiFlash max replicas: %d", MaxReplicaCount.Int64)
			if len(meta.Instances[string(constants.ComponentIDTiFlash)])-1 < int(MaxReplicaCount.Int64) {
				return errors.NewError(errors.TIUNIMANAGER_CHECK_TIFLASH_MAX_REPLICAS_ERROR,
					"the number of remaining TiFlash instances is less than the maximum copies of data tables")
			}
		}
	}

	return nil
}

// ClonePreCheck
// When use CDCSyncClone strategy to clone cluster, source cluster must have CDC
func ClonePreCheck(ctx context.Context, sourceMeta *ClusterMeta, meta *ClusterMeta, cloneStrategy string) error {
	if cloneStrategy == string(constants.CDCSyncClone) {
		masters, err := models.GetClusterReaderWriter().GetMasters(ctx, sourceMeta.Cluster.ID)
		if err != nil {
			return err
		}
		if len(masters) > 0 {
			return errors.NewErrorf(errors.TIUNIMANAGER_CLONE_SLAVE_ERROR,
				"cluster %s is slave, which can not be cloned by %s", sourceMeta.Cluster.ID, cloneStrategy)
		}

		cmp, err := CompareTiDBVersion(sourceMeta.Cluster.Version, "v5.2.2")
		if err != nil {
			return err
		}
		if !cmp {
			return errors.NewErrorf(errors.TIUNIMANAGER_CHECK_CLUSTER_VERSION_ERROR,
				"cluster %s version must be greater than or equal to v5.2.2", sourceMeta.Cluster.ID)
		}
		if _, ok := sourceMeta.Instances[string(constants.ComponentIDCDC)]; !ok {
			return errors.NewErrorf(errors.TIUNIMANAGER_CDC_NOT_FOUND,
				"cluster %s not found CDC, which cloned by %s", sourceMeta.Cluster.ID, cloneStrategy)
		}
	}

	if len(meta.Instances[string(constants.ComponentIDTiKV)]) < meta.Cluster.Copies {
		errMsg := "the number of TiKV instances is less than the copies"
		framework.LogWithContext(ctx).Errorf(errMsg)
		return errors.NewError(errors.TIUNIMANAGER_CLONE_TIKV_ERROR, errMsg)
	}
	return nil
}

// WaitWorkflow
// @Description wait workflow done
// @Parameter	workflowID
// @Parameter	timeout
// @Return		error
func WaitWorkflow(ctx context.Context, workflowID string, interval, timeout time.Duration) error {
	index := int(timeout.Seconds() / interval.Seconds())
	ticker := time.NewTicker(interval)
	for range ticker.C {
		response, err := workflow.GetWorkFlowService().DetailWorkFlow(ctx,
			message.QueryWorkFlowDetailReq{WorkFlowID: workflowID})
		if err != nil {
			return err
		}
		if response.Info.Status == constants.WorkFlowStatusFinished {
			framework.LogWithContext(ctx).Infof("workflow %s runs successfully!", workflowID)
			return nil
		} else if response.Info.Status == constants.WorkFlowStatusError {
			framework.LogWithContext(ctx).Errorf("workflow %s runs failed!", workflowID)
			return errors.NewError(errors.TIUNIMANAGER_WORKFLOW_DETAIL_FAILED,
				fmt.Sprintf("wait workflow %s, which runs failed!", workflowID))
		}
		index -= 1
		if index == 0 {
			return errors.NewError(errors.TIUNIMANAGER_WORKFLOW_NODE_POLLING_TIME_OUT,
				fmt.Sprintf("wait workflow %s timeout", workflowID))
		}
	}

	return nil
}

const retainedPortCount = 2

// getRetainedPortRange
// @Description: get retained port range for all clusters
// @Parameter ctx
// @return []int
// @return error
func getRetainedPortRange(ctx context.Context) ([]int, error) {
	configResp, err := config.NewSystemConfigManager().GetSystemConfig(ctx, message.GetSystemConfigReq{
		ConfigKey: constants.ConfigKeyRetainedPortRange,
	})
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get config %s failed, err = %s", constants.ConfigKeyRetainedPortRange, err.Error())
		return nil, err
	}

	if len(configResp.ConfigValue) == 0 {
		err = errors.NewErrorf(errors.TIUNIMANAGER_SYSTEM_MISSING_CONFIG, "missing config %s", constants.ConfigKeyRetainedPortRange)
		framework.LogWithContext(ctx).Error(err)
		return nil, err
	}

	portRange := make([]int, retainedPortCount)
	err = json.Unmarshal([]byte(configResp.ConfigValue), &portRange)
	if err != nil {
		err = errors.NewErrorf(errors.TIUNIMANAGER_SYSTEM_MISSING_CONFIG, "invalid value for config %s, value = %s", constants.ConfigKeyRetainedPortRange, configResp.ConfigValue)
		framework.LogWithContext(ctx).Error(err)
		return nil, err
	}
	return portRange, nil
}

// GetRandomString get random password
func GetRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}
