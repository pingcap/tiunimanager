
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

package uuidutil

import (
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid"
)

const (
	ENTITY_UUID_LENGTH int = 22
)

// GenerateID Encode with Base64
// URLEncoding Set contains '-' and '_'
func GenerateID() string {
	var encoded []byte = nil

	for encoded == nil || strings.HasPrefix(string(encoded), "-") || strings.HasPrefix(string(encoded), "_"){
		uuid := uuid.New()

		encoded = make([]byte, ENTITY_UUID_LENGTH)
		base64.URLEncoding.WithPadding(base64.NoPadding).Encode(encoded, uuid[0:16])
	}

	return string(encoded)
}

// ShortId Encode with Base57
func ShortId() string {
	return shortuuid.New()
}
