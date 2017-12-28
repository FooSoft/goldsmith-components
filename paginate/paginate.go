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

package paginate

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type namer func(path string, index int) string

type page struct {
	Index int
	Items interface{}
	File  goldsmith.File

	Next *page
	Prev *page
}

type pager struct {
	PagesAll []page
	PageCurr *page
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
		return fmt.Sprintf("%s-%d%s", body, index, ext)
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

	valueCount, err := sliceLength(values)
	if err != nil {
		return err
	}

	pageCount := valueCount/p.limit + 1
	pages := make([]page, pageCount, pageCount)

	for i := 0; i < pageCount; i++ {
		page := &pages[i]
		page.Index = i + 1

		if i > 0 {
			page.Prev = &pages[i-1]
		}
		if i+1 < pageCount {
			page.Next = &pages[i+1]
		}

		indexStart := i * p.limit
		indexEnd := indexStart + p.limit
		if indexEnd > valueCount {
			indexEnd = valueCount
		}

		if page.Items, err = sliceCrop(values, indexStart, indexEnd); err != nil {
			return err
		}

		if i == 0 {
			page.File = f
		} else {
			page.File = goldsmith.NewFileFromData(p.callback(f.Path(), page.Index), buff.Bytes())
			page.File.CopyValues(f)
		}
		page.File.SetValue(p.pagerKey, pager{pages, page})

		p.files = append(p.files, page.File)
	}

	return nil
}

func (p *paginate) Finalize(ctx goldsmith.Context) error {
	for _, f := range p.files {
		ctx.DispatchFile(f)
	}

	return nil
}

func sliceLength(slice interface{}) (int, error) {
	sliceVal := reflect.Indirect(reflect.ValueOf(slice))
	if sliceVal.Kind() != reflect.Slice {
		return -1, errors.New("invalid slice")
	}

	return sliceVal.Len(), nil

}

func sliceCrop(slice interface{}, start, end int) (interface{}, error) {
	sliceVal := reflect.Indirect(reflect.ValueOf(slice))
	if sliceVal.Kind() != reflect.Slice {
		return nil, errors.New("invalid slice")
	}
	if start < 0 || start > end {
		return nil, errors.New("invalid slice range")
	}

	sliceValNew := reflect.Indirect(reflect.New(sliceVal.Type()))
	for i := start; i < end; i++ {
		sliceElemNew := sliceVal.Index(i)
		sliceValNew.Set(reflect.Append(sliceValNew, sliceElemNew))
	}

	return sliceValNew.Interface(), nil
}
