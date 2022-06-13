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
	"strings"

	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	rp_consts "github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/resourcepool/constants"
)

func (p *FileHostInitiator) autoFix(ctx context.Context, h *structs.HostInfo, sortedResult map[string]*[]checkHostResult) (err error) {
	log := framework.LogWithContext(ctx)
	var needInstallNumaCtl = false
	var needSetSwap = false
	var needRemount = false

	// mount point -> adding remount opts
	remount := map[string]map[string]struct{}{}

	fails, hasFails := sortedResult[string(Fail)]
	if hasFails {
		for _, fail := range *fails {
			// xx.xx.xx.xx  command         Fail    numactl not usable, bash: numactl: command not found
			if fail.Name == "command" {
				needInstallNumaCtl = true
			}
			// xx.xx.xx.xx  swap            Fail    swap is enabled, please disable it for best performance
			if fail.Name == "swap" {
				needSetSwap = true
			}
			// xx.xx.xx.xx  disk            Fail    mount point /data does not have 'nodelalloc' option set
			if fail.Name == "disk" && strings.HasPrefix(fail.Message, "mount point") {
				needRemount = true
				path, opt, err := p.getRemountInfoFromMsg(ctx, fail.Message)
				if err != nil {
					log.Errorf("extract remount info for host %s %s failed, %v", h.HostName, h.IP, err)
					return err
				}
				p.addRemountOpts(remount, path, opt)
			}
		}
	}

	warns, hasWarns := sortedResult[string(Warn)]
	if hasWarns {
		for _, warning := range *warns {
			// xx.xx.xx.xx  disk            Warn    mount point /data does not have 'noatime' option set
			if warning.Name == "disk" && strings.HasPrefix(warning.Message, "mount point") {
				needRemount = true
				path, opt, err := p.getRemountInfoFromMsg(ctx, warning.Message)
				if err != nil {
					log.Errorf("extract remount info for host %s %s failed, %v", h.HostName, h.IP, err)
					return err
				}
				p.addRemountOpts(remount, path, opt)
			}
		}
	}
	log.Infof("after apply host %s %s, needInstallNumaCtl: %v, needSetSwap: %v, needRemount: %v", h.HostName, h.IP, needInstallNumaCtl, needSetSwap, needRemount)

	if !needInstallNumaCtl && !needSetSwap && !needRemount {
		log.Infof("no need to auto fix for host %s %s", h.HostName, h.IP)
		return nil
	}

	if needInstallNumaCtl {
		if err = p.installNumaCtl(ctx, h); err != nil {
			errMsg := fmt.Sprintf("install numactl on host %s %s failed, %v", h.HostName, h.IP, err)
			log.Errorln(errMsg)
			return errors.NewError(errors.TIUNIMANAGER_RESOURCE_PREPARE_HOST_ERROR, errMsg)
		}
	}

	if needSetSwap {
		if err = p.setOffSwap(ctx, h); err != nil {
			errMsg := fmt.Sprintf("set off swap on host %s %s failed, %v", h.HostName, h.IP, err)
			log.Errorln(errMsg)
			return errors.NewError(errors.TIUNIMANAGER_RESOURCE_PREPARE_HOST_ERROR, errMsg)
		}
	}

	if needRemount {
		log.Infof("remount host %s %s with %v", h.HostName, h.IP, remount)
		for path, opts := range remount {
			var remountOpts []string
			for opt := range opts {
				remountOpts = append(remountOpts, opt)
			}
			if err = p.remountFS(ctx, h, path, remountOpts); err != nil {
				errMsg := fmt.Sprintf("remount host %s %s path %s adding opts [%v] failed, %v", h.HostName, h.IP, path, remountOpts, err)
				log.Errorln(errMsg)
				return errors.NewError(errors.TIUNIMANAGER_RESOURCE_PREPARE_HOST_ERROR, errMsg)
			}
		}
	}

	return nil
}

func (p *FileHostInitiator) setOffSwap(ctx context.Context, h *structs.HostInfo) (err error) {
	framework.LogWithContext(ctx).Infof("begin to set swap off on host %s %s", h.HostName, h.IP)
	checkExisted := "count=`cat /etc/sysctl.conf | grep -E '^vm.swappiness = 0$' | wc -l`"
	changeConf := "if [ $count -eq 0 ] ; then echo 'vm.swappiness = 0'>> /etc/sysctl.conf; fi"
	flushCmd := "swapoff -a"
	updateCmd := "sysctl -p"
	fstabCmd := "sed -i '/swap/s/^\\(.*\\)$/#\\1/g' /etc/fstab"
	authenticate := p.getEMAuthenticateToHost(ctx)
	result, err := p.sshClient.RunCommandsInRemoteHost(h.IP, int(h.SSHPort), *authenticate, true, rp_consts.DefaultCopySshIDTimeOut, []string{checkExisted, changeConf, flushCmd, updateCmd, fstabCmd})
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Infof("host %s [%s] set off swap, %v", h.HostName, h.IP, result)
	return nil
}

func (p *FileHostInitiator) installNumaCtl(ctx context.Context, h *structs.HostInfo) (err error) {
	framework.LogWithContext(ctx).Infof("begin to install numactl on host %s %s", h.HostName, h.IP)
	installNumaCtrlCmd := "yum install -y numactl"
	authenticate := p.getEMAuthenticateToHost(ctx)
	result, err := p.sshClient.RunCommandsInRemoteHost(h.IP, int(h.SSHPort), *authenticate, true, rp_consts.DefaultCopySshIDTimeOut, []string{installNumaCtrlCmd})
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Infof("host %s [%s] install numactl, %v", h.HostName, h.IP, result)
	return nil
}

func (p *FileHostInitiator) remountFS(ctx context.Context, h *structs.HostInfo, path string, opts []string) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("begin to remount path %s by adding opts %s on host %s %s", path, opts, h.HostName, h.IP)
	addingOpts := strings.Join(opts, ",")
	getMountInfoCmd := fmt.Sprintf("sed -n '\\# %s #p' /etc/fstab", path)
	authenticate := p.getEMAuthenticateToHost(ctx)
	result, err := p.sshClient.RunCommandsInRemoteHost(h.IP, int(h.SSHPort), *authenticate, true, rp_consts.DefaultCopySshIDTimeOut, []string{getMountInfoCmd})
	if err != nil {
		log.Errorf("host %s %s execute command %s failed, %v", h.HostName, h.IP, getMountInfoCmd, err)
		return err
	}
	// result should be "/dev/mapper/centos-root /data    xfs     defaults        0 0"
	mountInfo := strings.Fields(result)
	if len(mountInfo) != 6 {
		errMsg := fmt.Sprintf("get mount point %s option failed, mountInfo: %v", path, mountInfo)
		log.Errorln(errMsg)
		return errors.NewError(errors.TIUNIMANAGER_RESOURCE_PREPARE_HOST_ERROR, errMsg)
	}
	originOpts := mountInfo[3]
	targetOpts := fmt.Sprintf("%s,%s", originOpts, addingOpts)
	updateFsTabCmd := fmt.Sprintf("sed -i '\\# %s #s#%s#%s#g' /etc/fstab", path, originOpts, targetOpts)
	log.Infof("update fstab on host %s %s, using %s", h.HostName, h.IP, updateFsTabCmd)
	remountCMD := fmt.Sprintf("mount -o remount %s", path)
	result, err = p.sshClient.RunCommandsInRemoteHost(h.IP, int(h.SSHPort), *authenticate, true, rp_consts.DefaultCopySshIDTimeOut, []string{updateFsTabCmd, remountCMD})
	if err != nil {
		return err
	}
	log.Infof("host %s [%s] remout %s by adding %s, %v", h.HostName, h.IP, path, addingOpts, result)
	return nil

}

func (p *FileHostInitiator) getRemountInfoFromMsg(ctx context.Context, msg string) (mountPoint string, option string, err error) {
	warnExample := "mount point /xx/xx does not have 'xxx' option set, auto fixing not supported"
	exampleFields := strings.Fields(warnExample)
	msgFields := strings.Fields(msg)
	if len(msgFields) != len(exampleFields) {
		errMsg := fmt.Sprintf("remount warning message [%s] has a different format as expected [%s]", msg, warnExample)
		framework.LogWithContext(ctx).Errorln(errMsg)
		return "", "", errors.NewError(errors.TIUNIMANAGER_RESOURCE_PREPARE_HOST_ERROR, errMsg)
	}
	mountPoint = msgFields[2]
	option = strings.Trim(msgFields[6], "'")
	if option != "noatime" && option != "nodelalloc" {
		errMsg := fmt.Sprintf("remount option %s is not expected, should be 'noatime' or 'nodelalloc'", option)
		framework.LogWithContext(ctx).Errorln(errMsg)
		return "", "", errors.NewError(errors.TIUNIMANAGER_RESOURCE_PREPARE_HOST_ERROR, errMsg)
	}
	return
}

func (p *FileHostInitiator) addRemountOpts(remount map[string]map[string]struct{}, path string, opt string) {
	if opts, ok := remount[path]; !ok {
		remount[path] = make(map[string]struct{})
		remount[path][opt] = struct{}{}
	} else {
		if _, exist := opts[opt]; !exist {
			opts[opt] = struct{}{}
		}
	}
}
