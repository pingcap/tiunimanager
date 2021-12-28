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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
	"time"
)

type ChangeFeedTask struct {
	dbCommon.Entity
	TaskStatus        constants.ChangeFeedStatus `gorm:"-"`
	Name              string                     `gorm:"type:varchar(32)"`
	ClusterId         string                     `gorm:"not null;type:varchar(22);index"`
	Type              constants.DownstreamType   `gorm:"not null;type:varchar(16)"`
	StartTS           int64                      `gorm:"column:start_ts"`
	TargetTS          int64                      `gorm:"column:target_ts"`
	FilterRules       []string                   `gorm:"-"`
	FilterRulesConfig string                     `gorm:"type:text"`
	Downstream        ChangeFeedDownStream       `gorm:"-"`
	DownstreamConfig  string                     `gorm:"type:text"`
	StatusLock        sql.NullTime               `gorm:"column:status_lock"`
}

func (t *ChangeFeedTask) GetStatusLock() sql.NullTime {
	return t.StatusLock
}

func UnmarshalDownstream(dt constants.DownstreamType, cc string) (ChangeFeedDownStream, error) {
	switch dt {
	case constants.DownstreamTypeTiDB:
		downstream := &TiDBDownstream{}
		err := json.Unmarshal([]byte(cc), downstream)
		return downstream, err
	case constants.DownstreamTypeKafka:
		downstream := &KafkaDownstream{}
		err := json.Unmarshal([]byte(cc), downstream)
		return downstream, err
	case constants.DownstreamTypeMysql:
		downstream := &MysqlDownstream{}
		err := json.Unmarshal([]byte(cc), downstream)
		return downstream, err
	}
	return nil, errors.NewError(errors.TIEM_CHANGE_FEED_UNSUPPORTED_DOWNSTREAM, "")
}

func (t *ChangeFeedTask) BeforeSave(tx *gorm.DB) (err error) {
	if t.Downstream != nil {
		b, jsonErr := json.Marshal(t.Downstream)
		if jsonErr == nil {
			t.DownstreamConfig = string(b)
		} else {
			return errors.NewError(errors.TIEM_PARAMETER_INVALID, jsonErr.Error())
		}
	}
	if t.FilterRules == nil {
		t.FilterRules = make([]string, 0)
	}

	b, jsonErr := json.Marshal(t.FilterRules)
	if jsonErr == nil {
		t.FilterRulesConfig = string(b)
	} else {
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, jsonErr.Error())
	}

	if len(t.ID) == 0 {
		return t.Entity.BeforeCreate(tx)
	}
	return nil
}

func (t *ChangeFeedTask) AfterFind(tx *gorm.DB) (err error) {
	if len(t.DownstreamConfig) > 0 {
		downstream, err := UnmarshalDownstream(t.Type, t.DownstreamConfig)
		if err != nil {
			return err
		}
		t.Downstream = downstream
	}
	if len(t.FilterRulesConfig) > 0 {
		t.FilterRules = make([]string, 0)
		json.Unmarshal([]byte(t.FilterRulesConfig), &t.FilterRules)
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

type ChangeFeedDownStream interface {
	GetSinkURI() string
}

// GetSinkURI todo
func (p *MysqlDownstream) GetSinkURI() string {
	return "todo"
}

func (p *TiDBDownstream) GetSinkURI() string {
	return "todo"
}

func (p *KafkaDownstream) GetSinkURI() string {
	return "todo"
}