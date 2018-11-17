// Package layout transforms content with Go templates.
package layout

import (
	"bytes"
	"html/template"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/bmatcuk/doublestar"
)

// Layout chainable context.
type Layout interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor
	goldsmith.Finalizer

	// LayoutKey sets the metadata key used to access the layout identifier (default: "Layout").
	LayoutKey(layoutKey string) Layout

	// ContentKey sets the metadata key used to access the source content (default: "Content").
	ContentKey(key string) Layout

	// Helpers sets the function map used to lookup template helper functions.
	Helpers(helpers template.FuncMap) Layout
}

// New creates a new instance of the Layout plugin.
func New(globs ...string) Layout {
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

type layout struct {
	layoutKey, contentKey string

	files    []*goldsmith.File
	filesMtx sync.Mutex

	paths   []string
	helpers template.FuncMap
	tmpl    *template.Template
}

func (lay *layout) LayoutKey(key string) Layout {
	lay.layoutKey = key
	return lay
}

func (lay *layout) ContentKey(key string) Layout {
	lay.contentKey = key
	return lay
}

func (lay *layout) Helpers(helpers template.FuncMap) Layout {
	lay.helpers = helpers
	return lay
}

func (*layout) Name() string {
	return "layout"
}

func (lay *layout) Initialize(ctx *goldsmith.Context) ([]goldsmith.Filter, error) {
	var err error
	if lay.tmpl, err = template.New("").Funcs(lay.helpers).ParseFiles(lay.paths...); err != nil {
		return nil, err
	}

	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (lay *layout) Process(ctx *goldsmith.Context, f *goldsmith.File) error {
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

func (lay *layout) Finalize(ctx *goldsmith.Context) error {
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
		nf.InheritValues(f)
		ctx.DispatchFile(nf)
	}

	return nil
}
