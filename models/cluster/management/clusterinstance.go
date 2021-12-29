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

package management

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

// ClusterInstance the component instances of cluster
type ClusterInstance struct {
	common.Entity

	Type         string `gorm:"not null;default:null"`
	Version      string `gorm:"not null;default:null"`
	ClusterID    string `gorm:"not null;default:null"`
	Role         string
	DiskType     string
	DiskCapacity int32

	// instance resource info
	CpuCores int8
	Memory   int8
	HostID   string
	Zone     string
	Rack     string
	HostIP   []string `gorm:"-"`
	Ports    []int32  `gorm:"-"`
	DiskID   string
	DiskPath string

	// marshal HostIP, never use
	HostInfo string `gorm:"type:varchar(128)"`
	// marshal PortInfo, never use
	PortInfo string `gorm:"type:varchar(128)"`
}

func (t *ClusterInstance) BeforeSave(tx *gorm.DB) (err error) {
	if t.Status == "" {
		t.Status = string(constants.ClusterInitializing)
	}

	if t.Ports == nil {
		t.Ports = make([]int32, 0)
	}
	p, jsonErr := json.Marshal(t.Ports)
	if jsonErr == nil {
		t.PortInfo = string(p)
	} else {
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, jsonErr.Error())
	}

	if t.HostIP == nil {
		t.HostIP = make([]string, 0)
	}
	h, jsonErr := json.Marshal(t.HostIP)
	if jsonErr == nil {
		t.HostInfo = string(h)
	} else {
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, jsonErr.Error())
	}

	if len(t.ID) == 0 {
		return t.Entity.BeforeCreate(tx)
	}
	return nil
}

func (t *ClusterInstance) AfterFind(tx *gorm.DB) (err error) {
	if len(t.PortInfo) > 0 {
		t.Ports = make([]int32, 0)
		err = json.Unmarshal([]byte(t.PortInfo), &t.Ports)
	}
	if err == nil && len(t.HostInfo) > 0 {
		t.HostIP = make([]string, 0)
		err = json.Unmarshal([]byte(t.HostInfo), &t.HostIP)
	}
	return err
}

func (t *ClusterInstance) GetDeployDir() string {
	return fmt.Sprintf("%s/%s/%s-deploy", t.DiskPath, t.ClusterID, strings.ToLower(t.Type))
}

func (t *ClusterInstance) GetDataDir() string {
	return fmt.Sprintf("%s/%s/%s-data", t.DiskPath, t.ClusterID, strings.ToLower(t.Type))
}

func (t *ClusterInstance) GetLogDir() string {
	return fmt.Sprintf("%s/%s/tidb-log", t.GetDeployDir(), t.ClusterID)
}
