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
 * @File:  scp.go
 * @Version: 1.0.0
 * @Date: 2021/12/7 10:00
 */

package cmd

import (
	"context"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/pingcap-inc/tiem/library/framework"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"time"
)

type ScpSessionConfig struct {
	Sess SessionConfig
	//file name with file path
	SrcFileName string
	//file name with file path
	DstFileName string
	Permissions string
}

func (s ScpSessionConfig) sshConnectRetClient() (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		err          error
	)
	//set connect timeout 2s when param timeout  is zero
	if s.Sess.Timeout == 0 {
		s.Sess.Timeout = 20000
	}
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(s.Sess.Password))

	hostKeyCb := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}

	clientConfig = &ssh.ClientConfig{
		User:            s.Sess.User,
		Auth:            auth,
		Timeout:         time.Duration(s.Sess.Timeout) * time.Second,
		HostKeyCallback: hostKeyCb,
	}

	addr = fmt.Sprintf("%s:%d", s.Sess.Host, s.Sess.Port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *ScpSessionConfig) ScpFile(ctx context.Context) error {

	sess, err := s.sshConnectRetClient()
	if err != nil {
		return err
	}

	// Create a new SCP client, note that this function might
	// return an error, as a new SSH session is established using the existing connection
	client, err := scp.NewClientBySSH(sess)
	if err != nil {
		return err
	}
	// Connect to the remote server
	err = client.Connect()
	if err != nil {
		return err
	}

	// Close client connection after the file has been copied
	defer client.Close()

	// Open a file
	f, err := os.Open(s.SrcFileName)
	if err != nil {
		return err
	}

	// Close the file after it has been copied
	defer func() {
		err = f.Close()
		if err != nil {
			framework.LogWithContext(ctx).Errorf("close file %s fail , %v", s.SrcFileName, err.Error())
		}
	}()

	// Finally, copy the file over
	// Usage: CopyFile(fileReader, remotePath, permission)

	err = client.CopyFile(f, s.DstFileName, s.Permissions)
	if err != nil {
		return err
	}

	return nil
}
