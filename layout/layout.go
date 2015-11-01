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
	"path"
	"strings"

	"github.com/FooSoft/goldsmith"
)

type layout struct {
	tmpl *template.Template
	def  string
}

func New(glob, def string) goldsmith.Context {
	t, err := template.ParseGlob(glob)
	if err != nil {
		return goldsmith.Context{nil, err}
	}

	return goldsmith.Context{&layout{t, def}, nil}
}

func (t *layout) TaskSingle(ctx goldsmith.Context, file goldsmith.File) goldsmith.File {
	ext := strings.ToLower(path.Ext(file.Path))
	if ext != ".html" {
		return file
	}

	name, ok := file.Meta["Template"]
	if !ok {
		name = t.def
	}

	file.Meta["Content"] = template.HTML(file.Buff.Bytes())

	var buff bytes.Buffer
	if file.Err = t.tmpl.ExecuteTemplate(&buff, name.(string), file.Meta); file.Err == nil {
		file.Buff = buff
	}

	return file
}
