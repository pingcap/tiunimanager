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
 * @File:  command.go
 * @Version: 1.0.0
 * @Date: 2021/12/3 16:12
 */
package cmd

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	WaitCmdEnd uint16 = iota
	NoWaitCmdEnd
)

type SessionConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Timeout  int
}

func (s *SessionConfig) GetUser() string {
	return s.User
}

func (s *SessionConfig) GetPassword() string {
	return s.Password
}

func (s *SessionConfig) GetHost() string {
	return s.Host
}

func (s *SessionConfig) GetPort() int {
	return s.Port
}

func (s *SessionConfig) GetTimeout() int {
	return s.Timeout
}

func (s *SessionConfig) SetTimeout(timeout int) {
	s.Timeout = timeout
}

func NewSessionConfig(user, password, host string, port, timeout int) SessionConfig {
	return SessionConfig{
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
		Timeout:  timeout,
	}
}

//type execSessionBuilder = func() (ExecSession, error)

type ExecSession interface {
	Exec(cmd string, runType uint16) error
	Close() error
	Output() string
	ErrMsg() string
}

type RemoteSession struct {
	session *ssh.Session
	msgBuf  *bytes.Buffer
	errBuf  *bytes.Buffer
}

func (config SessionConfig) sshConnectRetSess() (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		session      *ssh.Session
		err          error
	)
	//set connect timeout 2s when param timeout  is zero
	if config.GetTimeout() == 0 {
		config.SetTimeout(20000)
	}
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(config.GetPassword()))

	hostKeyCb := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}

	clientConfig = &ssh.ClientConfig{
		User:            config.GetUser(),
		Auth:            auth,
		Timeout:         time.Duration(config.GetTimeout()) * time.Second,
		HostKeyCallback: hostKeyCb,
	}

	addr = fmt.Sprintf("%s:%d", config.GetHost(), config.GetPort())

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}
	return session, nil
}

func NewRemoteSession(config SessionConfig) (ExecSession, error) {
		r := RemoteSession{
			msgBuf: &bytes.Buffer{},
			errBuf: &bytes.Buffer{},
		}
		session, err := config.sshConnectRetSess()
		if err != nil {
			return nil, err
		}
		r.session = session
		r.session.Stdout = r.msgBuf
		r.session.Stderr = r.errBuf
		return &r, nil
}

func (r *RemoteSession) Exec(cmd string, runType uint16) error {
	var err error

	if runType == WaitCmdEnd {
		err = r.session.Run(cmd)
		if err != nil {
			r.errBuf.WriteString(err.Error())
		}
	}

	if runType == NoWaitCmdEnd {
		err = r.session.Start(cmd)
		if err != nil {
			r.errBuf.WriteString(err.Error())
		}
	}
	return err
}

func (r *RemoteSession) Output() string {
	return r.msgBuf.String()
}

func (r *RemoteSession) ErrMsg() string {
	return r.errBuf.String()
}

func (r *RemoteSession) Close() error {
	return r.session.Close()
}
