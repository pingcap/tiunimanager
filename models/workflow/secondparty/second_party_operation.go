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
	"time"

	"github.com/pingcap-inc/tiunimanager/util/uuidutil"

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
	OperationStatusInit       OperationStatus = "init"
	OperationStatusProcessing OperationStatus = "processing"
	OperationStatusFinished   OperationStatus = "finished"
	OperationStatusError      OperationStatus = "error"
)

type OperationType string

const (
	OperationTypeClusterDeploy             OperationType = "cluster deploy"
	OperationTypeClusterStart              OperationType = "cluster start"
	OperationTypeClusterDestroy            OperationType = "cluster destroy"
	OperationTypeClusterList               OperationType = "cluster list"
	OperationTypeClusterRestart            OperationType = "cluster restart"
	OperationTypeClusterStop               OperationType = "cluster stop"
	OperationTypeClusterUpgrade            OperationType = "cluster upgrade"
	OperationTypeClusterScaleOut           OperationType = "cluster scale out"
	OperationTypeClusterScaleIn            OperationType = "cluster scale in"
	OperationTypeClusterPrune              OperationType = "cluster prune"
	OperationTypeClusterEditGlobalConfig   OperationType = "cluster edit global config"
	OperationTypeClusterEditInstanceConfig OperationType = "cluster edit instance config"
	OperationTypeClusterReload             OperationType = "cluster reload"
	OperationTypeClusterExec               OperationType = "cluster exec"
	OperationTypeTransfer                  OperationType = "transfer"
	OperationTypeDumpling                  OperationType = "dumpling"
	OperationType_Lightning                OperationType = "lightning"
	OperationTypeBackup                    OperationType = "backup"
	OperationTypeRestore                   OperationType = "restore"
)

const (
	ColumnType           = "type"
	ColumnWorkFlowNodeID = "work_flow_node_id"
)
