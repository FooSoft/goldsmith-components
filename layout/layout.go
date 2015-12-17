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
	"strings"

	"github.com/FooSoft/goldsmith"
)

type layout struct {
	srcKey, dstKey string
	defVal         string
	paths          []string
	funcs          template.FuncMap
	tmpl           *template.Template
}

func New(paths []string, srcKey, dstKey, defVal string, funcs template.FuncMap) goldsmith.Plugin {
	return &layout{
		srcKey: srcKey,
		dstKey: dstKey,
		defVal: defVal,
		paths:  paths,
		funcs:  funcs,
	}
}

func (*layout) Accept(file *goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(file.Path)) {
	case ".html", ".htm":
		return true
	default:
		return false
	}
}

func (t *layout) Initialize(ctx goldsmith.Context) (err error) {
	t.tmpl, err = template.New("").Funcs(t.funcs).ParseFiles(t.paths...)
	return
}

func (t *layout) Process(ctx goldsmith.Context, file *goldsmith.File) bool {
	name, ok := file.Meta[t.srcKey]
	if !ok {
		name = t.defVal
	}

	nameStr, ok := name.(string)
	if !ok {
		name = t.defVal
	}

	file.Meta[t.dstKey] = template.HTML(file.Buff.Bytes())

	var buff bytes.Buffer
	if file.Err = t.tmpl.ExecuteTemplate(&buff, nameStr, file); file.Err == nil {
		file.Buff = buff
	}

	return true
}
