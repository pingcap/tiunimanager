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
 * @File: scp.go
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/6
*******************************************************************************/

package scp

import (
	"context"
	"fmt"
	"regexp"
	"time"

	expect "github.com/google/goexpect"
	"github.com/pingcap-inc/tiem/library/framework"
)

// CopySSHID
// @Description: wrap the command 'ssh-copy-id <user>@<ip>' without interaction
// @Parameter ip
// @Parameter user
// @Parameter password
// @return error
func CopySSHID(ctx context.Context, ip string, user string, password string, timeoutS int) (err error) {
	logInFunc := framework.LogWithContext(ctx)
	timeout := time.Duration(timeoutS) * time.Second
	cmd := fmt.Sprintf("ssh-copy-id %s@%s", user, ip)
	logInFunc.Infof("copysshid: %s", cmd)

	e, _, err := expect.Spawn(cmd, timeout, expect.Verbose(true))
	if err != nil {
		logInFunc.Errorf("cmd(%s) spawned return with err: %+v", cmd, err)
		return
	}
	defer e.Close()

	_, _, idx, err := e.ExpectSwitchCase(
		[]expect.Caser{
			// All keys were skipped because they already exist on the remote system
			&expect.Case{R: regexp.MustCompile(`.*already exist.*`), T: expect.OK()},
			// Are you sure you want to continue connecting (yes/no/[fingerprint])?
			&expect.Case{R: regexp.MustCompile(`.*yes/no.*`), S: "yes\n", T: expect.Next(), Rt: 1},
			// \rroot@172.16.6.216's password:
			&expect.Case{R: regexp.MustCompile(`.*password.*`), S: fmt.Sprintf("%s\n", password), T: expect.Next(), Rt: 1},
			// Number of key(s) added:
			&expect.Case{R: regexp.MustCompile(`.*added.*`), T: expect.OK()},
		}, timeout)

	if err != nil {
		logInFunc.Errorf("cmd(%s) expect return with err: %+v", cmd, err)
		return
	}

	if idx == 0 {
		logInFunc.Infof("cmd(%s) already copied\n", cmd)
	} else {
		logInFunc.Infof("cmd(%s) keys successfully added\n", cmd)
	}

	return nil
}
