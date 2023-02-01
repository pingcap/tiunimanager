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
	dbCommon "github.com/pingcap/tiunimanager/models/common"
	"gorm.io/gorm"
)

import (
	"context"

	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
)

type SqlEditorFileReadWrite struct {
	dbCommon.GormDB
}

func NewSqlEditorReadWrite(db *gorm.DB) *SqlEditorFileReadWrite {
	return &SqlEditorFileReadWrite{
		dbCommon.WrapDB(db),
	}
}

func (arw *SqlEditorFileReadWrite) Create(ctx context.Context, sqlEditorFile *SqlEditorFile, name string) (string, error) {

	if "" == name {
		framework.LogWithContext(ctx).Errorf("create sqleditorfile %s, name is not allow empty", name)
		return "", errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "create sqleditorfile %s, parameter invalid", name)
	}
	// create user
	err := arw.DB(ctx).Create(sqlEditorFile).Error
	if err != nil {
		return "", err
	}

	return sqlEditorFile.ID, nil
}

func (arw *SqlEditorFileReadWrite) Update(ctx context.Context, sqlEditorFile *SqlEditorFile) error {
	// create user
	err := arw.DB(ctx).Model(&SqlEditorFile{}).Where("id = ? AND created_by = ? and is_deleted=0", sqlEditorFile.ID, sqlEditorFile.UpdatedBy).
		Update("name", sqlEditorFile.Name).Update("content", sqlEditorFile.Content).
		Update("database", sqlEditorFile.Database).
		Update("updated_by", sqlEditorFile.UpdatedBy).Error
	if err != nil {
		return err
	}

	return nil
}

func (arw *SqlEditorFileReadWrite) Delete(ctx context.Context, ID string, createdBy string) error {
	if "" == ID {
		framework.LogWithContext(ctx).Errorf("delete sqleditorfile %s, parameter invalid", ID)
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "delete sqleditorfile %s, parameter invalid", ID)
	}

	err := arw.DB(ctx).Model(&SqlEditorFile{}).Where("id = ? and created_by = ?", ID, createdBy).Update("is_deleted", 1).Error
	if err != nil {
		return err
	}
	return nil
}

func (arw *SqlEditorFileReadWrite) GetSqlFileByID(ctx context.Context, ID string, createdBy string) (sqlfile *SqlEditorFile, err error) {
	sqlfile = &SqlEditorFile{}
	return sqlfile, arw.DB(ctx).Model(sqlfile).Where("id = ? and created_by=? and is_deleted = 0", ID, createdBy).First(sqlfile).Error

}

func (arw *SqlEditorFileReadWrite) GetSqlFileList(ctx context.Context, clusterID string, createdBy string, pageSize int, offset int) (sqlFileList []*SqlEditorFile, total int64, err error) {

	sqlFileList = make([]*SqlEditorFile, 0)
	query := arw.DB(ctx).Model(&SqlEditorFile{})
	query = query.Where("cluster_id = ? and created_by = ? and is_deleted = 0", clusterID, createdBy)
	err = query.Count(&total).Offset(offset).Limit(pageSize).Find(&sqlFileList).Error
	return sqlFileList, total, err

}
