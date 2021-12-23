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
 ******************************************************************************/

/**
 * @Author: guobob
 * @Description:
 * @File:  common.go
 * @Version: 1.0.0
 * @Date: 2021/12/9 21:27
 */

package utils

const (
	EventHandshake uint64 = iota
	EventQuit
	EventQuery
	EventStmtPrepare
	EventStmtExecute
	EventStmtClose
)

const (
	Success  int = iota
	ErrorNoNotEqual
	ExecTimeCompareFail
	ResultRowCountNotEqual
	ResultRowDetailNotEqual
)

func TypeString(t uint64) string {
	switch t {
	case EventHandshake:
		return "EventHandshake"
	case EventQuit:
		return "EventQuit"
	case EventQuery:
		return "EventQuery"
	case EventStmtPrepare:
		return "EventStmtPrepare"
	case EventStmtExecute:
		return "EventStmtExecute"
	case EventStmtClose:
		return "EventStmtClose"
	default:
		return "UnKnownType"
	}
}
