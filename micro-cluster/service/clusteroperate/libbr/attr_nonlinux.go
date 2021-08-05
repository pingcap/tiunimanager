// +build !linux

package libbr

import "syscall"

func genSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
