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
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: readerwriter
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/28
*******************************************************************************/

package mirror

import "context"

type ReaderWriter interface {
	// Create
	// @Description: create new mirror record
	// @Receiver m
	// @Parameter ctx
	// @Parameter tiUPComponentType
	// @Parameter mirrorAddr
	// @Return *Mirror
	// @Return error
	Create(ctx context.Context, tiUPComponentType string, mirrorAddr string) (*Mirror, error)

	// Update
	// @Description: update mirror record
	// @Receiver m
	// @Parameter ctx
	// @Parameter updateTemplate
	// @Return error
	Update(ctx context.Context, updateTemplate *Mirror) error

	// Get
	// @Description: get mirror record for given id
	// @Receiver m
	// @Parameter ctx
	// @Parameter id
	// @Return *Mirror
	// @Return error
	Get(ctx context.Context, id string) (*Mirror, error)

	// QueryByComponentType
	// @Description: get mirror record for given tiUPComponentType
	// @Receiver m
	// @Parameter ctx
	// @Parameter tiUPComponentType
	// @Return *Mirror
	// @Return error
	QueryByComponentType(ctx context.Context, tiUPComponentType string) (*Mirror, error)
}
