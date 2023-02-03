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

package dbagent

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_cronClearExpireSession(t *testing.T) {
	nowTime := time.Now().Unix()
	storage.sessionMap.Store("uuid1", &SessionItem{
		Conn:   nil,
		EndSec: uint64(nowTime - DelayCloseSessionSec*2),
	},
	)
	storage.sessionMap.Store("uuid2", &SessionItem{
		Conn:   nil,
		EndSec: uint64(nowTime - DelayCloseSessionSec*2),
	},
	)
	Init()
	time.Sleep(ClearInterval * time.Millisecond)
	_, ok1 := storage.sessionMap.Load("uuid1")
	assert.Equal(t, ok1, false)
	_, ok2 := storage.sessionMap.Load("uuid2")
	assert.Equal(t, ok2, false)
}

func TestManager_CreateSession(t *testing.T) {
	conn := &sql.Conn{}

	t.Run("normal", func(t *testing.T) {
		uuid := CreateSessionCache(conn, 100)
		assert.NotEmpty(t, uuid)
	})

}

func TestManager_CloseSession(t *testing.T) {
	conn := &sql.Conn{}
	uuid := CreateSessionCache(conn, 100)
	assert.NotEmpty(t, uuid)

	t.Run("normal", func(t *testing.T) {
		err := CloseSessionCache(uuid)
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		err := CloseSessionCache("failed")
		assert.Error(t, err)
	})

}

func TestManager_GetSessionFromCache(t *testing.T) {
	conn := &sql.Conn{}
	uuid := CreateSessionCache(conn, 100)
	assert.NotEmpty(t, uuid)

	t.Run("normal", func(t *testing.T) {
		_, err := GetSessionFromCache(uuid)
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		_, err := GetSessionFromCache("failed")
		assert.Error(t, err)
	})

}
