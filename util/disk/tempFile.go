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
 * @File: tempFile
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/11
*******************************************************************************/

package disk

import (
	"fmt"
	"io/ioutil"
	"os"
)

// NewTmpFileWithContent
// @Description:
// @Parameter dir
// @Parameter prefix
// @Parameter suffix
// @Parameter content
// @return fileName
// @return err
func NewTmpFileWithContent(dir, prefix, suffix string, content []byte) (fileName string, err error) {
	tmpfile, err := ioutil.TempFile(dir, fmt.Sprintf("%s-*.%s", prefix, suffix))
	if err != nil {
		err = fmt.Errorf("fail to create temp file err: %v", err)
		return "", err
	}
	fileName = tmpfile.Name()
	var ct int
	ct, err = tmpfile.Write(content)
	if err != nil || ct != len(content) {
		tmpfile.Close()
		os.Remove(fileName)
		err = fmt.Errorf("fail to write content to temp file %s, err: %v, length of content: %d, writed: %d", fileName, err, len(content), ct)
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		panic(fmt.Sprintln("fail to close temp file ", fileName))
	}
	return fileName, nil
}

//func Overwrite(fileName string, content []byte) error {
//
//}
//
//func Read(fileName string) (data string, err error) {
//
//}
