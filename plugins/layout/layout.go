// Package layout transforms content with Go templates.
package layout

import (
	"bytes"
	"html/template"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
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
func New() Layout {
	return &layout{layoutKey: "Layout", contentKey: "Content", helpers: nil}
}

type layout struct {
	layoutKey  string
	contentKey string
	helpers    template.FuncMap

	inputFiles    []*goldsmith.File
	templateFiles []*goldsmith.File
	mutex         sync.Mutex

	template *template.Template
}

func (layout *layout) LayoutKey(key string) Layout {
	layout.layoutKey = key
	return layout
}

func (layout *layout) ContentKey(key string) Layout {
	layout.contentKey = key
	return layout
}

func (layout *layout) Helpers(helpers template.FuncMap) Layout {
	layout.helpers = helpers
	return layout
}

func (*layout) Name() string {
	return "layout"
}

func (layout *layout) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	layout.template = template.New("").Funcs(layout.helpers)
	return extension.New(".html", ".htm", ".tmpl", ".gohtml"), nil
}

func (layout *layout) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	layout.mutex.Lock()
	defer layout.mutex.Unlock()

	switch inputFile.Ext() {
	case ".html", ".htm":
		if _, ok := inputFile.Meta[layout.layoutKey]; ok {
			var buff bytes.Buffer
			if _, err := inputFile.WriteTo(&buff); err != nil {
				return err
			}

			inputFile.Meta[layout.contentKey] = template.HTML(buff.Bytes())
			layout.inputFiles = append(layout.inputFiles, inputFile)
		} else {
			context.DispatchFile(inputFile)
		}
	case ".tmpl", ".gohtml":
		layout.templateFiles = append(layout.templateFiles, inputFile)
	}

	return nil
}

func (layout *layout) Finalize(context *goldsmith.Context) error {
	for _, templateFile := range layout.templateFiles {
		var buff bytes.Buffer
		if _, err := templateFile.WriteTo(&buff); err != nil {
			return err
		}

		if _, err := layout.template.Parse(string(buff.Bytes())); err != nil {
			return err
		}
	}

	for _, inputFile := range layout.inputFiles {
		name, ok := inputFile.Meta[layout.layoutKey].(string)
		if !ok {
			context.DispatchFile(inputFile)
			continue
		}

		var buff bytes.Buffer
		if err := layout.template.ExecuteTemplate(&buff, name, inputFile); err != nil {
			return err
		}

		outputFile := context.CreateFileFromData(inputFile.Path(), buff.Bytes())
		outputFile.Meta = inputFile.Meta
		context.DispatchFile(outputFile)
	}

	return nil
}
