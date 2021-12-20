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

package backuprestore

const (
	contextClusterMetaKey             string = "clusterMeta"
	contextBackupRecordKey            string = "backupRecord"
	contextMaintenanceStatusChangeKey string = "maintenanceStatusChange"
	contextBackupTiupTaskIDKey        string = "backupTaskId"
)

const (
	Sunday    string = "Sunday"
	Monday    string = "Monday"
	Tuesday   string = "Tuesday"
	Wednesday string = "Wednesday"
	Thursday  string = "Thursday"
	Friday    string = "Friday"
	Saturday  string = "Saturday"
)

var WeekDayMap = map[string]int{
	Sunday:    0,
	Monday:    1,
	Tuesday:   2,
	Wednesday: 3,
	Thursday:  4,
	Friday:    5,
	Saturday:  6}

func checkWeekDayValid(day string) bool {
	_, exist := WeekDayMap[day]
	return exist
}
