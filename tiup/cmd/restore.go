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

package main

import (
	"github.com/pingcap-inc/tiem/tiup/manager"
	"github.com/spf13/cobra"
)

/* Add a command to restore EM metadata information */

func restoreCmd() *cobra.Command {
	opt := manager.RestoreOptions{}
	cmd := &cobra.Command{
		Use:    "restore <cluster-name> <source-path>",
		Short:  "Restore tiem cluster metadata information",
		Hidden: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return cmd.Help()
			}

			clusterName := args[0]
			opt.Source = args[1]

			return cm.Restore(clusterName, opt, gOpt)
		},
	}

	cmd.Flags().StringSliceVarP(&gOpt.Roles, "role", "R", nil, "Only exec on host with specified roles")
	cmd.Flags().StringSliceVarP(&gOpt.Nodes, "node", "N", nil, "Only exec on host with specified nodes")

	return cmd
}
