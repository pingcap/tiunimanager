package sqleditor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message/dataapps/sqleditor"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/cluster/management"
	"github.com/pingcap/tiunimanager/models/common"
	sqleditorfile "github.com/pingcap/tiunimanager/models/dataapps/sqleditor/sqlfile"
	mock_sqleditorfile "github.com/pingcap/tiunimanager/test/mockmodels/mocksqleditor"
	workflow "github.com/pingcap/tiunimanager/workflow2"
	"github.com/stretchr/testify/assert"
)

func TestManager_GetManager(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		manager := GetManager()
		assert.NotNil(t, manager)
	})
}

func TestManager_CreateSQLFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		c := &gin.Context{}
		userID := "userID"
		c.Set(framework.TiUniManager_X_USER_ID_KEY, userID)
		ctx := framework.NewMicroCtxFromGinCtx(c)

		sqleditorRW := mock_sqleditorfile.NewMockReaderWriter(ctrl)
		models.SetSqlEditorFileReaderWriter(sqleditorRW)

		sqleditorRW.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return("test-id", nil)

		got, err := manager.CreateSQLFile(ctx, sqleditor.SQLFileReq{ClusterID: "123", Database: "test",
			Name: "test", Content: "test"})
		assert.NoError(t, err)
		assert.Equal(t, got.ID, "test-id")
	})

	t.Run("fail", func(t *testing.T) {
		c := &gin.Context{}
		userID := "userID"
		c.Set(framework.TiUniManager_X_USER_ID_KEY, userID)
		ctx := framework.NewMicroCtxFromGinCtx(c)

		sqleditorRW := mock_sqleditorfile.NewMockReaderWriter(ctrl)
		models.SetSqlEditorFileReaderWriter(sqleditorRW)

		sqleditorRW.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("create failed"))

		got, err := manager.CreateSQLFile(ctx, sqleditor.SQLFileReq{ClusterID: "123", Database: "test",
			Name: "test", Content: "test"})
		assert.NotNil(t, err)
		assert.Equal(t, got.ID, "")
	})

}

func TestManager_UpdateSQLFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		c := &gin.Context{}
		userID := "userID"
		c.Set(framework.TiUniManager_X_USER_ID_KEY, userID)
		ctx := framework.NewMicroCtxFromGinCtx(c)

		sqleditorRW := mock_sqleditorfile.NewMockReaderWriter(ctrl)
		models.SetSqlEditorFileReaderWriter(sqleditorRW)

		sqleditorRW.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		_, err := manager.UpdateSQLFile(ctx, sqleditor.SQLFileUpdateReq{ClusterID: "123", Database: "test",
			Name: "test", Content: "test"})
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		c := &gin.Context{}
		userID := "userID"
		c.Set(framework.TiUniManager_X_USER_ID_KEY, userID)
		ctx := framework.NewMicroCtxFromGinCtx(c)

		sqleditorRW := mock_sqleditorfile.NewMockReaderWriter(ctrl)
		models.SetSqlEditorFileReaderWriter(sqleditorRW)

		sqleditorRW.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))

		_, err := manager.UpdateSQLFile(ctx, sqleditor.SQLFileUpdateReq{ClusterID: "123", Database: "test",
			Name: "test", Content: "test"})
		assert.NotNil(t, err)
	})

}

func TestManager_DeleteSQLFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		c := &gin.Context{}
		userID := "userID"
		c.Set(framework.TiUniManager_X_USER_ID_KEY, userID)
		ctx := framework.NewMicroCtxFromGinCtx(c)

		sqleditorRW := mock_sqleditorfile.NewMockReaderWriter(ctrl)
		models.SetSqlEditorFileReaderWriter(sqleditorRW)

		sqleditorRW.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		_, err := manager.DeleteSQLFile(ctx, sqleditor.SQLFileDeleteReq{ClusterID: "123", SqlEditorFileID: "afsadf-adfasdf"})
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		c := &gin.Context{}
		userID := "userID"
		c.Set(framework.TiUniManager_X_USER_ID_KEY, userID)
		ctx := framework.NewMicroCtxFromGinCtx(c)

		sqleditorRW := mock_sqleditorfile.NewMockReaderWriter(ctrl)
		models.SetSqlEditorFileReaderWriter(sqleditorRW)

		sqleditorRW.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("delete failed"))

		_, err := manager.DeleteSQLFile(ctx, sqleditor.SQLFileDeleteReq{SqlEditorFileID: "afsadf-adfasdf"})
		assert.NotNil(t, err)
	})

}

func TestManager_ShowSQLFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		c := &gin.Context{}
		userID := "userID"
		c.Set(framework.TiUniManager_X_USER_ID_KEY, userID)
		ctx := framework.NewMicroCtxFromGinCtx(c)

		sqleditorRW := mock_sqleditorfile.NewMockReaderWriter(ctrl)
		models.SetSqlEditorFileReaderWriter(sqleditorRW)

		sqleditorRW.EXPECT().GetSqlFileByID(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&sqleditorfile.SqlEditorFile{
			ID:       "test",
			Name:     "name",
			Content:  "Content",
			Database: "test",
		}, nil)

		res, err := manager.ShowSQLFile(ctx, sqleditor.ShowSQLFileReq{ClusterID: "123", SqlEditorFileID: "test"})
		assert.NoError(t, err)
		assert.Equal(t, res.ID, "test")
	})

}

func TestManager_ListSqlFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		c := &gin.Context{}
		userID := "userID"
		c.Set(framework.TiUniManager_X_USER_ID_KEY, userID)
		ctx := framework.NewMicroCtxFromGinCtx(c)

		sqleditorRW := mock_sqleditorfile.NewMockReaderWriter(ctrl)
		models.SetSqlEditorFileReaderWriter(sqleditorRW)
		sqleditorRW.EXPECT().GetSqlFileList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*sqleditorfile.SqlEditorFile{
			{
				ID:       "test",
				Name:     "name",
				Content:  "Content",
				Database: "test",
			},
		}, int64(1), nil)

		res, err := manager.ListSqlFile(ctx, sqleditor.ListSQLFileReq{ClusterID: "123"})
		assert.NoError(t, err)
		assert.Equal(t, res.Total, int64(1))
		assert.Equal(t, len(res.List), 1)
	})

}

func TestManager_ShowClusterMeta(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		_, err := manager.ShowClusterMeta(getClusterMetaContext(), sqleditor.ShowClusterMetaReq{ClusterID: "123", IsBrief: true})
		assert.Error(t, err)
	})
}
func TestManager_ShowTableMeta(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		_, err := manager.ShowTableMeta(getClusterMetaContext(), sqleditor.ShowTableMetaReq{ClusterID: "test-table", DbName: "test", TableName: "test"})
		assert.Error(t, err)
	})
}

func TestManager_CreateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		_, err := manager.CreateSession(getClusterMetaContext(), sqleditor.CreateSessionReq{ClusterID: "test-table", Database: "test"})
		assert.Error(t, err)
	})
}

func TestManager_CloseSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := &Manager{}

	t.Run("normal", func(t *testing.T) {
		_, err := manager.CloseSession(getClusterMetaContext(), sqleditor.CloseSessionReq{ClusterID: "test-table", SessionID: "test"})
		assert.Error(t, err)
	})
}

func getClusterMetaContext() context.Context {
	flowContext := workflow.NewFlowContext(context.TODO(), make(map[string]string))
	flowContext.SetData("ClusterMeta", &meta.ClusterMeta{
		Cluster: &management.Cluster{
			Entity: common.Entity{
				ID:        "2145635758",
				TenantId:  "324567",
				Status:    string(constants.ClusterInitializing),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:              "koojdafij",
			Type:              "TiDB",
			Version:           "v5.0.0",
			Tags:              []string{"111", "333"},
			OwnerId:           "436534636u",
			ParameterGroupID:  "352467890",
			Copies:            4,
			Region:            "Region1",
			Exclusive:         false,
			CpuArchitecture:   "x86_64",
			MaintenanceStatus: constants.ClusterMaintenanceCreating,
		},
		Instances: map[string][]*management.ClusterInstance{
			"TiDB": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "TiDB",
					Version:      "v5.0.0",
					Ports:        []int32{10001, 10002, 10003, 10004},
					HostIP:       []string{"127.0.0.1", "127.0.0.6"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
			"TiKV": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "TiKV",
					Version:      "v5.0.0",
					Ports:        []int32{20001, 20002, 20003, 20004},
					HostIP:       []string{"127.0.0.2"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
			"PD": {
				{
					Entity: common.Entity{
						Status: string(constants.ClusterInstanceInitializing),
					},
					Zone:         "zone1",
					CpuCores:     4,
					Memory:       8,
					Type:         "PD",
					Version:      "v5.0.0",
					Ports:        []int32{30001, 30002, 30003, 30004},
					HostIP:       []string{"127.0.0.3"},
					DiskType:     "SSD",
					DiskCapacity: 128,
				},
			},
		},
	})
	return flowContext
}
