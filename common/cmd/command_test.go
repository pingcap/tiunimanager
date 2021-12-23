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
 ******************************************************************************/

/**
 * @Author: guobob
 * @Description:
 * @File:  cmmand_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/7 15:53
 */

package cmd

import (
	"bytes"
	"github.com/agiledragon/gomonkey"
	"github.com/pingcap/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
	"reflect"
	"testing"
)

func Test_NewSessionConfig(t *testing.T) {
	user := "root"
	password := "pingcap!@#"
	host := "172.16.5.189"
	port := 22
	timeout := 0
	s := NewSessionConfig(user, password, host, port, timeout)

	ast := assert.New(t)

	ast.Equal(user, s.User)
	ast.Equal(password, s.Password)
	ast.Equal(host, s.Host)
	ast.Equal(port, s.Port)
	ast.Equal(timeout, s.Timeout)
}

func Test_sshConnectRetSess_ssh_dial_fail(t *testing.T) {
	user := "root"
	password := "pingcap!@#"
	host := "172.16.5.189"
	port := 22
	timeout := 0
	s := NewSessionConfig(user, password, host, port, timeout)

	err := errors.New("ssh connect server fail")
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return nil, err
	})
	defer patch.Reset()

	sess, err1 := s.sshConnectRetSess()

	ast := assert.New(t)
	ast.Equal(err1, err)
	ast.Nil(sess)
}

func Test_sshConnectRetSess_NewSession_fail(t *testing.T) {
	user := "root"
	password := "pingcap!@#"
	host := "172.16.5.189"
	port := 22
	timeout := 0
	s := NewSessionConfig(user, password, host, port, timeout)

	client := new(ssh.Client)
	err := errors.New("new connect session fail")
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return client, nil
	})
	defer patch.Reset()

	patches := gomonkey.ApplyMethod(reflect.TypeOf(client), "NewSession",
		func(_ *ssh.Client) (*ssh.Session, error) {
			return nil, err
		})
	defer patches.Reset()

	sess, err1 := s.sshConnectRetSess()

	ast := assert.New(t)
	ast.Equal(err1, err)
	ast.Nil(sess)
}

func Test_sshConnectRetSess_success(t *testing.T) {
	user := "root"
	password := "pingcap!@#"
	host := "172.16.5.189"
	port := 22
	timeout := 0
	s := NewSessionConfig(user, password, host, port, timeout)

	client := new(ssh.Client)
	sess1 := new(ssh.Session)
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return client, nil
	})
	defer patch.Reset()

	patches := gomonkey.ApplyMethod(reflect.TypeOf(client), "NewSession",
		func(_ *ssh.Client) (*ssh.Session, error) {
			return sess1, nil
		})
	defer patches.Reset()

	sess, err := s.sshConnectRetSess()

	ast := assert.New(t)
	ast.Equal(sess1, sess)
	ast.Nil(err)
}

func Test_NewRemoteSession_sshConnectRetSess_fail(t *testing.T) {

	User := "root"
	Password := "pingcap!@#"
	Host := "172.16.5.189"
	Port := 22
	Timeout := 0
	config := NewSessionConfig(User, Password, Host, Port, Timeout)

	client := new(ssh.Client)
	err := errors.New("new connect session fail")
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return client, nil
	})
	defer patch.Reset()

	patches := gomonkey.ApplyMethod(reflect.TypeOf(client), "NewSession",
		func(_ *ssh.Client) (*ssh.Session, error) {
			return nil, err
		})
	defer patches.Reset()

	r, err1 := NewRemoteSession(config)

	ast := assert.New(t)
	ast.Nil(r)
	ast.Equal(err, err1)

}

func Test_NewRemoteSession_success(t *testing.T) {

	User := "root"
	Password := "pingcap!@#"
	Host := "172.16.5.189"
	Port := 22
	Timeout := 0
	config := NewSessionConfig(User, Password, Host, Port, Timeout)

	client := new(ssh.Client)
	sess1 := new(ssh.Session)
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return client, nil
	})
	defer patch.Reset()

	patches := gomonkey.ApplyMethod(reflect.TypeOf(client), "NewSession",
		func(_ *ssh.Client) (*ssh.Session, error) {
			return sess1, nil
		})
	defer patches.Reset()

	r, err := NewRemoteSession(config)

	ast := assert.New(t)

	ast.Nil(err)
	ast.NotNil(r)
}

func Test_RemoteSession_Exec_WaitCmdEnd_fail(t *testing.T) {
	r := &RemoteSession{
		session: new(ssh.Session),
		msgBuf:  &bytes.Buffer{},
		errBuf:  &bytes.Buffer{},
	}

	err := errors.New("session run fail")

	patches := gomonkey.ApplyMethod(reflect.TypeOf(r.session), "Run",
		func(_ *ssh.Session, cmd string) error {
			return err
		})
	defer patches.Reset()

	err1 := r.Exec("ls -al", WaitCmdEnd)

	assert.New(t).Equal(err, err1)

}

func Test_RemoteSession_Exec_WaitCmdEnd_success(t *testing.T) {
	r := &RemoteSession{
		session: new(ssh.Session),
		msgBuf:  &bytes.Buffer{},
		errBuf:  &bytes.Buffer{},
	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(r.session), "Run",
		func(_ *ssh.Session, cmd string) error {
			return nil
		})
	defer patches.Reset()

	err := r.Exec("ls -al", WaitCmdEnd)

	assert.New(t).Nil(err)

}

func Test_RemoteSession_Exec_NoWaitCmdEnd_fail(t *testing.T) {
	r := &RemoteSession{
		session: new(ssh.Session),
		msgBuf:  &bytes.Buffer{},
		errBuf:  &bytes.Buffer{},
	}

	err := errors.New("session run fail")

	patches := gomonkey.ApplyMethod(reflect.TypeOf(r.session), "Start",
		func(_ *ssh.Session, cmd string) error {
			return err
		})
	defer patches.Reset()

	err1 := r.Exec("ls -al", NoWaitCmdEnd)

	assert.New(t).Equal(err, err1)

}

func Test_RemoteSession_Exec_NoWaitCmdEnd_success(t *testing.T) {
	r := &RemoteSession{
		session: new(ssh.Session),
		msgBuf:  &bytes.Buffer{},
		errBuf:  &bytes.Buffer{},
	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(r.session), "Start",
		func(_ *ssh.Session, cmd string) error {
			return nil
		})
	defer patches.Reset()

	err := r.Exec("ls -al", NoWaitCmdEnd)

	assert.New(t).Nil(err)

}

func Test_RemoteSession_Output(t *testing.T) {
	o := "cmd exec success"
	w := &bytes.Buffer{}
	w.WriteString(o)
	r := &RemoteSession{
		session: new(ssh.Session),
		msgBuf:  w,
		errBuf:  &bytes.Buffer{},
	}

	output := r.Output()

	assert.New(t).Equal(o, output)
}

func Test_RemoteSession_ErrMsg(t *testing.T) {
	o := "cmd exec success"
	w := &bytes.Buffer{}
	w.WriteString(o)
	r := &RemoteSession{
		session: new(ssh.Session),
		msgBuf:  &bytes.Buffer{},
		errBuf:  w,
	}

	output := r.ErrMsg()

	assert.New(t).Equal(o, output)
}

func Test_RemoteSession_Close(t *testing.T) {

	r := &RemoteSession{
		session: new(ssh.Session),
		msgBuf:  &bytes.Buffer{},
		errBuf:  &bytes.Buffer{},
	}
	patches := gomonkey.ApplyMethod(reflect.TypeOf(r.session), "Close",
		func(_ *ssh.Session) error {
			return nil
		})
	defer patches.Reset()

	err := r.Close()

	assert.New(t).Nil(err)
}
