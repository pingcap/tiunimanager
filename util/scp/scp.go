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

	if CheckCopiedSSHID(ctx, ip, user, password) {
		logInFunc.Infof("cmd(%s) already copied", cmd)
		return nil
	}

	e, _, err := expect.Spawn(cmd, timeout, expect.Verbose(true))
	if err != nil {
		logInFunc.Errorf("cmd(%s) spawned return with err: %+v", cmd, err)
		return
	}
	defer e.Close()

	_, _, err = e.Expect(regexp.MustCompile(".*yes/no.*"), timeout)
	if err != nil {
		logInFunc.Warnf("cmd(%s) expect(yes/no req) return with err: %+v", cmd, err)
	} else {
		err = e.Send("yes\n")
		if err != nil {
			logInFunc.Errorf("cmd(%s) respond yes to (yes/no req) return with err: %+v", cmd, err)
			return
		}
	}

	_, _, err = e.Expect(regexp.MustCompile(".*password.*"), timeout)
	if err != nil {
		logInFunc.Errorf("cmd(%s) expect(password req) return with err: %+v", cmd, err)
		return
	}
	err = e.Send(fmt.Sprintf("%s\n", password))
	if err != nil {
		logInFunc.Errorf("cmd(%s) respond password to (password req) return with err: %+v", cmd, err)
		return
	}

	_, _, err = e.Expect(regexp.MustCompile(".*added.*"), timeout)
	if err != nil {
		logInFunc.Errorf("cmd(%s) expect(keys added) return with err: %+v", cmd, err)
		return
	}

	logInFunc.Infof("cmd(%s) keys successfully added", cmd)
	return nil
}

// CheckCopiedSSHID
// @Description: check 'ssh-copy-id <user>@<ip>' has been done
// @Parameter ip
// @Parameter user
// @Parameter password
// @return bool
func CheckCopiedSSHID(ctx context.Context, ip string, user string, password string) bool {
	logInFunc := framework.LogWithContext(ctx)
	timeout := time.Duration(1) * time.Second
	cmd := fmt.Sprintf("ssh-copy-id %s@%s", user, ip)
	logInFunc.Infof("checkcopiedsshid: %s", cmd)

	e, _, err := expect.Spawn(cmd, timeout, expect.Verbose(true))
	if err != nil {
		logInFunc.Errorf("cmd(%s) spawned return with err: %+v", cmd, err)
		return false
	}
	defer e.Close()

	_, _, err = e.Expect(regexp.MustCompile(".*already exist.*"), timeout)
	if err != nil {
		logInFunc.Infof("cmd(%s) expect(keys added) return with err: %+v, which means not copied yet", cmd, err)
		return false
	} else {
		return true
	}
}
