// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"bytes"
	"crypto/x509/pkix"
	"github.com/pingcap-inc/tiem/library/util/fastrand"
	"github.com/pingcap/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunWithRetry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		cnt := 0
		err := RunWithRetry(3, 1, func() (bool, error) {
			cnt++
			if cnt < 2 {
				return true, errors.New("err")
			}
			return true, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 2, cnt)
	})

	t.Run("retry exceeds", func(t *testing.T) {
		t.Parallel()
		cnt := 0
		err := RunWithRetry(3, 1, func() (bool, error) {
			cnt++
			if cnt < 4 {
				return true, errors.New("err")
			}
			return true, nil
		})
		assert.NotNil(t, err)
		assert.Equal(t, 3, cnt)
	})

	t.Run("failed result", func(t *testing.T) {
		t.Parallel()
		cnt := 0
		err := RunWithRetry(3, 1, func() (bool, error) {
			cnt++
			if cnt < 2 {
				return false, errors.New("err")
			}
			return true, nil
		})
		assert.NotNil(t, err)
		assert.Equal(t, 1, cnt)
	})
}

func TestX509NameParseMatch(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", X509NameOnline(pkix.Name{}))

	check := pkix.Name{
		Names: []pkix.AttributeTypeAndValue{
			MockPkixAttribute(Country, "SE"),
			MockPkixAttribute(Province, "Stockholm2"),
			MockPkixAttribute(Locality, "Stockholm"),
			MockPkixAttribute(Organization, "MySQL demo client certificate"),
			MockPkixAttribute(OrganizationalUnit, "testUnit"),
			MockPkixAttribute(CommonName, "client"),
			MockPkixAttribute(Email, "client@example.com"),
		},
	}
	result := "/C=SE/ST=Stockholm2/L=Stockholm/O=MySQL demo client certificate/OU=testUnit/CN=client/emailAddress=client@example.com"
	assert.Equal(t, result, X509NameOnline(check))
}

func TestBasicFuncGetStack(t *testing.T) {
	t.Parallel()
	b := GetStack()
	assert.Less(t, len(b), 4096)
}

func TestBasicFuncWithRecovery(t *testing.T) {
	t.Parallel()
	var recovery interface{}
	WithRecovery(func() {
		panic("test")
	}, func(r interface{}) {
		recovery = r
	})
	assert.Equal(t, "test", recovery)
}

func TestBasicFuncRandomBuf(t *testing.T) {
	t.Parallel()
	buf := fastrand.Buf(5)
	assert.Len(t, buf, 5)
	assert.False(t, bytes.Contains(buf, []byte("$")))
	assert.False(t, bytes.Contains(buf, []byte{0}))
}
