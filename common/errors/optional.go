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

type Optional struct {
	err       error
	broken    bool
}

func OfNullable(err error) *Optional {
	return &Optional{
		err:    err,
		broken: false,
	}
}

func (p *Optional) BreakIf(executor func() error) *Optional {
	if p.broken {
		return p
	}
	if err := executor(); err != nil {
		p.err = err
		p.broken = true
	}

	return p
}

func (p *Optional) ContinueIf(executor func() error) *Optional {
	if p.broken {
		return p
	}
	if err := executor(); err != nil {
		p.err = err
	}
	return p
}

func (p *Optional) If(handle func(err error)) *Optional {
	if p.err != nil && handle != nil {
		handle(p.err)
	}
	return p
}

func (p *Optional) Else(handle func()) *Optional {
	if p.err == nil && handle != nil {
		handle()
	}
	return p
}

func (p *Optional) Handle(errHandler func(err error), nilHandler func()) *Optional {
	return p.If(errHandler).Else(nilHandler)
}

func (p *Optional) Present() error {
	return p.err
}
