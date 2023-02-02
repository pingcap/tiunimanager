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

package sqleditor

import (
	"context"
	"sync"

	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message/dataapps/sqleditor"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/dataapps/sqleditor/dbagent"
	sqleditorfile "github.com/pingcap/tiunimanager/models/dataapps/sqleditor/sqlfile"
)

var manager *Manager
var once sync.Once

const (
	EXPIRESEC = 2 * 14400
)

type Manager struct{}

func GetManager() *Manager {
	once.Do(func() {
		if manager == nil {
			manager = &Manager{}
		}
	})
	return manager
}

// CreateSQLFile
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return sqleditor.CreateSQLFileRes
// @return error
func (p *Manager) CreateSQLFile(ctx context.Context, request sqleditor.SQLFileReq) (res sqleditor.CreateSQLFileRes, err error) {
	createdBy := framework.GetUserIDFromContext(ctx)

	createEntity := sqleditorfile.SqlEditorFile{
		ClusterID: request.ClusterID,
		Database:  request.Database,
		Name:      request.Name,
		Content:   request.Content,
		IsDeleted: 0,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}
	rw := models.GetSqlEditorFileReaderWriter()
	id, err := rw.Create(ctx, &createEntity, request.Name)
	if err != nil {
		return res, err
	}

	return sqleditor.CreateSQLFileRes{
		ID: id,
	}, nil
}

// UpdateSQLFile
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return sqleditor.UpdateSQLFileRes
// @return error
func (p *Manager) UpdateSQLFile(ctx context.Context, request sqleditor.SQLFileUpdateReq) (res sqleditor.UpdateSQLFileRes, err error) {
	updatedBy := framework.GetUserIDFromContext(ctx)

	createEntity := sqleditorfile.SqlEditorFile{
		Database:  request.Database,
		Name:      request.Name,
		Content:   request.Content,
		UpdatedBy: updatedBy,
		ID:        request.SqlEditorFileID,
	}
	rw := models.GetSqlEditorFileReaderWriter()
	err = rw.Update(ctx, &createEntity)
	if err != nil {
		return res, err
	}

	return sqleditor.UpdateSQLFileRes{}, nil
}

// DeleteSQLFile
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return sqleditor.DeleteSQLFileRes
// @return error
func (p *Manager) DeleteSQLFile(ctx context.Context, request sqleditor.SQLFileDeleteReq) (res sqleditor.DeleteSQLFileRes, err error) {
	userID := framework.GetUserIDFromContext(ctx)

	rw := models.GetSqlEditorFileReaderWriter()
	err = rw.Delete(ctx, request.SqlEditorFileID, userID)
	if err != nil {
		return res, err
	}

	return sqleditor.DeleteSQLFileRes{}, nil
}

// ShowSQLFile
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return sqleditor.ShowSQLFileRes
// @return error
func (p *Manager) ShowSQLFile(ctx context.Context, request sqleditor.ShowSQLFileReq) (res sqleditor.ShowSQLFileRes, err error) {
	userID := framework.GetUserIDFromContext(ctx)

	rw := models.GetSqlEditorFileReaderWriter()
	sqlfile, err := rw.GetSqlFileByID(ctx, request.SqlEditorFileID, userID)
	if err != nil {
		return res, err
	}

	return sqleditor.ShowSQLFileRes{
		parse(sqlfile),
	}, nil
}

func parse(sqlfile *sqleditorfile.SqlEditorFile) sqleditor.SqlEditorFile {
	return sqleditor.SqlEditorFile{
		ID:        sqlfile.ID,
		ClusterID: sqlfile.ClusterID,
		Database:  sqlfile.Database,
		Name:      sqlfile.Name,
		Content:   sqlfile.Content,
		IsDeleted: sqlfile.IsDeleted,
		CreatedBy: sqlfile.CreatedBy,
		UpdatedBy: sqlfile.UpdatedBy,
		CreatedAt: sqlfile.CreatedAt,
		UpdatedAt: sqlfile.UpdatedAt,
	}

}

// ListSqlFile
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return sqleditor.ListSQLFileRes
// @return error
func (p *Manager) ListSqlFile(ctx context.Context, request sqleditor.ListSQLFileReq) (res sqleditor.ListSQLFileRes, err error) {
	userID := framework.GetUserIDFromContext(ctx)

	rw := models.GetSqlEditorFileReaderWriter()
	if request.PageSize <= 0 {
		request.PageSize = sqleditor.DEFAULTPAGESIZE
	}
	offset := request.PageRequest.CalcOffset()
	sqlfileList, total, err := rw.GetSqlFileList(ctx, request.ClusterID, userID, request.PageSize, offset)
	if err != nil {
		return res, err
	}
	sqllist := make([]sqleditor.SqlEditorFile, 0)
	for _, sqlfile := range sqlfileList {
		if sqlfile != nil {
			sqllist = append(sqllist, parse(sqlfile))
		}
	}

	return sqleditor.ListSQLFileRes{
		Total:    total,
		List:     sqllist,
		Page:     request.Page,
		PageSize: request.PageSize,
	}, nil
}

// ShowTableMeta
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return *sqleditor.MetaRes
// @return error
func (p *Manager) ShowTableMeta(ctx context.Context, request sqleditor.ShowTableMetaReq) (res *sqleditor.MetaRes, err error) {
	clusterMeta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		return
	}
	db, err := meta.CreateSQLLink(ctx, clusterMeta)
	if err != nil {
		return res, errors.WrapError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, err.Error(), err)
	}
	defer db.Close()
	dbModel := dbagent.NewDBAgent(db)
	return dbModel.GetTableMetaData(ctx, request.ClusterID, request.DbName, request.TableName)
}

// ShowClusterMeta
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return []*sqleditor.DBMeta
// @return error
func (p *Manager) ShowClusterMeta(ctx context.Context, request sqleditor.ShowClusterMetaReq) (res []*sqleditor.DBMeta, err error) {
	clusterMeta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		return
	}
	db, err := meta.CreateSQLLink(ctx, clusterMeta)
	if err != nil {
		return res, errors.WrapError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, err.Error(), err)
	}
	defer db.Close()
	dbModel := dbagent.NewDBAgent(db)
	return dbModel.GetClusterMetaData(ctx, request.IsBrief, request.ShowSystemDB)
}

// CreateSession
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return []*sqleditor.CreateSessionRes
// @return error
func (p *Manager) CreateSession(ctx context.Context, request sqleditor.CreateSessionReq) (res *sqleditor.CreateSessionRes, err error) {
	clusterMeta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		return
	}
	db, err := meta.CreateSQLLinkWithDatabase(ctx, clusterMeta, request.Database)
	if err != nil {
		return res, errors.WrapError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, err.Error(), err)
	}
	defer db.Close()
	dbModel := dbagent.NewDBAgent(db)
	sessionID, err := dbModel.CreateSession(ctx, request.ClusterID, EXPIRESEC, request.Database)
	if err != nil {
		return
	}

	return &sqleditor.CreateSessionRes{
		SessionID: sessionID,
	}, nil
}

// CloseSession
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return sqleditor.CloseSessionRes
// @return error
func (p *Manager) CloseSession(ctx context.Context, request sqleditor.CloseSessionReq) (res *sqleditor.CloseSessionRes, err error) {
	clusterMeta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		return
	}
	db, err := meta.CreateSQLLink(ctx, clusterMeta)
	if err != nil {
		return res, errors.WrapError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, err.Error(), err)
	}
	defer db.Close()
	dbModel := dbagent.NewDBAgent(db)

	return &sqleditor.CloseSessionRes{}, dbModel.CloseSession(ctx, request.SessionID)
}

// Statements
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return sqleditor.StatementsRes
// @return error
func (p *Manager) Statements(ctx context.Context, request sqleditor.StatementParam) (res *sqleditor.StatementsRes, err error) {
	if request.SessionId == "" {
		return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "sessionId is empty")
	}
	clusterMeta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		return
	}
	db, err := meta.CreateSQLLink(ctx, clusterMeta)
	if err != nil {
		return res, errors.WrapError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, err.Error(), err)
	}
	defer db.Close()
	dbModel := dbagent.NewDBAgent(db)

	return dbModel.ExecSqlWithSession(ctx, request.SessionId, request.Sql)
}
