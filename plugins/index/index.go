// Package index creates pages for displaying directory listings.
package index

import (
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/FooSoft/goldsmith"
)

// Index chainable plugin context.
type Index interface {
	goldsmith.Plugin
	goldsmith.Processor
	goldsmith.Finalizer

	// IndexFilename sets the name of the file to be created as the directory index (default: "index.html").
	IndexFilename(filename string) Index

	// FilesKey sets the metadata key used to access the files in the current directory (default: "Files").
	FilesKey(filename string) Index
}

// New creates a new instance of the Index plugin.
func New(meta map[string]interface{}) Index {
	return &indexPlugin{
		filename:     "index.html",
		filesKey:     "Files",
		meta:         meta,
		pathsHandled: make(map[string]bool),
		dirIndices:   make(map[string]*dirIndex),
	}
}

type indexPlugin struct {
	filename string
	filesKey string
	meta     map[string]interface{}

	dirIndices   map[string]*dirIndex
	pathsHandled map[string]bool

	mutex sync.Mutex
}

func (plugin *indexPlugin) IndexFilename(filename string) Index {
	plugin.filename = filename
	return plugin
}

func (plugin *indexPlugin) FilesKey(key string) Index {
	plugin.filesKey = key
	return plugin
}

func (*indexPlugin) Name() string {
	return "index"
}

func (plugin *indexPlugin) Process(context goldsmith.Context, f goldsmith.File) error {
	plugin.mutex.Lock()
	defer plugin.mutex.Unlock()

	currPath := f.Path()
	currIsLeaf := true

	for {
		if handled, _ := plugin.pathsHandled[currPath]; handled {
			break
		}

		plugin.pathsHandled[currPath] = true

		currDir := path.Dir(currPath)
		currBase := path.Base(currPath)

		currDirIndex, ok := plugin.dirIndices[currDir]
		if !ok {
			currDirIndex = &dirIndex{modTime: time.Now()}
			plugin.dirIndices[currDir] = currDirIndex
		}

		if currIsLeaf {
			if currBase == plugin.filename {
				currDirIndex.indexFile = f
			} else {
				context.DispatchFile(f)
			}
		}

		currDirIndex.entries = append(
			currDirIndex.entries,
			dirEntry{
				Name:  currBase,
				Path:  currPath,
				IsDir: !currIsLeaf,
				File:  f,
			},
		)

		if currDir == "." {
			break
		}

		currPath = currDir
		currIsLeaf = false
	}

	return nil
}

func (plugin *indexPlugin) Finalize(context goldsmith.Context) error {
	for name, index := range plugin.dirIndices {
		sort.Sort(index.entries)

		f := index.indexFile
		if f == nil {
			f = goldsmith.NewFileFromData(path.Join(name, plugin.filename), make([]byte, 0), index.modTime)
			for name, value := range plugin.meta {
				f.SetValue(name, value)
			}
		}

		f.SetValue(plugin.filesKey, index.entries)
		context.DispatchFile(f)
	}

	return nil
}

type dirIndex struct {
	entries   dirIndices
	modTime   time.Time
	indexFile goldsmith.File
}

type dirEntry struct {
	Name  string
	Path  string
	IsDir bool
	File  goldsmith.File
}

type dirIndices []dirEntry

func (d dirIndices) Len() int {
	return len(d)
}

func (d dirIndices) Less(i, j int) bool {
	d1, d2 := d[i], d[j]

	if d1.IsDir && !d2.IsDir {
		return true
	}
	if !d1.IsDir && d2.IsDir {
		return false
	}

	return strings.Compare(d1.Name, d2.Name) == -1
}

func (d dirIndices) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
