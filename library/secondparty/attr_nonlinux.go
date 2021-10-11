// +build !linux

package secondparty

import "syscall"

func genSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
