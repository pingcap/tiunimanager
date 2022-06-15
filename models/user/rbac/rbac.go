/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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

package rbac

type RBAC struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"size:128;uniqueIndex:unique_index"`
	V0    string `gorm:"size:128;uniqueIndex:unique_index"`
	V1    string `gorm:"size:128;uniqueIndex:unique_index"`
	V2    string `gorm:"size:128;uniqueIndex:unique_index"`
	V3    string `gorm:"size:128;uniqueIndex:unique_index"`
	V4    string `gorm:"size:128;uniqueIndex:unique_index"`
	V5    string `gorm:"size:128;uniqueIndex:unique_index"`
}
