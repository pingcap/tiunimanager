/******************************************************************************
 * Copyright (c)  2023 PingCAP, Inc.                                          *
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

package sqleditorfile

import (
	"context"
)

type ReaderWriter interface {
	Create(ctx context.Context, sqlEditorFile *SqlEditorFile, name string) (string, error)
	Update(ctx context.Context, sqlEditorFile *SqlEditorFile) error
	Delete(ctx context.Context, ID string, createBy string) error
	GetSqlFileByID(ctx context.Context, ID string, createdBy string) (sqlfile *SqlEditorFile, err error)
	GetSqlFileList(ctx context.Context, clusterID string, createdBy string, pageSize int, offset int) (sqlFileList []*SqlEditorFile, total int64, err error)
}
