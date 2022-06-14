/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package errors

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestOfNullable(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want *Optional
	}{
		{"normal", args{err: NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "")}, &Optional{
			last:  NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, ""),
			broken: true,
			allErrors: []error{ NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "")},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OfNullable(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OfNullable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptional_BreakIf(t *testing.T) {
	type fields struct {
		err    error
		broken bool
	}
	type args struct {
		executor func() error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Optional
	}{
		{"with error", fields{broken: false}, args{executor: func() error {
			return NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "")
		}}, &Optional{
			last:   NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, ""),
			broken: true,
			allErrors: []error{ NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "")},
		}},
		{"without error", fields{broken: false}, args{executor: func() error {
			return nil
		}}, &Optional{
			last:   nil,
			broken: false,
			allErrors: nil,

		}},
		{"broken", fields{broken: true}, args{executor: func() error {
			return nil
		}}, &Optional{
			last:   nil,
			broken: true,
			allErrors: nil,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Optional{
				last:   tt.fields.err,
				broken: tt.fields.broken,
			}
			if got := p.BreakIf(tt.args.executor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BreakIf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptional_ContinueIf(t *testing.T) {
	type fields struct {
		err    error
		broken bool
	}
	type args struct {
		executor func() error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Optional
	}{
		{"with error", fields{}, args{executor: func() error {
			return NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "")
		}}, &Optional{
			last:   NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, ""),
			broken: false,
			allErrors: []error{ NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "")},

		}},
		{"without error", fields{}, args{executor: func() error {
			return nil
		}}, &Optional{
			last:   nil,
			broken: false,
			allErrors: nil,
		}},
		{"broken", fields{broken: true}, args{executor: func() error {
			return nil
		}}, &Optional{
			last:   nil,
			broken: true,
			allErrors: nil,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Optional{
				last:   tt.fields.err,
				broken: tt.fields.broken,
			}
			if got := p.ContinueIf(tt.args.executor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContinueIf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptional_Else(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		a := 0
		optional := &Optional{
			last: NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, ""),
		}
		optional.Else(func() {
			a = 5
		})
		assert.Equal(t, 0, a)
	})
	t.Run("without error", func(t *testing.T) {
		a := 0
		optional := &Optional{
			last: nil,
		}
		optional.Else(func() {
			a = 5
		})
		assert.Equal(t, 5, a)
	})
}

func TestOptional_Handle(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		a := ""
		optional := &Optional{
			last: NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "aaa"),
		}
		optional.If(func(err error) {
			a = err.Error()
		})
		assert.Contains(t, a, "[10000]")
	})
	t.Run("without error", func(t *testing.T) {
		a := ""
		optional := &Optional{
			last: nil,
		}
		optional.If(func(err error) {
			a = err.Error()
		})
		assert.Equal(t, "", a)
	})
}

func TestOptional_Map(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		assert.Contains(t, OfNullable(NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "aaa")).
			Map(func(err error) error {
				return NewError(TIUNIMANAGER_MARSHAL_ERROR, "bbb")
			}).Present().Error(), "[10004]")
	})
	t.Run("without error", func(t *testing.T) {
		assert.NoError(t, OfNullable(nil).
			Map(func(err error) error {
				return NewError(TIUNIMANAGER_MARSHAL_ERROR, "bbb")
			}).
			Present())
	})
}

func TestOptional_If(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		a := ""
		b := 0
		optional := &Optional{
			last: NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "aaa"),
		}
		optional.IfElse(func(err error) {
			a = err.Error()
		}, func() {
			b = 5
		})
		assert.Contains(t, a, "[10000]")
		assert.Contains(t, a, "aaa")
		assert.Equal(t, 0, b)
	})
	t.Run("without error", func(t *testing.T) {
		a := ""
		b := 0
		optional := &Optional{
			last: nil,
		}
		optional.IfElse(func(err error) {
			a = err.Error()
		}, func() {
			b = 5
		})
		assert.Equal(t, "", a)
		assert.Equal(t, 5, b)
	})
}

func TestOptional_Present(t *testing.T) {
	type fields struct {
		err    error
		broken bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"error", fields{err: NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "")}, true},
		{"error", fields{err: nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Optional{
				last:   tt.fields.err,
				broken: tt.fields.broken,
			}
			if err := p.Present(); (err != nil) != tt.wantErr {
				t.Errorf("Present() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


func TestOptional_IfPresent(t *testing.T) {
	assert.True(t, OfNullable(NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "")).IfPresent())
	assert.False(t, OfNullable(NewError(TIUNIMANAGER_UNRECOGNIZED_ERROR, "")).IfNil())
	assert.True(t, OfNullable(nil).IfNil())
	assert.False(t, OfNullable(nil).IfPresent())

}

