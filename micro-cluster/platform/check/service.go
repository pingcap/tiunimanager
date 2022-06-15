/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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

package check

import (
	"context"
	"github.com/pingcap/tiunimanager/message"
)

type CheckService interface {
	// Check
	// @Description: check platform
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.CheckPlatformRsp
	// @Return error
	Check(ctx context.Context, request message.CheckPlatformReq) (resp message.CheckPlatformRsp, err error)

	// CheckCluster
	// @Description: check cluster
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.CheckClusterRsp
	// @Return error
	CheckCluster(ctx context.Context, request message.CheckClusterReq) (resp message.CheckClusterRsp, err error)

	// QueryCheckReports
	// @Description: query check reports
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryCheckReportsRsp
	// @Return error
	QueryCheckReports(ctx context.Context, request message.QueryCheckReportsReq) (resp message.QueryCheckReportsRsp, err error)

	// GetCheckReport
	// @Description: get check report
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.GetCheckReportRsp
	// @Return error
	GetCheckReport(ctx context.Context, request message.GetCheckReportReq) (resp message.GetCheckReportRsp, err error)
}
