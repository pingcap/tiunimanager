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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqleditorReadWrite_Create(t *testing.T) {
	editorFile := &SqlEditorFile{
		Name:      "test",
		Content:   "set names utf8",
		Database:  "test",
		ClusterID: "test",
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	t.Run("suc", func(t *testing.T) {
		ID, err := testRW.Create(context.TODO(), editorFile, "test")
		assert.NotEmpty(t, ID)
		assert.NoError(t, err)
	})

	t.Run("name empty", func(t *testing.T) {
		editorFile.Name = ""
		ID, err := testRW.Create(context.TODO(), editorFile, "")
		assert.Empty(t, ID)
		assert.Error(t, err)
	})

	t.Run("insert error", func(t *testing.T) {
		editorFile.UpdatedBy = ""
		ID, err := testRW.Create(context.TODO(), editorFile, "test")
		assert.Empty(t, ID)
		assert.Error(t, err)
	})

}

func TestSqleditorReadWrite_Update(t *testing.T) {
	editorFile := &SqlEditorFile{
		Name:      "test",
		Content:   "set names utf8",
		Database:  "test",
		ClusterID: "test",
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	var ID string
	var err error
	t.Run("create suc", func(t *testing.T) {
		ID, err = testRW.Create(context.TODO(), editorFile, "test")
		assert.NotEmpty(t, ID)
		assert.NoError(t, err)
	})

	t.Run("update suc", func(t *testing.T) {
		editorFile.ID = ID
		editorFile.Content = "test"
		err := testRW.Update(context.TODO(), editorFile)
		assert.NoError(t, err)
	})
}

func TestSqleditorReadWrite_Delete(t *testing.T) {
	editorFile := &SqlEditorFile{
		Name:      "test",
		Content:   "set names utf8",
		Database:  "test",
		ClusterID: "test",
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	var ID string
	var err error
	t.Run("create suc", func(t *testing.T) {
		ID, err = testRW.Create(context.TODO(), editorFile, "test")
		assert.NotEmpty(t, ID)
		assert.NoError(t, err)
	})

	t.Run("ID empty", func(t *testing.T) {
		err := testRW.Delete(context.TODO(), "", "test")
		assert.Error(t, err)
	})
	t.Run("delete suc", func(t *testing.T) {
		err := testRW.Delete(context.TODO(), ID, "test")
		assert.NoError(t, err)
	})
}

func TestSqleditorReadWrite_GetSqlFileByID(t *testing.T) {
	editorFile := &SqlEditorFile{
		Name:      "test",
		Content:   "set names utf8",
		Database:  "test",
		ClusterID: "test",
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	var ID string
	var err error
	t.Run("create suc", func(t *testing.T) {
		ID, err = testRW.Create(context.TODO(), editorFile, "test")
		assert.NotEmpty(t, ID)
		assert.NoError(t, err)
	})

	t.Run("query suc", func(t *testing.T) {
		res, err := testRW.GetSqlFileByID(context.TODO(), ID, "test")
		assert.NoError(t, err)
		assert.Equal(t, res.Name, "test")
	})
}

func TestSqleditorReadWrite_GetSqlFileList(t *testing.T) {
	editorFile := &SqlEditorFile{
		Name:      "test",
		Content:   "set names utf8",
		Database:  "test",
		ClusterID: "test",
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	var ID string
	var err error
	t.Run("create suc", func(t *testing.T) {
		ID, err = testRW.Create(context.TODO(), editorFile, "test")
		assert.NotEmpty(t, ID)
		assert.NoError(t, err)
	})

	t.Run("query suc", func(t *testing.T) {
		_, total, err := testRW.GetSqlFileList(context.TODO(), "test", "test", 1, 100)
		assert.NoError(t, err)
		assert.Equal(t, total, int64(4))
	})

}
