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

package hostinitiator

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	sshclient "github.com/pingcap-inc/tiem/util/ssh"
)

const (
	defaultShell = "/bin/bash"
	// SSH public key file
	publicKeyPath = "~/.ssh/id_rsa.pub"
	// SSH authorized_keys file
	defaultSSHAuthorizedKeys = "~/.ssh/authorized_keys"

	// UserActionAdd add user.
	userActionAdd = "add"
	// UserActionDel delete user.
	userActionDel = "del"

	// TODO: in RHEL/CentOS, the commands are in /usr/sbin, but in some
	// other distros they may be in other location such as /usr/bin, we'll
	// need to check and find the proper path of commands in the future.
	useraddCmd  = "/usr/sbin/useradd"
	userdelCmd  = "/usr/sbin/userdel"
	groupaddCmd = "/usr/sbin/groupadd"
)

type UserModuleConfig struct {
	Action string // add, del
	Name   string // username
	Group  string // group name
	Home   string // home directory of user
	Shell  string // login shell of the user
	Sudoer bool   // when true, the user will be added to sudoers list
}

// NewUserModule builds and returns a UserModule object base on given config.
func (config *UserModuleConfig) buildUserCommand() string {
	cmd := ""

	switch config.Action {
	case userActionAdd:
		cmd = useraddCmd
		// You have to use -m, otherwise no home directory will be created. If you want to specify the path of the home directory, use -d and specify the path
		// useradd -m -d /PATH/TO/FOLDER
		cmd += " -m"
		if config.Home != "" {
			cmd += " -d" + config.Home
		}

		// set user's login shell
		if config.Shell != "" {
			cmd = fmt.Sprintf("%s -s %s", cmd, config.Shell)
		} else {
			cmd = fmt.Sprintf("%s -s %s", cmd, defaultShell)
		}

		// set user's group
		if config.Group == "" {
			config.Group = config.Name
		}

		// groupadd -f <group-name>
		groupAdd := fmt.Sprintf("%s -f %s", groupaddCmd, config.Group)

		// useradd -g <group-name> <user-name>
		cmd = fmt.Sprintf("%s -g %s %s", cmd, config.Group, config.Name)

		// prevent errors when username already in use
		cmd = fmt.Sprintf("id -u %s > /dev/null 2>&1 || (%s && %s)", config.Name, groupAdd, cmd)

		// add user to sudoers list
		if config.Sudoer {
			sudoLine := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL",
				config.Name)
			cmd = fmt.Sprintf("%s && %s",
				cmd,
				fmt.Sprintf("echo '%s' > /etc/sudoers.d/%s", sudoLine, config.Name))
		}

	case userActionDel:
		cmd = fmt.Sprintf("%s -r %s", userdelCmd, config.Name)
		// prevent errors when user does not exist
		cmd = fmt.Sprintf("%s || [ $? -eq 6 ]", cmd)
	}

	return cmd
}

func (p *FileHostInitiator) createDeployUser(ctx context.Context, deployUser, userGroup string, h *structs.HostInfo) error {
	um := UserModuleConfig{
		Action: userActionAdd,
		Name:   deployUser,
		Group:  userGroup,
		Sudoer: true,
	}
	createUserCmd := um.buildUserCommand()
	_, err := p.sshClient.RunCommandsInRemoteHost(h.IP, rp_consts.HostSSHPort, sshclient.Passwd, h.UserName, h.Passwd, rp_consts.DefaultCopySshIDTimeOut, []string{createUserCmd})
	if err != nil {
		errMsg := fmt.Sprintf("create user %s on remote host %s failed by cmd %s, err: %v", deployUser, h.IP, createUserCmd, err)
		return errors.NewError(errors.TIEM_RESOURCE_INIT_DEPLOY_USER_ERROR, errMsg)
	}

	return nil
}

func (p *FileHostInitiator) appendAuthorizedKeysFile(ctx context.Context, deployUser string, h *structs.HostInfo) error {
	pubKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		errMsg := fmt.Sprintf("read user em user public key file %s failed, %v", publicKeyPath, err)
		return errors.NewError(errors.TIEM_RESOURCE_INIT_HOST_AUTH_ERROR, errMsg)
	}

	// Authorize
	cmd := `su - ` + deployUser + ` -c 'mkdir -p ~/.ssh && chmod 700 ~/.ssh'`
	_, err = p.sshClient.RunCommandsInRemoteHost(h.IP, rp_consts.HostSSHPort, sshclient.Passwd, h.UserName, h.Passwd, rp_consts.DefaultCopySshIDTimeOut, []string{cmd})
	if err != nil {
		errMsg := fmt.Sprintf("create '~/.ssh' directory for user '%s' on host %s failed, %v", deployUser, h.IP, err)
		return errors.NewError(errors.TIEM_RESOURCE_INIT_HOST_AUTH_ERROR, errMsg)
	}

	pk := strings.TrimSpace(string(pubKey))
	sshAuthorizedKeys := p.findSSHAuthorizedKeysFile(ctx, h)
	cmd = fmt.Sprintf(`su - %[1]s -c 'grep $(echo %[2]s) %[3]s || echo %[2]s >> %[3]s && chmod 600 %[3]s'`,
		deployUser, pk, sshAuthorizedKeys)
	_, err = p.sshClient.RunCommandsInRemoteHost(h.IP, rp_consts.HostSSHPort, sshclient.Passwd, h.UserName, h.Passwd, rp_consts.DefaultCopySshIDTimeOut, []string{cmd})
	if err != nil {
		errMsg := fmt.Sprintf("write public keys '%s' to host '%s' for user '%s'", sshAuthorizedKeys, h.IP, deployUser)
		return errors.NewError(errors.TIEM_RESOURCE_INIT_HOST_AUTH_ERROR, errMsg)
	}

	return nil
}

func (p *FileHostInitiator) AuthHost(ctx context.Context, deployUser, userGroup string, h *structs.HostInfo) (err error) {
	log := framework.LogWithContext(ctx)
	err = p.createDeployUser(ctx, deployUser, userGroup, h)
	if err != nil {
		log.Errorf("auth host failed, %v", err)
		return err
	}

	err = p.appendAuthorizedKeysFile(ctx, deployUser, h)
	if err != nil {
		log.Errorf("auth host failed after user created, %v", err)
		return err
	}

	return nil
}

// FindSSHAuthorizedKeysFile finds the correct path of SSH authorized keys file
func (p *FileHostInitiator) findSSHAuthorizedKeysFile(ctx context.Context, h *structs.HostInfo) string {
	// detect if custom path of authorized keys file is set
	// NOTE: we do not yet support:
	//   - custom config for user (~/.ssh/config)
	//   - sshd started with custom config (other than /etc/ssh/sshd_config)
	//   - ssh server implementations other than OpenSSH (such as dropbear)
	sshAuthorizedKeys := defaultSSHAuthorizedKeys
	cmd := "grep -Ev '^\\s*#|^\\s*$' /etc/ssh/sshd_config"
	// error ignored as we have default value
	stdout, _ := p.sshClient.RunCommandsInRemoteHost(h.IP, rp_consts.HostSSHPort, sshclient.Passwd, h.UserName, h.Passwd, rp_consts.DefaultCopySshIDTimeOut, []string{cmd})

	for _, line := range strings.Split(string(stdout), "\n") {
		if !strings.Contains(line, "AuthorizedKeysFile") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			sshAuthorizedKeys = fields[1]
			break
		}
	}

	if !strings.HasPrefix(sshAuthorizedKeys, "/") && !strings.HasPrefix(sshAuthorizedKeys, "~") {
		sshAuthorizedKeys = fmt.Sprintf("~/%s", sshAuthorizedKeys)
	}
	return sshAuthorizedKeys
}
