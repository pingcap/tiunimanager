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
	"reflect"
	"strings"
	"testing"
)

type Foo1 struct {
	Bar                string
}

type Foo2 struct {
	Bar                string `yaml:"bar"`
}

func Test_assert_false(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	assert(false)
}

func Test_assert_true(t *testing.T) {
	assert(true)
}

func Test_myPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	myPanic("panic info")
}

func Test_newTmpFileWithContent(t *testing.T) {
	content := []byte("temp info")
	_, err := newTmpFileWithContent("test-prefix", content)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test_findName_IncorrectFieldKey(t *testing.T) {
	foo2 := Foo2 {
		"bar",
	}
	v := reflect.ValueOf(&foo2).Elem()
	typeField := v.Type().Field(0)
	tag := typeField.Tag
	_, err := findName(tag, FieldKey_Json)
	if err == nil || !strings.Contains(err.Error(), "tag provided does not define") {
		t.Errorf("nil error for incorrectfieldkey or incorrectfieldkey: %v", err)
	}
}

func Test_SetField_NonPointer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	foo1 := Foo1 {
		"bar",
	}
	SetField(foo1, FieldKey_Yaml, "bar", "baz")
}

func Test_SetField_FieldNotExist(t *testing.T) {
	foo2 := Foo2 {
		"bar",
	}
	err := SetField(&foo2, FieldKey_Yaml, "nobar", "baz")
	if err == nil || !strings.Contains(err.Error(), "does not exist within the provided item") {
		t.Errorf("nil error for incorrectfieldkey or not tag exist: %s", err.Error())
	}
}

func Test_SetField_Success(t *testing.T) {
	foo2 := Foo2 {
		"bar",
	}
	err := SetField(&foo2, FieldKey_Yaml, "bar", "newbar")
	if err != nil{
		t.Errorf(err.Error())
	}
	if foo2.Bar != "newbar" {
		t.Errorf("fail set, it is %s now", foo2.Bar)
	}
}