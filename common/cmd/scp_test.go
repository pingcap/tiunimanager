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
 * @File:  scp_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/7 17:31
 */

package cmd

import (
	"context"
	"github.com/agiledragon/gomonkey"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/pingcap/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"reflect"
	"testing"
)

func Test_sshConnectRetClient_ssh_dial_fail(t *testing.T) {

	s := ScpSessionConfig{
		Sess: SessionConfig{
			User:     "root",
			Password: "pingcap!@#",
			Host:     "172.16.5.189",
			Port:     22,
			Timeout:  0,
		},
		SrcFileName: "./test",
		DstFileName: "./test1",
		Permissions: "0666",
	}

	err := errors.New("ssh connect server fail")
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return nil, err
	})
	defer patch.Reset()

	sess, err1 := s.sshConnectRetClient()

	ast := assert.New(t)
	ast.Equal(err1, err)
	ast.Nil(sess)
}

func Test_ScpFile_sshConnectRetClient_fail(t *testing.T) {

	s := ScpSessionConfig{
		Sess: SessionConfig{
			User:     "root",
			Password: "pingcap!@#",
			Host:     "172.16.5.189",
			Port:     22,
			Timeout:  0,
		},
		SrcFileName: "./test",
		DstFileName: "./test1",
		Permissions: "0666",
	}

	err := errors.New("ssh connect server fail")
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return nil, err
	})
	defer patch.Reset()

	err1 := s.ScpFile(context.Background())

	ast := assert.New(t)
	ast.Equal(err1, err)

}

func Test_ScpFile_NewClientBySSH_fail(t *testing.T) {

	s := ScpSessionConfig{
		Sess: SessionConfig{
			User:     "root",
			Password: "pingcap!@#",
			Host:     "172.16.5.189",
			Port:     22,
			Timeout:  0,
		},
		SrcFileName: "./test",
		DstFileName: "./test1",
		Permissions: "0666",
	}

	c := new(ssh.Client)
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return c, nil
	})
	defer patch.Reset()

	err := errors.New("new client by ssh fail")
	patch1 := gomonkey.ApplyFunc(scp.NewClientBySSH, func(ssh *ssh.Client) (scp.Client, error) {
		return scp.Client{}, err
	})
	defer patch1.Reset()

	err1 := s.ScpFile(context.Background())

	ast := assert.New(t)
	ast.Equal(err1, err)

}

func Test_ScpFile_Connect_fail(t *testing.T) {

	s := ScpSessionConfig{
		Sess: SessionConfig{
			User:     "root",
			Password: "pingcap!@#",
			Host:     "172.16.5.189",
			Port:     22,
			Timeout:  0,
		},
		SrcFileName: "./test",
		DstFileName: "./test1",
		Permissions: "0666",
	}

	c := new(ssh.Client)
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return c, nil
	})
	defer patch.Reset()

	patch1 := gomonkey.ApplyFunc(scp.NewClientBySSH, func(ssh *ssh.Client) (scp.Client, error) {
		return scp.Client{}, nil
	})
	defer patch1.Reset()

	err := errors.New("client server fail")
	patch2 := gomonkey.ApplyMethod(reflect.TypeOf(&scp.Client{}), "Connect",
		func(_ *scp.Client) error {
			return err
		})
	defer patch2.Reset()

	err1 := s.ScpFile(context.Background())

	ast := assert.New(t)
	ast.Equal(err1, err)

}

func Test_ScpFile_Open_fail(t *testing.T) {

	s := ScpSessionConfig{
		Sess: SessionConfig{
			User:     "root",
			Password: "pingcap!@#",
			Host:     "172.16.5.189",
			Port:     22,
			Timeout:  0,
		},
		SrcFileName: "./test",
		DstFileName: "./test1",
		Permissions: "0666",
	}

	c := new(ssh.Client)
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return c, nil
	})
	defer patch.Reset()

	patch1 := gomonkey.ApplyFunc(scp.NewClientBySSH, func(ssh *ssh.Client) (scp.Client, error) {
		return scp.Client{}, nil
	})
	defer patch1.Reset()

	patch2 := gomonkey.ApplyMethod(reflect.TypeOf(&scp.Client{}), "Connect",
		func(_ *scp.Client) error {
			return nil
		})
	defer patch2.Reset()

	err := errors.New("open file fail")
	patch3 := gomonkey.ApplyFunc(os.Open, func(name string) (*os.File, error) {
		return nil, err
	})
	defer patch3.Reset()

	err1 := s.ScpFile(context.Background())

	ast := assert.New(t)
	ast.Equal(err1, err)

}

func Test_ScpFile_CopyFile_fail(t *testing.T) {

	s := ScpSessionConfig{
		Sess: SessionConfig{
			User:     "root",
			Password: "pingcap!@#",
			Host:     "172.16.5.189",
			Port:     22,
			Timeout:  0,
		},
		SrcFileName: "./test",
		DstFileName: "./test1",
		Permissions: "0666",
	}

	c := new(ssh.Client)
	patch := gomonkey.ApplyFunc(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return c, nil
	})
	defer patch.Reset()

	patch1 := gomonkey.ApplyFunc(scp.NewClientBySSH, func(ssh *ssh.Client) (scp.Client, error) {
		return scp.Client{}, nil
	})
	defer patch1.Reset()

	patch2 := gomonkey.ApplyMethod(reflect.TypeOf(&scp.Client{}), "Connect",
		func(_ *scp.Client) error {
			return nil
		})
	defer patch2.Reset()

	patch3 := gomonkey.ApplyFunc(os.Open, func(name string) (*os.File, error) {
		return new(os.File), nil
	})
	defer patch3.Reset()

	err := errors.New("copy file fail")
	patch4 := gomonkey.ApplyMethod(reflect.TypeOf(&scp.Client{}), "CopyFile",
		func(_ *scp.Client, fileReader io.Reader, remotePath string, permissions string) error {
			return err
		})
	defer patch4.Reset()

	err1 := s.ScpFile(context.Background())

	ast := assert.New(t)
	ast.Equal(err1, err)

}
