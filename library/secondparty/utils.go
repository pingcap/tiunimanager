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
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"io/ioutil"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

type FieldKey string

const (
	FieldKey_Yaml FieldKey = "yaml"
	FieldKey_Json FieldKey = "json"
)

func assert(b bool) {
	if b {
	} else {
		framework.Log().Error("unexpected panic with stack trace:", string(debug.Stack()))
		panic("unexpected")
	}
}

func myPanic(v interface{}) {
	s := fmt.Sprint(v)
	framework.Log().Errorf("panic: %s, with stack trace: %s", s, string(debug.Stack()))
	panic("unexpected" + s)
}

func newTmpFileWithContent(filePrefix string, content []byte) (fileName string, err error) {
	tmpfile, err := ioutil.TempFile("", fmt.Sprintf("%s-*.yaml", filePrefix))
	if err != nil {
		err = fmt.Errorf("fail to create temp file err: %s", err)
		return "", err
	}
	fileName = tmpfile.Name()
	var ct int
	ct, err = tmpfile.Write(content)
	if err != nil || ct != len(content) {
		tmpfile.Close()
		os.Remove(fileName)
		err = fmt.Errorf(fmt.Sprint("fail to write content to temp file ", fileName, "err:", err, "length of content:", "writed:", ct))
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		myPanic(fmt.Sprintln("fail to close temp file ", fileName))
	}
	return fileName, nil
}

func jsonMustMarshal(v interface{}) []byte {
	bs, err := json.Marshal(v)
	assert(err == nil)
	return bs
}

func SetField(ctx context.Context, item interface{}, fieldKey FieldKey, fieldName string, value interface{}) {
	v := reflect.ValueOf(item).Elem()

	// key: fieldName, value: index of fieldName in struct
	fieldNames := map[string]int{}
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		tag := typeField.Tag
		fname, _ := findName(tag, fieldKey)
		fieldNames[fname] = i
	}

	fieldNum, ok := fieldNames[fieldName]
	if !ok {
		framework.LogWithContext(ctx).Infof("field %s does not exist within the provided item", fieldName)
		return
	}
	fieldVal := v.Field(fieldNum)
	fieldVal.Set(reflect.ValueOf(value))
}

// It's possible we can cache this, which is why precompute all these ahead of time.
func findName(t reflect.StructTag, fieldKey FieldKey) (string, error) {
	if yt, ok := t.Lookup(string(fieldKey)); ok {
		return strings.Split(yt, ",")[0], nil
	}
	return "", fmt.Errorf("tag provided does not define a tag %s", string(fieldKey))
}
