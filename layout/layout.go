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
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/bmatcuk/doublestar"
)

type layout struct {
	srcKey, dstKey string
	defVal         string

	files    []goldsmith.File
	filesMtx sync.Mutex

	paths []string
	funcs template.FuncMap
	tmpl  *template.Template
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

func NewGlob(glob, srcKey, dstKey, defVal string, funcs template.FuncMap) goldsmith.Plugin {
	paths, _ := doublestar.Glob(glob)
	return New(paths, srcKey, dstKey, defVal, funcs)
}

func (t *layout) Initialize(ctx goldsmith.Context) (err error) {
	t.tmpl, err = template.New("").Funcs(t.funcs).ParseFiles(t.paths...)
	return
}

func (*layout) Accept(ctx goldsmith.Context, f goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(f.Path())) {
	case ".html", ".htm":
		return true
	default:
		return false
	}
}

func (t *layout) Process(ctx goldsmith.Context, f goldsmith.File) error {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(f); err != nil {
		return err
	}

	f.SetValue(t.dstKey, template.HTML(buff.Bytes()))

	t.filesMtx.Lock()
	t.files = append(t.files, f)
	t.filesMtx.Unlock()

	return nil
}

func (t *layout) Finalize(ctx goldsmith.Context) error {
	for _, f := range t.files {
		name, ok := f.Value(t.srcKey)
		if !ok {
			name = t.defVal
		}

		nameStr, ok := name.(string)
		if !ok {
			nameStr = t.defVal
		}

		var buff bytes.Buffer
		if err := t.tmpl.ExecuteTemplate(&buff, nameStr, f); err != nil {
			return err
		}

		nf := goldsmith.NewFileFromData(f.Path(), buff.Bytes())
		nf.CopyValues(f)
		ctx.DispatchFile(nf)
	}

	return nil
}
