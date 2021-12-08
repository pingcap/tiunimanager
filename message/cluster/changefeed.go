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

package cluster

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"time"
)

type CreateChangeFeedTaskReq struct {
	ChangeFeedTask
}

type CreateChangeFeedTaskResp struct {
	ID string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
}

type QueryChangeFeedTaskReq struct {
	ClusterId string `json:"clusterId" form:"clusterId" example:"CLUSTER_ID_IN_TIEM__22"`
	controller.Page
}

type QueryChangeFeedTaskResp struct {
	ChangeFeedTaskInfo
}

type DetailChangeFeedTaskReq struct {
	ID string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
}

type DetailChangeFeedTaskResp struct {
	ChangeFeedTaskInfo
}

type PauseChangeFeedTaskReq struct {
	ID string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
}

type PauseChangeFeedTaskResp struct {
	Status int `json:"status" form:"status" example:"1" enums:"0,1,2,3,4,5"`
}

type ResumeChangeFeedTaskReq struct {
	Id string `json:"id" form:"id" example:"CLUSTER_ID_IN_TIEM__22"`
}

type ResumeChangeFeedTaskResp struct {
	Status int `json:"status" form:"status" example:"1" enums:"0,1,2,3,4,5"`
}

type UpdateChangeFeedTaskReq struct {
	ChangeFeedTask
}

type UpdateChangeFeedTaskResp struct {
	Status int `json:"status" form:"status" example:"1" enums:"0,1,2,3,4,5"`
}

type DeleteChangeFeedTaskReq struct {
	ID string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
}

type DeleteChangeFeedTaskResp struct {
	ID     string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
	Status int    `json:"status" form:"status" example:"1" enums:"0,1,2,3,4,5"`
}

type ChangeFeedTask struct {
	Id             string      `json:"id" form:"id" example:"CLUSTER_ID_IN_TIEM__22"`
	Name           string      `json:"name" form:"name" example:"my_sync_name"`
	ClusterId      string      `json:"clusterId" form:"clusterId" example:"CLUSTER_ID_IN_TIEM__22"`
	StartTS        int64       `json:"startTS" form:"startTS" example:"415241823337054209"`
	FilterRules    []string    `json:"rules" form:"rules" example:"*.*"`
	Status         int         `json:"status" form:"status" example:"1" enums:"0,1,2,3,4,5"`
	DownstreamType string      `json:"downstreamType"  form:"downstreamType" example:"tidb" enums:"tidb,kafka,mysql"`
	Downstream     interface{} `json:"downstream" form:"downstream"`
	CreateTime     time.Time   `json:"createTime" form:"createTime"`
	UpdateTime     time.Time   `json:"updateTime" form:"updateTime"`
}

type MysqlDownstream struct {
	Ip                string `json:"ip" form:"ip" example:"127.0.0.1"`
	Port              int    `json:"port" form:"port" example:"8001"`
	Username          string `json:"username" form:"username" example:"root"`
	Password          string `json:"password" form:"password" example:"my_password"`
	ConcurrentThreads int    `json:"concurrentThreads" form:"concurrentThreads" example:"7"`
	WorkerCount       int    `json:"workerCount" form:"workerCount" example:"2"`
	MaxTxnRow         int    `json:"maxTxnRow" form:"maxTxnRow" example:"5"`
	Tls               bool   `json:"tls" form:"tls" example:"false"`
}

type KafkaDownstream struct {
	Ip                string       `json:"ip" form:"ip" example:"127.0.0.1"`
	Port              int          `json:"port" form:"port" example:"9001"`
	Version           string       `json:"version" form:"version" example:"2.4.0"`
	ClientId          string       `json:"clientId" form:"clientId" example:"213"`
	TopicName         string       `json:"topicName" form:"topicName" example:"my_topic"`
	Protocol          string       `json:"protocol" form:"protocol" example:"default" enums:"default,canal,avro,maxwell"`
	Partitions        int          `json:"partitions" form:"partitions" example:"1"`
	ReplicationFactor int          `json:"replicationFactor" form:"replicationFactor" example:"1"`
	MaxMessageBytes   int          `json:"maxMessageBytes" form:"maxMessageBytes" example:"16"`
	MaxBatchSize      int          `json:"maxBatchSize" form:"maxBatchSize" example:"5"`
	Dispatchers       []Dispatcher `json:"dispatchers" form:"dispatchers"`
	Tls               bool         `json:"tls" form:"tls" example:"false"`
}

type Dispatcher struct {
	Matcher    string `json:"matcher" form:"matcher" example:"test1.*"`
	Dispatcher string `json:"dispatcher" form:"dispatcher" example:"ts"`
}

type TiDBDownstream struct {
	Ip                string `json:"ip" form:"ip" example:"127.0.0.1"`
	Port              int    `json:"port" form:"port" example:"4534"`
	Username          string `json:"username" form:"username" example:"tidb"`
	Password          string `json:"password" form:"password" example:"my_password"`
	ConcurrentThreads int    `json:"concurrentThreads" form:"concurrentThreads" example:"5"`
	WorkerCount       int    `json:"workerCount" form:"workerCount" example:"2"`
	MaxTxnRow         int    `json:"maxTxnRow" form:"maxTxnRow" example:"4"`
	Tls               bool   `json:"tls" form:"tls" example:"false"`
	TargetClusterId   string `json:"targetClusterId" form:"targetClusterId" example:"CLUSTER_ID_IN_TIEM__22"`
}

type ChangeFeedTaskInfo struct {
	ChangeFeedTask
	UnSteady          bool `json:"unsteady" form:"unsteady" example:"false"`
	UpstreamUpdateTs  uint `json:"upstreamUpdateTs" form:"upstreamUpdateTs" example:"415241823337054209"`
	DownstreamFetchTs uint `json:"downstreamFetchTs" form:"downstreamFetchTs" example:"415241823337054209"`
	DownstreamSyncTs  uint `json:"downstreamSyncTs" form:"downstreamSyncTs" example:"415241823337054209"`
}
