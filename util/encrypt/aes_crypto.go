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
 *                                                                            *
 ******************************************************************************/

package encrypt

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/pingcap/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// key should be 16、24 or 32 length [] byte, conresponding to AES-128, AES-192 或 AES-256.
var key []byte

func init() {
	// left the key blank and wait user call `InitKey` to init
	key = []byte("")
}

func InitKey(key []byte) error {
	l := len(key)
	switch l {
	case 16, 24, 32:
		break
	default:
		return fmt.Errorf("invalid key size %d", l)
	}
	return nil
}

func aesEncryptCFB(plain []byte) (encrypted []byte, err error) {
	encrypted = make([]byte, aes.BlockSize+len(plain))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, status.Errorf(codes.Internal, "init vector err, %s", err)
	}
	crypted, err := AESEncryptWithCFB(plain, key, iv)
	if err != nil {
		return nil, err
	}
	cnt := copy(encrypted[aes.BlockSize:], crypted)
	if cnt != len(plain) {
		return nil, errors.Errorf("copy crypt to encrypted failed, expect %d, but %d", len(plain), cnt)
	}
	return encrypted, nil
}

func aesDecryptCFB(encrypted []byte) (decrypted []byte, err error) {
	if len(encrypted) < aes.BlockSize {
		return nil, status.Errorf(codes.Internal, "ciphertext too short, %d < aes.BlockSize(%d)", len(encrypted), aes.BlockSize)
	}
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	return AESDecryptWithCFB(encrypted, key, iv)
}

func AesEncryptCFB(plainStr string) (encryptedStr string, err error) {
	encrypted, err := aesEncryptCFB([]byte(plainStr))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(encrypted), err
}

func AesDecryptCFB(encryptedStr string) (decryptedStr string, err error) {
	encrypted, err := hex.DecodeString(encryptedStr)
	if err != nil {
		return "", err
	}
	decrypted, err := aesDecryptCFB(encrypted)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
