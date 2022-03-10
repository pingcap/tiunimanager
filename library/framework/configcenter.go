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

package framework

import (
	"errors"
)

type Configuration map[Key]Instance

var LocalConfig map[Key]Instance

type Instance struct {
	Key     Key
	Value   interface{}
	Version int
}

type Key string

const (
	UsingSpecifiedKeyPair Key = "UsingSpecifiedKeyPair"
)

func SetLocalConfig(key Key, value interface{}) {
	instance := CreateInstance(key, value)
	LocalConfig[key] = instance
}

func CreateInstance(key Key, value interface{}) Instance {
	return Instance{
		key, value, 0,
	}
}

func Get(key Key) (interface{}, error) {
	instance, ok := LocalConfig[key]
	if !ok {
		return nil, errors.New("undefined config")
	}

	return instance.Value, nil
}

func GetInstance(key Key) (Instance, error) {
	instance, ok := LocalConfig[key]
	if !ok {
		return instance, errors.New("undefined config")
	}

	return instance, nil
}

func GetWithDefault(key Key, value interface{}) interface{} {
	instance, ok := LocalConfig[key]
	if !ok {
		return value
	}

	return instance.Value
}

func GetStringWithDefault(key Key, value string) string {
	result := GetWithDefault(key, value)
	return result.(string)
}

func GetIntegerWithDefault(key Key, value int) int {
	result := GetWithDefault(key, value)
	return result.(int)
}

func GetBoolWithDefault(key Key, value bool) bool {
	result := GetWithDefault(key, value)
	return result.(bool)
}

func UpdateLocalConfig(key Key, value interface{}, newVersion int) (bool, int) {
	instance, err := GetInstance(key)
	if err != nil {
		Log().Error(err)
		return false, -1
	}
	if newVersion < instance.Version {
		return false, instance.Version
	}
	LocalConfig[key] = Instance{key, value, newVersion}
	return true, newVersion
}

func ModifyLocalServiceConfig(key Key, value interface{}) bool {
	return true
}

func ModifyGlobalConfig(key Key, value interface{}) bool {
	return true
}
