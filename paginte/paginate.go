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
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type namer func(path string, index int) string

type pager struct {
	HasNext bool
	UrlNext string
	HasPrev bool
	UrlPrev string
	Urls    []string
	Count   int
	Index   int
}

type paginate struct {
	key      string
	pagerKey string
	limit    int
	callback namer
	files    []goldsmith.File
	mtx      sync.Mutex
}

func New(key string) *paginate {
	callback := func(path string, index int) string {
		ext := filepath.Ext(path)
		body := strings.TrimSuffix(path, ext)
		return fmt.Sprintf("%s-%d.%s", body, index, ext)
	}

	return &paginate{
		key:      key,
		pagerKey: "Pager",
		limit:    10,
		callback: callback,
	}
}

func (p *paginate) PagerKey(key string) *paginate {
	p.pagerKey = key
	return p
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
	p.mtx.Lock()
	defer p.mtx.Unlock()

	var buff bytes.Buffer
	if _, err := buff.ReadFrom(f); err != nil {
		return err
	}

	values, ok := f.Value(p.key)
	if !ok {
		p.files = append(p.files, f)
		return nil
	}

	valueArr, ok := values.([]interface{})
	if !ok {
		return errors.New("invalid pagination array")
	}

	valueArrLen := len(valueArr)
	pageCount := valueArrLen/p.limit + 1
	pageUrls := []string{f.Path()}
	for i := 1; i < pageCount; i++ {
		pageUrls = append(pageUrls, p.callback(f.Path(), i))
	}

	for i := 0; i < pageCount; i++ {
		pager := pager{
			HasNext: i+1 < pageCount,
			HasPrev: i > 0,
			Urls:    pageUrls,
			Count:   pageCount,
			Index:   i,
		}

		if pager.HasNext {
			pager.UrlPrev = pageUrls[i+1]
		}
		if pager.HasPrev {
			pager.UrlNext = pageUrls[i-1]
		}

		fc := f
		if i > 0 {
			fc = goldsmith.NewFileFromData(p.callback(f.Path(), i), buff.Bytes())
			fc.CopyValues(f)
		}

		indexStart := i * p.limit
		indexEnd := indexStart + p.limit
		if indexEnd > valueArrLen {
			indexEnd = valueArrLen
		}

		fc.SetValue(p.key, valueArr[indexStart:indexEnd])
		fc.SetValue(p.pagerKey, pager)

		p.files = append(p.files, fc)
	}

	return nil
}

func (p *paginate) Finalize(ctx goldsmith.Context) error {
	for _, f := range p.files {
		ctx.DispatchFile(f)
	}

	return nil
}
