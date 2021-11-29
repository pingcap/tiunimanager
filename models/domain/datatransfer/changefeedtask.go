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

package datatransfer

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/metadb/models"
	"gorm.io/gorm"
	"time"
)

type ChangeFeedTask struct {
	models.Entity
	Name              string         `gorm:"type:varchar(32)"`
	ClusterId         string         `gorm:"not null;type:varchar(22);index"`
	Type              DownstreamType `gorm:"not null;type:varchar(16)"`
	StartTS           int64          `gorm:"column:start_ts"`
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
		b , jsonErr := json.Marshal(t.Downstream)
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

type GormChangeFeedReadWrite struct {
	db *gorm.DB
}

func NewGormChangeFeedReadWrite(d *gorm.DB) *GormChangeFeedReadWrite {
	m := new(GormChangeFeedReadWrite)
	m.SetDb(d)
	return m
}

func (m *GormChangeFeedReadWrite) SetDb(db *gorm.DB) {
	m.db = db
}

func (m *GormChangeFeedReadWrite) Db(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

func (m *GormChangeFeedReadWrite) Create(ctx context.Context, task *ChangeFeedTask) (*ChangeFeedTask, error) {
	task.StatusLock = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	return task, m.Db(ctx).Create(task).Error
}

func (m *GormChangeFeedReadWrite) Delete(ctx context.Context, taskId string) (err error) {
	if "" == taskId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}
	task := &ChangeFeedTask{}

	return m.Db(ctx).First(task, "id = ?", taskId).Delete(task).Error
}

func (m *GormChangeFeedReadWrite) LockStatus(ctx context.Context, taskId string) error {
	if "" == taskId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	task := &ChangeFeedTask{}
	err := m.Db(ctx).First(task, "id = ?", taskId).Error

	if err != nil {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_NOT_FOUND)
	}

	if task.Locked() {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_STATUS_CONFLICT)
	}

	return m.Db(ctx).Model(task).
		Update("status_lock", sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}).Error
}

func (m *GormChangeFeedReadWrite) UnlockStatus(ctx context.Context, taskId string, targetStatus int8) error {
	if "" == taskId {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	task := &ChangeFeedTask{}
	err := m.Db(ctx).First(task, "id = ?", taskId).Error

	if err != nil {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_NOT_FOUND)
	}

	if !task.Locked() {
		return framework.SimpleError(common.TIEM_CHANGE_FEED_LOCK_EXPIRED)
	}

	return m.Db(ctx).Model(task).
		Update("status", targetStatus).
		Update("status_lock", sql.NullTime{
			Time:  time.Now(),
			Valid: false,
		}).Error
}

func (m *GormChangeFeedReadWrite) UpdateConfig(ctx context.Context, updateTemplate *ChangeFeedTask) error {
	if "" == updateTemplate.ID {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	return m.Db(ctx).Omit("status_lock", "status", "cluster_id").
		Save(updateTemplate).Error
}

func (m *GormChangeFeedReadWrite) Get(ctx context.Context, taskId string) (*ChangeFeedTask, error) {
	if "" == taskId {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	task := &ChangeFeedTask{}
	err := m.Db(ctx).First(task, "id = ?", taskId).Error

	if err != nil {
		return nil, framework.SimpleError(common.TIEM_CHANGE_FEED_NOT_FOUND)
	} else {
		return task, nil
	}
}

func (m *GormChangeFeedReadWrite) QueryByClusterId(ctx context.Context, clusterId string, offset int, length int) (tasks []*ChangeFeedTask, total int64, err error) {
	if "" == clusterId {
		return nil, 0, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	tasks = make([]*ChangeFeedTask, length)

	return tasks, total, m.Db(ctx).Table(models.TABLE_NAME_CHANGE_FEED_TASKS).
		Where("cluster_id = ?", clusterId).
		Order("created_at").Offset(offset).Limit(length).Find(&tasks).Count(&total).Error
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
	Ip                string   `json:"ip"`
	Port              int      `json:"port"`
	Version           string   `json:"version"`
	ClientId          string   `json:"clientId"`
	TopicName         string   `json:"topicName"`
	Protocol          string   `json:"protocol"`
	Partitions        int      `json:"partitions"`
	ReplicationFactor int      `json:"replicationFactor"`
	MaxMessageBytes   int      `json:"maxMessageBytes"`
	MaxBatchSize      int      `json:"maxBatchSize"`
	Dispatchers       []Dispatcher `json:"dispatchers"`
	Tls               bool     `json:"tls"`
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
