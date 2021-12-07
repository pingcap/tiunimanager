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

package changefeed

import (
	"database/sql"
	"encoding/json"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
	"time"
)

type Status string

const (
	Initial  Status = "Initial"
	Normal   Status = "Normal"
	Stopped  Status = "Stopped"
	Finished Status = "Finished"
	Error    Status = "Error"
	Failed   Status = "Failed"
	Unknown  Status = "Unknown"
)

func (s Status) IsFinal() bool {
	return Finished == s || Failed == s
}

func (s Status) ToString() string {
	return string(s)
}

func invalidStatus(s string) bool {
	return Initial.ToString() == s ||
		Normal.ToString() == s ||
		Stopped.ToString() == s ||
		Finished.ToString() == s ||
		Error.ToString() == s ||
		Failed.ToString() == s
}

func ConvertStatus(s string) (status Status, err error) {
	if invalidStatus(s) {
		return Status(s), nil
	} else {
		return Unknown, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}
}

type ChangeFeedTask struct {
	dbCommon.Entity
	TaskStatus        Status         `gorm:"-"`
	Name              string         `gorm:"type:varchar(32)"`
	ClusterId         string         `gorm:"not null;type:varchar(22);index"`
	Type              DownstreamType `gorm:"not null;type:varchar(16)"`
	StartTS           int64          `gorm:"column:start_ts"`
	TargetTS          int64          `gorm:"column:target_ts"`
	FilterRulesConfig string         `gorm:"type:text"`
	Downstream        interface{}    `gorm:"-"`
	DownstreamConfig  string         `gorm:"type:text"`
	StatusLock        sql.NullTime   `gorm:"column:status_lock"`
}

func (t ChangeFeedTask) GetStatusLock() sql.NullTime {
	return t.StatusLock
}

type DownstreamType string

const (
	DownstreamTypeTiDB  DownstreamType = "tidb"
	DownstreamTypeKafka DownstreamType = "kafka"
	DownstreamTypeMysql DownstreamType = "mysql"
)

func (dt DownstreamType) unmarshal(cc string) (interface{}, error) {
	switch dt {
	case DownstreamTypeTiDB:
		downstream := &TiDBDownstream{}
		err := json.Unmarshal([]byte(cc), downstream)
		return downstream, err
	case DownstreamTypeKafka:
		downstream := &KafkaDownstream{}
		err := json.Unmarshal([]byte(cc), downstream)
		return downstream, err
	case DownstreamTypeMysql:
		downstream := &MysqlDownstream{}
		err := json.Unmarshal([]byte(cc), downstream)
		return downstream, err
	}
	return nil, framework.SimpleError(common.TIEM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM)
}

func (t *ChangeFeedTask) BeforeSave(tx *gorm.DB) (err error) {
	if t.Downstream != nil {
		b, jsonErr := json.Marshal(t.Downstream)
		if jsonErr == nil {
			t.DownstreamConfig = string(b)
		} else {
			return framework.NewTiEMErrorf(common.TIEM_PARAMETER_INVALID, jsonErr.Error())
		}
	}

	if len(t.ID) == 0 {
		return t.Entity.BeforeCreate(tx)
	}
	return nil
}

func (t *ChangeFeedTask) AfterFind(tx *gorm.DB) (err error) {
	if len(t.DownstreamConfig) > 0 {
		downstream, err := t.Type.unmarshal(t.DownstreamConfig)
		if err != nil {
			return err
		}
		t.Downstream = downstream
	}
	return nil
}

func (t *ChangeFeedTask) Locked() bool {
	return t.StatusLock.Valid &&
		t.StatusLock.Time.Add(time.Minute).After(time.Now())
}

type MysqlDownstream struct {
	Ip                string `json:"ip"`
	Port              int    `json:"port"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	ConcurrentThreads int    `json:"concurrentThreads"`
	WorkerCount       int    `json:"workerCount"`
	MaxTxnRow         int    `json:"maxTxnRow"`
	Tls               bool   `json:"tls"`
}

type KafkaDownstream struct {
	Ip                string       `json:"ip"`
	Port              int          `json:"port"`
	Version           string       `json:"version"`
	ClientId          string       `json:"clientId"`
	TopicName         string       `json:"topicName"`
	Protocol          string       `json:"protocol"`
	Partitions        int          `json:"partitions"`
	ReplicationFactor int          `json:"replicationFactor"`
	MaxMessageBytes   int          `json:"maxMessageBytes"`
	MaxBatchSize      int          `json:"maxBatchSize"`
	Dispatchers       []Dispatcher `json:"dispatchers"`
	Tls               bool         `json:"tls"`
}

type TiDBDownstream struct {
	Ip                string `json:"ip"`
	Port              int    `json:"port"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	ConcurrentThreads int    `json:"concurrentThreads"`
	WorkerCount       int    `json:"workerCount"`
	MaxTxnRow         int    `json:"maxTxnRow"`
	Tls               bool   `json:"tls"`
	TargetClusterId   string `json:"targetClusterId"`
}

type Dispatcher struct {
	Matcher    string `json:"matcher"`
	Dispatcher string `json:"dispatcher"`
}
