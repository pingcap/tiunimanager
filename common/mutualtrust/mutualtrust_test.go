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
 * @File:  mutualtrust_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/7 18:04
 */

package mutualtrust

import (
	"github.com/agiledragon/gomonkey"
	"github.com/cosiner/socker"
	"github.com/pingcap/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_Connect_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}
	err := errors.New("connect host fail")
	patch := gomonkey.ApplyFunc(socker.Dial, func(addr string, auth *socker.Auth, gate ...*socker.SSH) (*socker.SSH, error) {
		return nil, err
	})
	defer patch.Reset()
	agent, err1 := h.Connect()
	ast := assert.New(t)
	ast.Equal(err, err1)
	ast.Nil(agent)
}

func Test_CheckKeyExist_id_rsa_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}
	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("run Check if the key exists fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.CheckKeyExit(agent)

	ast := assert.New(t)
	ast.Equal(err, err1)
}

func Test_CheckKeyExist_authorized_keys_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}
	agent := new(socker.SSH)

	outputs := make([]gomonkey.OutputCell, 0)
	v1 := gomonkey.OutputCell{Values: nil, Times: 1}
	v2 := gomonkey.OutputCell{Values: nil, Times: 2}
	outputs = append(outputs, v1, v2)
	patches := gomonkey.ApplyMethodSeq(reflect.TypeOf(agent), "Rcmd",
		outputs)
	defer patches.Reset()

	err := errors.New("run Check if the key exists fail")
	outputs1 := make([]gomonkey.OutputCell, 0)

	a := gomonkey.OutputCell{
		Values: gomonkey.Params{nil},
		Times:  1,
	}

	b := gomonkey.OutputCell{
		Values: gomonkey.Params{err},
		Times:  2,
	}
	outputs1 = append(outputs1, a, b)

	patches1 := gomonkey.ApplyMethodSeq(reflect.TypeOf(agent), "Error",
		outputs1)
	defer patches1.Reset()

	err1 := h.CheckKeyExit(agent)

	ast := assert.New(t)
	ast.Equal(err, err1)
}

func Test_CheckKeyExist_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}
	agent := new(socker.SSH)

	outputs := make([]gomonkey.OutputCell, 0)
	v1 := gomonkey.OutputCell{Values: nil, Times: 1}
	v2 := gomonkey.OutputCell{Values: nil, Times: 2}
	outputs = append(outputs, v1, v2)
	patches := gomonkey.ApplyMethodSeq(reflect.TypeOf(agent), "Rcmd",
		outputs)
	defer patches.Reset()

	outputs1 := make([]gomonkey.OutputCell, 0)

	a := gomonkey.OutputCell{
		Values: gomonkey.Params{nil},
		Times:  1,
	}

	b := gomonkey.OutputCell{
		Values: gomonkey.Params{nil},
		Times:  2,
	}
	outputs1 = append(outputs1, a, b)

	patches1 := gomonkey.ApplyMethodSeq(reflect.TypeOf(agent), "Error",
		outputs1)
	defer patches1.Reset()

	err := h.CheckKeyExit(agent)

	ast := assert.New(t)
	ast.Nil(err)
}

func Test_GetKey_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}
	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("run Check if the key exists fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	ret, err1 := h.GetKey(agent)

	ast := assert.New(t)
	ast.Equal(err, err1)
	ast.Equal(ret, "")

}

func Test_GetKey_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}
	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	ret, err := h.GetKey(agent)

	ast := assert.New(t)
	ast.Nil(err)
	ast.Equal(ret, "")

}

func Test_SelectSecret_Connect_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	err := errors.New("connect host fail")
	patches := gomonkey.ApplyMethod(reflect.TypeOf(h), "Connect",
		func(_ *Host) (*socker.SSH, error) {
			return nil, err
		})
	defer patches.Reset()

	key, err1 := h.SelectSecret()

	ast := assert.New(t)
	ast.Equal("", key)
	ast.Equal(err, err1)
}

func Test_SelectSecret_CheckKeyExit_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(h), "Connect",
		func(_ *Host) (*socker.SSH, error) {
			return agent, nil
		})
	defer patches.Reset()

	err := errors.New("connect host fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(h), "CheckKeyExit",
		func(_ *Host, agent *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	key, err1 := h.SelectSecret()
	ast := assert.New(t)
	ast.Equal(key, "")
	ast.Equal(err, err1)
}

func Test_SelectSecret_GetKey_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(h), "Connect",
		func(_ *Host) (*socker.SSH, error) {
			return agent, nil
		})
	defer patches.Reset()

	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(h), "CheckKeyExit",
		func(_ *Host, agent *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err := errors.New("connect host fail")
	patches2 := gomonkey.ApplyMethod(reflect.TypeOf(h), "GetKey",
		func(_ *Host, agent *socker.SSH) (string, error) {
			return "", err
		})
	defer patches2.Reset()

	key, err1 := h.SelectSecret()
	ast := assert.New(t)
	ast.Equal(key, "")
	ast.Equal(err, err1)
}

func Test_SelectSecret_GetKey_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(h), "Connect",
		func(_ *Host) (*socker.SSH, error) {
			return agent, nil
		})
	defer patches.Reset()

	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(h), "CheckKeyExit",
		func(_ *Host, agent *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	key := "AAAAAAAA"
	patches2 := gomonkey.ApplyMethod(reflect.TypeOf(h), "GetKey",
		func(_ *Host, agent *socker.SSH) (string, error) {
			return key, nil
		})
	defer patches2.Reset()

	key1, err1 := h.SelectSecret()
	ast := assert.New(t)
	ast.Equal(key, key1)
	ast.Nil(err1)
}

func Test_CreateTempAuthFile_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("create temp auth file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.CreateTempAuthFile(agent)

	assert.New(t).Equal(err, err1)
}

func Test_CreateTempAuthFile_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("create temp auth file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.CreateTempAuthFile(agent)

	assert.New(t).Nil(err1)
}

func Test_ClearTempAuthFile_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("clear temp auth file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.ClearTempAuthFile(agent)

	assert.New(t).Equal(err1, err)
}

func Test_ClearTempAuthFile_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("clear temp auth file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.ClearTempAuthFile(agent)

	assert.New(t).Nil(err1)
}

func Test_WriteKeyToTempFile_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("write key to temp file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.WriteKeyToTempFile(agent, "aaaaaa")

	assert.New(t).Equal(err, err1)
}

func Test_WriteKeyToTempFile_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("write key to temp file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.WriteKeyToTempFile(agent, "aaaaaa")

	assert.New(t).Nil(err1)
}

func Test_MergeOriginalKeyFilesAndTmpFile_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("merge original key files and temp file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.MergeOriginalKeyFilesAndTmpFile(agent)

	assert.New(t).Equal(err, err1)
}

func Test_MergeOriginalKeyFilesAndTmpFile_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("merge original key files and temp file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.MergeOriginalKeyFilesAndTmpFile(agent)

	assert.New(t).Nil(err1)
}

func Test_CreateSshConfigFile_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.CreateSshConfigFile(agent)

	assert.New(t).Equal(err1, err)
}

func Test_CreateSshConfigFile_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.CreateSshConfigFile(agent)

	assert.New(t).Nil(err1)
}

func Test_SetNotStrictHostKeyCheck_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.SetNotStrictHostKeyCheck(agent)

	assert.New(t).Equal(err1, err)
}

func Test_SetNotStrictHostKeyCheck_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.SetNotStrictHostKeyCheck(agent)

	assert.New(t).Nil(err1)
}

func Test_ChangConfigFileMode_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.ChangConfigFileMode(agent)

	assert.New(t).Equal(err1, err)
}

func Test_ChangConfigFileMode_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.ChangConfigFileMode(agent)

	assert.New(t).Nil(err1)
}

func Test_WriteKeyBack_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.WriteKeyBack(agent)

	assert.New(t).Equal(err1, err)
}

func Test_WriteKeyBack_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.WriteKeyBack(agent)

	assert.New(t).Nil(err1)
}

func Test_ChangeKeyFileOwner_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.ChangeKeyFileOwner(agent)

	assert.New(t).Equal(err1, err)
}

func Test_ChangeKeyFileOwner_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.ChangeKeyFileOwner(agent)

	assert.New(t).Nil(err1)
}

func Test_ChangeKeyFileMode_fail(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	err := errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return err
		})
	defer patches1.Reset()

	err1 := h.ChangeKeyFileMode(agent)

	assert.New(t).Equal(err1, err)
}

func Test_ChangeKeyFileMode_success(t *testing.T) {
	h := &Host{
		IP:       "172.16.5.190",
		Port:     22,
		User:     "root",
		Password: "pingcap!@#",
	}

	agent := new(socker.SSH)

	patches := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Rcmd",
		func(_ *socker.SSH, cmd string, env ...string) {
			return
		})
	defer patches.Reset()

	//err:=errors.New("create ssh config file fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(agent), "Error",
		func(_ *socker.SSH) error {
			return nil
		})
	defer patches1.Reset()

	err1 := h.ChangeKeyFileMode(agent)

	assert.New(t).Nil(err1)
}
