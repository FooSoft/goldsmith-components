/*
 * Copyright (c) 2015 Alex Yatskov <alex@foosoft.net>
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

package layout

import (
	"bytes"
	"html/template"
	"path/filepath"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type layout struct {
	tmpl *template.Template
	def  string
}

func New(glob, def string, funcs template.FuncMap) (goldsmith.Chainer, error) {
	tmpl, err := template.New("").Funcs(funcs).ParseGlob(glob)
	if err != nil {
		return nil, err
	}

	return &layout{tmpl, def}, nil
}

func (*layout) Filter(path string) bool {
	if ext := filepath.Ext(path); ext == ".html" {
		return false
	}

	return true
}

func (t *layout) Chain(ctx goldsmith.Context, input, output chan *goldsmith.File) {
	defer close(output)

	var wg sync.WaitGroup
	for file := range input {
		wg.Add(1)
		go func(f *goldsmith.File) {
			defer wg.Done()

			name, ok := f.Meta["layout"]
			if !ok {
				name = t.def
			}

			f.Meta["content"] = template.HTML(f.Buff.Bytes())

			var buff bytes.Buffer
			if f.Err = t.tmpl.ExecuteTemplate(&buff, name.(string), f); f.Err == nil {
				f.Buff = &buff
			}

			output <- f
		}(file)
	}

	wg.Wait()
}
