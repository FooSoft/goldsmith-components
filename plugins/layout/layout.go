// Package layout transforms content with Go templates.
package layout

import (
	"bytes"
	"html/template"
	"os"
	"sync"
	"time"

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
	var (
		paths   []string
		modTime time.Time
	)

	for _, glob := range globs {
		matches, _ := doublestar.Glob(glob)
		for _, match := range matches {
			matchInfo, err := os.Stat(match)
			if err == nil {
				paths = append(paths, match)
				if matchModTime := matchInfo.ModTime(); matchModTime.Unix() > modTime.Unix() {
					modTime = matchModTime
				}
			}
		}
	}

	return &layoutPlugin{
		layoutKey:  "Layout",
		contentKey: "Content",
		paths:      paths,
		helpers:    nil,
		modTime:    modTime,
	}
}

type layoutPlugin struct {
	layoutKey, contentKey string

	files    []goldsmith.File
	filesMtx sync.Mutex

	paths   []string
	helpers template.FuncMap
	tmpl    *template.Template

	modTime time.Time
}

func (plugin *layoutPlugin) LayoutKey(key string) Layout {
	plugin.layoutKey = key
	return plugin
}

func (plugin *layoutPlugin) ContentKey(key string) Layout {
	plugin.contentKey = key
	return plugin
}

func (plugin *layoutPlugin) Helpers(helpers template.FuncMap) Layout {
	plugin.helpers = helpers
	return plugin
}

func (*layoutPlugin) Name() string {
	return "layout"
}

func (plugin *layoutPlugin) Initialize(context goldsmith.Context) ([]goldsmith.Filter, error) {
	var err error
	if plugin.tmpl, err = template.New("").Funcs(plugin.helpers).ParseFiles(plugin.paths...); err != nil {
		return nil, err
	}

	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (plugin *layoutPlugin) Process(context goldsmith.Context, f goldsmith.File) error {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(f); err != nil {
		return err
	}

	if _, ok := f.Value(plugin.layoutKey); ok {
		f.SetValue(plugin.contentKey, template.HTML(buff.Bytes()))

		plugin.filesMtx.Lock()
		plugin.files = append(plugin.files, f)
		plugin.filesMtx.Unlock()
	} else {
		context.DispatchFile(f)
	}

	return nil
}

func (plugin *layoutPlugin) Finalize(context goldsmith.Context) error {
	for _, f := range plugin.files {
		name, ok := f.Value(plugin.layoutKey)
		if !ok {
			context.DispatchFile(f)
			continue
		}

		nameStr, ok := name.(string)
		if !ok {
			context.DispatchFile(f)
			continue
		}

		var buff bytes.Buffer
		if err := plugin.tmpl.ExecuteTemplate(&buff, nameStr, f); err != nil {
			return err
		}

		modTime := f.ModTime()
		if plugin.modTime.Unix() > modTime.Unix() {
			modTime = plugin.modTime
		}

		nf := goldsmith.NewFileFromData(f.Path(), buff.Bytes(), modTime)
		nf.InheritValues(f)
		context.DispatchFile(nf)
	}

	return nil
}
