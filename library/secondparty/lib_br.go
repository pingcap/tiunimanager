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

package secondparty

import (
	_ "github.com/go-sql-driver/mysql"
)

type DbConnParam struct {
	Username string
	Password string
	IP       string
	Port     string
}

type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeS3    StorageType = "s3"
)

type ClusterFacade struct {
	TaskID          uint64 // do not pass this value for br command
	DbConnParameter DbConnParam
	DbName          string
	TableName       string
	ClusterId       string // todo: need to know the usage
	ClusterName     string // todo: need to know the usage
	RateLimitM      string
	Concurrency     string
	CheckSum        string
}

type BrStorage struct {
	StorageType StorageType
	Root        string // "/tmp/backup"
}