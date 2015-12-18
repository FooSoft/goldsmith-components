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

func (*minify) Name() string {
	return "Minify"
}

func (*minify) Accept(f goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(f.Path())) {
	case ".css", ".html", ".htm", ".js", ".svg", ".json", ".xml":
		return true
	default:
		return false
	}
}

func (*minify) Process(ctx goldsmith.Context, f goldsmith.File) error {
	var (
		buff bytes.Buffer
		err  error
	)

	switch m := min.New(); filepath.Ext(strings.ToLower(f.Path())) {
	case ".css":
		err = css.Minify(m, &buff, f, nil)
	case ".html", ".htm":
		err = html.Minify(m, &buff, f, nil)
	case ".js":
		err = js.Minify(m, &buff, f, nil)
	case ".json":
		err = json.Minify(m, &buff, f, nil)
	case ".svg":
		err = svg.Minify(m, &buff, f, nil)
	case ".xml":
		err = xml.Minify(m, &buff, f, nil)
	}

	// f.Rewrite(buff.Bytes())
	return err
}
