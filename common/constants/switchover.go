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

package constants

import (
	"time"
)

const SwitchoverCheckClusterReadWriteHealthRetriesCount = 3
const SwitchoverCheckClusterReadWriteHealthRetryWait = 500 * time.Millisecond

const SwitchoverCheckMasterSlaveMaxLagTime = 30 * time.Second
const SwitchoverCheckMasterSlaveMaxLagTimeRetriesCount = 3
const SwitchoverCheckMasterSlaveMaxLagTimeRetryWait = 3000 * time.Millisecond

const SwitchoverCheckSyncChangeFeedTaskHealthTimeInterval = 500 * time.Millisecond
const SwitchoverCheckSyncChangeFeedTaskHealthRetriesCount = 3
const SwitchoverCheckSyncChangeFeedTaskHealthRetryWait = 500 * time.Millisecond

const SwitchoverCheckSyncChangeFeedTaskCaughtUpRetriesCount = 3
const SwitchoverCheckSyncChangeFeedTaskCaughtUpRetryWait = 1000 * time.Millisecond
const SwitchoverCheckSyncChangeFeedTaskCaughtUpMaxLagTime = 1000 * time.Millisecond

const SwitchoverExclusiveDBNameForReadWriteHealthTest = "em_private_switchover_db"
