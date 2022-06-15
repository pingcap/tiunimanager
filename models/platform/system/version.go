/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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

/*******************************************************************************
 * @File: version.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/16
*******************************************************************************/

package system

type VersionInfo struct {
	ID          string `gorm:"primaryKey;"`
	Desc        string
	ReleaseNote string
}

type VersionInitializer struct {
	VersionID       string
	DataInitializer func() error
}