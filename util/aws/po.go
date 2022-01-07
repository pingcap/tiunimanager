/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

/*******************************************************************************
 * @File: InstancesInfo
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/7 16:23
*******************************************************************************/

package aws

type State string

const (
	running State = "running"
)

type InstancesInfo struct {
	Reservations []Reservation `json:"Reservations"`
}

type Reservation struct {
	Instances []Instance `json:"Instances"`
}

type Instance struct {
	ImageId          string        `json:"ImageId"`
	InstanceId       string        `json:"InstanceId"`
	InstanceType     string        `json:"InstanceType"`
	KeyName          string        `json:"KeyName"`
	PrivateIpAddress string        `json:"PrivateIpAddress"`
	State            InstanceState `json:"State"`
}

type InstanceState struct {
	Code int    `json:"Code"`
	Name string `json:"Name"`
}
