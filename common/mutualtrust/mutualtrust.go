/*
 * Copyright (c)  2021 PingCAP, Inc.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/**
 * @Author: guobob
 * @Description:
 * @File:  mutualtrust.go
 * @Version: 1.0.0
 * @Date: 2021/12/6 21:38
 */

package mutualtrust

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosiner/socker"
	//"github.com/pingcap/log"
)

var strictHostKeyChecking = false

type Host struct {
	IP       string
	Port     uint16
	User     string
	Password string
}

func (h *Host) Connect() (*socker.SSH, error) {
	var sshConfig = &socker.Auth{User: h.User, Password: h.Password}

	agent, err := socker.Dial(h.IP+":"+strconv.Itoa(int(h.Port)), sshConfig)
	if err != nil {
		//log.Error(fmt.Sprintf("dial agent failed: %v", err))
		return nil, err
	}
	return agent, nil
}

func (h *Host) CheckKeyExit(agent *socker.SSH) error {
	var err error
	//Check if the key exists on the host, if not it will create
	agent.Rcmd("if [ ! -f ~/.ssh/id_rsa ]; then ssh-keygen -t rsa -P '' -f ~/.ssh/id_rsa;fi")
	if err = agent.Error(); err != nil {
		return err
	}
	agent.Rcmd("if [ ! -f ~/.ssh/authorized_keys ]; then touch ~/.ssh/authorized_keys;fi")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) GetKey(agent *socker.SSH) (string, error) {
	var err error
	agent.Rcmd("cat ~/.ssh/id_rsa.pub")
	if err = agent.Error(); err != nil {
		return "", err
	}
	return string(agent.Output()), nil
}

func (h *Host) SelectSecret() (string, error) {
	agent, err := h.Connect()
	if err != nil {
		return "", err
	}
	defer agent.Close()

	err = h.CheckKeyExit(agent)
	if err != nil {
		return "", err
	}
	key, err := h.GetKey(agent)
	if err != nil {
		return "", err
	}
	return key, nil
}

func (h *Host) CreateTempAuthFile(agent *socker.SSH) error {
	var err error
	agent.Rcmd("if [ ! -f ~/.ssh/auth.tmp ]; then touch ~/.ssh/auth.tmp;fi")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) ClearTempAuthFile(agent *socker.SSH) error {
	var err error
	agent.Rcmd("> ~/.ssh/auth.tmp")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) WriteKeyToTempFile(agent *socker.SSH, key string) error {
	var err error
	str := fmt.Sprintf(`echo "%s" >> ~/.ssh/auth.tmp`, strings.Trim(key, "\n"))
	agent.Rcmd(str)
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) MergeOriginalKeyFilesAndTmpFile(agent *socker.SSH) error {
	var err error
	agent.Rcmd("awk '{print $0}' ~/.ssh/auth.tmp ~/.ssh/authorized_keys|sort | uniq > ~/.ssh/auth.tmp2")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) CreateSshConfigFile(agent *socker.SSH) error {
	var err error
	agent.Rcmd("if [ ! -f ~/.ssh/config ]; then touch ~/.ssh/config;fi")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) SetNotStrictHostKeyCheck(agent *socker.SSH) error {
	var err error
	agent.Rcmd("if ! grep 'StrictHostKeyChecking' ~/.ssh/config >/dev/null; then echo 'StrictHostKeyChecking no' > ~/.ssh/config; else sed -i 's#^StrictHostKeyChecking.*$#StrictHostKeyChecking no#g' ~/.ssh/config;fi")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) ChangConfigFileMode(agent *socker.SSH) error {
	var err error
	agent.Rcmd("chmod 0600 ~/.ssh/config")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) WriteKeyBack(agent *socker.SSH) error {
	var err error
	agent.Rcmd("cat ~/.ssh/auth.tmp2 > ~/.ssh/authorized_keys")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) ChangeKeyFileOwner(agent *socker.SSH) error {
	var err error
	agent.Rcmd("chown " + h.User + " ~/.ssh/authorized_keys")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) ChangeKeyFileMode(agent *socker.SSH) error {
	var err error
	agent.Rcmd("chmod 0600 ~/.ssh/authorized_keys")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) ClearTempFile(agent *socker.SSH) error {
	var err error
	agent.Rcmd("rm -f ~/.ssh/auth.tmp")
	if err = agent.Error(); err != nil {
		return err
	}
	agent.Rcmd("rm -f ~/.ssh/auth.tmp2")
	if err = agent.Error(); err != nil {
		return err
	}
	return nil
}

func (h *Host) SecretWrite(authKey string) error {

	agent, err := h.Connect()
	if err != nil {
		return err
	}

	err = h.CreateTempAuthFile(agent)
	if err != nil {
		return err
	}

	err = h.ClearTempAuthFile(agent)
	if err != nil {
		return err
	}

	err = h.WriteKeyToTempFile(agent, authKey)
	if err != nil {
		return err
	}

	err = h.MergeOriginalKeyFilesAndTmpFile(agent)
	if err != nil {
		return err
	}

	if !strictHostKeyChecking {

		err = h.CreateSshConfigFile(agent)
		if err != nil {
			return err
		}
		err = h.SetNotStrictHostKeyCheck(agent)
		if err != nil {
			return err
		}
		err = h.ChangConfigFileMode(agent)
		if err != nil {
			return err
		}
	}

	err = h.WriteKeyBack(agent)
	if err != nil {
		return err
	}

	err = h.ChangeKeyFileOwner(agent)
	if err != nil {
		return err
	}

	err = h.ChangeKeyFileMode(agent)
	if err != nil {
		return err
	}

	err = h.ClearTempFile(agent)
	if err != nil {
		return err
	}

	agent.Close()
	return nil
}
