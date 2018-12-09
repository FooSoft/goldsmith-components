// Package paginate splits arrays of metadata into standalone pages.
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
	"github.com/FooSoft/goldsmith-components/filters/extension"
)

// The Namer callback allows the user to generate custom filenames for generated pages.
type Namer func(path string, index int) string

// Paginate chainable context.
type Paginate interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor
	goldsmith.Finalizer

	PagerKey(key string) Paginate
	PaginateKey(key string) Paginate
	ItemsPerPage(limit int) Paginate
	Namer(namer Namer) Paginate
	InheritKeys(keys ...string) Paginate
}

// New creates a new instance of the Paginate plugin.
func New(key string) Paginate {
	namer := func(path string, index int) string {
		ext := filepath.Ext(path)
		body := strings.TrimSuffix(path, ext)
		return fmt.Sprintf("%s-%d%s", body, index, ext)
	}

	return &paginate{
		key:          key,
		pagerKey:     "Pager",
		paginateKey:  "Paginate",
		itemsPerPage: 10,
		namer:        namer,
	}
}

type page struct {
	Index int
	Items interface{}
	File  *goldsmith.File

	Next *page
	Prev *page
}

type pager struct {
	PagesAll []page
	PageCurr *page
	Paged    bool
}

type paginate struct {
	key, pagerKey, paginateKey string

	itemsPerPage int
	namer        Namer
	inheritKeys  []string

	files []*goldsmith.File
	mutex sync.Mutex
}

func (p *paginate) PagerKey(key string) Paginate {
	p.pagerKey = key
	return p
}

func (p *paginate) PaginateKey(key string) Paginate {
	p.paginateKey = key
	return p
}

func (p *paginate) ItemsPerPage(limit int) Paginate {
	p.itemsPerPage = limit
	return p
}

func (p *paginate) Namer(namer Namer) Paginate {
	p.namer = namer
	return p
}

func (p *paginate) InheritKeys(keys ...string) Paginate {
	p.inheritKeys = keys
	return p
}

func (*paginate) Name() string {
	return "paginate"
}

func (*paginate) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".html", ".htm"), nil
}

func (p *paginate) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var buff bytes.Buffer
	if _, err := buff.ReadFrom(inputFile); err != nil {
		return err
	}

	paginate, ok := inputFile.Meta[p.paginateKey]
	if !ok {
		p.files = append(p.files, inputFile)
		return nil
	}

	if paginateBool, ok := paginate.(bool); !ok || !paginateBool {
		return errors.New("invalid pagination setting")
	}

	values, ok := inputFile.Meta[p.key]
	if !ok {
		p.files = append(p.files, inputFile)
		return nil
	}

	valueCount, err := sliceLength(values)
	if err != nil {
		return err
	}

	pageCount := valueCount / p.itemsPerPage
	if valueCount%p.itemsPerPage > 0 {
		pageCount++
	}

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

		indexStart := i * p.itemsPerPage
		indexEnd := indexStart + p.itemsPerPage
		if indexEnd > valueCount {
			indexEnd = valueCount
		}

		if page.Items, err = sliceCrop(values, indexStart, indexEnd); err != nil {
			return err
		}

		if i == 0 {
			page.File = inputFile
		} else {
			page.File = context.CreateFileFromData(p.namer(inputFile.Path(), page.Index), buff.Bytes())
			if len(p.inheritKeys) == 0 {
				page.File.Meta = inputFile.Meta
			} else {
				for _, key := range p.inheritKeys {
					if value, ok := inputFile.Meta[key]; ok {
						page.File.Meta[key] = value
					}
				}
			}
		}

		page.File.Meta[p.pagerKey] = pager{pages, page, pageCount > 1}
		p.files = append(p.files, page.File)
	}

	return nil
}

func (p *paginate) Finalize(ctx *goldsmith.Context) error {
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
