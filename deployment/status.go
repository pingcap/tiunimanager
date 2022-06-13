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
 * @File: status
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/20
*******************************************************************************/

package deployment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/pingcap/tiunimanager/util/disk"
)

type Status string

const (
	Init       Status = "init"
	Processing Status = "processing"
	Finished   Status = "finished"
	Error      Status = "error"
)

// Operation Record information about each TiUP operation
type Operation struct {
	Type       string `json:"type"`         // operation of type, eg: deploy, start, stop...
	Operation  string `json:"operation"`    // full amount of Command record, which can be used for playback after subsequent system reboots.
	WorkFlowID string `json:"work_flow_id"` // workflow ID which operation belongs to
	Status     Status `json:"status"`       // operation status
	Result     string `json:"result"`       // operation error message
	ErrorStr   string `json:"error_str"`
}

// Create an operation record
func Create(tiUPHome string, op Operation) (fileName string, err error) {
	b, err := json.Marshal(op)
	if err != nil {
		return "", err
	}

	fileName, err = disk.CreateWithContent(fmt.Sprintf("%s/storage", tiUPHome), fmt.Sprintf("operation-%s-%s", time.Now().Format("20060102-150405"), op.Type), "json", b)
	return
}

// Update an operation record
func Update(fileName string, op Operation) error {
	b, err := json.Marshal(op)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, b, 0600)
}

// Read an operation record
func Read(fileName string) (op Operation, err error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &op)
	return
}

// Delete an operation Record
func Delete(filename string) error {
	return os.Remove(filename)
}
