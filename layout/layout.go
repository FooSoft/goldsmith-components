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
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/bmatcuk/doublestar"
)

type layout struct {
	layoutKey, contentKey string

	files    []goldsmith.File
	filesMtx sync.Mutex

	paths   []string
	helpers template.FuncMap
	tmpl    *template.Template
}

func New(globs ...string) *layout {
	var paths []string
	for _, glob := range globs {
		matches, _ := doublestar.Glob(glob)
		paths = append(paths, matches...)
	}

	return &layout{
		layoutKey:  "Layout",
		contentKey: "Content",
		paths:      paths,
		helpers:    nil,
	}
}

func (lay *layout) LayoutKey(key string) *layout {
	lay.layoutKey = key
	return lay
}

func (lay *layout) ContentKey(key string) *layout {
	lay.contentKey = key
	return lay
}

func (lay *layout) Helpers(helpers template.FuncMap) *layout {
	lay.helpers = helpers
	return lay
}

func (*layout) Name() string {
	return "layout"
}

func (lay *layout) Initialize(ctx goldsmith.Context) ([]string, error) {
	var err error
	lay.tmpl, err = template.New("").Funcs(lay.helpers).ParseFiles(lay.paths...)
	return []string{"**/*.html", "**/*.htm"}, err
}

func (lay *layout) Process(ctx goldsmith.Context, f goldsmith.File) error {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(f); err != nil {
		return err
	}

	if _, ok := f.Value(lay.layoutKey); ok {
		f.SetValue(lay.contentKey, template.HTML(buff.Bytes()))

		lay.filesMtx.Lock()
		lay.files = append(lay.files, f)
		lay.filesMtx.Unlock()
	} else {
		ctx.DispatchFile(f)
	}

	return nil
}

func (lay *layout) Finalize(ctx goldsmith.Context) error {
	for _, f := range lay.files {
		name, ok := f.Value(lay.layoutKey)
		if !ok {
			ctx.DispatchFile(f)
			continue
		}

		nameStr, ok := name.(string)
		if !ok {
			ctx.DispatchFile(f)
			continue
		}

		var buff bytes.Buffer
		if err := lay.tmpl.ExecuteTemplate(&buff, nameStr, f); err != nil {
			return err
		}

		nf := goldsmith.NewFileFromData(f.Path(), buff.Bytes())
		nf.CopyValues(f)
		ctx.DispatchFile(nf)
	}

	return nil
}
