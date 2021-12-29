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
	ctx "context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/net/context"
)

type Foo1 struct {
	Bar string
}

type Foo2 struct {
	Bar string `yaml:"bar"`
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
	foo2 := Foo2{
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
	foo1 := Foo1{
		"bar",
	}
	SetField(ctx.TODO(), foo1, FieldKey_Yaml, "bar", "baz")
}

func Test_SetField_FieldNotExist(t *testing.T) {
	foo2 := Foo2{
		"bar",
	}
	SetField(ctx.TODO(), &foo2, FieldKey_Yaml, "nobar", "baz")
	if foo2.Bar != "bar" {
		t.Errorf("bar value should not change, expect: bar, actual: %s", foo2.Bar)
	}
}

func Test_SetField_Success(t *testing.T) {
	foo2 := Foo2{
		"bar",
	}
	SetField(ctx.TODO(), &foo2, FieldKey_Yaml, "bar", "newbar")
	if foo2.Bar != "newbar" {
		t.Errorf("fail set, it is %s now", foo2.Bar)
	}
}

func Test_execShowRestoreInfoThruSQL_Fail(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SHOW RESTORES").
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	resp := execShowRestoreInfoThruSQL(context.TODO(), db, "SHOW RESTORES")
	if resp.Destination == "" && resp.ErrorStr != "some error" {
		t.Errorf("case: show restore info. Destination(%s) should have zero value, and ErrorStr(%v) should be 'some error'", resp.Destination, resp.ErrorStr)
	}
}

func Test_execShowRestoreInfoThruSQL_Success(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SHOW RESTORES").
		WillReturnError(fmt.Errorf("sql: no rows in result set"))
	mock.ExpectRollback()

	resp := execShowRestoreInfoThruSQL(context.TODO(), db, "SHOW RESTORES")
	if resp.Progress != 100 && resp.ErrorStr != "" {
		t.Errorf("case: show restore info. Progress(%f) should be 100, and ErrorStr(%v) should have zero value", resp.Progress, resp.ErrorStr)
	}
}

//func Test_setTiUPMirror(t *testing.T) {
//	res, err := setTiUPMirror(context.TODO(), constants.TiUPBinPath, "https://tiup-mirrors.pingcap.com")
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Println(res)
//
//	mirror, err := showTiUPMirror(context.TODO(), constants.TiUPBinPath)
//	if err != nil || mirror != "https://tiup-mirrors.pingcap.com\n" {
//		t.Error(err)
//	}
//
//	res, err = setTiUPMirror(context.TODO(), constants.TiUPBinPath, "http://127.0.0.2:8080/tiup-repo/")
//	if err == nil || res != "" {
//		t.Error(err)
//	}
//
//	mirror, err = showTiUPMirror(context.TODO(), constants.TiUPBinPath)
//	if err != nil || mirror != "https://tiup-mirrors.pingcap.com\n" {
//		t.Error(err)
//	}
//}

func Test_setTiUPMirror(t *testing.T) {
	_, err := setTiUPMirror(context.TODO(), "mock-tiup", "https://tiup-mirrors.pingcap.com")
	if err == nil {
		t.Error("err nil")
	}

	_, err = showTiUPMirror(context.TODO(), "mock-tiup")
	if err == nil {
		t.Error("err nil")
	}
}
