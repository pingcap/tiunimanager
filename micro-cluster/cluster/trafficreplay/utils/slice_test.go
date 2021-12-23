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
 * @File:  slice_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/10 12:01
 */

package utils
import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUtil_Sort2DSlice_Int (t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,1,2,3,4,5)
	row2 := make([]interface{},0)
	row2 = append(row2,2,3,4,5,6)
	row3 :=make([]interface{},0)
	row3 =append(row3,3,4,5,6,7)
	row4 := make([]interface{},0)
	row4= append(row4,4,5,6,7,8)
	a=append(a,row2,row3,row4,row1)

	SortSlice(a)

	ast:= assert.New(t)

	ast.Equal(a[0][0],1)
	ast.Equal(a[1][0],2)
	ast.Equal(a[2][0],3)
	ast.Equal(a[3][0],4)
}

func TestUtil_Sort2DSlice_Uint (t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,uint(1),uint(2),uint(3),uint(4),uint(5))
	row2 := make([]interface{},0)
	row2 = append(row2,uint(2),uint(3),uint(4),uint(5),uint(6))
	row3 :=make([]interface{},0)
	row3 =append(row3,uint(3),uint(4),uint(5),uint(6),uint(7))
	row4 := make([]interface{},0)
	row4= append(row4,uint(4),uint(5),uint(6),uint(7),uint(8))
	a=append(a,row2,row3,row4,row1)

	SortSlice(a)

	ast:= assert.New(t)

	ast.Equal(a[0][0],uint(1))
	ast.Equal(a[1][0],uint(2))
	ast.Equal(a[2][0],uint(3))
	ast.Equal(a[3][0],uint(4))
}

func TestUtil_Sort2DSlice_Int32 (t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,int32(1),int32(2),int32(3),int32(4),int32(5))
	row2 := make([]interface{},0)
	row2 = append(row2,int32(2),int32(3),int32(4),int32(5),int32(6))
	row3 :=make([]interface{},0)
	row3 =append(row3,int32(3),int32(4),int32(5),int32(6),int32(7))
	row4 := make([]interface{},0)
	row4= append(row4,int32(4),int32(5),int32(6),int32(7),int32(8))
	a=append(a,row2,row3,row4,row1)

	SortSlice(a)

	ast:= assert.New(t)

	ast.Equal(a[0][0],int32(1))
	ast.Equal(a[1][0],int32(2))
	ast.Equal(a[2][0],int32(3))
	ast.Equal(a[3][0],int32(4))
}

func TestUtil_Sort2DSlice_Uint32 (t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,uint32(1),uint32(2),uint32(3),uint32(4),uint32(5))
	row2 := make([]interface{},0)
	row2 = append(row2,uint32(2),uint32(3),uint32(4),uint32(5),uint32(6))
	row3 :=make([]interface{},0)
	row3 =append(row3,uint32(3),uint32(4),uint32(5),uint32(6),uint32(7))
	row4 := make([]interface{},0)
	row4= append(row4,uint32(4),uint32(5),uint32(6),uint32(7),uint32(8))
	a=append(a,row2,row3,row4,row1)

	SortSlice(a)

	ast:= assert.New(t)

	ast.Equal(a[0][0],uint32(1))
	ast.Equal(a[1][0],uint32(2))
	ast.Equal(a[2][0],uint32(3))
	ast.Equal(a[3][0],uint32(4))
}


func TestUtil_Sort2DSlice_Int64 (t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,int64(1),int64(2),int64(3),int64(4),int64(5))
	row2 := make([]interface{},0)
	row2 = append(row2,int64(2),int64(3),int64(4),int64(5),int64(6))
	row3 :=make([]interface{},0)
	row3 =append(row3,int64(3),int64(4),int64(5),int64(6),int64(7))
	row4 := make([]interface{},0)
	row4= append(row4,int64(4),int64(5),int64(6),int64(7),int64(8))
	a=append(a,row2,row3,row4,row1)

	SortSlice(a)

	ast:= assert.New(t)

	ast.Equal(a[0][0],int64(1))
	ast.Equal(a[1][0],int64(2))
	ast.Equal(a[2][0],int64(3))
	ast.Equal(a[3][0],int64(4))
}


func TestUtil_Sort2DSlice_Uint64 (t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,uint64(1),uint64(2),uint64(3),uint64(4),uint64(5))
	row2 := make([]interface{},0)
	row2 = append(row2,uint64(2),uint64(3),uint64(4),uint64(5),uint64(6))
	row3 :=make([]interface{},0)
	row3 =append(row3,uint64(3),uint64(4),uint64(5),uint64(6),uint64(7))
	row4 := make([]interface{},0)
	row4= append(row4,uint64(4),uint64(5),uint64(6),uint64(7),uint64(8))
	a=append(a,row2,row3,row4,row1)

	SortSlice(a)

	ast:= assert.New(t)

	ast.Equal(a[0][0],uint64(1))
	ast.Equal(a[1][0],uint64(2))
	ast.Equal(a[2][0],uint64(3))
	ast.Equal(a[3][0],uint64(4))
}


func TestUtil_Sort2DSlice_String (t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,"a","b","c","d","e")
	row2 := make([]interface{},0)
	row2 = append(row2,"b","c","d","e","f")
	row3 :=make([]interface{},0)
	row3 =append(row3,"c","d","e","f","g")
	row4 := make([]interface{},0)
	row4= append(row4,"d","e","f","g","h")
	a=append(a,row2,row3,row4,row1)

	SortSlice(a)

	ast:= assert.New(t)

	ast.Equal(a[0][0],"a")
	ast.Equal(a[1][0],"b")
	ast.Equal(a[2][0],"c")
	ast.Equal(a[3][0],"d")
}

//The type sorting algorithm will not take effect if it is not supported
func TestUtil_Sort2DSlice_Byte (t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,uint16(1),uint16(2),uint16(3),uint16(4),uint16(5))
	row2 := make([]interface{},0)
	row2 = append(row2,uint16(2),uint16(3),uint16(4),uint16(5),uint16(6))
	row3 :=make([]interface{},0)
	row3 =append(row3,uint16(3),uint16(4),uint16(5),uint16(6),uint16(7))
	row4 := make([]interface{},0)
	row4= append(row4,uint16(4),uint16(5),uint16(6),uint16(7),uint16(8))
	a=append(a,row2,row3,row4,row1)

	SortSlice(a)

	ast:= assert.New(t)


	ast.Equal(a[0][0],uint16(2))
	ast.Equal(a[1][0],uint16(3))
	ast.Equal(a[2][0],uint16(4))
	ast.Equal(a[3][0],uint16(1))
}

func TestUtil_Sore2DSlice_Row1_Nil(t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,nil)
	row2 := make([]interface{},0)
	row2 = append(row2,3)

	a = append(a,row1,row2)

	SortSlice(a)
}

func TestUtil_Sore2DSlice_Row2_Nil(t *testing.T){

	a :=make([][]interface{},0)
	row1 := make([]interface{} ,0)
	row1 = append(row1,1)
	row2 := make([]interface{},0)
	row2 = append(row2,nil)

	a = append(a,row1,row2)

	SortSlice(a)
}