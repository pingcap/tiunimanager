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
package linux_test

import (
	"github.com/pingcap/tiunimanager/util/sys/linux"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetOSVersion(t *testing.T) {
	t.Parallel()

	osRelease, err := linux.OSVersion()
	require.NoError(t, err)
	require.NotEmpty(t, osRelease)
}
