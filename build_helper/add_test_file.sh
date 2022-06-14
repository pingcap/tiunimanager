#
# Copyright (c)  2021 PingCAP, Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#

find . -name "*.go" | grep -vE "proto|docs|library/util" | while read fname; do
    fpath=`dirname $fname`
    firstline=$(sed -n -e '/^package/p' $fname)
#    if [ ! -f "$fpath/main_test.go" ]; then
    if ls $fpath/*_test.go 1> /dev/null 2>&1; then
      echo "no need to add main_test.go file in $fpath"
    else
      echo $firstline > $fpath/main_test.go
      echo "import \"testing\"" >> $fpath/main_test.go
      echo "func Test_Main(t *testing.T) {}" >> $fpath/main_test.go
    fi
done