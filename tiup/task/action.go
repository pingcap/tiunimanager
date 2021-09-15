// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	"context"
	"crypto/tls"
	"fmt"

	operator "github.com/pingcap-inc/tiem/tiup/operation"
	"github.com/pingcap-inc/tiem/tiup/spec"
	"github.com/pingcap/errors"
)

// ClusterOperate represents the cluster operation task.
type ClusterOperate struct {
	spec    *spec.Specification
	op      operator.Operation
	options operator.Options
	tlsCfg  *tls.Config
}

// Execute implements the Task interface
func (c *ClusterOperate) Execute(ctx context.Context) error {
	var (
		err         error
		printStatus bool = true
	)
	var opErrMsg = [...]string{
		"failed to start",
		"failed to stop",
		"failed to restart",
		"failed to destroy",
		"failed to upgrade",
		"failed to scale in",
		"failed to scale out",
	}
	switch c.op {
	case operator.StartOperation:
		err = operator.Start(ctx, c.spec, c.options, c.tlsCfg)
	case operator.StopOperation:
		err = operator.Stop(ctx, c.spec, c.options, c.tlsCfg)
	case operator.RestartOperation:
		err = operator.Restart(ctx, c.spec, c.options, c.tlsCfg)
	case operator.DestroyOperation:
		err = operator.Destroy(ctx, c.spec, c.options)
	case operator.UpgradeOperation:
		err = operator.Upgrade(ctx, c.spec, c.options, c.tlsCfg)
	case operator.ScaleInOperation:
		err = operator.ScaleIn(ctx, c.spec, c.options, c.tlsCfg)
		printStatus = false
	default:
		return errors.Errorf("nonsupport %s", c.op)
	}

	if err != nil {
		return errors.Annotatef(err, opErrMsg[c.op])
	}

	if printStatus {
		operator.PrintClusterStatus(ctx, c.spec)
	}

	return nil
}

// Rollback implements the Task interface
func (c *ClusterOperate) Rollback(ctx context.Context) error {
	return ErrUnsupportedRollback
}

// String implements the fmt.Stringer interface
func (c *ClusterOperate) String() string {
	return fmt.Sprintf("ClusterOperate: operation=%s, options=%+v", c.op, c.options)
}
