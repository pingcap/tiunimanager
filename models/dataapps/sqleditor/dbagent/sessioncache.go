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
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SessionItem struct {
	Conn   *sql.Conn
	EndSec uint64
}

// Database configuration
type SessionCache struct {
	maxNum     uint64
	sessionMap sync.Map
}

const (
	DelayCloseSessionSec = 30
	ClearInterval        = 100
	MaxNum               = 100
)

var storage = &SessionCache{}

func cronClearExpireSession(storage *SessionCache) {
	for {
		curTimestamp := time.Now().Unix()
		storage.sessionMap.Range(func(k, v interface{}) bool {
			val := v.(*SessionItem)
			if curTimestamp > int64(val.EndSec+DelayCloseSessionSec) {
				storage.sessionMap.Delete(k)
			}
			return true
		})
		time.Sleep(time.Millisecond * ClearInterval)
	}
}
func CloseSessionCache(sessionId string) error {
	val, ok := storage.sessionMap.Load(sessionId)
	if !ok {
		return errors.New("no sessionId found in cache: " + sessionId)
	}
	sessionItem := val.(*SessionItem)
	sessionItem.EndSec = uint64(time.Now().Unix() - DelayCloseSessionSec)
	//conn will be closed by sql.DB automatically
	storage.sessionMap.Store(sessionId, sessionItem)
	return nil
}

func GetSessionFromCache(sessionId string) (*sql.Conn, error) {
	val, ok := storage.sessionMap.Load(sessionId)
	if !ok {
		return nil, errors.New("no sessionId found in cache: " + sessionId)
	}
	sessionItem := val.(*SessionItem)
	return sessionItem.Conn, nil
}

func CreateSessionCache(conn *sql.Conn, expireSec uint64) string {
	uuid := uuid.NewString()
	item := &SessionItem{
		Conn:   conn,
		EndSec: uint64(time.Now().Unix()) + expireSec,
	}
	storage.sessionMap.Store(uuid, item)
	return uuid
}

func Init() {
	storage.maxNum = MaxNum
	go cronClearExpireSession(storage)
}
