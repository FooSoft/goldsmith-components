/*
* Copyright (c) 2016 Alex Yatskov <alex@foosoft.net>
* Author: Alex Yatskov <alex@foosoft.net>
*
* Permission is hereby granted, free of charge, to any person obtaining a copy of
* this software and associated documentation files (the "Software"), to deal in
* the Software without restriction, including without limitation the rights to
* use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
* the Software, and to permit persons to whom the Software is furnished to do so,
* subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all
* copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
* FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
* COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
* IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
* CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package abs

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/FooSoft/goldsmith"
)

type namer func(path string, index int) string

type paginate struct {
	key      string
	limit    int
	callback namer
}

func New(key string) *paginate {
	callback := func(path string, index int) string {
		ext := filepath.Ext(path)
		body := strings.TrimSuffix(path, ext)
		return fmt.Sprintf("%s-%d.%s", body, index, ext)
	}

	return &paginate{key, 10, callback}
}

func (p *paginate) Limit(limit int) *paginate {
	p.limit = limit
	return p
}

func (p *paginate) Namer(callback namer) *paginate {
	p.callback = callback
	return p
}

func (*paginate) Name() string {
	return "paginate"
}

func (*paginate) Initialize(ctx goldsmith.Context) ([]string, error) {
	return []string{"**/*.html", "**/*.htm"}, nil
}

func (p *paginate) Process(ctx goldsmith.Context, f goldsmith.File) error {
	defer ctx.DispatchFile(f)

	var buff bytes.Buffer
	if _, err := buff.ReadFrom(f); err != nil {
		return err
	}

	values, ok := f.Value(p.key)
	if !ok {
		return nil
	}

	valuesArr, ok := values.([]interface{})
	if !ok {
		return nil
	}

	count := len(valuesArr)
	if count < p.limit {
		return nil
	}

	f.SetValue(p.key, valuesArr[:p.limit])

	for i := p.limit; i < count; i += p.limit {
		limit := i + p.limit
		if limit > count {
			limit = count
		}

		nf := goldsmith.NewFileFromData(p.callback(f.Path(), i), buff.Bytes())
		nf.CopyValues(f)
		nf.SetValue(p.key, valuesArr[i:limit])

		ctx.DispatchFile(nf)
	}

	return nil
}
