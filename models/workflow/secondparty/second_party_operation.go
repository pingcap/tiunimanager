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

/*******************************************************************************
 * @File: second_party_operation
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/9
*******************************************************************************/

package secondparty

import (
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"time"

	"gorm.io/gorm"
)

// SecondPartyOperation Record information about each second party operation, eg: TiUP, br...
type SecondPartyOperation struct {
	ID             string          `gorm:"primaryKey;"`
	Type           OperationType   `gorm:"not null;comment:'second party operation of type, eg: secondparty cluster deploy,start,stop,...BACKUP, RESTORE as SQL cmd;'"`
	WorkFlowNodeID string          `gorm:"not null;index;comment:'TiEM of the workflow ID'"`
	Status         OperationStatus `gorm:"default:null"`
	Result         string          `gorm:"default:null"`
	ErrorStr       string          `gorm:"size:8192;comment:'second party operation error msg'"`
	CreatedAt      time.Time       `gorm:"<-:create"`
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (s *SecondPartyOperation) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuidutil.GenerateID()
	return nil
}

type OperationStatus string

const (
	OperationStatus_Init       OperationStatus = "init"
	OperationStatus_Processing OperationStatus = "processing"
	OperationStatus_Finished   OperationStatus = "finished"
	OperationStatus_Error      OperationStatus = "error"
)

type OperationType string

const (
	OperationType_ClusterDeploy             OperationType = "cluster deploy"
	OperationType_ClusterStart              OperationType = "cluster start"
	OperationType_ClusterDestroy            OperationType = "cluster destroy"
	OperationType_ClusterList               OperationType = "cluster list"
	OperationType_ClusterRestart            OperationType = "cluster restart"
	OperationType_ClusterStop               OperationType = "cluster stop"
	OperationType_ClusterUpgrade            OperationType = "cluster upgrade"
	OperationType_ClusterScaleOut           OperationType = "cluster scale out"
	OperationType_ClusterScaleIn            OperationType = "cluster scale in"
	OperationType_ClusterPrune              OperationType = "cluster prune"
	OperationType_ClusterEditGlobalConfig   OperationType = "cluster edit global config"
	OperationType_ClusterEditInstanceConfig OperationType = "cluster edit instance config"
	OperationType_ClusterReload             OperationType = "cluster reload"
	OperationType_ClusterExec               OperationType = "cluster exec"
	OperationType_Transfer                  OperationType = "transfer"
	OperationType_Dumpling                  OperationType = "dumpling"
	OperationType_Lightning                 OperationType = "lightning"
	OperationType_Backup                    OperationType = "backup"
	OperationType_Restore                   OperationType = "restore"
)

const (
	Column_Type           = "type"
	Column_WorkFlowNodeID = "work_flow_node_id"
)
