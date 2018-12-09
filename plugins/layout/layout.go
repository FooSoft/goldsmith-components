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
		layoutKey:       "Layout",
		contentKey:      "Content",
		templatePaths:   paths,
		templateHelpers: nil,
	}
}

type layout struct {
	layoutKey  string
	contentKey string

	files      []*goldsmith.File
	filesMutex sync.Mutex

	templatePaths   []string
	templateHelpers template.FuncMap
	template        *template.Template
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
	lay.templateHelpers = helpers
	return lay
}

func (*layout) Name() string {
	return "layout"
}

func (lay *layout) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	var err error
	if lay.template, err = template.New("").Funcs(lay.templateHelpers).ParseFiles(lay.templatePaths...); err != nil {
		return nil, err
	}

	return extension.New(".html", ".htm"), nil
}

func (lay *layout) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(inputFile); err != nil {
		return err
	}

	if _, ok := inputFile.Meta[lay.layoutKey]; ok {
		inputFile.Meta[lay.contentKey] = template.HTML(buff.Bytes())

		lay.filesMutex.Lock()
		lay.files = append(lay.files, inputFile)
		lay.filesMutex.Unlock()
	} else {
		context.DispatchFile(inputFile)
	}

	return nil
}

func (lay *layout) Finalize(context *goldsmith.Context) error {
	for _, inputFile := range lay.files {
		name, ok := inputFile.Meta[lay.layoutKey]
		if !ok {
			context.DispatchFile(inputFile)
			continue
		}

		nameStr, ok := name.(string)
		if !ok {
			context.DispatchFile(inputFile)
			continue
		}

		var buff bytes.Buffer
		if err := lay.template.ExecuteTemplate(&buff, nameStr, inputFile); err != nil {
			return err
		}

		outputFile := context.CreateFileFromData(inputFile.Path(), buff.Bytes())
		outputFile.Meta = inputFile.Meta
		context.DispatchFile(outputFile)
	}

	return nil
}
