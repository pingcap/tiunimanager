/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
 * @Date: 2021/12/30
*******************************************************************************/

package tiup

import "context"

type ReaderWriter interface {
	// Create
	// @Description: create new record
	// @Receiver m
	// @Parameter ctx
	// @Parameter tiUPComponentType
	// @Parameter tiUPHome
	// @Return *TiupConfig
	// @Return error
	Create(ctx context.Context, tiUPComponentType string, tiUPHome string) (*TiupConfig, error)

	// Update
	// @Description: update record
	// @Receiver m
	// @Parameter ctx
	// @Parameter updateTemplate
	// @Return error
	Update(ctx context.Context, updateTemplate *TiupConfig) error

	// Get
	// @Description: get record for given id
	// @Receiver m
	// @Parameter ctx
	// @Parameter id
	// @Return *TiupConfig
	// @Return error
	Get(ctx context.Context, id string) (*TiupConfig, error)

	// QueryByComponentType
	// @Description: get record for given tiUPComponentType
	// @Receiver m
	// @Parameter ctx
	// @Parameter tiUPComponentType
	// @Return *TiupConfig
	// @Return error
	QueryByComponentType(ctx context.Context, tiUPComponentType string) (*TiupConfig, error)
}
