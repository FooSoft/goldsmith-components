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

package minify

import (
	"bytes"
	"path/filepath"
	"strings"

	min "github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"

	"github.com/FooSoft/goldsmith"
)

type minify struct {
}

func New() goldsmith.Plugin {
	return new(minify)
}

func (*minify) Accept(file *goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(file.Path)) {
	case ".css", ".html", ".htm", ".js", ".svg", ".json", ".xml":
		return true
	default:
		return false
	}
}

func (*minify) Process(ctx goldsmith.Context, file *goldsmith.File) bool {
	var buff bytes.Buffer

	switch m := min.New(); filepath.Ext(strings.ToLower(file.Path)) {
	case ".css":
		file.Err = css.Minify(m, &buff, &file.Buff, nil)
	case ".html", ".htm":
		file.Err = html.Minify(m, &buff, &file.Buff, nil)
	case ".js":
		file.Err = js.Minify(m, &buff, &file.Buff, nil)
	case ".json":
		file.Err = json.Minify(m, &buff, &file.Buff, nil)
	case ".svg":
		file.Err = svg.Minify(m, &buff, &file.Buff, nil)
	case ".xml":
		file.Err = xml.Minify(m, &buff, &file.Buff, nil)
	}

	file.Buff = buff
	return true
}
