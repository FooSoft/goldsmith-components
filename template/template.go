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

package template

import (
	"bytes"
	tmpl "html/template"
	"path"
	"strings"

	"github.com/FooSoft/goldsmith"
)

type template struct {
	tmpl *tmpl.Template
	def  string
}

func New(glob, def string) *template {
	t, err := tmpl.ParseGlob(glob)
	if err != nil {
		panic(err)
	}

	return &template{t, def}
}

func (t *template) TaskSingle(ctx goldsmith.Context, file goldsmith.File) goldsmith.File {
	ext := strings.ToLower(path.Ext(file.Path()))
	if ext != ".html" {
		return file
	}

	name := file.Property("template", t.def).(string)
	params := make(map[string]interface{})
	params["Content"] = tmpl.HTML(file.Data())

	var buff bytes.Buffer
	if err := t.tmpl.ExecuteTemplate(&buff, name, params); err != nil {
		file.SetError(err)
	} else {
		file.SetData(buff.Bytes())
	}

	return file
}
