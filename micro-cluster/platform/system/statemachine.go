/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

/*******************************************************************************
 * @File: statemachine.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/18
*******************************************************************************/

package system

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
)

type action func(ctx context.Context, event constants.SystemEvent, originalState constants.SystemState) error

// pushToServiceReady
// @Description: push system state to system ready, then trigger data initializer
// @Parameter ctx
// @Parameter event
// @Parameter originalState
// @return error
func pushToServiceReady(ctx context.Context, event constants.SystemEvent, originalState constants.SystemState) error {
	return errors.OfNullable(nil).BreakIf(func() error {
		return models.GetSystemReaderWriter().UpdateState(ctx, originalState, constants.SystemServiceReady)
	}).BreakIf(func() error {
		systemInfo, err := models.GetSystemReaderWriter().GetSystemInfo(ctx)
		if err == nil {
			return models.IncrementVersionData(systemInfo.CurrentVersionID, framework.Current.GetClientArgs().EMVersion)
		}
		return err
	}).BreakIf(func() error {
		acceptSystemEvent(ctx, constants.SystemDataInitialized)
		return nil
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("push system state to ServiceReady failed, event = %s, originalState = %s, err = %s", event, originalState, err.Error())
	}).Present()
}

func pushToDataReady(ctx context.Context, event constants.SystemEvent, originalState constants.SystemState) error {
	return errors.OfNullable(nil).BreakIf(func() error {
		return models.GetSystemReaderWriter().UpdateState(ctx, originalState, constants.SystemDataReady)
	}).BreakIf(func() error {
		models.GetSystemReaderWriter().UpdateVersion(ctx, framework.Current.GetClientArgs().EMVersion)
		return nil
	}).BreakIf(func() error {
		acceptSystemEvent(ctx, constants.SystemServe)
		return nil
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("push system state to DataReady failed, event = %s, originalState = %s, err = %s", event, originalState, err.Error())
	}).Present()
}

func pushToRunning(ctx context.Context, event constants.SystemEvent, originalState constants.SystemState) error {
	return errors.OfNullable(nil).BreakIf(func() error {
		return models.GetSystemReaderWriter().UpdateState(ctx, originalState, constants.SystemRunning)
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("push system state to Running failed, event = %s, originalState = %s, err = %s", event, originalState, err.Error())
	}).Present()
}

func pushToUnserviceable(ctx context.Context, event constants.SystemEvent, originalState constants.SystemState) error {
	return errors.OfNullable(nil).BreakIf(func() error {
		return models.GetSystemReaderWriter().UpdateState(ctx, originalState, constants.SystemUnserviceable)
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("push system state to Unserviceable failed, event = %s, originalState = %s, err = %s", event, originalState, err.Error())
	}).Present()
}

func pushToUpgrading(ctx context.Context, event constants.SystemEvent, originalState constants.SystemState) error {
	return errors.OfNullable(nil).BreakIf(func() error {
		return models.GetSystemReaderWriter().UpdateState(ctx, originalState, constants.SystemUpgrading)
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("push system state to Upgrading failed, event = %s, originalState = %s, err = %s", event, originalState, err.Error())
	}).Present()
}

func pushToFailure(ctx context.Context, event constants.SystemEvent, originalState constants.SystemState) error {
	return errors.OfNullable(nil).BreakIf(func() error {
		return models.GetSystemReaderWriter().UpdateState(ctx, originalState, constants.SystemFailure)
	}).If(func(err error) {
		framework.LogWithContext(ctx).Errorf("push system state to Failure failed, event = %s, originalState = %s, err = %s", event, originalState, err.Error())
	}).Present()
}
