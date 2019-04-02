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

type Namer func(path string, index int) string
type Lister func(file *goldsmith.File) interface{}

type Page struct {
	Index int
	Items interface{}
	File  *goldsmith.File

	Next *Page
	Prev *Page
}

type Pager struct {
	AllPages    []Page
	CurrentPage *Page
	Paged       bool
}

type Paginate struct {
	pagerKey  string
	enableKey string

	namer        Namer
	lister       Lister
	inheritKeys  []string
	itemsPerPage int

	files []*goldsmith.File
	mutex sync.Mutex
}

// New creates a new instance of the Paginate plugin.
func New(lister Lister) *Paginate {
	namer := func(path string, index int) string {
		ext := filepath.Ext(path)
		body := strings.TrimSuffix(path, ext)
		return fmt.Sprintf("%s-%d%s", body, index, ext)
	}

	return &Paginate{
		pagerKey:     "Pager",
		enableKey:    "PagerEnable",
		namer:        namer,
		lister:       lister,
		itemsPerPage: 10,
	}
}

func (plugin *Paginate) PagerKey(key string) *Paginate {
	plugin.pagerKey = key
	return plugin
}

func (plugin *Paginate) ItemsPerPage(limit int) *Paginate {
	plugin.itemsPerPage = limit
	return plugin
}

func (plugin *Paginate) Namer(namer Namer) *Paginate {
	plugin.namer = namer
	return plugin
}

func (p *Paginate) InheritKeys(keys ...string) *Paginate {
	p.inheritKeys = keys
	return p
}

func (*Paginate) Name() string {
	return "paginate"
}

func (*Paginate) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".html", ".htm"), nil
}

func (plugin *Paginate) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	plugin.mutex.Lock()
	defer plugin.mutex.Unlock()

	enabled, err := plugin.isEnabledForFile(inputFile)
	if err != nil {
		return err
	}

	if !enabled {
		plugin.files = append(plugin.files, inputFile)
		return nil
	}

	var buff bytes.Buffer
	if _, err := buff.ReadFrom(inputFile); err != nil {
		return err
	}

	values := plugin.lister(inputFile)
	valueCount, err := sliceLength(values)
	if err != nil {
		return err
	}

	pageCount := valueCount / plugin.itemsPerPage
	if valueCount%plugin.itemsPerPage > 0 {
		pageCount++
	}

	pages := make([]Page, pageCount, pageCount)
	for i := 0; i < pageCount; i++ {
		page := &pages[i]
		page.Index = i + 1

		if i > 0 {
			page.Prev = &pages[i-1]
		}
		if i+1 < pageCount {
			page.Next = &pages[i+1]
		}

		var (
			indexStart = i * plugin.itemsPerPage
			indexEnd   = indexStart + plugin.itemsPerPage
		)

		if indexEnd > valueCount {
			indexEnd = valueCount
		}

		if page.Items, err = sliceCrop(values, indexStart, indexEnd); err != nil {
			return err
		}

		if i == 0 {
			page.File = inputFile
		} else {
			page.File = context.CreateFileFromData(plugin.namer(inputFile.Path(), page.Index), buff.Bytes())
			if len(plugin.inheritKeys) == 0 {
				for key, value := range inputFile.Meta {
					page.File.Meta[key] = value
				}
			} else {
				for _, key := range plugin.inheritKeys {
					if value, ok := inputFile.Meta[key]; ok {
						page.File.Meta[key] = value
					}
				}
			}
		}

		page.File.Meta[plugin.pagerKey] = Pager{
			AllPages:    pages,
			CurrentPage: page,
			Paged:       pageCount > 1,
		}

		plugin.files = append(plugin.files, page.File)
	}

	return nil
}

func (plugin *Paginate) Finalize(ctx *goldsmith.Context) error {
	for _, f := range plugin.files {
		ctx.DispatchFile(f)
	}

	return nil
}

func (plugin *Paginate) isEnabledForFile(file *goldsmith.File) (bool, error) {
	enableRaw, ok := file.Meta[plugin.enableKey]
	if !ok {
		return false, nil
	}

	enable, ok := enableRaw.(bool)
	if !ok {
		return false, errors.New("invalid paginate enable setting")
	}

	return enable, nil
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
