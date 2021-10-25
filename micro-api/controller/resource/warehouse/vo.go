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
 *                                                                            *
 ******************************************************************************/

package warehouse

type ZoneHostStock struct {
	ZoneBaseInfo
	SpecBaseInfo
	Count int
}

type RegionBaseInfo struct {
	RegionCode string `json:"regionCode"`
	RegionName string `json:"regionName"`
}
type ZoneBaseInfo struct {
	ZoneCode string `json:"zoneCode"`
	ZoneName string `json:"zoneName"`
}

type RackBaseInfo struct {
	RackCode string `json:"rackCode"`
	RackName string `json:"rackName"`
}

type SpecBaseInfo struct {
	SpecCode string `json:"specCode"`
	SpecName string `json:"specName"`
}

type SpecStock struct {
	SpecBaseInfo
	Count int `json:"count"`
}

type RackStock struct {
	RackBaseInfo
	SpecStock []SpecStock `json:"specStock"`
}

type ZoneStock struct {
	ZoneBaseInfo
	RackStock []RackStock `json:"rackStock"`
}

type RegionStock struct {
	RegionBaseInfo
	ZoneStock []ZoneStock `json:"zoneStock"`
}
